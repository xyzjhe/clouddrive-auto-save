# CloudSaver & PanSou 资源搜索对接实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 完整实现 CloudSaver 和 PanSou 搜索引擎对接，支持认证、配置管理、结果去重排序

**Architecture:** 重构现有 `internal/core/search/` 骨架，拆分为独立源文件，通过 Setting 表持久化配置，前端新增搜索源配置 Tab

**Tech Stack:** Go (net/http, sync, encoding/json), GORM (Setting 表), Vue 3 + Element Plus

---

## 文件结构

| 文件 | 操作 | 职责 |
|------|------|------|
| `internal/core/search/sources.go` | 修改 | Source 接口 + SearchItem/SearchResult 类型定义 |
| `internal/core/search/config.go` | 新建 | 配置结构体 + LoadConfig/SaveConfig |
| `internal/core/search/config_test.go` | 新建 | 配置加载/保存测试 |
| `internal/core/search/cloudsaver.go` | 新建 | CloudSaver 认证 + 搜索 + 结果清洗 |
| `internal/core/search/cloudsaver_test.go` | 新建 | CloudSaver Mock 测试 |
| `internal/core/search/pansou.go` | 新建 | PanSou 搜索 + 结果格式化 |
| `internal/core/search/pansou_test.go` | 新建 | PanSou Mock 测试 |
| `internal/core/search/client.go` | 重构 | 配置化源创建 + 去重 + 排序 |
| `internal/core/search/client_test.go` | 重写 | 完整测试覆盖 |
| `internal/api/search.go` | 修改 | 新增 GetConfig/UpdateConfig 端点 |
| `internal/api/router.go` | 修改 | 注册新路由 |
| `cmd/server/main.go` | 修改 | 搜索客户端初始化方式变更 |
| `web/src/api/search.js` | 新建 | 搜索相关 API 封装 |
| `web/src/views/Search.vue` | 修改 | 适配新字段，增加标签显示 |
| `web/src/views/Settings.vue` | 修改 | 新增搜索源配置 Tab |

---

### Task 1: 更新类型定义 (sources.go)

**Files:**
- Modify: `internal/core/search/sources.go`

- [ ] **Step 1: 更新 SearchItem 和 SearchResult 结构体**

```go
// internal/core/search/sources.go
package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Source 搜索源接口
type Source interface {
	Name() string
	Search(query string, page int) (*SearchResult, error)
}

// SearchResult 搜索结果
type SearchResult struct {
	Total      int          `json:"total"`
	Page       int          `json:"page"`
	Items      []SearchItem `json:"items"`
	NextCursor string       `json:"next_cursor,omitempty"`
}

// SearchItem 搜索结果项
type SearchItem struct {
	Title     string   `json:"title" binding:"required"`
	Source    string   `json:"source" binding:"required"`
	Platform  string   `json:"platform" binding:"required"`
	URL       string   `json:"url" binding:"required"`
	Summary   string   `json:"summary"`
	UpdatedAt string   `json:"updated_at"`
	Tags      []string `json:"tags,omitempty"`
	Channel   string   `json:"channel,omitempty"`
}

// cstLocation 中国标准时间 UTC+8
var cstLocation = func() *time.Location {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	if loc == nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	return loc
}()

// toCST 将 ISO 时间字符串转换为 CST 格式 YYYY-MM-DD HH:MM:SS
func toCST(isoTime string) string {
	if isoTime == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, isoTime)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05-07:00", isoTime)
		if err != nil {
			return isoTime
		}
	}
	cst := t.In(cstLocation)
	if cst.Year() < 1970 {
		return ""
	}
	return cst.Format("2006-01-02 15:04:05")
}
```

- [ ] **Step 2: 删除 sources.go 中的旧 CloudSaverSource 和 PanSouSource 实现**

删除 `CloudSaverSource`、`NewCloudSaverSource`、`PanSouSource`、`NewPanSouSource` 及其方法，只保留接口定义和类型。

- [ ] **Step 3: 运行编译检查**

```bash
cd /home/zcq/Github/clouddrive-auto-save && go build ./internal/core/search/
```

Expected: 编译失败（因为 client.go 引用了已删除的类型，这是预期的）

- [ ] **Step 4: 提交**

```bash
git add internal/core/search/sources.go
git commit -m "refactor(search): 更新 SearchItem 类型定义，新增 Tags/Channel 字段"
```

---

### Task 2: 配置管理 (config.go)

**Files:**
- Create: `internal/core/search/config.go`
- Create: `internal/core/search/config_test.go`

- [ ] **Step 1: 编写配置测试**

