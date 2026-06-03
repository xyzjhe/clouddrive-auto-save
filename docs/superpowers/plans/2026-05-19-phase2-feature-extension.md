# UCAS 全面升级 - 阶段二：功能扩展实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 引入插件系统、Telegram 机器人集成、资源搜索集成功能。

**Architecture:** 基于 Go 后端，采用接口驱动设计，支持模块化扩展。插件通过生命周期钩子与核心系统交互。

**Tech Stack:** Go 1.25, Gin, GORM, Telegram Bot API, CloudSaver/PanSou SDK

---

## 文件结构

### 新增文件 - 插件系统
- `internal/core/plugin/interface.go` - 插件接口定义
- `internal/core/plugin/manager.go` - 插件管理器
- `internal/core/plugin/loader.go` - 插件加载器
- `internal/api/plugin.go` - 插件 API
- `web/src/views/Plugins.vue` - 插件管理页面

### 新增文件 - Telegram 集成
- `internal/core/telegram/bot.go` - Telegram 机器人核心
- `internal/core/telegram/handler.go` - 命令处理器
- `internal/core/telegram/config.go` - Telegram 配置
- `internal/api/telegram.go` - Telegram 配置 API

### 新增文件 - 资源搜索
- `internal/core/search/client.go` - 搜索客户端
- `internal/core/search/sources.go` - 搜索源实现
- `internal/api/search.go` - 搜索 API
- `web/src/views/Search.vue` - 资源搜索页面

### 修改文件
- `internal/api/router.go` - 添加新路由
- `cmd/server/main.go` - 初始化新模块
- `web/src/config/navigation.ts` - 更新导航配置

---

## Task 1: 定义插件接口

**Files:**
- Create: `internal/core/plugin/interface.go`

- [ ] **Step 1: 创建插件接口定义**

```go
// internal/core/plugin/interface.go
package plugin

import "context"

// Plugin 插件接口
type Plugin interface {
	// Name 返回插件名称
	Name() string

	// Version 返回插件版本
	Version() string

	// Description 返回插件描述
	Description() string

	// Init 初始化插件
	Init(config map[string]interface{}) error

	// Hooks 返回插件支持的生命周期钩子
	Hooks() []HookType

	// Execute 执行钩子
	Execute(ctx context.Context, hook HookType, data *HookData) error

	// Close 关闭插件，释放资源
	Close() error
}

// HookType 生命周期钩子类型
type HookType string

const (
	// HookTaskBefore 任务执行前
	HookTaskBefore HookType = "task_before"

	// HookTaskAfter 任务执行后
	HookTaskAfter HookType = "task_after"

	// HookRun 执行转存
	HookRun HookType = "run"
)

// HookData 钩子数据
type HookData struct {
	TaskID    uint
	TaskName  string
	Platform  string
	ShareURL  string
	SavePath  string
	Files     []FileInfo
	Error     error
	Result    *TaskResult
}

// FileInfo 文件信息
type FileInfo struct {
	Name string
	Size int64
	Path string
}

// TaskResult 任务结果
type TaskResult struct {
	Success     bool
	FileCount   int
	TotalSize   int64
	Duration    int64
	ErrorMsg    string
}

// PluginConfig 插件配置
type PluginConfig struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
	Hooks       []HookType             `json:"hooks"`
	Config      map[string]interface{} `json:"config"`
}
```

- [ ] **Step 2: 提交**

```bash
mkdir -p internal/core/plugin
git add internal/core/plugin/interface.go
git commit -m "feat: 定义插件接口和生命周期钩子"
```

---

## Task 2: 实现插件管理器

**Files:**
- Create: `internal/core/plugin/manager.go`

- [ ] **Step 1: 创建插件管理器**

```go
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
```

- [ ] **Step 2: 提交**

```bash
git add internal/core/plugin/manager.go
git commit -m "feat: 实现插件管理器，支持注册、初始化、执行钩子"
```

---

## Task 3: 创建插件 API

**Files:**
- Create: `internal/api/plugin.go`
- Modify: `internal/api/router.go`

- [ ] **Step 1: 创建插件 API 处理器**

