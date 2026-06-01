package scheduler

import (
	"encoding/json"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/zcq/clouddrive-auto-save/internal/core/worker"
	"github.com/zcq/clouddrive-auto-save/internal/db"
)

type Scheduler struct {
	cron           *cron.Cron
	CustomEntryIDs map[uint]cron.EntryID
	GlobalEntryID  cron.EntryID
	wm             *worker.Manager
	mu             sync.Mutex
}

var Global *Scheduler

func Init(wm *worker.Manager) {
	Global = New(wm)
}

func New(wm *worker.Manager) *Scheduler {
	return &Scheduler{
		cron:           cron.New(cron.WithSeconds()),
		CustomEntryIDs: make(map[uint]cron.EntryID),
		wm:             wm,
	}
}

func (s *Scheduler) Start() {
	s.cron.Start()
	slog.Info("调度器已启动")
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
	slog.Info("调度器已停止")
}

func (s *Scheduler) RemoveTask(taskID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, ok := s.CustomEntryIDs[taskID]; ok {
		s.cron.Remove(entryID)
		delete(s.CustomEntryIDs, taskID)
		slog.Info("已移除任务自定义调度", "task_id", taskID)
	}
}

func (s *Scheduler) UpdateGlobalSchedule(cronExpr string, enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.GlobalEntryID != 0 {
		s.cron.Remove(s.GlobalEntryID)
		s.GlobalEntryID = 0
		slog.Info("已移除旧的全局调度")
	}

	if enabled && cronExpr != "" {
		entryID, err := s.cron.AddFunc(cronExpr, func() {
			slog.Info("触发全局调度任务")
			var tasks []db.Task
			if err := db.DB.Preload("Account").Where("schedule_mode = ?", "global").Find(&tasks).Error; err != nil {
				slog.Error("全局调度：获取任务列表失败", "error", err)
				return
			}

			for _, task := range tasks {
				if task.Status == "running" {
					slog.Info("全局调度：任务正在运行，跳过", "task_id", task.ID)
					continue
				}
				if strings.Contains(task.Message, "[Fatal]") {
					slog.Info("全局调度：任务存在致命错误，跳过", "task_id", task.ID)
					continue
				}
				if !isRunDayAllowed(task.RunDays) {
					slog.Info("全局调度：今天不是任务运行日，跳过", "task_id", task.ID, "run_days", task.RunDays)
					continue
				}

				slog.Info("全局调度：正在触发任务", "task_id", task.ID)
				if err := s.wm.Submit(worker.Job{Task: &task}); err != nil {
					slog.Warn("全局调度提交失败：队列已满", "task_id", task.ID, "error", err)
				}
			}
		})

		if err != nil {
			slog.Error("更新全局调度失败", "error", err)
			return
		}
		s.GlobalEntryID = entryID
	}

	slog.Info("全局调度配置已更新", "cron", cronExpr, "enabled", enabled)
}

func (s *Scheduler) UpdateTask(taskID uint, mode string, customCron string) {
	s.RemoveTask(taskID)

	if mode == "custom" && customCron != "" {
		s.mu.Lock()
		defer s.mu.Unlock()

		entryID, err := s.cron.AddFunc(customCron, func() {
			var task db.Task
			if err := db.DB.Preload("Account").First(&task, taskID).Error; err != nil {
				slog.Error("自定义调度：任务未找到，正在移除调度", "task_id", taskID)
				s.RemoveTask(taskID)
				return
			}

			if task.Status == "running" {
				slog.Info("自定义调度：任务正在运行，跳过", "task_id", taskID)
				return
			}
			if strings.Contains(task.Message, "[Fatal]") {
				slog.Info("自定义调度：任务存在致命错误，跳过", "task_id", taskID)
				return
			}
			if !isRunDayAllowed(task.RunDays) {
				slog.Info("自定义调度：今天不是任务运行日，跳过", "task_id", taskID, "run_days", task.RunDays)
				return
			}

			slog.Info("自定义调度：正在触发任务", "task_id", taskID)
		if err := s.wm.Submit(worker.Job{Task: &task}); err != nil {
			slog.Warn("自定义调度提交失败：队列已满", "task_id", task.ID, "error", err)
		}
	})

		if err != nil {
			slog.Error("更新任务自定义调度失败", "task_id", taskID, "error", err)
			return
		}
		s.CustomEntryIDs[taskID] = entryID
		slog.Info("已添加任务自定义调度", "task_id", taskID, "cron", customCron)
	}
}

func ValidateCron(expr string) error {
	_, err := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow).Parse(expr)
	return err
}

// isRunDayAllowed 检查今天是否是任务允许运行的日期
// runDays 格式为 JSON 数组，如 "[1,2,3,4,5]" 表示周一到周五
// 空字符串表示每天都可以运行
// 1=周一, 7=周日
func isRunDayAllowed(runDays string) bool {
	if runDays == "" || runDays == "[]" {
		return true
	}

	var days []int
	if err := json.Unmarshal([]byte(runDays), &days); err != nil {
		return true // 解析失败默认允许运行
	}
	if len(days) == 0 {
		return true
	}

	today := int(time.Now().Weekday())
	if today == 0 {
		today = 7 // Go 的 Sunday=0 转换为 7
	}

	for _, d := range days {
		if d == today {
			return true
		}
	}
	return false
}
