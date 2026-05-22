// internal/core/notify/manager.go
package notify

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

// Manager 通知管理器
type Manager struct {
	notifiers map[string]Notifier
	configs   map[string]*NotifierConfig
	mu        sync.RWMutex
}

// NewManager 创建通知管理器
func NewManager() *Manager {
	return &Manager{
		notifiers: make(map[string]Notifier),
		configs:   make(map[string]*NotifierConfig),
	}
}

// Register 注册通知渠道
func (m *Manager) Register(notifier Notifier) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := notifier.Name()
	if _, exists := m.notifiers[name]; exists {
		return fmt.Errorf("通知渠道 %s 已存在", name)
	}

	m.notifiers[name] = notifier
	slog.Info("通知渠道已注册", "name", name, "type", notifier.Type())
	return nil
}

// Init 初始化所有通知渠道
func (m *Manager) Init(configs map[string]*NotifierConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, config := range configs {
		notifier, exists := m.notifiers[name]
		if !exists {
			slog.Warn("通知渠道不存在，跳过初始化", "name", name)
			continue
		}

		if err := notifier.Init(config.Config); err != nil {
			return fmt.Errorf("初始化通知渠道 %s 失败: %w", name, err)
		}

		m.configs[name] = config
		slog.Info("通知渠道已初始化", "name", name)
	}

	return nil
}

// Send 发送通知到所有启用的渠道
func (m *Manager) Send(ctx context.Context, message *Message) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var lastErr error

	for name, notifier := range m.notifiers {
		config, exists := m.configs[name]
		if !exists || !config.Enabled {
			continue
		}

		// 检查是否应该发送
		if !m.shouldSend(config, message.Level) {
			continue
		}

		if err := notifier.Send(ctx, message); err != nil {
			slog.Error("发送通知失败",
				"notifier", name,
				"error", err,
			)
			lastErr = err
		}
	}

	return lastErr
}

// shouldSend 检查是否应该发送通知
func (m *Manager) shouldSend(config *NotifierConfig, level MessageLevel) bool {
	switch level {
	case LevelSuccess:
		return config.NotifyOnSuccess
	case LevelError, LevelWarning:
		return config.NotifyOnFailure
	default:
		return true
	}
}

// Test 测试指定通知渠道
func (m *Manager) Test(ctx context.Context, name string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	notifier, exists := m.notifiers[name]
	if !exists {
		return fmt.Errorf("通知渠道 %s 不存在", name)
	}

	return notifier.Test(ctx)
}

// ListNotifiers 列出所有通知渠道
func (m *Manager) ListNotifiers() []NotifierInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var notifiers []NotifierInfo
	for name, notifier := range m.notifiers {
		config, _ := m.configs[name]
		notifiers = append(notifiers, NotifierInfo{
			Name:    name,
			Type:    notifier.Type(),
			Enabled: config != nil && config.Enabled,
		})
	}

	return notifiers
}

// NotifierInfo 通知渠道信息
type NotifierInfo struct {
	Name    string       `json:"name"`
	Type    NotifierType `json:"type"`
	Enabled bool         `json:"enabled"`
}

// Close 关闭所有通知渠道
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, notifier := range m.notifiers {
		if err := notifier.Close(); err != nil {
			slog.Error("关闭通知渠道失败", "name", name, "error", err)
		}
	}

	return nil
}
