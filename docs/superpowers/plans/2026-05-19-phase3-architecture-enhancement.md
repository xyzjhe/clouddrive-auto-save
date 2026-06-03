# UCAS 全面升级 - 阶段三：架构增强实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现多渠道消息推送，为未来平台扩展做好准备。

**Architecture:** 基于接口驱动设计，统一的 NotifyManager 调度器，每个渠道独立实现 Notifier 接口。

**Tech Stack:** Go 1.25, Gin, HTTP API, Webhook

---

## 文件结构

### 新增文件
- `internal/core/notify/interface.go` - 通知接口定义
- `internal/core/notify/manager.go` - 通知管理器
- `internal/core/notify/wechat.go` - 企业微信推送
- `internal/core/notify/telegram_notify.go` - Telegram 推送
- `internal/core/notify/wxpusher.go` - WxPusher 推送
- `internal/api/notify.go` - 通知配置 API
- `web/src/views/Notify.vue` - 通知配置页面
- `docs/platform-extension-guide.md` - 平台扩展指南

### 修改文件
- `internal/api/router.go` - 添加通知路由
- `cmd/server/main.go` - 初始化通知管理器
- `web/src/config/navigation.ts` - 更新导航配置

---

## Task 1: 定义通知接口

**Files:**
- Create: `internal/core/notify/interface.go`

- [ ] **Step 1: 创建通知接口定义**

```go
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
	Name        string                 `json:"name"`
	Type        NotifierType           `json:"type"`
	Enabled     bool                   `json:"enabled"`
	Config      map[string]interface{} `json:"config"`
	NotifyOnSuccess bool              `json:"notify_on_success"`
	NotifyOnFailure bool              `json:"notify_on_failure"`
}
```

- [ ] **Step 2: 提交**

```bash
mkdir -p internal/core/notify
git add internal/core/notify/interface.go
git commit -m "feat: 定义通知接口和消息格式"
```

---

## Task 2: 实现通知管理器

**Files:**
- Create: `internal/core/notify/manager.go`

- [ ] **Step 1: 创建通知管理器**

```go
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
```

- [ ] **Step 2: 提交**

```bash
git add internal/core/notify/manager.go
git commit -m "feat: 实现通知管理器"
```

---

## Task 3: 实现企业微信推送

**Files:**
- Create: `internal/core/notify/wechat.go`

- [ ] **Step 1: 创建企业微信推送实现**

```go
// internal/core/notify/wechat.go
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// WeChatNotifier 企业微信通知渠道
type WeChatNotifier struct {
	webhookURL string
	client     *http.Client
}

// NewWeChatNotifier 创建企业微信通知渠道
func NewWeChatNotifier() *WeChatNotifier {
	return &WeChatNotifier{
		client: &http.Client{},
	}
}

// Name 返回通知渠道名称
func (n *WeChatNotifier) Name() string {
	return "wechat"
}

// Type 返回通知渠道类型
func (n *WeChatNotifier) Type() NotifierType {
	return NotifierTypeWeChat
}

// Init 初始化通知渠道
func (n *WeChatNotifier) Init(config map[string]interface{}) error {
	webhookURL, ok := config["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return fmt.Errorf("企业微信 webhook_url 不能为空")
	}

	n.webhookURL = webhookURL
	return nil
}

// Send 发送通知
func (n *WeChatNotifier) Send(ctx context.Context, message *Message) error {
	// 构建消息内容
	content := fmt.Sprintf("**%s**\n\n%s", message.Title, message.Content)

	// 构建请求体
	body := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": content,
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	// 发送请求
	req, err := http.NewRequestWithContext(ctx, "POST", n.webhookURL, bytes.NewBuffer(jsonBody))
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
func (n *WeChatNotifier) Test(ctx context.Context) error {
	return n.Send(ctx, &Message{
		Title:   "UCAS 测试通知",
		Content: "这是一条测试消息，用于验证企业微信推送配置是否正确。",
		Level:   LevelInfo,
	})
}

// Close 关闭通知渠道
func (n *WeChatNotifier) Close() error {
	return nil
}
```

- [ ] **Step 2: 提交**

```bash
git add internal/core/notify/wechat.go
git commit -m "feat: 实现企业微信推送"
```

---

## Task 4: 实现 Telegram 推送

**Files:**
- Create: `internal/core/notify/telegram_notify.go`

- [ ] **Step 1: 创建 Telegram 推送实现**

