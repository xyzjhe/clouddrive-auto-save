// internal/core/plugin/manager_test.go
package plugin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockPlugin Mock 插件实现
type MockPlugin struct {
	name        string
	version     string
	description string
	hooks       []HookType
	config      map[string]interface{}
	initErr     error
	executeErr  error
	closeErr    error
}

func NewMockPlugin(name string, hooks []HookType) *MockPlugin {
	return &MockPlugin{
		name:        name,
		version:     "1.0.0",
		description: "Mock 插件",
		hooks:       hooks,
	}
}

func (p *MockPlugin) Name() string                                  { return p.name }
func (p *MockPlugin) Version() string                               { return p.version }
func (p *MockPlugin) Description() string                           { return p.description }
func (p *MockPlugin) Init(config map[string]interface{}) error      { return p.initErr }
func (p *MockPlugin) Hooks() []HookType                             { return p.hooks }
func (p *MockPlugin) Execute(ctx context.Context, hook HookType, data *HookData) error { return p.executeErr }
func (p *MockPlugin) Close() error                                  { return p.closeErr }

func TestManager_Register(t *testing.T) {
	manager := NewManager()
	plugin := NewMockPlugin("test", []HookType{HookTaskBefore})

	// 注册插件应成功
	err := manager.Register(plugin)
	require.NoError(t, err)

	// 注册重复插件应返回错误
	err = manager.Register(plugin)
	assert.Error(t, err)
}

func TestManager_Init(t *testing.T) {
	manager := NewManager()
	plugin := NewMockPlugin("test", []HookType{HookTaskBefore})
	manager.Register(plugin)

	// 初始化存在的插件应成功
	configs := map[string]*PluginConfig{
		"test": {
			Name:   "test",
			Config: map[string]interface{}{"key": "value"},
		},
	}
	err := manager.Init(configs)
	require.NoError(t, err)

	// 初始化不存在的插件应跳过
	configs["not_exist"] = &PluginConfig{
		Name:   "not_exist",
		Config: map[string]interface{}{},
	}
	err = manager.Init(configs)
	require.NoError(t, err)
}

func TestManager_ExecuteHook(t *testing.T) {
	manager := NewManager()
	plugin := NewMockPlugin("test", []HookType{HookTaskBefore, HookTaskAfter})
	manager.Register(plugin)

	configs := map[string]*PluginConfig{
		"test": {
			Name:   "test",
			Config: map[string]interface{}{},
		},
	}
	manager.Init(configs)

	// 执行钩子应成功
	data := &HookData{TaskID: 1, TaskName: "test"}
	err := manager.ExecuteHook(context.Background(), HookTaskBefore, data)
	require.NoError(t, err)

	// 执行不支持的钩子应跳过
	err = manager.ExecuteHook(context.Background(), HookRun, data)
	require.NoError(t, err)
}

func TestManager_ListPlugins(t *testing.T) {
	manager := NewManager()
	plugin1 := NewMockPlugin("test1", []HookType{HookTaskBefore})
	plugin2 := NewMockPlugin("test2", []HookType{HookTaskAfter})
	manager.Register(plugin1)
	manager.Register(plugin2)

	plugins := manager.ListPlugins()
	assert.Len(t, plugins, 2)
}

func TestManager_Close(t *testing.T) {
	manager := NewManager()
	plugin := NewMockPlugin("test", []HookType{HookTaskBefore})
	manager.Register(plugin)

	err := manager.Close()
	require.NoError(t, err)
}
