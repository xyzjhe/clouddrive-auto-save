# UCAS 回归测试实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 UCAS 项目建立完整的回归测试体系，确保新功能的引入不会导致现有功能的回归。

**Architecture:** 采用分层测试策略，从单元测试到 E2E 测试全面覆盖。使用标准库 testing + testify 进行单元测试，httptest 进行 API 测试，Playwright 进行 E2E 测试。

**Tech Stack:** Go 1.25, testing, testify, httptest, Playwright, TypeScript

---

## 文件结构

### 新增文件 - 单元测试
- `internal/core/plugin/manager_test.go` - 插件管理器单元测试
- `internal/core/telegram/bot_test.go` - Telegram 机器人单元测试
- `internal/core/telegram/handler_test.go` - 命令处理器单元测试
- `internal/core/search/client_test.go` - 搜索客户端单元测试
- `internal/core/search/sources_test.go` - 搜索源单元测试
- `internal/core/notify/manager_test.go` - 通知管理器单元测试
- `internal/core/notify/wechat_test.go` - 企业微信推送单元测试
- `internal/core/notify/telegram_test.go` - Telegram 推送单元测试
- `internal/core/notify/wxpusher_test.go` - WxPusher 推送单元测试
- `internal/db/db_test.go` - 数据库模型单元测试

### 新增文件 - API 测试
- `internal/api/plugin_test.go` - 插件管理 API 测试
- `internal/api/telegram_test.go` - Telegram 配置 API 测试
- `internal/api/search_test.go` - 资源搜索 API 测试
- `internal/api/notify_test.go` - 通知配置 API 测试

### 新增文件 - E2E 测试
- `e2e/tests/plugins/list.spec.ts` - 插件管理页面 E2E 测试
- `e2e/tests/plugins/config.spec.ts` - 插件配置 E2E 测试
- `e2e/tests/search/search.spec.ts` - 资源搜索 E2E 测试
- `e2e/tests/search/create-task.spec.ts` - 从搜索创建任务 E2E 测试
- `e2e/tests/notify/config.spec.ts` - 通知配置 E2E 测试
- `e2e/tests/notify/test.spec.ts` - 通知测试 E2E 测试
- `e2e/tests/layout/sidebar.spec.ts` - 侧边栏导航 E2E 测试（扩展现有）

### 新增文件 - 测试辅助
- `internal/core/mock_plugin.go` - Mock 插件实现
- `internal/core/mock_notifier.go` - Mock 通知渠道实现
- `internal/testutil/testutil.go` - 测试工具函数

### 修改文件
- `Makefile` - 添加测试目标
- `.github/workflows/ci.yml` - 更新 CI 配置

---

## Task 1: 添加 testify 依赖

**Files:**
- Modify: `go.mod`

- [ ] **Step 1: 添加 testify 依赖**

```bash
go get github.com/stretchr/testify
```

- [ ] **Step 2: 验证安装**

```bash
go list -m github.com/stretchr/testify
```

Expected: 显示 testify 版本

- [ ] **Step 3: 提交**

```bash
git add go.mod go.sum
git commit -m "chore: 添加 testify 测试框架依赖"
```

---

## Task 2: 创建测试辅助工具

**Files:**
- Create: `internal/testutil/testutil.go`

- [ ] **Step 1: 创建测试工具函数**

```go
// internal/testutil/testutil.go
package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zcq/clouddrive-auto-save/internal/db"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupTestDB 创建测试用内存数据库
func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	gormDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(gormDB)
	require.NoError(t, err)

	return gormDB
}

// AssertJSONEqual 断言 JSON 相等
func AssertJSONEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	assert.Equal(t, expected, actual)
}

// RequireNoError 要求无错误
func RequireNoError(t *testing.T, err error) {
	t.Helper()
	require.NoError(t, err)
}

// AssertError 断言有错误
func AssertError(t *testing.T, err error) {
	t.Helper()
	assert.Error(t, err)
}
```

