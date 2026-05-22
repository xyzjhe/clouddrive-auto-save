// internal/core/notify/manager_test.go
package notify

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockNotifier Mock 通知渠道实现
type MockNotifier struct {
	name         string
	notifierType NotifierType
	config       map[string]interface{}
	initErr      error
	sendErr      error
	testErr      error
	closeErr     error
	sendCalled   bool
	lastMessage  *Message
}

func NewMockNotifier(name string, notifierType NotifierType) *MockNotifier {
	return &MockNotifier{
		name:         name,
		notifierType: notifierType,
	}
}

func (n *MockNotifier) Name() string                             { return n.name }
func (n *MockNotifier) Type() NotifierType                       { return n.notifierType }
func (n *MockNotifier) Init(config map[string]interface{}) error { return n.initErr }
func (n *MockNotifier) Send(ctx context.Context, message *Message) error {
	if n.sendErr != nil {
		return n.sendErr
	}
	n.sendCalled = true
	n.lastMessage = message
	return nil
}
func (n *MockNotifier) Test(ctx context.Context) error { return n.testErr }
func (n *MockNotifier) Close() error                   { return n.closeErr }

func TestManager_Register(t *testing.T) {
	manager := NewManager()
	notifier := NewMockNotifier("test", NotifierTypeWeChat)

	// 注册通知渠道应成功
	err := manager.Register(notifier)
	require.NoError(t, err)

	// 注册重复渠道应返回错误
	err = manager.Register(notifier)
	assert.Error(t, err)
}

func TestManager_Init(t *testing.T) {
	manager := NewManager()
	notifier := NewMockNotifier("test", NotifierTypeWeChat)
	manager.Register(notifier)

	// 初始化存在的渠道应成功
	configs := map[string]*NotifierConfig{
		"test": {
			Name:    "test",
			Type:    NotifierTypeWeChat,
			Enabled: true,
			Config:  map[string]interface{}{"webhook_url": "https://example.com"},
		},
	}
	err := manager.Init(configs)
	require.NoError(t, err)

	// 初始化不存在的渠道应跳过
	configs["not_exist"] = &NotifierConfig{
		Name:   "not_exist",
		Config: map[string]interface{}{},
	}
	err = manager.Init(configs)
	require.NoError(t, err)
}

func TestManager_Send(t *testing.T) {
	manager := NewManager()
	notifier := NewMockNotifier("test", NotifierTypeWeChat)
	manager.Register(notifier)

	configs := map[string]*NotifierConfig{
		"test": {
			Name:            "test",
			Enabled:         true,
			Config:          map[string]interface{}{},
			NotifyOnSuccess: true,
			NotifyOnFailure: true,
		},
	}
	manager.Init(configs)

	// 发送成功通知应成功
	message := &Message{
		Title:   "测试",
		Content: "测试内容",
		Level:   LevelSuccess,
	}
	err := manager.Send(context.Background(), message)
	require.NoError(t, err)
	assert.True(t, notifier.sendCalled)

	// 发送到禁用的渠道应跳过
	notifier2 := NewMockNotifier("disabled", NotifierTypeTelegram)
	manager.Register(notifier2)
	configs["disabled"] = &NotifierConfig{
		Name:    "disabled",
		Enabled: false,
		Config:  map[string]interface{}{},
	}
	manager.Init(configs)

	notifier2.sendCalled = false
	err = manager.Send(context.Background(), message)
	require.NoError(t, err)
	assert.False(t, notifier2.sendCalled)
}

func TestManager_ListNotifiers(t *testing.T) {
	manager := NewManager()
	notifier1 := NewMockNotifier("test1", NotifierTypeWeChat)
	notifier2 := NewMockNotifier("test2", NotifierTypeTelegram)
	manager.Register(notifier1)
	manager.Register(notifier2)

	notifiers := manager.ListNotifiers()
	assert.Len(t, notifiers, 2)
}

func TestManager_Close(t *testing.T) {
	manager := NewManager()
	notifier := NewMockNotifier("test", NotifierTypeWeChat)
	manager.Register(notifier)

	err := manager.Close()
	require.NoError(t, err)
}
