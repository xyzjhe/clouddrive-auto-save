// internal/core/notify/manager.go
package notify

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/zcq/clouddrive-auto-save/internal/db"
	"gorm.io/gorm"
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

// Global 全局通知管理器
var Global = NewManager()

// InitGlobal 从数据库加载通知配置并初始化全局通知管理器
func InitGlobal(dbInst *gorm.DB) error {
	// 注册默认支持的渠道
	Global.mu.Lock()
	_, wechatExists := Global.notifiers["wechat"]
	Global.mu.Unlock()
	if !wechatExists {
		_ = Global.Register(NewWeChatNotifier())
		_ = Global.Register(NewTelegramNotifier())
		_ = Global.Register(NewWxPusherNotifier())
		_ = Global.Register(NewBarkNotifier())
	}

	configs := make(map[string]*NotifierConfig)
	channels := []string{"wechat", "telegram", "wxpusher", "bark"}

	for _, name := range channels {
		var setting db.Setting
		err := dbInst.Where("key = ?", "notify_config_"+name).First(&setting).Error
		if err == nil {
			var config NotifierConfig
			if err := json.Unmarshal([]byte(setting.Value), &config); err == nil {
				configs[name] = &config
			} else {
				slog.Error("反序列化通知配置失败", "name", name, "error", err)
			}
		}
	}

	return Global.Init(configs)
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

	// 清空历史 config，避免禁用某个渠道后还残留配置
	m.configs = make(map[string]*NotifierConfig)

	for name, config := range configs {
		notifier, exists := m.notifiers[name]
		if !exists {
			slog.Warn("通知渠道不存在，跳过初始化", "name", name)
			continue
		}

		if err := notifier.Init(config.Config); err != nil {
			slog.Error("初始化通知渠道失败，跳过该渠道", "name", name, "error", err)
			continue
		}

		m.configs[name] = config
		slog.Info("通知渠道已初始化", "name", name)
	}

	return nil
}

// Send 发送通知到所有启用的渠道
func (m *Manager) Send(ctx context.Context, message *Message) error {
	// 在读锁内只做快照，避免网络调用阻塞锁
	type sendJob struct {
		name     string
		notifier Notifier
		config   *NotifierConfig
	}

	m.mu.RLock()
	jobs := make([]sendJob, 0, len(m.notifiers))
	for name, notifier := range m.notifiers {
		config, exists := m.configs[name]
		if !exists || !config.Enabled {
			continue
		}
		// 检查是否应该发送
		if !m.shouldSend(config, message.Level) {
			continue
		}
		jobs = append(jobs, sendJob{name: name, notifier: notifier, config: config})
	}
	m.mu.RUnlock()

	var lastErr error
	for _, job := range jobs {
		if err := job.notifier.Send(ctx, message); err != nil {
			slog.Error("发送通知失败",
				"notifier", job.name,
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
	notifier, exists := m.notifiers[name]
	m.mu.RUnlock()

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
