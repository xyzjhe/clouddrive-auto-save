# 搜索结果渐进式链接验证设计

> 日期：2026-06-02
> 状态：待实现

## 背景

当前资源搜索将全量结果直接展示，超过 30% 的链接已失效（分享取消、过期、提取码错误等），用户体验差。此前尝试过逐条串行验证，200 条需 160s+，不可接受。

## 目标

- 搜索结果零延迟展示
- 后台并发验证链接有效性
- 通过 SSE 实时推送验证状态，前端逐条更新
- 用户无需等待全部验证完成即可操作有效链接

## 整体流程

```
用户搜索 "斗破苍穹"
        │
        ▼
GET /api/search?q=斗破苍穹
        │
        ├─ 立即返回搜索结果（含 search_id）
        │
        └─ 后端异步启动并发验证（15 goroutine）
              │
              ▼
         对每条结果调用 GetDriverByURL → ParseShare
              │
              ▼
         通过 SSE 推送验证结果
              │
              ▼
         前端监听 SSE，逐条更新 ✅/❌
```

## 后端设计

### 搜索 API 变更

`GET /api/search` 响应新增 `search_id` 字段：

```json
{
  "total": 85,
  "page": 1,
  "items": [...],
  "search_id": "srch_a1b2c3"
}
```

`search_id` 生成规则：`srch_` + 8 位随机 hex，用于前端匹配 SSE 事件。

### 并发验证流水线

在 `SearchHandler.Search` 返回响应后，异步启动 `go validateSearchResults(searchID, items)`：

- **并发控制**：`semaphore(15)` goroutine，平衡速度和云盘 API 压力
- **单条验证**：`GetDriverByURL(url)` 判断平台 → `driver.ParseShare(ctx, url, "", "")` 验证
- **单条超时**：5s（ParseShare 正常 1-2s）
- **结果推送**：每条完成即通过 `Broadcaster` 发布 SSE 事件
- **全局开关**：Setting 表 `search_validate_enabled`，默认开启

### SSE 事件格式

复用现有 `[EVENT:...]` 协议：

```
event: message
data: [EVENT:search_validate|{"search_id":"srch_a1b2c3","index":0,"valid":true}]

data: [EVENT:search_validate|{"search_id":"srch_a1b2c3","index":1,"valid":false,"message":"分享链接不存在或已被取消"}]
```

字段说明：

| 字段 | 类型 | 说明 |
|------|------|------|
| `search_id` | string | 关联搜索会话 |
| `index` | int | items 数组下标 |
| `valid` | bool | 链接是否有效 |
| `message` | string | 失败原因（仅 valid=false 时存在） |

## 前端设计

### Item 三态

| `valid` 值 | 含义 | 图标 | 可点击 |
|---|---|---|---|
| `null` / `undefined` | 未验证 | ⏳ 灰色 | ✅ |
| `true` | 有效 | ✅ 绿色 | ✅ |
| `false` | 失效 | ❌ 红色 + tooltip | ❌ 禁用 |

### SSE 监听

`Search.vue` 监听全局 SSE 事件，过滤 `search_validate` 类型，严格匹配当前 `search_id`：

```
SSE 事件到达
     ├─ type !== "search_validate" → 忽略
     └─ search_id !== 当前 search_id → 忽略（防串扰）
           │
           ▼
     results.value[index].valid = event.valid
     results.value[index].validMessage = event.message
```

### 验证进度

搜索栏下方轻量进度提示：

```
验证进度：85/200 ✅ 72 条有效 | ❌ 13 条失效
```

全部完成后自动隐藏。

### 多次搜索防串扰

每次搜索生成新 `search_id`，SSE 监听只处理当前 ID 的事件，旧 ID 事件静默丢弃。

## 涉及文件

| 文件 | 变更 |
|------|------|
| `internal/api/search.go` | `Search` handler 生成 `search_id`，异步启动验证 |
| `internal/api/search.go` | 新增 `validateSearchResults` 函数 |
| `web/src/views/Search.vue` | 监听 SSE 验证事件，更新 item 状态，显示进度 |
| `web/src/utils/sse.js` | 新增 `search_validate` 事件类型解析 |
| `internal/utils/broadcaster.go` | 无变更，复用现有广播机制 |

## 并发参数

| 参数 | 值 | 理由 |
|------|-----|------|
| 并发 goroutine 数 | 15 | 平衡速度和云盘 API 压力 |
| 单条验证超时 | 5s | ParseShare 正常 1-2s，5s 覆盖慢请求 |
| 全局验证开关 | `search_validate_enabled` | 可关闭，默认开启 |
