# CloudSaver & PanSou 资源搜索对接设计

**日期**: 2026-05-27
**状态**: 已批准
**参考项目**: [quark-auto-save](https://github.com/Cp0219/quark-auto-save)

## 1. 背景与目标

当前项目已有 `internal/core/search/` 模块骨架，包含 `Source` 接口、`CloudSaverSource`、`PanSouSource` 和 `Client`，但实现是占位性质的——API 端点是硬编码的占位 URL，请求格式与真实 API 不匹配，缺少认证、结果解析、去重等核心逻辑。

**目标**: 基于 quark-auto-save 的设计思路，在当前 Go 项目中完整实现 CloudSaver 和 PanSou 的对接，支持：
- CloudSaver 用户认证（JWT Token 登录 + 自动续期）
- PanSou 直接搜索（无需认证）
- 搜索结果去重 + 按时间排序
- 配置持久化（Setting 表 + Settings 页面管理）

## 2. 架构概览

```
┌─────────────────────────────────────────────────────┐
│                    Frontend (Vue 3)                   │
│  Search.vue ─── GET /api/search?q=&source=&page=     │
│  Settings.vue ─ GET/PUT /api/search/config            │
└───────────────────────┬─────────────────────────────┘
                        │
┌───────────────────────▼─────────────────────────────┐
│              internal/api/search.go                   │
│  SearchHandler: Search(), ListSources(), GetConfig()  │
└───────────────────────┬─────────────────────────────┘
                        │
┌───────────────────────▼─────────────────────────────┐
│             internal/core/search/                     │
│  config.go    ── LoadConfig() from Setting table      │
│  client.go    ── Client.Search() 并发 + 去重 + 排序   │
│  cloudsaver.go ── CloudSaverSource (认证 + 搜索)      │
│  pansou.go     ── PanSouSource (直接搜索)              │
│  sources.go    ── Source 接口 + SearchResult 定义      │
└───────────────────────┬─────────────────────────────┘
                        │
          ┌─────────────┼─────────────┐
          ▼                           ▼
   CloudSaver 服务              PanSou 服务
   (需 JWT 认证)               (无需认证)
```

## 3. 文件变更清单

| 文件 | 操作 | 说明 |
|------|------|------|
| `internal/core/search/sources.go` | 修改 | 保留接口定义，移除具体实现 |
| `internal/core/search/cloudsaver.go` | 新建 | CloudSaver 认证 + 搜索实现 |
| `internal/core/search/pansou.go` | 新建 | PanSou 搜索实现 |
| `internal/core/search/config.go` | 新建 | 配置加载与持久化 |
| `internal/core/search/client.go` | 重构 | 增加去重 + 排序，配置化源创建 |
| `internal/core/search/client_test.go` | 修改 | 补充测试用例 |
| `internal/api/search.go` | 修改 | 增加配置管理 API |
| `internal/api/router.go` | 修改 | 注册新路由 |
| `cmd/server/main.go` | 修改 | 搜索客户端初始化方式变更 |
| `web/src/views/Search.vue` | 修改 | 适配新字段，增加标签显示 |
| `web/src/views/Settings.vue` | 修改 | 新增搜索源配置 Tab |
| `web/src/api/search.js` | 新建 | 搜索相关 API 封装 |

## 4. 后端设计

### 4.1 Source 接口 (sources.go)

保留现有接口，`page` 参数语义统一为"页码"。CloudSaver 使用游标分页时，`page` 仅用于首次请求，后续翻页通过 `lastMessageId` 实现（存储在 `SearchResult.NextCursor` 中）：

```go
type Source interface {
    Name() string
    Search(query string, page int) (*SearchResult, error)
}
```

`SearchResult` 增加游标字段：
```go
type SearchResult struct {
    Total      int          `json:"total"`
    Page       int          `json:"page"`
    Items      []SearchItem `json:"items"`
    NextCursor string       `json:"next_cursor,omitempty"` // CloudSaver 游标分页用
}

`SearchResult` 和 `SearchItem` 结构体保留，新增 `Tags` 和 `Channel` 字段：

```go
type SearchItem struct {
    Title     string   `json:"title"`
    URL       string   `json:"url"`
    Source    string   `json:"source"`
    Platform  string   `json:"platform"`
    Summary   string   `json:"summary"`
    UpdatedAt string   `json:"updated_at"`
    Tags      []string `json:"tags,omitempty"`
    Channel   string   `json:"channel,omitempty"`
}
```

### 4.2 CloudSaver (cloudsaver.go)

**结构体**:
```go
type CloudSaverSource struct {
    baseURL  string
    username string
    password string
    token    string
    mu       sync.RWMutex
}
```

**认证流程**:
1. `Search()` 先用现有 Token 尝试请求
2. 若返回 `无效的 token` 或 `未提供 token`，自动调用 `login()`
3. `login()` 调用 `POST /api/user/login` 获取新 Token
4. 新 Token 存入结构体 + 通过 `OnTokenUpdate` 回调持久化到 Setting 表
5. 用新 Token 重试搜索

**Token 持久化机制**: `CloudSaverSource` 持有一个 `OnTokenUpdate func(token string)` 回调函数，由 `Client` 在创建源时注入，回调内部调用 `config.SaveConfig()` 更新 Setting 表。

**搜索接口**: `GET /api/search?keyword={query}&lastMessageId={page_cursor}`

**响应解析** (`cleanResults`):
- 遍历 `data[].list[].cloudLinks[]`
- 仅保留 `cloudType == "quark"`
- 从 `title` 提取纯标题（正则去除"名称："、"标题："前缀）
- 从 `content` 提取纯描述（正则去除"描述："、"简介："前缀，去除 HTML `<mark>` 标签）
- `pubDate` 转换为 CST 时间格式 `YYYY-MM-DD HH:MM:SS`
- 按链接去重

### 4.3 PanSou (pansou.go)

**结构体**:
```go
type PanSouSource struct {
    baseURL string
}
```

**搜索接口**: `GET /api/search?kw={query}&cloud_types=["quark"]&res=merge&refresh={deep}`

**响应解析** (`formatResults`):
- 从 `data.merged_by_type.quark[]` 提取结果
- 正则从 `note` 分离标题和描述：`r'^(.*?)(?:[【\[]?(?:简介|介绍|描述)[】\]]?[:：]?)(.*)$'`
- `datetime` 转换为 CST 时间格式
- 按 URL 去重

### 4.4 配置管理 (config.go)

**Setting 表存储**:

| Key | 说明 |
|-----|------|
| `search.cloudsaver.server` | CloudSaver 服务地址 |
| `search.cloudsaver.username` | 登录用户名 |
| `search.cloudsaver.password` | 登录密码 |
| `search.cloudsaver.token` | JWT Token (自动更新) |
| `search.pansou.server` | PanSou 服务地址 |

**配置结构体**:
```go
type SearchConfig struct {
    CloudSaver CloudSaverConfig `json:"cloudsaver"`
    PanSou     PanSouConfig     `json:"pansou"`
}

type CloudSaverConfig struct {
    Server   string `json:"server"`
    Username string `json:"username"`
    Password string `json:"password"`
    Token    string `json:"token"`
}

type PanSouConfig struct {
    Server string `json:"server"`
}
```

**函数**:
- `LoadConfig(db *gorm.DB) (*SearchConfig, error)` — 从 Setting 表加载
- `SaveConfig(db *gorm.DB, config *SearchConfig) error` — 保存到 Setting 表

### 4.5 Client 重构 (client.go)

**变更**:
- `NewClient(config *SearchConfig)` — 用配置创建源，非硬编码 URL
- `Client.Search()` 增加去重 + 排序逻辑：
  1. 并发调用各 `Source.Search()`
  2. 合并所有 `SearchItem`
  3. 按 `URL` 去重
  4. 按 `UpdatedAt` 降序排列
- `Client.UpdateConfig(config *SearchConfig)` — 热更新配置（重建源），持有 `*gorm.DB` 引用以便持久化 Token

### 4.6 API 端点

**现有端点** (保留):
- `GET /api/search?q=&source=&page=` — 搜索资源
- `GET /api/search/sources` — 列出搜索源

**新增端点**:
- `GET /api/search/config` — 获取搜索源配置（密码脱敏返回 `***`）
- `PUT /api/search/config` — 更新搜索源配置，保存后自动调用 `Client.UpdateConfig()` 热更新搜索源

### 4.7 初始化流程 (main.go)

```go
// 原来
searchClient := search.NewClient()

// 改为
searchConfig, _ := search.LoadConfig(db)
searchClient := search.NewClient(searchConfig)
api.InitSearchHandler(searchClient)
```

## 5. 前端设计

### 5.1 Search.vue 更新

**字段适配**:
- 保留 `title`、`url`、`source`、`platform`、`updated_at`、`summary`
- 新增 `tags` 显示（el-tag 组件）
- 新增 `channel` 显示

**创建任务联动**: `handleCreateTask(item)` 传递 `share_url`、`title` 到 Tasks 页面（已有逻辑不变）

### 5.2 Settings.vue 新增搜索源配置 Tab

在现有 Tab 列表中新增"搜索源"Tab，包含：
- CloudSaver 配置区：服务地址、用户名、密码、Token 状态 + 测试连接按钮
- PanSou 配置区：服务地址 + 测试连接按钮
- 保存配置按钮

### 5.3 API 封装 (search.js)

新建 `web/src/api/search.js`：
- `searchResources(params)` — 搜索
- `listSearchSources()` — 列出源
- `getSearchConfig()` — 获取配置
- `updateSearchConfig(config)` — 更新配置

## 6. 错误处理

| 场景 | 处理方式 |
|------|----------|
| CloudSaver Token 过期 | 自动重新登录，重试搜索 |
| CloudSaver 登录失败 | 返回错误消息，不影响其他源 |
| 搜索源超时 (10s) | 跳过该源，返回其他源结果 |
| 搜索源全部失败 | 返回空结果 + 错误消息 |
| 配置项缺失 | 源不参与搜索，不报错 |

## 7. 测试策略

- `cloudsaver_test.go` — Mock HTTP 服务器测试认证流程和响应解析
- `pansou_test.go` — Mock HTTP 服务器测试搜索和结果格式化
- `config_test.go` — 测试配置加载和保存
- `client_test.go` — 补充去重、排序、多源并发测试