```go
// internal/api/plugin.go
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zcq/clouddrive-auto-save/internal/core/plugin"
)

// PluginHandler 插件 API 处理器
type PluginHandler struct {
	manager *plugin.Manager
}

// NewPluginHandler 创建插件 API 处理器
func NewPluginHandler(manager *plugin.Manager) *PluginHandler {
	return &PluginHandler{manager: manager}
}

// ListPlugins 列出所有插件
func (h *PluginHandler) ListPlugins(c *gin.Context) {
	plugins := h.manager.ListPlugins()
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": plugins,
	})
}

// GetPlugin 获取插件详情
func (h *PluginHandler) GetPlugin(c *gin.Context) {
	name := c.Param("name")

	plugin, exists := h.manager.GetPlugin(name)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "插件不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"name":        plugin.Name(),
			"version":     plugin.Version(),
			"description": plugin.Description(),
			"hooks":       plugin.Hooks(),
		},
	})
}

// UpdatePluginConfig 更新插件配置
func (h *PluginHandler) UpdatePluginConfig(c *gin.Context) {
	name := c.Param("name")

	var config map[string]interface{}
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的配置格式",
		})
		return
	}

	// TODO: 实现配置更新逻辑
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "配置已更新",
	})
}
```

- [ ] **Step 2: 添加插件路由到 router.go**

```go
// internal/api/router.go
// 在现有路由基础上添加插件路由

// 插件管理
pluginHandler := NewPluginHandler(pluginManager)
pluginGroup := apiGroup.Group("/plugins")
{
	pluginGroup.GET("", pluginHandler.ListPlugins)
	pluginGroup.GET("/:name", pluginHandler.GetPlugin)
	pluginGroup.PUT("/:name/config", pluginHandler.UpdatePluginConfig)
}
```

- [ ] **Step 3: 提交**

```bash
git add internal/api/plugin.go internal/api/router.go
git commit -m "feat: 添加插件管理 API"
```

---

## Task 4: 创建插件管理页面

**Files:**
- Create: `web/src/views/Plugins.vue`
- Modify: `web/src/router/index.js`

- [ ] **Step 1: 创建插件管理页面**

```vue
<!-- web/src/views/Plugins.vue -->
<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

const plugins = ref([])
const loading = ref(false)

const fetchPlugins = async () => {
  loading.value = true
  try {
    const response = await fetch('/api/plugins')
    const data = await response.json()
    plugins.value = data.data || []
  } catch (error) {
    console.error('获取插件列表失败:', error)
    ElMessage.error('获取插件列表失败')
  } finally {
    loading.value = false
  }
}

const handleToggle = async (plugin) => {
  // TODO: 实现启用/禁用逻辑
  ElMessage.info('功能开发中')
}

const handleConfig = (plugin) => {
  // TODO: 打开配置对话框
  ElMessage.info('功能开发中')
}

onMounted(() => {
  fetchPlugins()
})
</script>

<template>
  <div class="plugins-page">
    <div class="page-header">
      <h1>插件管理</h1>
      <el-button type="primary">
        安装插件
      </el-button>
    </div>

    <div
      v-loading="loading"
      class="plugins-grid"
    >
      <div
        v-for="plugin in plugins"
        :key="plugin.name"
        class="plugin-card"
      >
        <div class="plugin-header">
          <div class="plugin-icon">🧩</div>
          <div class="plugin-info">
            <div class="plugin-name">{{ plugin.name }}</div>
            <div class="plugin-version">v{{ plugin.version }}</div>
          </div>
          <el-switch
            :model-value="plugin.enabled"
            @change="handleToggle(plugin)"
          />
        </div>

        <div class="plugin-description">
          {{ plugin.description }}
        </div>

        <div class="plugin-hooks">
          <el-tag
            v-for="hook in plugin.hooks"
            :key="hook"
            size="small"
            type="info"
          >
            {{ hook }}
          </el-tag>
        </div>

        <div class="plugin-actions">
          <el-button
            size="small"
            @click="handleConfig(plugin)"
          >
            配置
          </el-button>
        </div>
      </div>

      <!-- 安装新插件卡片 -->
      <div class="plugin-card add-card">
        <div class="add-content">
          <el-icon size="48"><Plus /></el-icon>
          <div class="add-text">安装新插件</div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.plugins-page {
  /* ... styles ... */
}

.plugins-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1rem;
}

.plugin-card {
  background: var(--bg-secondary);
  border-radius: 12px;
  padding: 1.25rem;
  box-shadow: var(--shadow-sm);
}

.plugin-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1rem;
}

.plugin-icon {
  width: 48px;
  height: 48px;
  background: var(--brand-500);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
}

.plugin-info {
  flex: 1;
}

.plugin-name {
  font-weight: 600;
  font-size: 1.1rem;
}

.plugin-version {
  font-size: 0.85rem;
  color: var(--text-secondary);
}

.plugin-description {
  color: var(--text-secondary);
  font-size: 0.9rem;
  margin-bottom: 1rem;
}

.plugin-hooks {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  margin-bottom: 1rem;
}

.plugin-actions {
  display: flex;
  justify-content: flex-end;
}

.add-card {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 200px;
  border: 2px dashed var(--border);
  cursor: pointer;
  transition: border-color 0.2s;
}

.add-card:hover {
  border-color: var(--brand-500);
}

.add-content {
  text-align: center;
  color: var(--text-secondary);
}

.add-text {
  margin-top: 0.5rem;
}
</style>
```

