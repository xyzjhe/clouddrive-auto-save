// internal/core/notify/wechat.go
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WeChatNotifier 企业微信通知渠道
type WeChatNotifier struct {
	webhookURL string
	client     *http.Client
}

// NewWeChatNotifier 创建企业微信通知渠道
func NewWeChatNotifier() *WeChatNotifier {
	return &WeChatNotifier{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Name 返回通知渠道名称
func (n *WeChatNotifier) Name() string {
	return "wechat"
}

// Type 返回通知渠道类型
func (n *WeChatNotifier) Type() NotifierType {
	return NotifierTypeWeChat
}

// Init 初始化通知渠道
func (n *WeChatNotifier) Init(config map[string]interface{}) error {
	webhookURL, ok := config["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return fmt.Errorf("企业微信 webhook_url 不能为空")
	}

	n.webhookURL = webhookURL
	return nil
}

// Send 发送通知
func (n *WeChatNotifier) Send(ctx context.Context, message *Message) error {
	// 构建消息内容
	content := fmt.Sprintf("**%s**\n\n%s", message.Title, message.Content)

	// 构建请求体
	body := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": content,
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	// 发送请求
	req, err := http.NewRequestWithContext(ctx, "POST", n.webhookURL, bytes.NewBuffer(jsonBody))
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
func (n *WeChatNotifier) Test(ctx context.Context) error {
	return n.Send(ctx, &Message{
		Title:   "UCAS 测试通知",
		Content: "这是一条测试消息，用于验证企业微信推送配置是否正确。",
		Level:   LevelInfo,
	})
}

// Close 关闭通知渠道
func (n *WeChatNotifier) Close() error {
	return nil
}
