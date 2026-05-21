// internal/core/plugin/manager.go
package plugin

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

// Manager 插件管理器
type Manager struct {
	plugins map[string]Plugin
	configs map[string]*PluginConfig
	mu      sync.RWMutex
}

// NewManager 创建插件管理器
func NewManager() *Manager {
	return &Manager{
		plugins: make(map[string]Plugin),
		configs: make(map[string]*PluginConfig),
	}
}

// Register 注册插件
func (m *Manager) Register(plugin Plugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := plugin.Name()
	if _, exists := m.plugins[name]; exists {
		return fmt.Errorf("插件 %s 已存在", name)
	}

	m.plugins[name] = plugin
	slog.Info("插件已注册", "name", name, "version", plugin.Version())
	return nil
}

// Init 初始化所有插件
func (m *Manager) Init(configs map[string]*PluginConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, config := range configs {
		plugin, exists := m.plugins[name]
		if !exists {
			slog.Warn("插件不存在，跳过初始化", "name", name)
			continue
		}

		if err := plugin.Init(config.Config); err != nil {
			return fmt.Errorf("初始化插件 %s 失败: %w", name, err)
		}

		m.configs[name] = config
		slog.Info("插件已初始化", "name", name)
	}

	return nil
}

// ExecuteHook 执行钩子
func (m *Manager) ExecuteHook(ctx context.Context, hook HookType, data *HookData) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for name, plugin := range m.plugins {
		// 检查插件是否支持该钩子
		if !m.supportsHook(plugin, hook) {
			continue
		}

		// 检查插件是否启用
		config, exists := m.configs[name]
		if !exists {
			continue
		}

		// 执行钩子
		if err := plugin.Execute(ctx, hook, data); err != nil {
			slog.Error("执行插件钩子失败",
				"plugin", name,
				"hook", hook,
				"error", err,
			)
			// 继续执行其他插件，不中断
		}
	}

	return nil
}

// supportsHook 检查插件是否支持指定钩子
func (m *Manager) supportsHook(plugin Plugin, hook HookType) bool {
	for _, h := range plugin.Hooks() {
		if h == hook {
			return true
		}
	}
	return false
}

// GetPlugin 获取插件
func (m *Manager) GetPlugin(name string) (Plugin, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugin, exists := m.plugins[name]
	return plugin, exists
}

// ListPlugins 列出所有插件
func (m *Manager) ListPlugins() []PluginInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var plugins []PluginInfo
	for name, plugin := range m.plugins {
		config, _ := m.configs[name]
		plugins = append(plugins, PluginInfo{
			Name:        name,
			Version:     plugin.Version(),
			Description: plugin.Description(),
			Hooks:       plugin.Hooks(),
			Enabled:     config != nil,
		})
	}

	return plugins
}

// PluginInfo 插件信息
type PluginInfo struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	Hooks       []HookType `json:"hooks"`
	Enabled     bool      `json:"enabled"`
}

// Close 关闭所有插件
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, plugin := range m.plugins {
		if err := plugin.Close(); err != nil {
			slog.Error("关闭插件失败", "name", name, "error", err)
		}
	}

	return nil
}
