// internal/core/telegram/bot.go
package telegram

import (
	"fmt"
	"log/slog"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot Telegram 机器人
type Bot struct {
	config  *Config
	api     *tgbotapi.BotAPI
	handler *Handler
	mu      sync.RWMutex
	running bool
}

// NewBot 创建 Telegram 机器人
func NewBot(config *Config) *Bot {
	return &Bot{
		config: config,
	}
}

// Start 启动机器人
func (b *Bot) Start() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.running {
		return fmt.Errorf("机器人已在运行")
	}

	if !b.config.Enabled {
		return fmt.Errorf("机器人未启用")
	}

	api, err := tgbotapi.NewBotAPI(b.config.BotToken)
	if err != nil {
		return fmt.Errorf("创建 Bot API 失败: %w", err)
	}

	b.api = api
	b.running = true

	slog.Info("Telegram 机器人已启动", "username", api.Self.UserName)

	// 启动消息处理
	go b.handleUpdates()

	return nil
}

// Stop 停止机器人
func (b *Bot) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.running = false
	slog.Info("Telegram 机器人已停止")
}

// handleUpdates 处理更新
func (b *Bot) handleUpdates() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// 检查权限
		if !b.isAllowed(update.Message.From.ID) {
			continue
		}

		// 处理命令
		if update.Message.IsCommand() {
			b.handleCommand(update.Message)
		}
	}
}

// isAllowed 检查用户是否被允许
func (b *Bot) isAllowed(userID int64) bool {
	// 如果没有设置白名单，允许所有用户
	if len(b.config.AllowedIDs) == 0 {
		return true
	}

	for _, id := range b.config.AllowedIDs {
		if id == userID {
			return true
		}
	}

	return false
}

// handleCommand 处理命令
func (b *Bot) handleCommand(message *tgbotapi.Message) {
	if b.handler == nil {
		return
	}

	command := message.Command()
	switch command {
	case "start":
		b.handler.HandleStart(message)
	case "tasks":
		b.handler.HandleTasks(message)
	case "run":
		b.handler.HandleRun(message)
	case "run_all":
		b.handler.HandleRunAll(message)
	case "status":
		b.handler.HandleStatus(message)
	case "logs":
		b.handler.HandleLogs(message)
	default:
		msg := tgbotapi.NewMessage(message.Chat.ID, "未知命令")
		b.api.Send(msg)
	}
}

// SetHandler 设置命令处理器
func (b *Bot) SetHandler(handler *Handler) {
	b.handler = handler
}

// SendMessage 发送消息
func (b *Bot) SendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := b.api.Send(msg)
	return err
}