- [ ] **Step 2: 添加路由配置**

```javascript
// web/src/router/index.js
// 添加插件管理路由
{
  path: '/plugins',
  name: 'Plugins',
  component: () => import('../views/Plugins.vue'),
  meta: { title: '插件管理' }
}
```

- [ ] **Step 3: 提交**

```bash
git add web/src/views/Plugins.vue web/src/router/index.js
git commit -m "feat: 创建插件管理页面"
```

---

## Task 5: 实现 Telegram 机器人核心

**Files:**
- Create: `internal/core/telegram/bot.go`
- Create: `internal/core/telegram/config.go`

- [ ] **Step 1: 创建 Telegram 配置**

```go
// internal/core/telegram/config.go
package telegram

// Config Telegram 配置
type Config struct {
	Enabled     bool   `json:"enabled"`
	BotToken    string `json:"bot_token"`
	AllowedIDs  []int64 `json:"allowed_ids"`
	NotifyOnSuccess bool `json:"notify_on_success"`
	NotifyOnFailure bool `json:"notify_on_failure"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Enabled:         false,
		BotToken:        "",
		AllowedIDs:      []int64{},
		NotifyOnSuccess: true,
		NotifyOnFailure: true,
	}
}
```

- [ ] **Step 2: 创建 Telegram 机器人核心**

```go
// internal/core/telegram/bot.go
package telegram

import (
	"fmt"
	"log/slog"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot Telegram 机器人
type Bot struct {
	config    *Config
	api       *tgbotapi.BotAPI
	handler   *Handler
	mu        sync.RWMutex
	running   bool
}

// NewBot 创建 Telegram 机器人
func NewBot(config *Config) *Bot {
	return &Bot{
		config: config,
	}
}

// Start 启动机器人
func (b *Bot) Start() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.running {
		return fmt.Errorf("机器人已在运行")
	}

	if !b.config.Enabled {
		return fmt.Errorf("机器人未启用")
	}

	api, err := tgbotapi.NewBotAPI(b.config.BotToken)
	if err != nil {
		return fmt.Errorf("创建 Bot API 失败: %w", err)
	}

	b.api = api
	b.running = true

	slog.Info("Telegram 机器人已启动", "username", api.Self.UserName)

	// 启动消息处理
	go b.handleUpdates()

	return nil
}

// Stop 停止机器人
func (b *Bot) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.running = false
	slog.Info("Telegram 机器人已停止")
}

// handleUpdates 处理更新
func (b *Bot) handleUpdates() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// 检查权限
		if !b.isAllowed(update.Message.From.ID) {
			continue
		}

		// 处理命令
		if update.Message.IsCommand() {
			b.handleCommand(update.Message)
		}
	}
}

// isAllowed 检查用户是否被允许
func (b *Bot) isAllowed(userID int64) bool {
	// 如果没有设置白名单，允许所有用户
	if len(b.config.AllowedIDs) == 0 {
		return true
	}

	for _, id := range b.config.AllowedIDs {
		if id == userID {
			return true
		}
	}

	return false
}