```go
// internal/core/search/config_test.go
package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	err = db.AutoMigrate(&Setting{})
	require.NoError(t, err)
	return db
}

type Setting struct {
	Key   string `gorm:"primaryKey"`
	Value string `gorm:"type:text"`
}

func (Setting) TableName() string { return "settings" }

func TestLoadConfig_Empty(t *testing.T) {
	db := setupTestDB(t)
	config, err := LoadConfig(db)
	require.NoError(t, err)
	assert.Equal(t, "", config.CloudSaver.Server)
	assert.Equal(t, "", config.PanSou.Server)
}

func TestSaveAndLoadConfig(t *testing.T) {
	db := setupTestDB(t)
	config := &SearchConfig{
		CloudSaver: CloudSaverConfig{
			Server:   "http://localhost:8080",
			Username: "admin",
			Password: "pass123",
		},
		PanSou: PanSouConfig{
			Server: "https://so.252035.xyz",
		},
	}
	err := SaveConfig(db, config)
	require.NoError(t, err)

	loaded, err := LoadConfig(db)
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8080", loaded.CloudSaver.Server)
	assert.Equal(t, "admin", loaded.CloudSaver.Username)
	assert.Equal(t, "pass123", loaded.CloudSaver.Password)
	assert.Equal(t, "https://so.252035.xyz", loaded.PanSou.Server)
}

func TestSaveConfig_TokenUpdate(t *testing.T) {
	db := setupTestDB(t)
	config := &SearchConfig{
		CloudSaver: CloudSaverConfig{
			Server: "http://localhost:8080",
			Token:  "old-token",
		},
	}
	err := SaveConfig(db, config)
	require.NoError(t, err)

	// 更新 token
	config.CloudSaver.Token = "new-token"
	err = SaveConfig(db, config)
	require.NoError(t, err)

	loaded, err := LoadConfig(db)
	require.NoError(t, err)
	assert.Equal(t, "new-token", loaded.CloudSaver.Token)
}
```

- [ ] **Step 2: 运行测试确认失败**

```bash
cd /home/zcq/Github/clouddrive-auto-save && go test ./internal/core/search/ -run TestLoad -v
```

Expected: FAIL — `LoadConfig` 未定义

- [ ] **Step 3: 实现配置管理**

```go
// internal/core/search/config.go
package search

import (
	"fmt"

	"gorm.io/gorm"
)

// SearchConfig 搜索源配置
type SearchConfig struct {
	CloudSaver CloudSaverConfig `json:"cloudsaver"`
	PanSou     PanSouConfig     `json:"pansou"`
}

// CloudSaverConfig CloudSaver 配置
type CloudSaverConfig struct {
	Server   string `json:"server"`
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

// PanSouConfig PanSou 配置
type PanSouConfig struct {
	Server string `json:"server"`
}

// configKeys 配置项在 Setting 表中的 key 列表
var configKeys = map[string]func(*SearchConfig) *string{
	"search.cloudsaver.server":   func(c *SearchConfig) *string { return &c.CloudSaver.Server },
	"search.cloudsaver.username": func(c *SearchConfig) *string { return &c.CloudSaver.Username },
	"search.cloudsaver.password": func(c *SearchConfig) *string { return &c.CloudSaver.Password },
	"search.cloudsaver.token":    func(c *SearchConfig) *string { return &c.CloudSaver.Token },
	"search.pansou.server":       func(c *SearchConfig) *string { return &c.PanSou.Server },
}

// Setting 模型引用（与 db 包中的 Setting 表名一致）
type searchSetting struct {
	Key   string `gorm:"primaryKey"`
	Value string `gorm:"type:text"`
}

func (searchSetting) TableName() string { return "settings" }

// LoadConfig 从 Setting 表加载搜索配置
func LoadConfig(db *gorm.DB) (*SearchConfig, error) {
	config := &SearchConfig{}
	for key, ptrFunc := range configKeys {
		var setting searchSetting
		result := db.Where("key = ?", key).First(&setting)
		if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("读取配置 %s 失败: %w", key, result.Error)
		}
		if result.Error == nil {
			*ptrFunc(config) = setting.Value
		}
	}
	return config, nil
}

// SaveConfig 保存搜索配置到 Setting 表
func SaveConfig(db *gorm.DB, config *SearchConfig) error {
	for key, ptrFunc := range configKeys {
		value := *ptrFunc(config)
		setting := searchSetting{Key: key, Value: value}
		result := db.Where("key = ?", key).Assign(searchSetting{Value: value}).FirstOrCreate(&setting)
		if result.Error != nil {
			return fmt.Errorf("保存配置 %s 失败: %w", key, result.Error)
		}
	}
	return nil
}
```

- [ ] **Step 4: 运行测试确认通过**

```bash
cd /home/zcq/Github/clouddrive-auto-save && go test ./internal/core/search/ -run TestLoad -v
```

Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/core/search/config.go internal/core/search/config_test.go
git commit -m "feat(search): 实现搜索配置管理，支持从 Setting 表加载/保存"
```

---

### Task 3: CloudSaver 源实现 (cloudsaver.go)

**Files:**
- Create: `internal/core/search/cloudsaver.go`
- Create: `internal/core/search/cloudsaver_test.go`

- [ ] **Step 1: 编写 CloudSaver 测试**

```go
// internal/core/search/cloudsaver_test.go
package search

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloudSaver_Search_Success(t *testing.T) {
	// Mock CloudSaver 服务
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/user/login":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"data":    map[string]string{"token": "test-token"},
			})
		case "/api/search":
			assert.Equal(t, "黑镜", r.URL.Query().Get("keyword"))
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"data": []map[string]interface{}{
					{
						"channelId": "ch1",
						"list": []map[string]interface{}{
							{
								"title":   "名称：黑镜 第七季",
								"content": "描述：科幻美剧 链接：https://pan.quark.cn/s/xxx 标签：#科幻",
								"pubDate": "2025-01-15T10:30:00+08:00",
								"tags":    []string{"科幻", "美剧"},
								"cloudLinks": []map[string]interface{}{
									{"cloudType": "quark", "link": "https://pan.quark.cn/s/abc123"},
									{"cloudType": "alipan", "link": "https://www.alipan.com/s/xyz"},
								},
							},
						},
					},
				},
			})
		}
	}))
	defer server.Close()

	src := NewCloudSaverSource(server.URL, "admin", "pass", "")
	tokenUpdated := ""
	src.OnTokenUpdate = func(token string) { tokenUpdated = token }

	result, err := src.Search("黑镜", 1)
	require.NoError(t, err)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, "黑镜 第七季", result.Items[0].Title)
	assert.Equal(t, "https://pan.quark.cn/s/abc123", result.Items[0].URL)
	assert.Equal(t, "科幻美剧", result.Items[0].Summary)
	assert.Equal(t, "CloudSaver", result.Items[0].Source)
	assert.Equal(t, "quark", result.Items[0].Platform)
	assert.Equal(t, "test-token", tokenUpdated)
}