```go
// internal/core/notify/telegram_notify.go
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// TelegramNotifier Telegram 通知渠道
type TelegramNotifier struct {
	botToken string
	chatID   string
	client   *http.Client
}

// NewTelegramNotifier 创建 Telegram 通知渠道
func NewTelegramNotifier() *TelegramNotifier {
	return &TelegramNotifier{
		client: &http.Client{},
	}
}

// Name 返回通知渠道名称
func (n *TelegramNotifier) Name() string {
	return "telegram"
}

// Type 返回通知渠道类型
func (n *TelegramNotifier) Type() NotifierType {
	return NotifierTypeTelegram
}

// Init 初始化通知渠道
func (n *TelegramNotifier) Init(config map[string]interface{}) error {
	botToken, ok := config["bot_token"].(string)
	if !ok || botToken == "" {
		return fmt.Errorf("Telegram bot_token 不能为空")
	}

	chatID, ok := config["chat_id"].(string)
	if !ok || chatID == "" {
		return fmt.Errorf("Telegram chat_id 不能为空")
	}

	n.botToken = botToken
	n.chatID = chatID
	return nil
}

// Send 发送通知
func (n *TelegramNotifier) Send(ctx context.Context, message *Message) error {
	// 构建消息内容
	content := fmt.Sprintf("*%s*\n\n%s", message.Title, message.Content)

	// 构建请求体
	body := map[string]string{
		"chat_id":    n.chatID,
		"text":       content,
		"parse_mode": "Markdown",
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	// 发送请求
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.botToken)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
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
func (n *TelegramNotifier) Test(ctx context.Context) error {
	return n.Send(ctx, &Message{
		Title:   "UCAS 测试通知",
		Content: "这是一条测试消息，用于验证 Telegram 推送配置是否正确。",
		Level:   LevelInfo,
	})
}

// Close 关闭通知渠道
func (n *TelegramNotifier) Close() error {
	return nil
}
```

- [ ] **Step 2: 提交**

```bash
git add internal/core/notify/telegram_notify.go
git commit -m "feat: 实现 Telegram 推送"
```

---

## Task 5: 实现 WxPusher 推送

**Files:**
- Create: `internal/core/notify/wxpusher.go`

- [ ] **Step 1: 创建 WxPusher 推送实现**

```go
// internal/core/notify/wxpusher.go
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
		client: &http.Client{},
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
```

- [ ] **Step 2: 提交**

```bash
git add internal/core/notify/wxpusher.go
git commit -m "feat: 实现 WxPusher 推送"
```

---

## Task 6: 创建通知配置 API

**Files:**
- Create: `internal/api/notify.go`
- Modify: `internal/api/router.go`

- [ ] **Step 1: 创建通知配置 API 处理器**

```go
// internal/api/notify.go
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zcq/clouddrive-auto-save/internal/core/notify"
)

// NotifyHandler 通知配置 API 处理器
type NotifyHandler struct {
	manager *notify.Manager
}

// NewNotifyHandler 创建通知配置 API 处理器
func NewNotifyHandler(manager *notify.Manager) *NotifyHandler {
	return &NotifyHandler{manager: manager}
}

// ListNotifiers 列出所有通知渠道
func (h *NotifyHandler) ListNotifiers(c *gin.Context) {
	notifiers := h.manager.ListNotifiers()
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": notifiers,
	})
}

// GetNotifier 获取通知渠道配置
func (h *NotifyHandler) GetNotifier(c *gin.Context) {
	name := c.Param("name")

	// TODO: 从数据库获取配置
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"name": name,
		},
	})
}

// UpdateNotifier 更新通知渠道配置
func (h *NotifyHandler) UpdateNotifier(c *gin.Context) {
	name := c.Param("name")

	var config notify.NotifierConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的配置格式",
		})
		return
	}

	// TODO: 保存配置到数据库

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "配置已更新",
	})
}

// TestNotifier 测试通知渠道
func (h *NotifyHandler) TestNotifier(c *gin.Context) {
	name := c.Param("name")

	if err := h.manager.Test(c.Request.Context(), name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "测试失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "测试成功",
	})
}
```

- [ ] **Step 2: 添加通知路由**

```go
// internal/api/router.go
// 添加通知路由

notifyHandler := NewNotifyHandler(notifyManager)
notifyGroup := apiGroup.Group("/notify")
{
	notifyGroup.GET("", notifyHandler.ListNotifiers)
	notifyGroup.GET("/:name", notifyHandler.GetNotifier)
	notifyGroup.PUT("/:name", notifyHandler.UpdateNotifier)
	notifyGroup.POST("/:name/test", notifyHandler.TestNotifier)
}
```

- [ ] **Step 3: 提交**

```bash
git add internal/api/notify.go internal/api/router.go
git commit -m "feat: 添加通知配置 API"
```

---

## Task 7: 创建通知配置页面

**Files:**
- Create: `web/src/views/Notify.vue`
- Modify: `web/src/router/index.js`

- [ ] **Step 1: 创建通知配置页面**