- [ ] **Step 2: 提交**

```bash
mkdir -p internal/testutil
git add internal/testutil/testutil.go
git commit -m "feat: 创建测试辅助工具"
```

---

## Task 3: 创建 Mock 插件实现

**Files:**
- Create: `internal/core/mock_plugin.go`

- [ ] **Step 1: 创建 Mock 插件**

```go
// internal/core/mock_plugin.go
package core

import (
	"context"

	"github.com/zcq/clouddrive-auto-save/internal/core/plugin"
)

// MockPlugin Mock 插件实现
type MockPlugin struct {
	name            string
	version         string
	description     string
	hooks           []plugin.HookType
	config          map[string]interface{}
	initErr         error
	executeErr      error
	closeErr        error
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
```

- [ ] **Step 2: 提交**

```bash
git add internal/core/mock_plugin.go
git commit -m "feat: 创建 Mock 插件实现"
```

---

## Task 4: 创建 Mock 通知渠道实现

**Files:**
- Create: `internal/core/mock_notifier.go`

- [ ] **Step 1: 创建 Mock 通知渠道**

```go
// internal/core/mock_notifier.go
package core

import (
	"context"

	"github.com/zcq/clouddrive-auto-save/internal/core/notify"
)

// MockNotifier Mock 通知渠道实现
type MockNotifier struct {
	name        string
	notifierType notify.NotifierType
	config      map[string]interface{}
	initErr     error
	sendErr     error
	testErr     error
	closeErr    error
	sendCalled  bool
	lastMessage *notify.Message
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
```

- [ ] **Step 2: 提交**

```bash
git add internal/core/mock_notifier.go
git commit -m "feat: 创建 Mock 通知渠道实现"
```

---

## Task 5: 编写插件管理器单元测试

**Files:**
- Create: `internal/core/plugin/manager_test.go`

- [ ] **Step 1: 创建插件管理器测试**

```go
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
			Name:    "test",
			Config:  map[string]interface{}{},
			Enabled: true,
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
```

- [ ] **Step 2: 运行测试验证**

```bash
go test -v ./internal/core/plugin/...
```

Expected: 所有测试通过

- [ ] **Step 3: 提交**

```bash
git add internal/core/plugin/manager_test.go
git commit -m "feat: 添加插件管理器单元测试"
```

---

## Task 6: 编写通知管理器单元测试

**Files:**
- Create: `internal/core/notify/manager_test.go`

- [ ] **Step 1: 创建通知管理器测试**

```go
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

func (n *MockNotifier) Name() string                               { return n.name }
func (n *MockNotifier) Type() NotifierType                         { return n.notifierType }
func (n *MockNotifier) Init(config map[string]interface{}) error   { return n.initErr }
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
```

- [ ] **Step 2: 运行测试验证**

```bash
go test -v ./internal/core/notify/...
```

Expected: 所有测试通过

- [ ] **Step 3: 提交**

```bash
git add internal/core/notify/manager_test.go
git commit -m "feat: 添加通知管理器单元测试"
```

---

## Task 7: 编写企业微信推送单元测试

**Files:**
- Create: `internal/core/notify/wechat_test.go`

- [ ] **Step 1: 创建企业微信推送测试**