func TestCloudSaver_Search_TokenExpired(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/user/login":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"data":    map[string]string{"token": "new-token"},
			})
		case "/api/search":
			callCount++
			if callCount == 1 {
				// 第一次返回 token 无效
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"message": "无效的 token",
				})
			} else {
				// 第二次返回成功
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": true,
					"data":    []map[string]interface{}{},
				})
			}
		}
	}))
	defer server.Close()

	src := NewCloudSaverSource(server.URL, "admin", "pass", "expired-token")
	result, err := src.Search("test", 1)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, callCount) // 验证重试了一次
}

func TestCloudSaver_CleanResults_FilterNonQuark(t *testing.T) {
	src := &CloudSaverSource{}
	raw := []map[string]interface{}{
		{
			"channelId": "ch1",
			"list": []map[string]interface{}{
				{
					"title":   "测试资源",
					"content": "描述内容",
					"pubDate": "2025-01-15T10:30:00+08:00",
					"cloudLinks": []map[string]interface{}{
						{"cloudType": "quark", "link": "https://pan.quark.cn/s/aaa"},
						{"cloudType": "alipan", "link": "https://www.alipan.com/s/bbb"},
						{"cloudType": "quark", "link": "https://pan.quark.cn/s/aaa"}, // 重复
					},
				},
			},
		},
	}
	items := src.cleanResults(raw)
	assert.Len(t, items, 1) // 去重后只有 1 条
	assert.Equal(t, "quark", items[0].Platform)
}
```

- [ ] **Step 2: 运行测试确认失败**

```bash
cd /home/zcq/Github/clouddrive-auto-save && go test ./internal/core/search/ -run TestCloudSaver -v
```

Expected: FAIL — `NewCloudSaverSource` 未定义

- [ ] **Step 3: 实现 CloudSaver 源**

```go
// internal/core/search/cloudsaver.go
package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

// CloudSaverSource CloudSaver 搜索源
type CloudSaverSource struct {
	baseURL        string
	username       string
	password       string
	token          string
	mu             sync.RWMutex
	OnTokenUpdate  func(token string) // Token 更新回调，用于持久化
}

// NewCloudSaverSource 创建 CloudSaver 搜索源
func NewCloudSaverSource(baseURL, username, password, token string) *CloudSaverSource {
	return &CloudSaverSource{
		baseURL:  strings.TrimRight(baseURL, "/"),
		username: username,
		password: password,
		token:    token,
	}
}

func (s *CloudSaverSource) Name() string {
	return "CloudSaver"
}

// login 登录获取 Token
func (s *CloudSaverSource) login() error {
	url := fmt.Sprintf("%s/api/user/login", s.baseURL)
	body := fmt.Sprintf(`{"username":"%s","password":"%s"}`, s.username, s.password)
	req, err := http.NewRequestWithContext(context.Background(), "POST", url, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("创建登录请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("登录请求失败: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("解析登录响应失败: %w", err)
	}
	if !result.Success {
		return fmt.Errorf("登录失败: %s", result.Message)
	}

	s.mu.Lock()
	s.token = result.Data.Token
	s.mu.Unlock()

	if s.OnTokenUpdate != nil {
		s.OnTokenUpdate(result.Data.Token)
	}
	return nil
}

// Search 搜索资源
func (s *CloudSaverSource) Search(query string, page int) (*SearchResult, error) {
	result, err := s.doSearch(query, "")
	if err != nil {
		return nil, err
	}

	// Token 过期自动重试
	if result.Message == "无效的 token" || result.Message == "未提供 token" {
		if loginErr := s.login(); loginErr != nil {
			return nil, fmt.Errorf("自动登录失败: %w", loginErr)
		}
		result, err = s.doSearch(query, "")
		if err != nil {
			return nil, err
		}
	}

	if !result.Success {
		return nil, fmt.Errorf("搜索失败: %s", result.Message)
	}

	items := s.cleanResults(result.Data)
	return &SearchResult{
		Total: len(items),
		Page:  page,
		Items: items,
	}, nil
}

// csSearchResponse CloudSaver 搜索响应
type csSearchResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Data    []map[string]interface{} `json:"data"`
}

// doSearch 执行搜索请求
func (s *CloudSaverSource) doSearch(query, lastMessageID string) (*csSearchResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/api/search?keyword=%s&lastMessageId=%s", s.baseURL, query, lastMessageID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建搜索请求失败: %w", err)
	}

	s.mu.RLock()
	token := s.token
	s.mu.RUnlock()

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 UCAS/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("搜索请求失败: %w", err)
	}
	defer resp.Body.Close()

	var result csSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析搜索响应失败: %w", err)
	}
	return &result, nil
}