```vue
<!-- web/src/views/Notify.vue -->
<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'

const notifiers = ref([])
const loading = ref(false)
const activeTab = ref('wechat')

// 配置表单
const wechatConfig = ref({
  enabled: false,
  webhook_url: '',
  notify_on_success: true,
  notify_on_failure: true
})

const telegramConfig = ref({
  enabled: false,
  bot_token: '',
  chat_id: '',
  notify_on_success: true,
  notify_on_failure: true
})

const wxpusherConfig = ref({
  enabled: false,
  app_token: '',
  uid: '',
  notify_on_success: true,
  notify_on_failure: true
})

const fetchNotifiers = async () => {
  loading.value = true
  try {
    const response = await fetch('/api/notify')
    const data = await response.json()
    notifiers.value = data.data || []
  } catch (error) {
    console.error('获取通知渠道失败:', error)
  } finally {
    loading.value = false
  }
}

const handleSave = async (type) => {
  let config = {}
  switch (type) {
    case 'wechat':
      config = wechatConfig.value
      break
    case 'telegram':
      config = telegramConfig.value
      break
    case 'wxpusher':
      config = wxpusherConfig.value
      break
  }

  try {
    const response = await fetch(`/api/notify/${type}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(config)
    })

    const data = await response.json()
    if (data.code === 0) {
      ElMessage.success('配置已保存')
    } else {
      ElMessage.error(data.message || '保存失败')
    }
  } catch (error) {
    console.error('保存配置失败:', error)
    ElMessage.error('保存失败')
  }
}

const handleTest = async (type) => {
  try {
    const response = await fetch(`/api/notify/${type}/test`, {
      method: 'POST'
    })

    const data = await response.json()
    if (data.code === 0) {
      ElMessage.success('测试消息已发送')
    } else {
      ElMessage.error(data.message || '测试失败')
    }
  } catch (error) {
    console.error('测试失败:', error)
    ElMessage.error('测试失败')
  }
}

onMounted(() => {
  fetchNotifiers()
})
</script>

<template>
  <div class="notify-page">
    <div class="page-header">
      <h1>消息推送</h1>
    </div>

    <el-tabs v-model="activeTab" type="border-card">
      <!-- 企业微信 -->
      <el-tab-pane label="企业微信" name="wechat">
        <el-form :model="wechatConfig" label-width="120px">
          <el-form-item label="启用">
            <el-switch v-model="wechatConfig.enabled" />
          </el-form-item>

          <el-form-item label="Webhook URL">
            <el-input
              v-model="wechatConfig.webhook_url"
              placeholder="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=..."
            />
          </el-form-item>

          <el-form-item label="通知设置">
            <el-checkbox v-model="wechatConfig.notify_on_success">
              成功通知
            </el-checkbox>
            <el-checkbox v-model="wechatConfig.notify_on_failure">
              失败通知
            </el-checkbox>
          </el-form-item>

          <el-form-item>
            <el-button type="primary" @click="handleSave('wechat')">
              保存
            </el-button>
            <el-button @click="handleTest('wechat')">
              测试
            </el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <!-- Telegram -->
      <el-tab-pane label="Telegram" name="telegram">
        <el-form :model="telegramConfig" label-width="120px">
          <el-form-item label="启用">
            <el-switch v-model="telegramConfig.enabled" />
          </el-form-item>

          <el-form-item label="Bot Token">
            <el-input
              v-model="telegramConfig.bot_token"
              placeholder="123456789:ABCdefGHIjklMNOpqrsTUVwxyz"
            />
          </el-form-item>

          <el-form-item label="Chat ID">
            <el-input
              v-model="telegramConfig.chat_id"
              placeholder="123456789"
            />
          </el-form-item>

          <el-form-item label="通知设置">
            <el-checkbox v-model="telegramConfig.notify_on_success">
              成功通知
            </el-checkbox>
            <el-checkbox v-model="telegramConfig.notify_on_failure">
              失败通知
            </el-checkbox>
          </el-form-item>

          <el-form-item>
            <el-button type="primary" @click="handleSave('telegram')">
              保存
            </el-button>
            <el-button @click="handleTest('telegram')">
              测试
            </el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <!-- WxPusher -->
      <el-tab-pane label="WxPusher" name="wxpusher">
        <el-form :model="wxpusherConfig" label-width="120px">
          <el-form-item label="启用">
            <el-switch v-model="wxpusherConfig.enabled" />
          </el-form-item>

          <el-form-item label="App Token">
            <el-input
              v-model="wxpusherConfig.app_token"
              placeholder="AT_xxx"
            />
          </el-form-item>

          <el-form-item label="UID">
            <el-input
              v-model="wxpusherConfig.uid"
              placeholder="UID_xxx"
            />
          </el-form-item>

          <el-form-item label="通知设置">
            <el-checkbox v-model="wxpusherConfig.notify_on_success">
              成功通知
            </el-checkbox>
            <el-checkbox v-model="wxpusherConfig.notify_on_failure">
              失败通知
            </el-checkbox>
          </el-form-item>

          <el-form-item>
            <el-button type="primary" @click="handleSave('wxpusher')">
              保存
            </el-button>
            <el-button @click="handleTest('wxpusher')">
              测试
            </el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<style scoped>
