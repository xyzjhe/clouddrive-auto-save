// internal/core/telegram/handler.go
package telegram

import (
	"fmt"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/zcq/clouddrive-auto-save/internal/core/worker"
	"github.com/zcq/clouddrive-auto-save/internal/db"
	"github.com/zcq/clouddrive-auto-save/internal/utils"
	"gorm.io/gorm"
)

// Handler 命令处理器
type Handler struct {
	bot *Bot
	db  *gorm.DB
	wm  *worker.Manager
}

// NewHandler 创建命令处理器
func NewHandler(bot *Bot, db *gorm.DB, wm *worker.Manager) *Handler {
	return &Handler{
		bot: bot,
		db:  db,
		wm:  wm,
	}
}

// HandleStart 处理 /start 命令
func (h *Handler) HandleStart(message *tgbotapi.Message) {
	text := `🤖 UCAS 机器人已就绪

可用命令：
/tasks - 查看所有任务
/run <任务ID> - 执行指定任务
/run_all - 批量执行所有任务
/status - 查看系统状态
/logs - 查看最近日志`

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	h.bot.api.Send(msg)
}

// HandleTasks 处理 /tasks 命令
func (h *Handler) HandleTasks(message *tgbotapi.Message) {
	var tasks []db.Task
	if err := h.db.Find(&tasks).Error; err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "获取任务列表失败")
		h.bot.api.Send(msg)
		return
	}

	if len(tasks) == 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "暂无任务")
		h.bot.api.Send(msg)
		return
	}

	text := "📋 任务列表：\n\n"
	for _, task := range tasks {
		text += fmt.Sprintf("ID: %d. %s [%s]\n", task.ID, task.Name, task.Status)
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	h.bot.api.Send(msg)
}

// HandleRun 处理 /run 命令
func (h *Handler) HandleRun(message *tgbotapi.Message) {
	args := message.CommandArguments()
	if args == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "请指定任务 ID，例如：/run 1")
		h.bot.api.Send(msg)
		return
	}

	id, err := strconv.Atoi(args)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "任务 ID 格式错误，应为整数")
		h.bot.api.Send(msg)
		return
	}

	var task db.Task
	if err := h.db.Preload("Account").First(&task, id).Error; err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "未找到该任务")
		h.bot.api.Send(msg)
		return
	}

	if task.Status == "running" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "该任务已在运行中")
		h.bot.api.Send(msg)
		return
	}

	// 立即更新状态并启动
	task.Status = "running"
	task.Stage = "Started"
	h.db.Model(&task).Updates(map[string]interface{}{
		"status": "running",
		"stage":  "Started",
	})
	utils.BroadcastTaskUpdate(&task)
	utils.BroadcastStatsUpdate()

	h.wm.Submit(worker.Job{Task: &task})

	msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("✅ 任务「%s」提交成功，开始后台转存", task.Name))
	h.bot.api.Send(msg)
}

// HandleRunAll 处理 /run_all 命令
func (h *Handler) HandleRunAll(message *tgbotapi.Message) {
	var tasks []db.Task
	err := h.db.Preload("Account").
		Where("status != ?", "running").
		Where("message NOT LIKE ? OR message IS NULL", "%[Fatal]%").
		Find(&tasks).Error

	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "查询任务列表失败")
		h.bot.api.Send(msg)
		return
	}

	if len(tasks) == 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "没有可运行的任务（全部在运行中或存在 Fatal 链接失效错误）")
		h.bot.api.Send(msg)
		return
	}

	batchID := fmt.Sprintf("tg_batch_%d", time.Now().Unix())
	h.wm.RegisterBatch(batchID, len(tasks))

	for i := range tasks {
		task := &tasks[i]
		task.Status = "running"
		task.Stage = "Started"
		h.db.Model(task).Updates(map[string]interface{}{
			"status": "running",
			"stage":  "Started",
		})
		utils.BroadcastTaskUpdate(task)
		h.wm.Submit(worker.Job{Task: task, BatchID: batchID})
	}
	utils.BroadcastStatsUpdate()

	msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("✅ 批量运行已启动，共提交了 %d 个任务", len(tasks)))
	h.bot.api.Send(msg)
}

// HandleStatus 处理 /status 命令
func (h *Handler) HandleStatus(message *tgbotapi.Message) {
	var runningCount int64
	h.db.Model(&db.Task{}).Where("status = ?", "running").Count(&runningCount)

	var totalCount int64
	h.db.Model(&db.Task{}).Count(&totalCount)

	// 统计今日成功/失败数（按 last_run 时间判断）
	var todaySuccess int64
	var todayFailed int64
	todayStart := time.Now().Local().Format("2006-01-02 00:00:00")

	h.db.Model(&db.Task{}).Where("status = ? AND last_run >= ?", "success", todayStart).Count(&todaySuccess)
	h.db.Model(&db.Task{}).Where("status = ? AND last_run >= ?", "failed", todayStart).Count(&todayFailed)

	text := "📊 系统状态快照：\n\n"
	text += fmt.Sprintf("• 运行中任务数：%d\n", runningCount)
	text += fmt.Sprintf("• 系统总任务数：%d\n", totalCount)
	text += fmt.Sprintf("• 今日转存成功：%d\n", todaySuccess)
	text += fmt.Sprintf("• 今日转存失败：%d\n", todayFailed)

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	h.bot.api.Send(msg)
}

// HandleLogs 处理 /logs 命令
func (h *Handler) HandleLogs(message *tgbotapi.Message) {
	logs := utils.GlobalBroadcaster.GetRecent()
	if len(logs) == 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "📭 暂无最新日志")
		h.bot.api.Send(msg)
		return
	}

	// 提取最近最多 10 条日志
	limit := 10
	if len(logs) < limit {
		limit = len(logs)
	}

	text := fmt.Sprintf("📜 最近 %d 条日志：\n\n", limit)
	startIdx := len(logs) - limit
	for i := startIdx; i < len(logs); i++ {
		text += fmt.Sprintf("• %s\n", logs[i])
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	h.bot.api.Send(msg)
}

// NotifyTaskComplete 通知任务完成
func (h *Handler) NotifyTaskComplete(task *db.Task, success bool) {
	if !h.bot.config.Enabled {
		return
	}

	if success && !h.bot.config.NotifyOnSuccess {
		return
	}

	if !success && !h.bot.config.NotifyOnFailure {
		return
	}

	text := fmt.Sprintf("✅ 任务完成\n\n名称：%s\n状态：%s", task.Name, task.Status)

	for _, chatID := range h.bot.config.AllowedIDs {
		h.bot.SendMessage(chatID, text)
	}
}