// cleanResults 清洗搜索结果
func (s *CloudSaverSource) cleanResults(data []map[string]interface{}) []SearchItem {
	var items []SearchItem
	seen := make(map[string]bool)

	patternTitle := regexp.MustCompile(`(?:名称|标题)[：:]?\s*(.*)`)
	patternContent := regexp.MustCompile(`(?:描述|简介)[：:]?\s*(.*?)(?:链接|标签|$)`)
	patternHTML := regexp.MustCompile(`<[^>]+>`)

	for _, channel := range data {
		channelID, _ := channel["channelId"].(string)
		list, _ := channel["list"].([]interface{})
		for _, item := range list {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			title, _ := itemMap["title"].(string)
			content, _ := itemMap["content"].(string)
			pubDate, _ := itemMap["pubDate"].(string)

			// 提取标签
			var tags []string
			if tagsRaw, ok := itemMap["tags"].([]interface{}); ok {
				for _, t := range tagsRaw {
					if s, ok := t.(string); ok {
						tags = append(tags, s)
					}
				}
			}

			// 清洗标题
			if m := patternTitle.FindStringSubmatch(title); len(m) > 1 {
				title = strings.TrimSpace(m[1])
			}

			// 清洗内容
			if m := patternContent.FindStringSubmatch(content); len(m) > 1 {
				content = strings.TrimSpace(m[1])
			}
			content = patternHTML.ReplaceAllString(content, "")
			content = strings.TrimSpace(content)

			// 遍历 cloudLinks
			cloudLinks, _ := itemMap["cloudLinks"].([]interface{})
			for _, link := range cloudLinks {
				linkMap, ok := link.(map[string]interface{})
				if !ok {
					continue
				}
				cloudType, _ := linkMap["cloudType"].(string)
				if cloudType != "quark" {
					continue
				}
				linkURL, _ := linkMap["link"].(string)
				if linkURL == "" || seen[linkURL] {
					continue
				}
				seen[linkURL] = true

				items = append(items, SearchItem{
					Title:     title,
					URL:       linkURL,
					Source:    "CloudSaver",
					Platform:  "quark",
					Summary:   content,
					UpdatedAt: toCST(pubDate),
					Tags:      tags,
					Channel:   channelID,
				})
			}
		}
	}
	return items
}
```

- [ ] **Step 4: 运行测试确认通过**

```bash
cd /home/zcq/Github/clouddrive-auto-save && go test ./internal/core/search/ -run TestCloudSaver -v
```

Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/core/search/cloudsaver.go internal/core/search/cloudsaver_test.go
git commit -m "feat(search): 实现 CloudSaver 搜索源，支持 JWT 认证和自动续期"
```

---

### Task 4: PanSou 源实现 (pansou.go)

**Files:**
- Create: `internal/core/search/pansou.go`
- Create: `internal/core/search/pansou_test.go`

- [ ] **Step 1: 编写 PanSou 测试**

```go
// internal/core/search/pansou_test.go
package search

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPanSou_Search_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "哪吒", r.URL.Query().Get("kw"))
		assert.Equal(t, "quark", r.URL.Query()["cloud_types"][0])
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"merged_by_type": map[string]interface{}{
					"quark": []map[string]interface{}{
						{
							"url":      "https://pan.quark.cn/s/abc123",
							"note":     "哪吒之魔童降世【简介】国产动画巅峰",
							"datetime": "2025-01-15T10:30:00+08:00",
							"source":   "channel1",
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	src := NewPanSouSource(server.URL)
	result, err := src.Search("哪吒", 1)
	require.NoError(t, err)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, "哪吒之魔童降世", result.Items[0].Title)
	assert.Equal(t, "国产动画巅峰", result.Items[0].Summary)
	assert.Equal(t, "https://pan.quark.cn/s/abc123", result.Items[0].URL)
	assert.Equal(t, "PanSou", result.Items[0].Source)
	assert.Equal(t, "quark", result.Items[0].Platform)
	assert.Equal(t, "channel1", result.Items[0].Channel)
}

func TestPanSou_Search_EmptyNote(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"merged_by_type": map[string]interface{}{
					"quark": []map[string]interface{}{
						{
							"url":      "https://pan.quark.cn/s/xyz",
							"note":     "简单标题没有简介",
							"datetime": "2025-02-01T12:00:00+08:00",
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	src := NewPanSouSource(server.URL)
	result, err := src.Search("test", 1)
	require.NoError(t, err)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, "简单标题没有简介", result.Items[0].Title)
	assert.Equal(t, "", result.Items[0].Summary)
}

func TestPanSou_Search_Dedup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"merged_by_type": map[string]interface{}{
					"quark": []map[string]interface{}{
						{"url": "https://pan.quark.cn/s/aaa", "note": "资源A", "datetime": "2025-01-01T00:00:00+08:00"},
						{"url": "https://pan.quark.cn/s/aaa", "note": "资源A重复", "datetime": "2025-01-02T00:00:00+08:00"},
						{"url": "https://pan.quark.cn/s/bbb", "note": "资源B", "datetime": "2025-01-03T00:00:00+08:00"},
					},
				},
			},
		})
	}))
	defer server.Close()

	src := NewPanSouSource(server.URL)
	result, err := src.Search("test", 1)
	require.NoError(t, err)
	assert.Len(t, result.Items, 2) // 去重后 2 条
}

func TestPanSou_Search_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	src := NewPanSouSource(server.URL)
	result, err := src.Search("test", 1)
	require.NoError(t, err) // PanSou 不返回错误，而是空结果
	assert.Len(t, result.Items, 0)
}
```