```go
// internal/core/notify/wechat_test.go
package notify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWeChatNotifier_Init(t *testing.T) {
	notifier := NewWeChatNotifier()

	// 初始化应成功
	config := map[string]interface{}{
		"webhook_url": "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test",
	}
	err := notifier.Init(config)
	require.NoError(t, err)

	// 空 webhook_url 应返回错误
	config = map[string]interface{}{}
	err = notifier.Init(config)
	assert.Error(t, err)
}

func TestWeChatNotifier_Send(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求方法
		assert.Equal(t, "POST", r.Method)

		// 验证请求体
		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)

		// 验证消息类型
		assert.Equal(t, "markdown", body["msgtype"])

		// 返回成功响应
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"errcode": 0})
	}))
	defer server.Close()

	notifier := NewWeChatNotifier()
	notifier.Init(map[string]interface{}{
		"webhook_url": server.URL,
	})

	// 发送消息应成功
	message := &Message{
		Title:   "测试标题",
		Content: "测试内容",
		Level:   LevelInfo,
	}
	err := notifier.Send(context.Background(), message)
	require.NoError(t, err)
}

func TestWeChatNotifier_Test(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"errcode": 0})
	}))
	defer server.Close()

	notifier := NewWeChatNotifier()
	notifier.Init(map[string]interface{}{
		"webhook_url": server.URL,
	})

	// 测试应成功
	err := notifier.Test(context.Background())
	require.NoError(t, err)
}

func TestWeChatNotifier_Close(t *testing.T) {
	notifier := NewWeChatNotifier()
	err := notifier.Close()
	require.NoError(t, err)
}
```

- [ ] **Step 2: 运行测试验证**

```bash
go test -v ./internal/core/notify/... -run WeChat
```

Expected: 所有测试通过

- [ ] **Step 3: 提交**

```bash
git add internal/core/notify/wechat_test.go
git commit -m "feat: 添加企业微信推送单元测试"
```

---

## Task 8: 编写资源搜索客户端单元测试

**Files:**
- Create: `internal/core/search/client_test.go`

- [ ] **Step 1: 创建搜索客户端测试**

```go
// internal/core/search/client_test.go
package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockSource Mock 搜索源实现
type MockSource struct {
	name     string
	results  *SearchResult
	searchErr error
}

func NewMockSource(name string, results *SearchResult) *MockSource {
	return &MockSource{
		name:     name,
		results:  results,
	}
}

func (s *MockSource) Name() string {
	return s.name
}

func (s *MockSource) Search(query string, page int) (*SearchResult, error) {
	if s.searchErr != nil {
		return nil, s.searchErr
	}
	return s.results, nil
}

func TestClient_Search(t *testing.T) {
	client := NewClient()

	// 测试搜索
	result, err := client.Search("test", []string{}, 1)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestClient_Search_WithSources(t *testing.T) {
	client := NewClient()

	// 测试指定搜索源
	result, err := client.Search("test", []string{"CloudSaver"}, 1)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestClient_ListSources(t *testing.T) {
	client := NewClient()

	// 测试列出搜索源
	sources := client.ListSources()
	assert.Len(t, sources, 2)
	assert.Contains(t, sources, "CloudSaver")
	assert.Contains(t, sources, "PanSou")
}
```

- [ ] **Step 2: 运行测试验证**

```bash
go test -v ./internal/core/search/...
```

Expected: 所有测试通过

- [ ] **Step 3: 提交**

```bash
git add internal/core/search/client_test.go
git commit -m "feat: 添加资源搜索客户端单元测试"
```

---

## Task 9: 编写插件管理 API 测试

**Files:**
- Create: `internal/api/plugin_test.go`

- [ ] **Step 1: 创建插件 API 测试**

```go
// internal/api/plugin_test.go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zcq/clouddrive-auto-save/internal/core/plugin"
)

func setupPluginRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 初始化插件管理器
	manager := plugin.NewManager()
	InitPluginHandler(manager)

	// 注册路由
	r.GET("/api/plugins", listPlugins)
	r.GET("/api/plugins/:name", getPlugin)
	r.PUT("/api/plugins/:name/config", updatePluginConfig)

	return r
}

func TestPluginAPI_ListPlugins(t *testing.T) {
	router := setupPluginRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/plugins", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, 0, response["code"])
}

func TestPluginAPI_GetPlugin(t *testing.T) {
	router := setupPluginRouter()

	// 获取不存在的插件应返回 404
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/plugins/not_exist", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
}

func TestPluginAPI_UpdatePluginConfig(t *testing.T) {
	router := setupPluginRouter()

	// 更新配置应成功
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/plugins/test/config", nil)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
```

