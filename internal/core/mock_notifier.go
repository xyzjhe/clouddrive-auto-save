//go:build e2e

// internal/core/mock_notifier.go
package core

import (
	"context"

	"github.com/zcq/clouddrive-auto-save/internal/core/notify"
)

// MockNotifier Mock 通知渠道实现
type MockNotifier struct {
	name         string
	notifierType notify.NotifierType
	config       map[string]interface{}
	initErr      error
	sendErr      error
	testErr      error
	closeErr     error
	sendCalled   bool
	lastMessage  *notify.Message
}

// NewMockNotifier 创建 Mock 通知渠道
func NewMockNotifier(name string, notifierType notify.NotifierType) *MockNotifier {
	return &MockNotifier{
		name:         name,
		notifierType: notifierType,
	}
}

// Name 返回通知渠道名称
func (n *MockNotifier) Name() string {
	return n.name
}

// Type 返回通知渠道类型
func (n *MockNotifier) Type() notify.NotifierType {
	return n.notifierType
}

// Init 初始化通知渠道
func (n *MockNotifier) Init(config map[string]interface{}) error {
	if n.initErr != nil {
		return n.initErr
	}
	n.config = config
	return nil
}

// Send 发送通知
func (n *MockNotifier) Send(ctx context.Context, message *notify.Message) error {
	if n.sendErr != nil {
		return n.sendErr
	}
	n.sendCalled = true
	n.lastMessage = message
	return nil
}

// Test 测试通知渠道
func (n *MockNotifier) Test(ctx context.Context) error {
	return n.testErr
}

// Close 关闭通知渠道
func (n *MockNotifier) Close() error {
	return n.closeErr
}

// SetInitError 设置初始化错误
func (n *MockNotifier) SetInitError(err error) {
	n.initErr = err
}

// SetSendError 设置发送错误
func (n *MockNotifier) SetSendError(err error) {
	n.sendErr = err
}

// SetTestError 设置测试错误
func (n *MockNotifier) SetTestError(err error) {
	n.testErr = err
}

// SetCloseError 设置关闭错误
func (n *MockNotifier) SetCloseError(err error) {
	n.closeErr = err
}

// IsSendCalled 检查是否调用了发送
func (n *MockNotifier) IsSendCalled() bool {
	return n.sendCalled
}

// GetLastMessage 获取最后发送的消息
func (n *MockNotifier) GetLastMessage() *notify.Message {
	return n.lastMessage
}