- [ ] **Step 2: 运行测试确认失败**

```bash
cd /home/zcq/Github/clouddrive-auto-save && go test ./internal/core/search/ -run TestPanSou -v
```

Expected: FAIL — `NewPanSouSource` 未定义

- [ ] **Step 3: 实现 PanSou 源**

```go
// internal/core/search/pansou.go
package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// PanSouSource PanSou 搜索源
type PanSouSource struct {
	baseURL string
}

// NewPanSouSource 创建 PanSou 搜索源
func NewPanSouSource(baseURL string) *PanSouSource {
	return &PanSouSource{baseURL: strings.TrimRight(baseURL, "/")}
}

func (s *PanSouSource) Name() string {
	return "PanSou"
}

// psSearchResponse PanSou 搜索响应
type psSearchResponse struct {
	Code int `json:"code"`
	Data struct {
		MergedByType struct {
			Quark []psItem `json:"quark"`
		} `json:"merged_by_type"`
	} `json:"data"`
}

type psItem struct {
	URL      string `json:"url"`
	Note     string `json:"note"`
	DateTime string `json:"datetime"`
	Source   string `json:"source"`
}

// Search 搜索资源
func (s *PanSouSource) Search(query string, page int) (*SearchResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	params := url.Values{}
	params.Set("kw", query)
	params.Set("cloud_types", `["quark"]`)
	params.Set("res", "merge")

	reqURL := fmt.Sprintf("%s/api/search?%s", s.baseURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return &SearchResult{Page: page}, nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 UCAS/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &SearchResult{Page: page}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &SearchResult{Page: page}, nil
	}

	var result psSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &SearchResult{Page: page}, nil
	}

	items := s.formatResults(result.Data.MergedByType.Quark)
	return &SearchResult{
		Total: len(items),
		Page:  page,
		Items: items,
	}, nil
}

// formatResults 格式化搜索结果
func (s *PanSouSource) formatResults(data []psItem) []SearchItem {
	// 正则：分离标题和描述
	pattern := regexp.MustCompile(`^(.*?)(?:[【\[]?(?:简介|介绍|描述)[】\]]?[:：]?)(.*)$`)

	var items []SearchItem
	seen := make(map[string]bool)

	for _, item := range data {
		if item.URL == "" || seen[item.URL] {
			continue
		}
		seen[item.URL] = true

		title := item.Note
		content := ""

		if m := pattern.FindStringSubmatch(item.Note); len(m) > 2 {
			title = strings.TrimSpace(m[1])
			content = strings.TrimSpace(m[2])
		}

		items = append(items, SearchItem{
			Title:     title,
			URL:       item.URL,
			Source:    "PanSou",
			Platform:  "quark",
			Summary:   content,
			UpdatedAt: toCST(item.DateTime),
			Channel:   item.Source,
		})
	}
	return items
}
```

- [ ] **Step 4: 运行测试确认通过**

```bash
cd /home/zcq/Github/clouddrive-auto-save && go test ./internal/core/search/ -run TestPanSou -v
```

Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/core/search/pansou.go internal/core/search/pansou_test.go
git commit -m "feat(search): 实现 PanSou 搜索源，支持结果格式化和去重"
```

---

### Task 5: Client 重构 (client.go)

**Files:**
- Modify: `internal/core/search/client.go`
- Modify: `internal/core/search/client_test.go`

- [ ] **Step 1: 重写 client.go**

```go
// internal/core/search/client.go
package search

import (
	"log/slog"
	"sort"
	"sync"

	"gorm.io/gorm"
)

// Client 搜索客户端
type Client struct {
	sources []Source
	config  *SearchConfig
	db      *gorm.DB
	mu      sync.RWMutex
}

// NewClient 创建搜索客户端
func NewClient(config *SearchConfig, db *gorm.DB) *Client {
	c := &Client{
		config: config,
		db:     db,
	}
	c.buildSources()
	return c
}

// buildSources 根据配置构建搜索源
func (c *Client) buildSources() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.sources = nil

	if c.config.CloudSaver.Server != "" {
		cs := NewCloudSaverSource(
			c.config.CloudSaver.Server,
			c.config.CloudSaver.Username,
			c.config.CloudSaver.Password,
			c.config.CloudSaver.Token,
		)
		cs.OnTokenUpdate = func(token string) {
			c.config.CloudSaver.Token = token
			if c.db != nil {
				if err := SaveConfig(c.db, c.config); err != nil {
					slog.Error("持久化 CloudSaver Token 失败", "error", err)
				}
			}
		}
		c.sources = append(c.sources, cs)
	}

	if c.config.PanSou.Server != "" {
		c.sources = append(c.sources, NewPanSouSource(c.config.PanSou.Server))
	}
}

// UpdateConfig 热更新配置
func (c *Client) UpdateConfig(config *SearchConfig) {
	c.config = config
	c.buildSources()
}