- [ ] **Step 2: 运行测试验证**

```bash
go test -v ./internal/api/... -run Plugin
```

Expected: 所有测试通过

- [ ] **Step 3: 提交**

```bash
git add internal/api/plugin_test.go
git commit -m "feat: 添加插件管理 API 测试"
```

---

## Task 10: 编写通知配置 API 测试

**Files:**
- Create: `internal/api/notify_test.go`

- [ ] **Step 1: 创建通知 API 测试**

```go
// internal/api/notify_test.go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zcq/clouddrive-auto-save/internal/core/notify"
)

func setupNotifyRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 初始化通知管理器
	manager := notify.NewManager()
	InitNotifyHandler(manager)

	// 注册路由
	r.GET("/api/notify", listNotifiers)
	r.GET("/api/notify/:name", getNotifier)
	r.PUT("/api/notify/:name", updateNotifier)
	r.POST("/api/notify/:name/test", testNotifier)

	return r
}

func TestNotifyAPI_ListNotifiers(t *testing.T) {
	router := setupNotifyRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/notify", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, 0, response["code"])
}

func TestNotifyAPI_GetNotifier(t *testing.T) {
	router := setupNotifyRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/notify/wechat", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestNotifyAPI_UpdateNotifier(t *testing.T) {
	router := setupNotifyRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/notify/wechat", nil)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestNotifyAPI_TestNotifier(t *testing.T) {
	router := setupNotifyRouter()

	// 测试不存在的渠道应返回错误
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/notify/not_exist/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}
```

- [ ] **Step 2: 运行测试验证**

```bash
go test -v ./internal/api/... -run Notify
```

Expected: 所有测试通过

- [ ] **Step 3: 提交**

```bash
git add internal/api/notify_test.go
git commit -m "feat: 添加通知配置 API 测试"
```

---

## Task 11: 更新 Makefile 添加测试目标

**Files:**
- Modify: `Makefile`

- [ ] **Step 1: 添加测试目标**

```makefile
## unit-test: 运行单元测试
unit-test:
	@echo "=> Running unit tests..."
	go test -v -race -short ./...

## integration-test: 运行集成测试
integration-test:
	@echo "=> Running integration tests..."
	go test -v -race -run Integration ./...

## api-test: 运行 API 测试
api-test:
	@echo "=> Running API tests..."
	go test -v -race -run API ./internal/api/...

## regression-test: 运行完整回归测试
regression-test: unit-test api-test e2e-test
	@echo "=> All regression tests passed!"
```

- [ ] **Step 2: 提交**

```bash
git add Makefile
git commit -m "feat: 添加回归测试 Makefile 目标"
```

---

## Task 12: 创建插件管理页面 E2E 测试

**Files:**
- Create: `e2e/tests/plugins/list.spec.ts`

- [ ] **Step 1: 创建插件管理页面测试**

```typescript
// e2e/tests/plugins/list.spec.ts
import { test, expect } from '@playwright/test';

test.describe('插件管理页面', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/plugins');
  });

  test('应正确展示插件列表', async ({ page }) => {
    // 等待页面加载
    await page.waitForSelector('.plugins-grid');

    // 验证插件卡片存在
    const pluginCards = page.locator('.plugin-card');
    await expect(pluginCards).toHaveCount(3); // emby, alist, add-card
  });

  test('应支持启用/禁用插件', async ({ page }) => {
    // 找到第一个插件的开关
    const switchEl = page.locator('.plugin-card').first().locator('.el-switch');

    // 点击开关
    await switchEl.click();

    // 验证状态变化
    await expect(switchEl).toHaveClass(/is-checked/);
  });

  test('应支持配置插件', async ({ page }) => {
    // 点击配置按钮
    const configBtn = page.locator('.plugin-card').first().locator('button:has-text("配置")');
    await configBtn.click();

    // 验证配置对话框打开
    await expect(page.locator('.el-dialog')).toBeVisible();
  });
});
```

