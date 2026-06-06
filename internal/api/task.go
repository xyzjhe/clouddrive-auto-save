package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zcq/clouddrive-auto-save/internal/core/scheduler"
	"github.com/zcq/clouddrive-auto-save/internal/core/worker"
	"github.com/zcq/clouddrive-auto-save/internal/db"
	"github.com/zcq/clouddrive-auto-save/internal/utils"
)

func listTasks(c *gin.Context) {
	var tasks []db.Task
	db.DB.Preload("Account").Find(&tasks)
	c.PureJSON(http.StatusOK, tasks)
}

func createTask(c *gin.Context) {
	var task db.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.PureJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 校验 Cron 表达式
	if task.ScheduleMode == "custom" {
		if err := scheduler.ValidateCron(task.Cron); err != nil {
			c.PureJSON(http.StatusBadRequest, gin.H{"error": "Cron 表达式格式错误: " + err.Error()})
			return
		}
	}

	slog.Info("创建任务", "name", task.Name)
	db.DB.Create(&task)

	// 推送实时事件
	utils.BroadcastTaskUpdate(&task)
	utils.BroadcastStatsUpdate()

	// 注册定时任务
	scheduler.Global.UpdateTask(task.ID, task.ScheduleMode, task.Cron)

	c.PureJSON(http.StatusOK, task)
}

