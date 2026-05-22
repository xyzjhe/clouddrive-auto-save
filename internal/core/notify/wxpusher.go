// internal/core/notify/wxpusher.go
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WxPusherNotifier WxPusher 通知渠道
type WxPusherNotifier struct {
	appToken string
	uid      string
	client   *http.Client
}

// NewWxPusherNotifier 创建 WxPusher 通知渠道
func NewWxPusherNotifier() *WxPusherNotifier {
	return &WxPusherNotifier{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Name 返回通知渠道名称
func (n *WxPusherNotifier) Name() string {
	return "wxpusher"
}

// Type 返回通知渠道类型
func (n *WxPusherNotifier) Type() NotifierType {
	return NotifierTypeWxPusher
}

// Init 初始化通知渠道
func (n *WxPusherNotifier) Init(config map[string]interface{}) error {
	appToken, ok := config["app_token"].(string)
	if !ok || appToken == "" {
		return fmt.Errorf("WxPusher app_token 不能为空")
	}

	uid, ok := config["uid"].(string)
	if !ok || uid == "" {
		return fmt.Errorf("WxPusher uid 不能为空")
	}

	n.appToken = appToken
	n.uid = uid
	return nil
}

// Send 发送通知
func (n *WxPusherNotifier) Send(ctx context.Context, message *Message) error {
	// 构建消息内容
	content := fmt.Sprintf("**%s**\n\n%s", message.Title, message.Content)

	// 构建请求体
	body := map[string]interface{}{
		"appToken":    n.appToken,
		"content":     content,
		"summary":     message.Title,
		"contentType": 3, // Markdown
		"uids":        []string{n.uid},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	// 发送请求
	req, err := http.NewRequestWithContext(ctx, "POST", "https://wxpusher.zjiecode.com/api/send/message", bytes.NewBuffer(jsonBody))
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
func (n *WxPusherNotifier) Test(ctx context.Context) error {
	return n.Send(ctx, &Message{
		Title:   "UCAS 测试通知",
		Content: "这是一条测试消息，用于验证 WxPusher 推送配置是否正确。",
		Level:   LevelInfo,
	})
}

// Close 关闭通知渠道
func (n *WxPusherNotifier) Close() error {
	return nil
}