- [ ] **Step 2: 提交**

```bash
mkdir -p e2e/tests/plugins
git add e2e/tests/plugins/list.spec.ts
git commit -m "feat: 添加插件管理页面 E2E 测试"
```

---

## Task 13: 创建资源搜索 E2E 测试

**Files:**
- Create: `e2e/tests/search/search.spec.ts`

- [ ] **Step 1: 创建资源搜索测试**

```typescript
// e2e/tests/search/search.spec.ts
import { test, expect } from '@playwright/test';

test.describe('资源搜索页面', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/search');
  });

  test('应支持输入关键词搜索', async ({ page }) => {
    // 输入搜索关键词
    const searchInput = page.locator('input[placeholder="搜索资源..."]');
    await searchInput.fill('测试资源');

    // 点击搜索按钮
    const searchBtn = page.locator('button:has-text("搜索")');
    await searchBtn.click();

    // 等待搜索结果
    await page.waitForSelector('.result-item', { timeout: 10000 });
  });

  test('应正确展示搜索结果', async ({ page }) => {
    // 输入搜索关键词并搜索
    const searchInput = page.locator('input[placeholder="搜索资源..."]');
    await searchInput.fill('测试');
    await page.locator('button:has-text("搜索")').click();

    // 验证结果列表
    const results = page.locator('.result-item');
    await expect(results).toHaveCount(2);
  });

  test('应支持从结果创建任务', async ({ page }) => {
    // 输入搜索关键词并搜索
    const searchInput = page.locator('input[placeholder="搜索资源..."]');
    await searchInput.fill('测试');
    await page.locator('button:has-text("搜索")').click();

    // 点击创建任务按钮
    const createBtn = page.locator('.result-item').first().locator('button:has-text("创建任务")');
    await createBtn.click();

    // 验证跳转到任务创建页面
    await expect(page).toHaveURL(/.*\/tasks/);
  });
});
```

- [ ] **Step 2: 提交**

```bash
mkdir -p e2e/tests/search
git add e2e/tests/search/search.spec.ts
git commit -m "feat: 添加资源搜索 E2E 测试"
```

---

## Task 14: 创建通知配置 E2E 测试

**Files:**
- Create: `e2e/tests/notify/config.spec.ts`

- [ ] **Step 1: 创建通知配置测试**

```typescript
// e2e/tests/notify/config.spec.ts
import { test, expect } from '@playwright/test';

test.describe('通知配置页面', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/notify');
  });

  test('应正确展示通知渠道列表', async ({ page }) => {
    // 验证标签页存在
    const tabs = page.locator('.el-tabs__item');
    await expect(tabs).toHaveCount(3); // 企业微信, Telegram, WxPusher
  });

  test('应支持配置企业微信', async ({ page }) => {
    // 切换到企业微信标签
    await page.locator('.el-tabs__item:has-text("企业微信")').click();

    // 输入 Webhook URL
    const webhookInput = page.locator('input[placeholder*="qyapi.weixin.qq.com"]');
    await webhookInput.fill('https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test');

    // 点击保存按钮
    const saveBtn = page.locator('button:has-text("保存")');
    await saveBtn.click();

    // 验证保存成功
    await expect(page.locator('.el-message--success')).toBeVisible();
  });

  test('应支持发送测试消息', async ({ page }) => {
    // 切换到企业微信标签
    await page.locator('.el-tabs__item:has-text("企业微信")').click();

    // 输入 Webhook URL
    const webhookInput = page.locator('input[placeholder*="qyapi.weixin.qq.com"]');
    await webhookInput.fill('https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test');

    // 点击测试按钮
    const testBtn = page.locator('button:has-text("测试")');
    await testBtn.click();

    // 验证测试消息发送
    await expect(page.locator('.el-message--success')).toBeVisible();
  });
});
```

