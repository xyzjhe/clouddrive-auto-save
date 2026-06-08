//go:build e2e

// internal/core/mock_plugin.go
package core

import (
	"context"

	"github.com/zcq/clouddrive-auto-save/internal/core/plugin"
)

// MockPlugin Mock 插件实现
type MockPlugin struct {
	name             string
	version          string
	description      string
	hooks            []plugin.HookType
	config           map[string]interface{}
	initErr          error
	executeErr       error
	closeErr         error
	taskBeforeCalled bool
	taskAfterCalled  bool
	runCalled        bool
}

// NewMockPlugin 创建 Mock 插件
func NewMockPlugin(name string, hooks []plugin.HookType) *MockPlugin {
	return &MockPlugin{
		name:        name,
		version:     "1.0.0",
		description: "Mock 插件",
		hooks:       hooks,
	}
}

// Name 返回插件名称
func (p *MockPlugin) Name() string {
	return p.name
}

// Version 返回插件版本
func (p *MockPlugin) Version() string {
	return p.version
}

// Description 返回插件描述
func (p *MockPlugin) Description() string {
	return p.description
}

// Init 初始化插件
func (p *MockPlugin) Init(config map[string]interface{}) error {
	if p.initErr != nil {
		return p.initErr
	}
	p.config = config
	return nil
}

// Hooks 返回插件支持的生命周期钩子
func (p *MockPlugin) Hooks() []plugin.HookType {
	return p.hooks
}

// Execute 执行钩子
func (p *MockPlugin) Execute(ctx context.Context, hook plugin.HookType, data *plugin.HookData) error {
	if p.executeErr != nil {
		return p.executeErr
	}

	switch hook {
	case plugin.HookTaskBefore:
		p.taskBeforeCalled = true
	case plugin.HookTaskAfter:
		p.taskAfterCalled = true
	case plugin.HookRun:
		p.runCalled = true
	}

	return nil
}

// Close 关闭插件
func (p *MockPlugin) Close() error {
	return p.closeErr
}

// SetInitError 设置初始化错误
func (p *MockPlugin) SetInitError(err error) {
	p.initErr = err
}

// SetExecuteError 设置执行错误
func (p *MockPlugin) SetExecuteError(err error) {
	p.executeErr = err
}

// SetCloseError 设置关闭错误
func (p *MockPlugin) SetCloseError(err error) {
	p.closeErr = err
}

// IsTaskBeforeCalled 检查 task_before 钩子是否被调用
func (p *MockPlugin) IsTaskBeforeCalled() bool {
	return p.taskBeforeCalled
}

// IsTaskAfterCalled 检查 task_after 钩子是否被调用
func (p *MockPlugin) IsTaskAfterCalled() bool {
	return p.taskAfterCalled
}

// IsRunCalled 检查 run 钩子是否被调用
func (p *MockPlugin) IsRunCalled() bool {
	return p.runCalled
}