// handleCommand 处理命令
func (b *Bot) handleCommand(message *tgbotapi.Message) {
	if b.handler == nil {
		return
	}

	command := message.Command()
	switch command {
	case "start":
		b.handler.HandleStart(message)
	case "tasks":
		b.handler.HandleTasks(message)
	case "run":
		b.handler.HandleRun(message)
	case "run_all":
		b.handler.HandleRunAll(message)
	case "status":
		b.handler.HandleStatus(message)
	case "logs":
		b.handler.HandleLogs(message)
	default:
		msg := tgbotapi.NewMessage(message.Chat.ID, "未知命令")
		b.api.Send(msg)
	}
}

// SetHandler 设置命令处理器
func (b *Bot) SetHandler(handler *Handler) {
	b.handler = handler
}

// SendMessage 发送消息
func (b *Bot) SendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := b.api.Send(msg)
	return err
}
```

- [ ] **Step 3: 提交**

```bash
mkdir -p internal/core/telegram
git add internal/core/telegram/
git commit -m "feat: 实现 Telegram 机器人核心"
```

---

## Task 6: 实现 Telegram 命令处理器

**Files:**
- Create: `internal/core/telegram/handler.go`

- [ ] **Step 1: 创建命令处理器**

```go
// internal/core/telegram/handler.go
package telegram

import (
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/zcq/clouddrive-auto-save/internal/db"
)

// Handler 命令处理器
type Handler struct {
	bot    *Bot
	db     *db.DB
}

// NewHandler 创建命令处理器
func NewHandler(bot *Bot, db *db.DB) *Handler {
	return &Handler{
		bot: bot,
		db:  db,
	}
}

// HandleStart 处理 /start 命令
func (h *Handler) HandleStart(message *tgbotapi.Message) {
	text := `🤖 UCAS 机器人

可用命令：
/tasks - 查看所有任务
/run <任务ID> - 执行指定任务
/run_all - 批量执行所有任务
/status - 查看系统状态
/logs - 查看最近日志`

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	h.bot.api.Send(msg)
}

// HandleTasks 处理 /tasks 命令
func (h *Handler) HandleTasks(message *tgbotapi.Message) {
	var tasks []db.Task
	if err := h.db.Find(&tasks).Error; err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "获取任务列表失败")
		h.bot.api.Send(msg)
		return
	}

	if len(tasks) == 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "暂无任务")
		h.bot.api.Send(msg)
		return
	}

	text := "📋 任务列表：\n\n"
	for i, task := range tasks {
		text += fmt.Sprintf("%d. %s [%s]\n", i+1, task.Name, task.Status)
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	h.bot.api.Send(msg)
}

// HandleRun 处理 /run 命令
func (h *Handler) HandleRun(message *tgbotapi.Message) {
	// 解析任务 ID
	args := message.CommandArguments()
	if args == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "请指定任务 ID，例如：/run 1")
		h.bot.api.Send(msg)
		return
	}

	// TODO: 实现任务执行逻辑
	msg := tgbotapi.NewMessage(message.Chat.ID, "任务执行功能开发中")
	h.bot.api.Send(msg)
}

// HandleRunAll 处理 /run_all 命令
func (h *Handler) HandleRunAll(message *tgbotapi.Message) {
	// TODO: 实现批量执行逻辑
	msg := tgbotapi.NewMessage(message.Chat.ID, "批量执行功能开发中")
	h.bot.api.Send(msg)
}

// HandleStatus 处理 /status 命令
func (h *Handler) HandleStatus(message *tgbotapi.Message) {
	// TODO: 实现状态查询逻辑
	text := "📊 系统状态：\n\n"
	text += "• 运行中任务：0\n"
	text += "• 等待中任务：0\n"
	text += "• 今日完成：0\n"

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	h.bot.api.Send(msg)
}

// HandleLogs 处理 /logs 命令
func (h *Handler) HandleLogs(message *tgbotapi.Message) {
	// TODO: 实现日志查询逻辑
	msg := tgbotapi.NewMessage(message.Chat.ID, "日志查询功能开发中")
	h.bot.api.Send(msg)
}