func updateTask(c *gin.Context) {
	id := c.Param("id")
	var task db.Task
	if err := db.DB.First(&task, id).Error; err != nil {
		c.PureJSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	// 记录更新前的关键参数，用于判断是否需要重置状态
	originalID := task.ID // 保存正确的 ID（来自 URL 路径的 DB 查询）
	oldURL := task.ShareURL
	oldCode := task.ExtractCode

	if err := c.ShouldBindJSON(&task); err != nil {
		c.PureJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 防御性：恢复 task.ID，防止 JSON body 中的 id 字段覆盖 URL 路径参数
	task.ID = originalID

	// 校验 Cron 表达式
	if task.ScheduleMode == "custom" {
		if err := scheduler.ValidateCron(task.Cron); err != nil {
			c.PureJSON(http.StatusBadRequest, gin.H{"error": "Cron 表达式格式错误: " + err.Error()})
			return
		}
	}

	slog.Info("更新任务", "name", task.Name)

	updateData := map[string]interface{}{
		"name":             task.Name,
		"account_id":       task.AccountID,
		"share_url":        task.ShareURL,
		"extract_code":     task.ExtractCode,
		"save_path":        task.SavePath,
		"pattern":          task.Pattern,
		"replacement":      task.Replacement,
		"start_file_id":    task.StartFileID,
		"start_file_name":  task.StartFileName,
		"share_parent_id":  task.ShareParentID,
		"cron":             task.Cron,
		"schedule_mode":    task.ScheduleMode,
		"max_retries":      task.MaxRetries,
		"ignore_extension": task.IgnoreExtension,
	}

	// 仅当分享链接或提取码发生变动时，才重置状态以解除 [Fatal] 封锁
	if task.ShareURL != oldURL || task.ExtractCode != oldCode {
		slog.Info("检测到关键参数变更，自动重置任务状态", "name", task.Name)
		updateData["status"] = "pending"
		updateData["message"] = ""
	}

	if err := db.DB.Model(&task).Updates(updateData).Error; err != nil {
		c.PureJSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}

	// 重新加载以获取关联的 Account 信息
	db.DB.Preload("Account").First(&task, task.ID)

	// 推送更新事件
	utils.BroadcastTaskUpdate(&task)

	// 刷新调度器
	scheduler.Global.UpdateTask(task.ID, task.ScheduleMode, task.Cron)

	c.PureJSON(http.StatusOK, task)
}

func deleteTask(c *gin.Context) {
	id := c.Param("id")
	slog.Info("删除任务", "task_id", id)

	idNum, err := strconv.Atoi(id)
	if err != nil {
		c.PureJSON(http.StatusBadRequest, gin.H{"error": "无效的任务 ID"})
		return
	}
	scheduler.Global.RemoveTask(uint(idNum))

	db.DB.Delete(&db.Task{}, id)

	// 推送实时事件
	utils.BroadcastTaskDelete(uint(idNum))
	utils.BroadcastStatsUpdate()

	c.PureJSON(http.StatusOK, gin.H{"message": "task deleted"})
}

func runTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.PureJSON(http.StatusBadRequest, gin.H{"error": "无效的任务 ID"})
		return
	}
	slog.Info("请求运行任务", "task_id", id)

	var task db.Task
	if err := db.DB.Preload("Account").First(&task, id).Error; err != nil {
		c.PureJSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	if task.Status == "running" {
		c.PureJSON(http.StatusBadRequest, gin.H{"error": "task is already running"})
		return
	}

	// 立即更新状态并推送
	task.Status = "running"
	task.Stage = "Started" // 重置 Dismissed 状态
	db.DB.Model(&task).Updates(map[string]interface{}{
		"status": "running",
		"stage":  "Started",
	})
	utils.BroadcastTaskUpdate(&task)
	utils.BroadcastStatsUpdate()

	if err := WorkerManager.Submit(worker.Job{Task: &task}); err != nil {
		// 队列满，回滚状态
		db.DB.Model(&task).Updates(map[string]interface{}{"status": "pending", "stage": ""})
		utils.BroadcastTaskUpdate(&task)
		c.PureJSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.PureJSON(http.StatusOK, gin.H{"message": "task submitted to worker pool"})
}

func runAllTasks(c *gin.Context) {
	slog.Info("请求批量运行所有任务")

	var tasks []db.Task
	// 筛选条件：1. status 不是 running; 2. message 中不包含 [Fatal]
	err := db.DB.Preload("Account").
		Where("status != ?", "running").
		Where("message NOT LIKE ? OR message IS NULL", "%[Fatal]%").
		Find(&tasks).Error

	if err != nil {
		slog.Error("获取批量运行任务列表失败", "error", err)
		c.PureJSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch tasks"})
		return
	}

	if len(tasks) == 0 {
		c.PureJSON(http.StatusOK, gin.H{"message": "没有可执行的任务", "count": 0})
		return
	}

	// 生成批次 ID（时间戳 + 随机后缀，确保唯一性）
	randBytes := make([]byte, 4)
	rand.Read(randBytes)
	batchID := fmt.Sprintf("batch_%d_%s", time.Now().UnixMilli(), hex.EncodeToString(randBytes))
	WorkerManager.RegisterBatch(batchID, len(tasks))

	count := 0
	for i := range tasks {
		task := &tasks[i]
		task.Status = "running"
		task.Stage = "Started"
		db.DB.Model(task).Updates(map[string]interface{}{
			"status": "running",
			"stage":  "Started",
		})
		utils.BroadcastTaskUpdate(task)

		if err := WorkerManager.Submit(worker.Job{Task: task, BatchID: batchID}); err != nil {
			// 队列满，回滚该任务状态并跳过
			db.DB.Model(task).Updates(map[string]interface{}{"status": "pending", "stage": ""})
			utils.BroadcastTaskUpdate(task)
			slog.Warn("批量提交跳过：队列已满", "task_id", task.ID)
			continue
		}
		count++
	}

	utils.BroadcastStatsUpdate()
	slog.Info("批量运行任务提交完成", "batch_id", batchID, "total_triggered", count)
	c.PureJSON(http.StatusOK, gin.H{"message": fmt.Sprintf("批量执行已开启，共触发 %d 个任务", count), "count": count})
}

func dismissTaskAPI(c *gin.Context) {
	id := c.Param("id")
	if err := db.DB.Model(&db.Task{}).Where("id = ?", id).Update("stage", "Dismissed").Error; err != nil {
		c.PureJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.PureJSON(http.StatusOK, gin.H{"message": "task dismissed"})
}
