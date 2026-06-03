// internal/core/notify/interface.go
package notify

import "context"

// Notifier 通知接口
type Notifier interface {
	// Name 返回通知渠道名称
	Name() string

	// Type 返回通知渠道类型
	Type() NotifierType

	// Init 初始化通知渠道
	Init(config map[string]interface{}) error

	// Send 发送通知
	Send(ctx context.Context, message *Message) error

	// Test 测试通知渠道
	Test(ctx context.Context) error

	// Close 关闭通知渠道
	Close() error
}

// NotifierType 通知渠道类型
type NotifierType string

const (
	NotifierTypeWeChat   NotifierType = "wechat"
	NotifierTypeTelegram NotifierType = "telegram"
	NotifierTypeWxPusher NotifierType = "wxpusher"
	NotifierTypeBark     NotifierType = "bark"
)

// Message 通知消息
type Message struct {
	Title    string       `json:"title"`
	Content  string       `json:"content"`
	Level    MessageLevel `json:"level"`
	TaskID   uint         `json:"task_id,omitempty"`
	TaskName string       `json:"task_name,omitempty"`
}

// MessageLevel 消息级别
type MessageLevel string

const (
	LevelInfo    MessageLevel = "info"
	LevelSuccess MessageLevel = "success"
	LevelWarning MessageLevel = "warning"
	LevelError   MessageLevel = "error"
)

// NotifierConfig 通知渠道配置
type NotifierConfig struct {
	Name            string                 `json:"name"`
	Type            NotifierType           `json:"type"`
	Enabled         bool                   `json:"enabled"`
	Config          map[string]interface{} `json:"config"`
	NotifyOnSuccess bool                   `json:"notify_on_success"`
	NotifyOnFailure bool                   `json:"notify_on_failure"`
}
