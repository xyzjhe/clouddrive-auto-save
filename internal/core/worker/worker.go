package worker

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/zcq/clouddrive-auto-save/internal/core"
	"github.com/zcq/clouddrive-auto-save/internal/core/notify"
	"github.com/zcq/clouddrive-auto-save/internal/core/openlist"
	"github.com/zcq/clouddrive-auto-save/internal/core/renamer"
	"github.com/zcq/clouddrive-auto-save/internal/db"
	"github.com/zcq/clouddrive-auto-save/internal/utils"
	"gorm.io/gorm"
)

// Job 代表一个待执行的转存任务
type Job struct {
	Task    *db.Task
	BatchID string // 为空表示单任务执行
}

// Manager 负责管理 Worker 池和任务分发
type Manager struct {
	workers  int
	jobQueue chan Job
	wg       sync.WaitGroup
	retryWg  sync.WaitGroup // 跟踪重试 goroutine，确保 Stop() 能优雅等待
	ctx      context.Context
	cancel   context.CancelFunc
	db       *gorm.DB
	tracker  *BatchTracker
}

func NewManager(numWorkers int, queueSize int, dbInst *gorm.DB) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	if queueSize <= 0 {
		queueSize = 100
	}
	return &Manager{
		workers:  numWorkers,
		jobQueue: make(chan Job, queueSize),
		ctx:      ctx,
		cancel:   cancel,
		db:       dbInst,
		tracker:  NewBatchTracker(),
	}
}

// Start 启动所有 Worker
func (m *Manager) Start() {
	for i := 1; i <= m.workers; i++ {
		m.wg.Add(1)
		go m.worker(i)
	}
}

// Stop 停止所有 Worker
func (m *Manager) Stop() {
	m.cancel()
	m.wg.Wait()
	m.retryWg.Wait() // 等待所有重试 goroutine 退出
}

// Submit 提交一个任务，队列满时返回错误而非阻塞
func (m *Manager) Submit(job Job) error {
	select {
	case m.jobQueue <- job:
		return nil
	default:
		return fmt.Errorf("任务队列已满（容量 %d），请稍后重试", cap(m.jobQueue))
	}
}

// RegisterBatch 注册一个批量执行批次
func (m *Manager) RegisterBatch(batchID string, total int) {
	m.tracker.RegisterBatch(batchID, total)
}

func (m *Manager) worker(id int) {
	defer m.wg.Done()
	slog.Info("Worker 启动", "id", id)
	for {
		select {
		case <-m.ctx.Done():
			slog.Info("Worker 正在停止", "id", id)
			return
		case job := <-m.jobQueue:
			m.execute(job)
		}
	}
}

func (m *Manager) updateProgress(task *db.Task, percent int, stage, message string) {
	task.Percent = percent
	task.Stage = stage
	task.Message = message
	m.db.Model(task).Updates(map[string]interface{}{
		"percent": percent,
		"stage":   stage,
		"message": message,
	})
	slog.Info(fmt.Sprintf("[PROGRESS:%d:%d:%s:%s]", task.ID, percent, stage, message))
	utils.BroadcastTaskUpdate(task)
}

// getExtension 获取文件扩展名（包含点号）
func getExtension(name string) string {
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '.' {
			return name[i:]
		}
		if name[i] == '/' || name[i] == '\\' {
			break
		}
	}
	return ""
}