// Search 搜索资源（并发 + 去重 + 排序）
func (c *Client) Search(query string, sources []string, page int) (*SearchResult, error) {
	c.mu.RLock()
	activeSources := make([]Source, len(c.sources))
	copy(activeSources, c.sources)
	c.mu.RUnlock()

	var allItems []SearchItem
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, source := range activeSources {
		if len(sources) > 0 && !contains(sources, source.Name()) {
			continue
		}

		wg.Add(1)
		go func(src Source) {
			defer wg.Done()
			result, err := src.Search(query, page)
			if err != nil {
				slog.Error("搜索源失败", "name", src.Name(), "error", err)
				return
			}
			if result != nil && len(result.Items) > 0 {
				mu.Lock()
				allItems = append(allItems, result.Items...)
				mu.Unlock()
			}
		}(source)
	}

	wg.Wait()

	// 去重
	seen := make(map[string]bool)
	var deduped []SearchItem
	for _, item := range allItems {
		if !seen[item.URL] {
			seen[item.URL] = true
			deduped = append(deduped, item)
		}
	}

	// 按时间降序排序
	sort.Slice(deduped, func(i, j int) bool {
		return deduped[i].UpdatedAt > deduped[j].UpdatedAt
	})

	return &SearchResult{
		Total: len(deduped),
		Page:  page,
		Items: deduped,
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

- [ ] **Step 2: 重写 client_test.go**

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
	name    string
	results *SearchResult
	searchErr error
}

func NewMockSource(name string, results *SearchResult) *MockSource {
	return &MockSource{name: name, results: results}
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

func TestClient_Search_MergeAndDedup(t *testing.T) {
	config := &SearchConfig{
		PanSou: PanSouConfig{Server: "http://test"},
	}
	client := NewClient(config, nil)

	// 替换为 mock 源
	client.sources = []Source{
		NewMockSource("CloudSaver", &SearchResult{
			Items: []SearchItem{
				{Title: "A", URL: "https://pan.quark.cn/s/aaa", UpdatedAt: "2025-01-01 10:00:00", Source: "CloudSaver", Platform: "quark"},
				{Title: "B", URL: "https://pan.quark.cn/s/bbb", UpdatedAt: "2025-01-02 10:00:00", Source: "CloudSaver", Platform: "quark"},
			},
		}),
		NewMockSource("PanSou", &SearchResult{
			Items: []SearchItem{
				{Title: "B-dup", URL: "https://pan.quark.cn/s/bbb", UpdatedAt: "2025-01-03 10:00:00", Source: "PanSou", Platform: "quark"},
				{Title: "C", URL: "https://pan.quark.cn/s/ccc", UpdatedAt: "2025-01-01 08:00:00", Source: "PanSou", Platform: "quark"},
			},
		}),
	}

	result, err := client.Search("test", []string{}, 1)
	require.NoError(t, err)
	assert.Len(t, result.Items, 3) // 去重后 3 条
	// 验证按时间降序
	assert.Equal(t, "https://pan.quark.cn/s/bbb", result.Items[0].URL) // 01-02
	assert.Equal(t, "https://pan.quark.cn/s/aaa", result.Items[1].URL) // 01-01 10:00
	assert.Equal(t, "https://pan.quark.cn/s/ccc", result.Items[2].URL) // 01-01 08:00
}

func TestClient_Search_FilterSources(t *testing.T) {
	config := &SearchConfig{}
	client := NewClient(config, nil)

	client.sources = []Source{
		NewMockSource("CloudSaver", &SearchResult{
			Items: []SearchItem{{Title: "CS", URL: "https://cs.test", Source: "CloudSaver", Platform: "quark"}},
		}),
		NewMockSource("PanSou", &SearchResult{
			Items: []SearchItem{{Title: "PS", URL: "https://ps.test", Source: "PanSou", Platform: "quark"}},
		}),
	}

	result, err := client.Search("test", []string{"PanSou"}, 1)
	require.NoError(t, err)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, "PanSou", result.Items[0].Source)
}

func TestClient_ListSources(t *testing.T) {
	config := &SearchConfig{}
	client := NewClient(config, nil)

	client.sources = []Source{
		NewMockSource("CloudSaver", nil),
		NewMockSource("PanSou", nil),
	}

	sources := client.ListSources()
	assert.Len(t, sources, 2)
	assert.Contains(t, sources, "CloudSaver")
	assert.Contains(t, sources, "PanSou")
}

func TestClient_UpdateConfig(t *testing.T) {
	config := &SearchConfig{}
	client := NewClient(config, nil)
	assert.Len(t, client.sources, 0)

	client.UpdateConfig(&SearchConfig{
		PanSou: PanSouConfig{Server: "https://so.252035.xyz"},
	})
	assert.Len(t, client.sources, 1)
	assert.Equal(t, "PanSou", client.sources[0].Name())
}
```

- [ ] **Step 3: 运行测试确认通过**

```bash
cd /home/zcq/Github/clouddrive-auto-save && go test ./internal/core/search/ -v
```

Expected: ALL PASS

- [ ] **Step 4: 提交**

```bash
git add internal/core/search/client.go internal/core/search/client_test.go
git commit -m "refactor(search): 重构 Client，支持配置化源创建、结果去重和排序"
```

---

### Task 6: API 层扩展 (search.go + router.go)

**Files:**
- Modify: `internal/api/search.go`
- Modify: `internal/api/router.go`

- [ ] **Step 1: 扩展 SearchHandler**

在 `internal/api/search.go` 中新增 `GetConfig` 和 `UpdateConfig` 方法：

```go
// GetConfig 获取搜索源配置（密码脱敏）
func (h *SearchHandler) GetConfig(c *gin.Context) {
	config := h.client.GetConfig()
	// 脱敏处理
	masked := *config
	if masked.CloudSaver.Password != "" {
		masked.CloudSaver.Password = "***"
	}
	if masked.CloudSaver.Token != "" {
		masked.CloudSaver.Token = "***"
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": masked})
}

// UpdateConfig 更新搜索源配置
func (h *SearchHandler) UpdateConfig(c *gin.Context) {
	var config search.SearchConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误"})
		return
	}
	if err := h.client.SaveAndUpdateConfig(&config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存配置失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "配置已更新"})
}
```

在 `Client` 中新增辅助方法（`client.go`）：

```go
// GetConfig 获取当前配置
func (c *Client) GetConfig() *SearchConfig {
	return c.config
}