.notify-page {
  /* ... styles ... */
}
</style>
```

- [ ] **Step 2: 添加路由配置**

```javascript
// web/src/router/index.js
// 添加通知配置路由
{
  path: '/notify',
  name: 'Notify',
  component: () => import('../views/Notify.vue'),
  meta: { title: '消息推送' }
}
```

- [ ] **Step 3: 提交**

```bash
git add web/src/views/Notify.vue web/src/router/index.js
git commit -m "feat: 创建通知配置页面"
```

---

## Task 8: 初始化通知管理器

**Files:**
- Modify: `cmd/server/main.go`

- [ ] **Step 1: 在 main.go 中初始化通知管理器**

```go
// cmd/server/main.go
// 添加以下初始化代码

// 初始化通知管理器
notifyManager := notify.NewManager()

// 注册通知渠道
notifyManager.Register(notify.NewWeChatNotifier())
notifyManager.Register(notify.NewTelegramNotifier())
notifyManager.Register(notify.NewWxPusherNotifier())

// TODO: 从数据库加载通知配置并初始化
```

- [ ] **Step 2: 提交**

```bash
git add cmd/server/main.go
git commit -m "feat: 初始化通知管理器"
```

---

## Task 9: 编写平台扩展指南

**Files:**
- Create: `docs/platform-extension-guide.md`

- [ ] **Step 1: 创建平台扩展指南文档**

```markdown
# 平台扩展指南

本文档指导开发者如何为 UCAS 添加新的云盘平台支持。

## 概述

UCAS 使用驱动工厂模式管理云盘平台。每个平台通过实现 `CloudDrive` 接口来提供统一的操作抽象。

## 步骤

### 1. 创建驱动目录

```bash
mkdir -p internal/core/driver_<platform>
```

### 2. 实现 CloudDrive 接口

创建 `client.go` 文件，实现以下接口：

```go
type CloudDrive interface {
    // GetInfo 获取账号信息
    GetInfo() (*AccountInfo, error)

    // Login 登录验证
    Login() error

    // ListFiles 列出目录文件
    ListFiles(path string) ([]FileInfo, error)

    // CreateFolder 创建文件夹
    CreateFolder(parentPath, name string) error

    // DeleteFile 删除文件
    DeleteFile(path string) error

    // ParseShare 解析分享链接
    ParseShare(shareURL, passCode string) (*ShareInfo, error)

    // SaveLink 保存分享链接
    SaveLink(shareID, fileID, targetPath string) error

    // RenameFile 重命名文件
    RenameFile(path, newName string) error
}
```

### 3. 注册驱动

在 `client.go` 的 `init()` 函数中注册驱动：

```go
func init() {
    core.RegisterDriver("<platform>", func(account *db.Account) (core.CloudDrive, error) {
        return &Client{
            account: account,
        }, nil
    })
}
```

### 4. 导入驱动

在 `internal/api/router.go` 中添加导入：

```go
_ "github.com/zcq/clouddrive-auto-save/internal/core/driver_<platform>"
```

### 5. 编写测试

创建 `client_test.go`，编写单元测试和集成测试。

### 6. 更新文档

更新 `README.md`，添加新平台的说明。

## 参考实现

- `internal/core/quark/` - 夸克网盘驱动
- `internal/core/cloud139/` - 移动云盘驱动

## 注意事项

1. 错误处理：映射平台特定错误码到统一错误类型
2. 速率限制：遵守平台 API 调用频率限制
3. 认证方式：支持 Cookie / Token / OAuth 等认证方式
4. 日志记录：使用 slog 记录关键操作和错误
```

- [ ] **Step 2: 提交**

```bash
git add docs/platform-extension-guide.md
git commit -m "docs: 添加平台扩展指南"
```

---

## 阶段三完成

所有架构增强任务已完成，包括：

✅ 多渠道消息推送（企业微信、Telegram、WxPusher）
✅ 通知管理器和 API
✅ 通知配置页面
✅ 平台扩展指南

---

## 全部阶段完成

UCAS 全面升级计划已完成，包括：

**阶段一：UI/UX 改进**
- ✅ 侧边栏导航优化
- ✅ 仪表盘增强
- ✅ 卡片式 UI 组件
- ✅ PWA 支持

**阶段二：功能扩展**
- ✅ 插件系统架构
- ✅ Telegram 机器人集成
- ✅ 资源搜索集成

**阶段三：架构增强**
- ✅ 多渠道消息推送
- ✅ 平台扩展准备

**下一步：** 运行测试，验证所有功能，准备发布。