func (m *Manager) execute(job Job) {
	task := job.Task
	startTime := time.Now()
	slog.Info("正在执行任务", "name", task.Name, "id", task.ID)
	m.updateProgress(task, 5, "Started", "任务已进入执行队列")

	// 1. 更新任务状态为 running
	m.db.Model(task).Update("status", "running")

	driver := core.GetDriver(&task.Account)
	if driver == nil {
		m.finishTask(job, "failed", "Driver not found", nil, startTime)
		return
	}

	// 2. 解析分享内容（独立超时，避免单个 API 挂起阻塞 worker）
	m.updateProgress(task, 15, "Parsing", "正在解析分享链接...")
	parseCtx, parseCancel := context.WithTimeout(m.ctx, 60*time.Second)
	defer parseCancel()
	files, err := driver.ParseShare(parseCtx, task.ShareURL, task.ExtractCode, task.ShareParentID)
	if err != nil {
		m.finishTask(job, "failed", "解析分享失败: "+err.Error(), nil, startTime)
		return
	}

	// 2.1 排序：按更新时间从新到旧
	sort.Slice(files, func(i, j int) bool {
		return files[i].UpdateTime.After(files[j].UpdateTime)
	})

	// 2.2 如果有起始文件，执行截断
	if task.StartFileID != "" {
		foundIdx := -1
		for i, f := range files {
			if f.ID == task.StartFileID {
				foundIdx = i
				break
			}
		}
		if foundIdx != -1 {
			// 仅保留该文件及其之后更新的文件 (即 0 到 foundIdx)
			files = files[:foundIdx+1]
			slog.Info("已应用起始文件截断", "task_id", task.ID, "start_file", task.StartFileName, "remaining_count", len(files))
		}
	}

	// 3. 列出目标目录文件，进行去重检查
	m.updateProgress(task, 35, "Checking", "正在检查目标目录是否存在同名文件...")
	checkCtx, checkCancel := context.WithTimeout(m.ctx, 30*time.Second)
	defer checkCancel()
	targetID, err := driver.PrepareTargetPath(checkCtx, task.SavePath)
	if err != nil {
		m.finishTask(job, "failed", "准备目标路径失败: "+err.Error(), nil, startTime)
		return
	}

	existingFiles, err := driver.ListFiles(checkCtx, targetID)
	if err != nil {
		m.finishTask(job, "failed", "列出目标目录失败: "+err.Error(), nil, startTime)
		return
	}

	existingNames := make(map[string]bool)
	for _, f := range existingFiles {
		if task.IgnoreExtension {
			// 忽略后缀模式：去掉扩展名后存储
			name := strings.TrimSuffix(f.Name, getExtension(f.Name))
			existingNames[name] = true
		} else {
			existingNames[f.Name] = true
		}
	}

	// 4. 计算预览名并应用过滤（正则匹配 + 智能去重）
	processor := renamer.NewProcessor()
	var filteredIDs []string
	var savedFileNames []string
	var skipCount int
	var regexSkipCount int
	renameMap := make(map[string]string) // 记录 原始文件名 -> 计算后新文件名 的对应关系

	// 解析预定义魔法匹配规则
	pattern := task.Pattern
	replacement := task.Replacement
	if strings.HasPrefix(pattern, "$") {
		if predefined := renamer.GetPredefinedPattern(pattern); predefined != nil {
			pattern = predefined.Pattern
			if replacement == "" {
				replacement = predefined.Replacement
			}
			slog.Info("使用预定义匹配规则", "name", task.Pattern, "pattern", pattern)
		}
	}

	var compiledReg *regexp.Regexp
	if pattern != "" {
		compiledReg, err = regexp.Compile(pattern)
		if err != nil {
			slog.Warn("正则表达式编译失败，跳过正则过滤", "pattern", pattern, "error", err)
		}
	}

	for _, f := range files {
		// a. 正则匹配过滤
		if compiledReg != nil {
			if !compiledReg.MatchString(f.Name) {
				regexSkipCount++
				continue
			}
		}

		// b. 计算预期的新名字
		newName := f.Name
		if replacement != "" {
			resName, err := processor.Process(renamer.RenameOptions{
				TaskName:        task.Name,
				FileName:        f.Name,
				Pattern:         pattern,
				Replacement:     replacement,
				CompiledPattern: compiledReg, // 复用外部已编译的正则，减少 CPU 开销
			})
			if err == nil {
				newName = resName
			}
		}

		// c. 智能去重：拿新名比对
		dedupName := newName
		if task.IgnoreExtension {
			dedupName = strings.TrimSuffix(newName, getExtension(newName))
		}
		if existingNames[dedupName] {
			skipCount++
			continue
		}

		filteredIDs = append(filteredIDs, f.ID)
		savedFileNames = append(savedFileNames, newName)
		if newName != f.Name {
			renameMap[f.Name] = newName
		}
	}

	if len(filteredIDs) == 0 {
		msg := fmt.Sprintf("没有需要转存的文件 (跳过 %d 个同名文件", skipCount)
		if regexSkipCount > 0 {
			msg += fmt.Sprintf(", 过滤 %d 个不匹配文件", regexSkipCount)
		}
		msg += ")"
		m.finishTask(job, "success", msg, nil, startTime)
		return
	}

	m.updateProgress(task, 60, "Saving", fmt.Sprintf("正在转存 %d 个文件...", len(filteredIDs)))
	saveCtx, saveCancel := context.WithTimeout(m.ctx, 120*time.Second)
	defer saveCancel()
	err = driver.SaveLink(saveCtx, task.ShareURL, task.ExtractCode, task.SavePath, filteredIDs, task.ShareParentID)
	if err != nil {
		m.finishTask(job, "failed", "转存失败: "+err.Error(), nil, startTime)
		return
	}

	// 5. 检查是否需要重命名 (如果有规则且有需要重命名的文件)
	if len(renameMap) > 0 {
		m.updateProgress(task, 85, "Renaming", "转存成功，正在执行重命名...")
		// 再次列出文件，找到刚才存入的文件进行重命名
		renameCtx, renameCancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer renameCancel()
		newFiles, listErr := driver.ListFiles(renameCtx, targetID)
		if listErr != nil {
			slog.Error("重命名阶段列出文件失败，跳过重命名", "task_id", task.ID, "error", listErr)
			m.updateProgress(task, 90, "Renaming", fmt.Sprintf("重命名失败：无法列出目录 (%v)", listErr))
		}
		for _, tf := range newFiles {
			// 如果当前文件名在待命名的映射表中，直接执行重命名
			if expectedNewName, ok := renameMap[tf.Name]; ok {
				if expectedNewName != tf.Name {
					slog.Info("正在执行重命名", "task_id", task.ID, "old_name", tf.Name, "new_name", expectedNewName)
					err = driver.RenameFile(m.ctx, tf.ID, expectedNewName)
					if err == nil {
						// 成功后将其从待命名 map 中移出
						delete(renameMap, tf.Name)
					}
				}
			}
			// 早期退出：待命名列表已清空，说明本次转存的新文件已全部重命名完毕，立刻中止无谓的遍历
			if len(renameMap) == 0 {
				break
			}
		}
	}

	m.finishTask(job, "success", fmt.Sprintf("转存成功 (新增 %d 个文件, 跳过 %d 个同名文件)", len(filteredIDs), skipCount), savedFileNames, startTime)
}