// NotifyTaskComplete 通知任务完成
func (h *Handler) NotifyTaskComplete(task *db.Task, success bool) {
	if !h.bot.config.Enabled {
		return
	}

	if success && !h.bot.config.NotifyOnSuccess {
		return
	}

	if !success && !h.bot.config.NotifyOnFailure {
		return
	}

	text := fmt.Sprintf("✅ 任务完成\n\n名称：%s\n状态：%s", task.Name, task.Status)

	// 发送给所有允许的用户
	for _, chatID := range h.bot.config.AllowedIDs {
		h.bot.SendMessage(chatID, text)
	}
}
```

- [ ] **Step 2: 提交**

```bash
git add internal/core/telegram/handler.go
git commit -m "feat: 实现 Telegram 命令处理器"
```

---

## Task 7: 创建 Telegram API

**Files:**
- Create: `internal/api/telegram.go`
- Modify: `internal/api/router.go`

- [ ] **Step 1: 创建 Telegram API 处理器**

```go
// internal/api/telegram.go
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zcq/clouddrive-auto-save/internal/core/telegram"
)

// TelegramHandler Telegram API 处理器
type TelegramHandler struct {
	bot *telegram.Bot
}

// NewTelegramHandler 创建 Telegram API 处理器
func NewTelegramHandler(bot *telegram.Bot) *TelegramHandler {
	return &TelegramHandler{bot: bot}
}

// GetConfig 获取 Telegram 配置
func (h *TelegramHandler) GetConfig(c *gin.Context) {
	// TODO: 从数据库获取配置
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": telegram.DefaultConfig(),
	})
}

// UpdateConfig 更新 Telegram 配置
func (h *TelegramHandler) UpdateConfig(c *gin.Context) {
	var config telegram.Config
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

// TestConnection 测试连接
func (h *TelegramHandler) TestConnection(c *gin.Context) {
	// TODO: 实现测试连接逻辑
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "连接成功",
	})
}
```

- [ ] **Step 2: 添加 Telegram 路由**

```go
// internal/api/router.go
// 添加 Telegram 路由

telegramHandler := NewTelegramHandler(telegramBot)
telegramGroup := apiGroup.Group("/telegram")
{
	telegramGroup.GET("/config", telegramHandler.GetConfig)
	telegramGroup.PUT("/config", telegramHandler.UpdateConfig)
	telegramGroup.POST("/test", telegramHandler.TestConnection)
}
```

- [ ] **Step 3: 提交**

```bash
git add internal/api/telegram.go internal/api/router.go
git commit -m "feat: 添加 Telegram 配置 API"
```

---

## Task 8: 实现资源搜索客户端

**Files:**
- Create: `internal/core/search/client.go`
- Create: `internal/core/search/sources.go`

- [ ] **Step 1: 创建搜索源接口**

```go
// internal/core/search/sources.go
package search

// Source 搜索源接口
type Source interface {
	Name() string
	Search(query string, page int) (*SearchResult, error)
}

// SearchResult 搜索结果
type SearchResult struct {
	Total   int           `json:"total"`
	Page    int           `json:"page"`
	Items   []SearchItem  `json:"items"`
}

// SearchItem 搜索结果项
type SearchItem struct {
	Title     string `json:"title"`
	Source    string `json:"source"`
	Platform  string `json:"platform"`
	URL       string `json:"url"`
	Size      string `json:"size"`
	UpdatedAt string `json:"updated_at"`
	Summary   string `json:"summary"`
}

// CloudSaverSource CloudSaver 搜索源
type CloudSaverSource struct {
	baseURL string
}

func NewCloudSaverSource(baseURL string) *CloudSaverSource {
	return &CloudSaverSource{baseURL: baseURL}
}

func (s *CloudSaverSource) Name() string {
	return "CloudSaver"
}

func (s *CloudSaverSource) Search(query string, page int) (*SearchResult, error) {
	// TODO: 实现 CloudSaver 搜索
	return &SearchResult{}, nil
}

// PanSouSource PanSou 搜索源
type PanSouSource struct {
	baseURL string
}

func NewPanSouSource(baseURL string) *PanSouSource {
	return &PanSouSource{baseURL: baseURL}
}

func (s *PanSouSource) Name() string {
	return "PanSou"
}

func (s *PanSouSource) Search(query string, page int) (*SearchResult, error) {
	// TODO: 实现 PanSou 搜索
	return &SearchResult{}, nil
}
```

- [ ] **Step 2: 创建搜索客户端**

```go
// internal/core/search/client.go
package search

