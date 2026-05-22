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

	if !b.running {
		return
	}
	b.running = false
	if b.api != nil {
		b.api.StopReceivingUpdates()
	}
	slog.Info("Telegram 机器人已停止")
}

// handleUpdates 处理更新
func (b *Bot) handleUpdates() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// 检查 api 是否初始化
	b.mu.RLock()
	api := b.api
	b.mu.RUnlock()
	if api == nil {
		return
	}

	updates := api.GetUpdatesChan(u)

	for update := range updates {
		// 检查运行状态，防止退出后继续处理
		b.mu.RLock()
		running := b.running
		b.mu.RUnlock()
		if !running {
			break
		}

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
	b.mu.RLock()
	defer b.mu.RUnlock()

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
	b.mu.RLock()
	handler := b.handler
	api := b.api
	running := b.running
	b.mu.RUnlock()

	if handler == nil || api == nil || !running {
		slog.Warn("机器人未完全初始化或已停止，忽略命令", "command", message.Command())
		return
	}

	command := message.Command()
	switch command {
	case "start":
		handler.HandleStart(message)
	case "tasks":
		handler.HandleTasks(message)
	case "run":
		handler.HandleRun(message)
	case "run_all":
		handler.HandleRunAll(message)
	case "status":
		handler.HandleStatus(message)
	case "logs":
		handler.HandleLogs(message)
	default:
		msg := tgbotapi.NewMessage(message.Chat.ID, "未知命令")
		_, _ = api.Send(msg)
	}
}

// SetHandler 设置命令处理器
func (b *Bot) SetHandler(handler *Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handler = handler
}

// SendMessage 发送消息
func (b *Bot) SendMessage(chatID int64, text string) error {
	b.mu.RLock()
	api := b.api
	running := b.running
	b.mu.RUnlock()

	if api == nil || !running {
		return fmt.Errorf("Telegram 机器人未运行")
	}

	msg := tgbotapi.NewMessage(chatID, text)
	_, err := api.Send(msg)
	return err
}