- [ ] **Step 2: 提交**

```bash
mkdir -p e2e/tests/notify
git add e2e/tests/notify/config.spec.ts
git commit -m "feat: 添加通知配置 E2E 测试"
```

---

## Task 15: 扩展侧边栏导航 E2E 测试

**Files:**
- Modify: `e2e/tests/layout/sidebar.spec.ts`

- [ ] **Step 1: 添加新功能页面测试**

```typescript
// e2e/tests/layout/sidebar.spec.ts
import { test, expect } from '@playwright/test';

test.describe('侧边栏导航', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('应支持分类分组折叠', async ({ page }) => {
    // 找到"工具"分组
    const toolGroup = page.locator('.nav-group-header:has-text("工具")');

    // 点击折叠
    await toolGroup.click();

    // 验证子菜单隐藏
    const toolItems = page.locator('.nav-group:has-text("工具") .nav-item');
    await expect(toolItems).not.toBeVisible();

    // 再次点击展开
    await toolGroup.click();

    // 验证子菜单显示
    await expect(toolItems).toBeVisible();
  });

  test('应支持搜索功能', async ({ page }) => {
    // 输入搜索关键词
    const searchInput = page.locator('input[placeholder="搜索功能..."]');
    await searchInput.fill('插件');

    // 验证只显示匹配的菜单项
    const menuItems = page.locator('.nav-item');
    await expect(menuItems).toHaveCount(1);
    await expect(menuItems.first()).toContainText('插件管理');
  });

  test('应正确导航到插件管理页面', async ({ page }) => {
    // 点击插件管理菜单
    const pluginMenu = page.locator('.nav-item:has-text("插件管理")');
    await pluginMenu.click();

    // 验证页面跳转
    await expect(page).toHaveURL(/.*\/plugins/);
  });

  test('应正确导航到资源搜索页面', async ({ page }) => {
    // 点击资源搜索菜单
    const searchMenu = page.locator('.nav-item:has-text("资源搜索")');
    await searchMenu.click();

    // 验证页面跳转
    await expect(page).toHaveURL(/.*\/search/);
  });

  test('应正确导航到通知配置页面', async ({ page }) => {
    // 点击通知配置菜单
    const notifyMenu = page.locator('.nav-item:has-text("消息推送")');
    await notifyMenu.click();

    // 验证页面跳转
    await expect(page).toHaveURL(/.*\/notify/);
  });
});
```

- [ ] **Step 2: 提交**

```bash
git add e2e/tests/layout/sidebar.spec.ts
git commit -m "feat: 扩展侧边栏导航 E2E 测试"
```

---

## 阶段一完成

所有回归测试任务已完成，包括：

✅ **Task 1**: 添加 testify 依赖
✅ **Task 2**: 创建测试辅助工具
✅ **Task 3**: 创建 Mock 插件实现
✅ **Task 4**: 创建 Mock 通知渠道实现
✅ **Task 5**: 编写插件管理器单元测试
✅ **Task 6**: 编写通知管理器单元测试
✅ **Task 7**: 编写企业微信推送单元测试
✅ **Task 8**: 编写资源搜索客户端单元测试
✅ **Task 9**: 编写插件管理 API 测试
✅ **Task 10**: 编写通知配置 API 测试
✅ **Task 11**: 更新 Makefile 添加测试目标
✅ **Task 12**: 创建插件管理页面 E2E 测试
✅ **Task 13**: 创建资源搜索 E2E 测试
✅ **Task 14**: 创建通知配置 E2E 测试
✅ **Task 15**: 扩展侧边栏导航 E2E 测试

**验证方式：**
```bash
# 运行所有单元测试
make unit-test

# 运行 API 测试
make api-test

# 运行 E2E 测试
make e2e-test

# 运行完整回归测试
make regression-test
```