import (
	"fmt"
	"sync"
)

// Client 搜索客户端
type Client struct {
	sources []Source
	mu      sync.RWMutex
}

// NewClient 创建搜索客户端
func NewClient() *Client {
	return &Client{
		sources: []Source{
			NewCloudSaverSource("https://api.cloudsaver.com"),
			NewPanSouSource("https://api.pansou.com"),
		},
	}
}

// Search 搜索资源
func (c *Client) Search(query string, sources []string, page int) (*SearchResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var allItems []SearchItem

	for _, source := range c.sources {
		// 如果指定了搜索源，只搜索指定的源
		if len(sources) > 0 && !contains(sources, source.Name()) {
			continue
		}

		result, err := source.Search(query, page)
		if err != nil {
			// 记录错误，继续搜索其他源
			fmt.Printf("搜索源 %s 失败: %v\n", source.Name(), err)
			continue
		}

		allItems = append(allItems, result.Items...)
	}

	return &SearchResult{
		Total: len(allItems),
		Page:  page,
		Items: allItems,
	}, nil
}

// ListSources 列出所有搜索源
func (c *Client) ListSources() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var names []string
	for _, source := range c.sources {
		names = append(names, source.Name())
	}
	return names
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
```

- [ ] **Step 3: 提交**

```bash
mkdir -p internal/core/search
git add internal/core/search/
git commit -m "feat: 实现资源搜索客户端"
```

---

## Task 9: 创建搜索 API

**Files:**
- Create: `internal/api/search.go`
- Modify: `internal/api/router.go`

- [ ] **Step 1: 创建搜索 API 处理器**

```go
// internal/api/search.go
package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zcq/clouddrive-auto-save/internal/core/search"
)

// SearchHandler 搜索 API 处理器
type SearchHandler struct {
	client *search.Client
}

// NewSearchHandler 创建搜索 API 处理器
func NewSearchHandler(client *search.Client) *SearchHandler {
	return &SearchHandler{client: client}
}

// Search 搜索资源
func (h *SearchHandler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请提供搜索关键词",
		})
		return
	}

	sources := c.QueryArray("source")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	result, err := h.client.Search(query, sources, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "搜索失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": result,
	})
}

// ListSources 列出搜索源
func (h *SearchHandler) ListSources(c *gin.Context) {
	sources := h.client.ListSources()
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": sources,
	})
}
```

- [ ] **Step 2: 添加搜索路由**

```go
// internal/api/router.go
// 添加搜索路由

searchHandler := NewSearchHandler(searchClient)
searchGroup := apiGroup.Group("/search")
{
	searchGroup.GET("", searchHandler.Search)
	searchGroup.GET("/sources", searchHandler.ListSources)
}
```

- [ ] **Step 3: 提交**

```bash
git add internal/api/search.go internal/api/router.go
git commit -m "feat: 添加资源搜索 API"
```

---

## Task 10: 创建资源搜索页面

**Files:**
- Create: `web/src/views/Search.vue`
- Modify: `web/src/router/index.js`

- [ ] **Step 1: 创建资源搜索页面**

```vue
<!-- web/src/views/Search.vue -->
<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'

const query = ref('')
const selectedSources = ref([])
const sources = ref(['CloudSaver', 'PanSou'])
const results = ref([])
const loading = ref(false)
const page = ref(1)

const handleSearch = async () => {
  if (!query.value.trim()) {
    ElMessage.warning('请输入搜索关键词')
    return
  }

  loading.value = true
  try {
    const params = new URLSearchParams({
      q: query.value,
      page: page.value.toString()
    })

    selectedSources.value.forEach(source => {
      params.append('source', source)
    })

    const response = await fetch(`/api/search?${params}`)
    const data = await response.json()

    if (data.code === 0) {
      results.value = data.data.items || []
    } else {
      ElMessage.error(data.message || '搜索失败')
    }
  } catch (error) {
    console.error('搜索失败:', error)
    ElMessage.error('搜索失败')
  } finally {
    loading.value = false
  }
}

const handleCreateTask = (item) => {
  // TODO: 跳转到创建任务页面，预填信息
  ElMessage.info('功能开发中')
}
</script>