// SaveAndUpdateConfig 保存配置并热更新
func (c *Client) SaveAndUpdateConfig(config *SearchConfig) error {
	if c.db != nil {
		if err := SaveConfig(c.db, config); err != nil {
			return err
		}
	}
	c.UpdateConfig(config)
	return nil
}
```

- [ ] **Step 2: 注册新路由**

在 `internal/api/router.go` 的搜索路由组中添加：

```go
// 资源搜索
api.GET("/search", searchResources)
api.GET("/search/sources", listSearchSources)
api.GET("/search/config", getSearchConfig)
api.PUT("/search/config", updateSearchConfig)
```

新增处理函数：

```go
func getSearchConfig(c *gin.Context) {
	if searchHandler == nil {
		c.PureJSON(http.StatusServiceUnavailable, gin.H{"error": "搜索服务未初始化"})
		return
	}
	searchHandler.GetConfig(c)
}

func updateSearchConfig(c *gin.Context) {
	if searchHandler == nil {
		c.PureJSON(http.StatusServiceUnavailable, gin.H{"error": "搜索服务未初始化"})
		return
	}
	searchHandler.UpdateConfig(c)
}
```

- [ ] **Step 3: 更新 main.go 初始化**

```go
// cmd/server/main.go 中搜索初始化部分改为：
slog.Info("Initializing search client...")
searchConfig, err := search.LoadConfig(db.DB)
if err != nil {
	slog.Warn("Failed to load search config, using defaults", "error", err)
	searchConfig = &search.SearchConfig{}
}
searchClient := search.NewClient(searchConfig, db.DB)
api.InitSearchHandler(searchClient)
```

- [ ] **Step 4: 运行完整测试**

```bash
cd /home/zcq/Github/clouddrive-auto-save && go test ./internal/... -v
```

Expected: ALL PASS

- [ ] **Step 5: 提交**

```bash
git add internal/api/search.go internal/api/router.go cmd/server/main.go internal/core/search/client.go
git commit -m "feat(search): 新增搜索配置管理 API，集成配置化初始化"
```

---

### Task 7: 前端 API 封装 (search.js)

**Files:**
- Create: `web/src/api/search.js`

- [ ] **Step 1: 创建 API 封装**

```javascript
// web/src/api/search.js
import request from './request'

export function searchResources(params) {
  return request({
    url: '/search',
    method: 'get',
    params
  })
}

export function listSearchSources() {
  return request({
    url: '/search/sources',
    method: 'get'
  })
}

export function getSearchConfig() {
  return request({
    url: '/search/config',
    method: 'get'
  })
}

export function updateSearchConfig(data) {
  return request({
    url: '/search/config',
    method: 'put',
    data
  })
}
```

- [ ] **Step 2: 提交**

```bash
git add web/src/api/search.js
git commit -m "feat(search): 新增搜索 API 封装"
```

---

### Task 8: Search.vue 更新

**Files:**
- Modify: `web/src/views/Search.vue`

- [ ] **Step 1: 更新 Search.vue**

将 `fetch` 调用替换为 API 封装，新增 `tags` 和 `channel` 显示：

```vue
<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search as SearchIcon, Link as LinkIcon, Clock as ClockIcon, FileText as FileTextIcon, Tag as TagIcon } from 'lucide-vue-next'
import { searchResources } from '../api/search'

const router = useRouter()
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
    const params = {
      q: query.value,
      page: page.value.toString()
    }
    if (selectedSources.value.length > 0) {
      params.source = selectedSources.value
    }
    const data = await searchResources(params)
    if (data.code === 0) {
      results.value = data.data.items || []
    } else {
      ElMessage.error(data.message || '搜索失败')
    }
  } catch (error) {
    console.error('搜索失败:', error)
  } finally {
    loading.value = false
  }
}

const handleCreateTask = (item) => {
  router.push({
    name: 'Tasks',
    query: {
      share_url: item.url,
      title: item.title,
      platform: item.platform
    }
  })
}
</script>

