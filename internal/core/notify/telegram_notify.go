// internal/core/notify/telegram_notify.go
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"time"
)

// TelegramNotifier Telegram 通知渠道
type TelegramNotifier struct {
	botToken string
	chatID   string
	client   *http.Client
}

// NewTelegramNotifier 创建 Telegram 通知渠道
func NewTelegramNotifier() *TelegramNotifier {
	return &TelegramNotifier{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Name 返回通知渠道名称
func (n *TelegramNotifier) Name() string {
	return "telegram"
}

// Type 返回通知渠道类型
func (n *TelegramNotifier) Type() NotifierType {
	return NotifierTypeTelegram
}

// Init 初始化通知渠道
func (n *TelegramNotifier) Init(config map[string]interface{}) error {
	botToken, ok := config["bot_token"].(string)
	if !ok || botToken == "" {
		return fmt.Errorf("Telegram bot_token 不能为空")
	}

	chatID, ok := config["chat_id"].(string)
	if !ok || chatID == "" {
		return fmt.Errorf("Telegram chat_id 不能为空")
	}

	n.botToken = botToken
	n.chatID = chatID
	return nil
}

// Send 发送通知
func (n *TelegramNotifier) Send(ctx context.Context, message *Message) error {
	// 构建消息内容
	content := fmt.Sprintf("<b>%s</b>\n\n%s", html.EscapeString(message.Title), html.EscapeString(message.Content))

	// 构建请求体
	body := map[string]string{
		"chat_id":    n.chatID,
		"text":       content,
		"parse_mode": "HTML",
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	// 发送请求
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.botToken)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// Test 测试通知渠道
func (n *TelegramNotifier) Test(ctx context.Context) error {
	return n.Send(ctx, &Message{
		Title:   "UCAS 测试通知",
		Content: "这是一条测试消息，用于验证 Telegram 推送配置是否正确。",
		Level:   LevelInfo,
	})
}

// Close 关闭通知渠道
func (n *TelegramNotifier) Close() error {
	return nil
}