// isFatalError 判断是否为致命错误（不应重试）
func isFatalError(message string) bool {
	// 驱动层已标记为致命错误（如 quarkErrorCodeMap / cloud139 命中），直接返回
	if strings.Contains(message, "[Fatal]") {
		return true
	}
	fatalPatterns := []string{
		"链接失效", "链接过期", "提取码错误", "提取码无效",
		"分享已删除", "分享已过期", "权限不足",
		"涉及违规", "取消了分享",
	}
	for _, p := range fatalPatterns {
		if strings.Contains(message, p) {
			return true
		}
	}
	// 限定子串匹配：这些词单独出现时可能是致命的，但需避免宽泛误判
	containsChecks := []string{
		"不存在", "cookie过期", "Cookie过期",
		"token无效", "token过期", "Token无效", "Token过期",
	}
	for _, p := range containsChecks {
		if strings.Contains(message, p) {
			return true
		}
	}
	return false
}

func (m *Manager) finishTask(job Job, status, message string, files []string, startTime time.Time) {
	task := job.Task
	task.LastRun = time.Now()
	task.Percent = 100
	duration := time.Since(startTime)

	// 重试逻辑：失败且非致命错误时，尝试重试
	if status == "failed" && !isFatalError(message) {
		maxRetries := task.MaxRetries
		if maxRetries == 0 {
			maxRetries = 3 // 默认最大重试 3 次
		}
		if task.RetryCount < maxRetries {
			task.RetryCount++
			// 指数退避：30s, 60s, 120s, 240s, ... 最大 3600s
			delay := 30 * (1 << (task.RetryCount - 1))
			if delay > 3600 {
				delay = 3600
			}
			m.db.Model(task).Updates(map[string]interface{}{
				"status":      "pending",
				"message":     fmt.Sprintf("[重试 %d/%d] %s", task.RetryCount, maxRetries, message),
				"last_run":    task.LastRun,
				"retry_count": task.RetryCount,
				"percent":     0,
				"stage":       "Retry",
			})
			slog.Warn("任务将自动重试", "id", task.ID, "retry", task.RetryCount, "max", maxRetries, "delay_s", delay)
			slog.Info(fmt.Sprintf("[PROGRESS:%d:0:Retry:将在 %ds 后重试 (%d/%d)]", task.ID, delay, task.RetryCount, maxRetries))
			utils.BroadcastTaskUpdate(task)
			utils.BroadcastStatsUpdate()

			// 延迟后重新入队（使用 select 等待，支持 context 取消）
			m.retryWg.Add(1)
			go func() {
				defer m.retryWg.Done()
				timer := time.NewTimer(time.Duration(delay) * time.Second)
				defer timer.Stop()
				select {
				case <-m.ctx.Done():
					slog.Info("重试已取消（服务关闭）", "task_id", task.ID)
				case <-timer.C:
					if err := m.Submit(Job{Task: task, BatchID: job.BatchID}); err != nil {
						slog.Warn("重试提交失败：队列已满", "task_id", task.ID, "error", err)
					}
				}
			}()
			return
		}
		// 重试次数用尽，标记为致命错误
		message = fmt.Sprintf("[Fatal] 重试 %d 次后仍然失败: %s", maxRetries, message)
	}

	task.Status = status
	task.Message = message
	if status == "success" {
		task.Stage = "Success"
		task.RetryCount = 0 // 成功后重置重试计数
	} else {
		task.Stage = "Failed"
	}

	m.db.Model(task).Updates(map[string]interface{}{
		"status":      status,
		"message":     message,
		"last_run":    task.LastRun,
		"percent":     task.Percent,
		"stage":       task.Stage,
		"retry_count": task.RetryCount,
	})
	slog.Info("任务完成", "id", task.ID, "status", status, "duration", duration)
	slog.Info(fmt.Sprintf("[PROGRESS:%d:100:%s:%s]", task.ID, task.Stage, message))
	utils.BroadcastTaskUpdate(task)
	utils.BroadcastStatsUpdate()

	// OpenList 扫描触发：单任务模式且有新文件时触发
	if job.BatchID == "" && status == "success" && len(files) > 0 {
		openlist.GlobalScanner.OnTaskComplete(true)
	}

	// Bark 通知：区分单任务和批量模式
	if job.BatchID != "" {
		m.tracker.ReportTask(job.BatchID, BatchResult{
			TaskName: task.Name, Status: status,
			Message: message, Files: files, Duration: duration,
		})
	} else {
		notify.SendTaskNotification(task.Name, status, message, files, duration)
	}
}