<template>
  <div class="search-page">
    <div class="page-header">
      <div class="title-section">
        <h2>资源搜索</h2>
        <p>搜索云盘资源，一键创建转存任务</p>
      </div>
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
            <el-icon><SearchIcon /></el-icon>
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

    <div v-loading="loading" class="search-results">
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
            <el-icon><LinkIcon /></el-icon>
            {{ item.source }}
          </span>
          <span v-if="item.channel" class="meta-item">
            <el-icon><FileTextIcon /></el-icon>
            {{ item.channel }}
          </span>
          <span class="meta-item">
            <el-icon><ClockIcon /></el-icon>
            {{ item.updated_at }}
          </span>
        </div>

        <div v-if="item.tags && item.tags.length > 0" class="result-tags">
          <el-tag
            v-for="tag in item.tags"
            :key="tag"
            size="small"
            type="info"
            class="tag-item"
          >
            {{ tag }}
          </el-tag>
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
/* ... 保留现有样式 ... */

.result-tags {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  margin-bottom: 0.5rem;
}

.tag-item {
  margin: 0;
}
</style>
```

- [ ] **Step 2: 提交**

```bash
git add web/src/views/Search.vue
git commit -m "feat(search): Search.vue 适配新字段，增加标签和频道显示"
```

---

### Task 9: Settings.vue 新增搜索源配置 Tab

**Files:**
- Modify: `web/src/views/Settings.vue`

- [ ] **Step 1: 新增搜索源 Tab**

在 `Settings.vue` 的 `el-tabs` 中，在"插件管理"Tab 之后新增"搜索源"Tab：

```vue
<!-- Tab: 搜索源配置 -->
<el-tab-pane name="search">
  <template #label>
    <div class="tab-label-inner">
      <el-icon><Search /></el-icon>
      <span>搜索源</span>
    </div>
  </template>

  <el-row :gutter="24">
    <el-col :xs="24" :lg="12">
      <el-card class="inner-settings-card">
        <template #header>
          <div class="card-header">
            <div class="header-title">
              <span>CloudSaver 配置</span>
            </div>
          </div>
        </template>
        <el-form label-position="top">
          <el-form-item label="服务地址">
            <el-input v-model="searchConfig.cloudsaver.server" placeholder="http://localhost:8080" />
          </el-form-item>
          <el-form-item label="用户名">
            <el-input v-model="searchConfig.cloudsaver.username" placeholder="用户名" />
          </el-form-item>
          <el-form-item label="密码">
            <el-input v-model="searchConfig.cloudsaver.password" type="password" show-password placeholder="密码" />
          </el-form-item>
          <el-form-item label="Token 状态">
            <el-tag :type="searchConfig.cloudsaver.token ? 'success' : 'info'">
              {{ searchConfig.cloudsaver.token ? '已获取' : '未获取' }}
            </el-tag>
          </el-form-item>
        </el-form>
      </el-card>
    </el-col>

    <el-col :xs="24" :lg="12">
      <el-card class="inner-settings-card">
        <template #header>
          <div class="card-header">
            <div class="header-title">
              <span>PanSou 配置</span>
            </div>
          </div>
        </template>
        <el-form label-position="top">
          <el-form-item label="服务地址">
            <el-input v-model="searchConfig.pansou.server" placeholder="https://so.252035.xyz" />
          </el-form-item>
        </el-form>
      </el-card>
    </el-col>
  </el-row>

  <div style="text-align: right; margin-top: 16px;">
    <el-button type="primary" @click="saveSearchConfig" :loading="searchConfigSaving">
      保存配置
    </el-button>
  </div>
</el-tab-pane>
```

- [ ] **Step 2: 新增脚本逻辑**

在 `<script setup>` 中添加：

```javascript
import { Search } from 'lucide-vue-next'
import { getSearchConfig, updateSearchConfig } from '../api/search'

const searchConfig = ref({
  cloudsaver: { server: '', username: '', password: '', token: '' },
  pansou: { server: '' }
})
const searchConfigSaving = ref(false)

const loadSearchConfig = async () => {
  try {
    const data = await getSearchConfig()
    if (data.code === 0 && data.data) {
      searchConfig.value = data.data
    }
  } catch (e) {
    console.error('加载搜索配置失败:', e)
  }
}

const saveSearchConfig = async () => {
  searchConfigSaving.value = true
  try {
    await updateSearchConfig(searchConfig.value)
    ElMessage.success('搜索配置已保存')
  } catch (e) {
    console.error('保存搜索配置失败:', e)
  } finally {
    searchConfigSaving.value = false
  }
}
```

在 `onMounted` 中调用 `loadSearchConfig()`。

- [ ] **Step 3: 提交**

```bash
git add web/src/views/Settings.vue
git commit -m "feat(search): Settings.vue 新增搜索源配置 Tab"
```

---

### Task 10: 编译验证与最终测试

- [ ] **Step 1: 运行后端完整测试**

```bash
cd /home/zcq/Github/clouddrive-auto-save && go test ./internal/... -v
```

Expected: ALL PASS

- [ ] **Step 2: 编译检查**

```bash
cd /home/zcq/Github/clouddrive-auto-save && go build ./cmd/server/
```

Expected: 编译成功

- [ ] **Step 3: 前端类型检查**

```bash
cd /home/zcq/Github/clouddrive-auto-save/web && npx vue-tsc --noEmit 2>&1 | head -20
```

Expected: 无新增类型错误

- [ ] **Step 4: 最终提交**

```bash
git add -A
git commit -m "feat(search): 完成 CloudSaver & PanSou 搜索对接，支持认证、配置管理和结果处理"
```