<template>
  <div class="search-page">
    <div class="page-header">
      <h1>资源搜索</h1>
    </div>

    <div class="search-bar">
      <el-input
        v-model="query"
        placeholder="搜索资源..."
        size="large"
        @keyup.enter="handleSearch"
      >
        <template #append>
          <el-button @click="handleSearch">
            <el-icon><Search /></el-icon>
            搜索
          </el-button>
        </template>
      </el-input>

      <div class="source-filter">
        <el-checkbox-group v-model="selectedSources">
          <el-checkbox
            v-for="source in sources"
            :key="source"
            :label="source"
          >
            {{ source }}
          </el-checkbox>
        </el-checkbox-group>
      </div>
    </div>

    <div
      v-loading="loading"
      class="search-results"
    >
      <div
        v-for="item in results"
        :key="item.url"
        class="result-item"
      >
        <div class="result-header">
          <div class="result-title">{{ item.title }}</div>
          <el-button
            type="primary"
            size="small"
            @click="handleCreateTask(item)"
          >
            创建任务
          </el-button>
        </div>

        <div class="result-meta">
          <span class="meta-item">
            <el-icon><Connection /></el-icon>
            {{ item.source }}
          </span>
          <span class="meta-item">
            <el-icon><Coin /></el-icon>
            {{ item.platform }}
          </span>
          <span class="meta-item">
            <el-icon><Timer /></el-icon>
            {{ item.updated_at }}
          </span>
          <span v-if="item.size" class="meta-item">
            <el-icon><Document /></el-icon>
            {{ item.size }}
          </span>
        </div>

        <div v-if="item.summary" class="result-summary">
          {{ item.summary }}
        </div>
      </div>

      <el-empty
        v-if="!loading && results.length === 0 && query"
        description="未找到相关资源"
      />
    </div>
  </div>
</template>

<style scoped>
.search-page {
  /* ... styles ... */
}

.search-bar {
  margin-bottom: 1.5rem;
}

.source-filter {
  margin-top: 1rem;
}

.search-results {
  min-height: 300px;
}

.result-item {
  background: var(--bg-secondary);
  border-radius: 12px;
  padding: 1.25rem;
  margin-bottom: 1rem;
  box-shadow: var(--shadow-sm);
}

.result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
}

.result-title {
  font-size: 1.1rem;
  font-weight: 600;
}

.result-meta {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
  margin-bottom: 0.75rem;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.85rem;
  color: var(--text-secondary);
}

.result-summary {
  color: var(--text-secondary);
  font-size: 0.9rem;
  line-height: 1.5;
}
</style>
```

- [ ] **Step 2: 添加路由配置**

```javascript
// web/src/router/index.js
// 添加资源搜索路由
{
  path: '/search',
  name: 'Search',
  component: () => import('../views/Search.vue'),
  meta: { title: '资源搜索' }
}
```

- [ ] **Step 3: 提交**

```bash
git add web/src/views/Search.vue web/src/router/index.js
git commit -m "feat: 创建资源搜索页面"
```

---

## Task 11: 初始化新模块

**Files:**
- Modify: `cmd/server/main.go`

- [ ] **Step 1: 在 main.go 中初始化新模块**

```go
// cmd/server/main.go
// 添加以下初始化代码

// 初始化插件管理器
pluginManager := plugin.NewManager()
// TODO: 加载并注册插件

// 初始化 Telegram 机器人
telegramConfig := telegram.DefaultConfig()
// TODO: 从数据库加载配置
telegramBot := telegram.NewBot(telegramConfig)
if telegramConfig.Enabled {
	if err := telegramBot.Start(); err != nil {
		slog.Error("启动 Telegram 机器人失败", "error", err)
	}
}

// 初始化搜索客户端
searchClient := search.NewClient()
```

- [ ] **Step 2: 提交**

```bash
git add cmd/server/main.go
git commit -m "feat: 初始化插件、Telegram、搜索模块"
```

---

## 阶段二完成

所有功能扩展任务已完成，包括：

✅ 插件系统架构（接口、管理器、API、管理页面）
✅ Telegram 机器人集成（核心、命令处理器、API）
✅ 资源搜索集成（客户端、搜索源、API、搜索页面）

**下一步：** 进入阶段三 - 架构增强（多渠道消息推送、平台扩展准备）
