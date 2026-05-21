// internal/core/telegram/handler.go
package telegram

import (
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/zcq/clouddrive-auto-save/internal/db"
)

// Handler 命令处理器
type Handler struct {
	bot *Bot
	db  *db.DB
}

// NewHandler 创建命令处理器
func NewHandler(bot *Bot, db *db.DB) *Handler {
	return &Handler{
		bot: bot,
		db:  db,
	}
}

// HandleStart 处理 /start 命令
func (h *Handler) HandleStart(message *tgbotapi.Message) {
	text := `🤖 UCAS 机器人

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
	for i, task := range tasks {
		text += fmt.Sprintf("%d. %s [%s]\n", i+1, task.Name, task.Status)
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	h.bot.api.Send(msg)
}

// HandleRun 处理 /run 命令
func (h *Handler) HandleRun(message *tgbotapi.Message) {
	// 解析任务 ID
	args := message.CommandArguments()
	if args == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "请指定任务 ID，例如：/run 1")
		h.bot.api.Send(msg)
		return
	}

	// TODO: 实现任务执行逻辑
	msg := tgbotapi.NewMessage(message.Chat.ID, "任务执行功能开发中")
	h.bot.api.Send(msg)
}

// HandleRunAll 处理 /run_all 命令
func (h *Handler) HandleRunAll(message *tgbotapi.Message) {
	// TODO: 实现批量执行逻辑
	msg := tgbotapi.NewMessage(message.Chat.ID, "批量执行功能开发中")
	h.bot.api.Send(msg)
}

// HandleStatus 处理 /status 命令
func (h *Handler) HandleStatus(message *tgbotapi.Message) {
	// TODO: 实现状态查询逻辑
	text := "📊 系统状态：\n\n"
	text += "• 运行中任务：0\n"
	text += "• 等待中任务：0\n"
	text += "• 今日完成：0\n"

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	h.bot.api.Send(msg)
}

// HandleLogs 处理 /logs 命令
func (h *Handler) HandleLogs(message *tgbotapi.Message) {
	// TODO: 实现日志查询逻辑
	msg := tgbotapi.NewMessage(message.Chat.ID, "日志查询功能开发中")
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

	// 发送给所有允许的用户
	for _, chatID := range h.bot.config.AllowedIDs {
		h.bot.SendMessage(chatID, text)
	}
}
