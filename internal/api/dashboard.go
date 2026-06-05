package api

import (
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zcq/clouddrive-auto-save/internal/db"
	"github.com/zcq/clouddrive-auto-save/internal/utils"
)

func getDashboardStats(c *gin.Context) {
	// 获取全局调度开关状态 (使用 Find 避免 record not found 日志噪音)
	var enabledSetting db.Setting
	db.DB.Where("key = ?", "global_schedule_enabled").Limit(1).Find(&enabledSetting)
	globalEnabled := enabledSetting.Value == "true"

	var scheduledTasks int64
	if globalEnabled {
		// 全局开启：统计模式为 global 的任务 + 模式为 custom 且有 cron 的任务
		db.DB.Model(&db.Task{}).Where("schedule_mode = ? OR (schedule_mode = ? AND cron != '')", "global", "custom").Count(&scheduledTasks)
	} else {
		// 全局关闭：仅统计模式为 custom 且有 cron 的任务
		db.DB.Model(&db.Task{}).Where("schedule_mode = ? AND cron != ''", "custom").Count(&scheduledTasks)
	}

	var capacityUsed int64
	// 使用 COALESCE 避免在无账号时 SUM 返回 NULL 导致 Scan 报错
	db.DB.Model(&db.Account{}).Where("status = 1").Select("COALESCE(SUM(capacity_used), 0)").Scan(&capacityUsed)

	var todayCompleted int64
	db.DB.Model(&db.Task{}).Where("status = ? AND DATE(last_run) = DATE('now', 'localtime')", "success").Count(&todayCompleted)

	var activeAccounts int64
	db.DB.Model(&db.Account{}).Where("status = 1").Count(&activeAccounts)

	var runningTasksList []db.Task
	// 获取：
	// 1. 正在运行的任务 (running)
	// 2. 8 秒内成功且消息不为空的任务 (success)
	// 3. 未被忽略且消息不为空的失败任务 (failed)
	db.DB.Where("status = ? OR (status = ? AND last_run >= ? AND message != '') OR (status = ? AND stage != ? AND message != '')",
		"running", "success", time.Now().Add(-8*time.Second), "failed", "Dismissed").Find(&runningTasksList)

	var recentTasks []db.Task
	// 仅显示有消息记录的任务（配合"清空日志"逻辑）
	db.DB.Where("message != ''").Order("last_run desc").Limit(15).Find(&recentTasks)

	c.PureJSON(http.StatusOK, gin.H{
		"scheduled_tasks":    scheduledTasks,
		"capacity_used":      capacityUsed,
		"today_completed":    todayCompleted,
		"active_accounts":    activeAccounts,
		"recent_activities":  recentTasks,
		"running_tasks_list": runningTasksList,
		"sys_info":           utils.GetSysInfo(),
	})
}

func streamLogs(c *gin.Context) {
	clientChan := utils.GlobalBroadcaster.Subscribe()
	defer utils.GlobalBroadcaster.Unsubscribe(clientChan)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	heartbeat := time.NewTicker(20 * time.Second)
	defer heartbeat.Stop()

	c.Stream(func(w io.Writer) bool {
		select {
		case <-c.Request.Context().Done():
			return false
		case msg, ok := <-clientChan:
			if !ok {
				return false
			}
			c.SSEvent("message", msg)
			return true
		case <-heartbeat.C:
			c.SSEvent("heartbeat", "keep-alive")
			return true
		}
	})
}

func getRecentLogs(c *gin.Context) {
	logs := utils.GlobalBroadcaster.GetRecent()
	c.PureJSON(http.StatusOK, logs)
}

func clearRecentLogs(c *gin.Context) {
	utils.GlobalBroadcaster.ClearRecent()

	// 同时清理数据库中所有任务的最后运行消息和阶段
	// 增加 Where("1 = 1") 以绕过 GORM 的 AllowGlobalUpdate 限制
	db.DB.Model(&db.Task{}).Where("1 = 1").Updates(map[string]interface{}{
		"message": "",
		"stage":   "",
	})

	// 通知前端刷新
	utils.BroadcastStatsUpdate()

	c.PureJSON(http.StatusOK, gin.H{"message": "logs and task summaries cleared"})
}
