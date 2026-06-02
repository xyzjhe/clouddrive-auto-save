# 搜索结果渐进式链接验证实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 搜索结果零延迟展示，后台并发验证链接有效性，通过 SSE 实时推送验证状态。

**Architecture:** 搜索 API 立即返回含 `search_id` 的全量结果，后端异步启动 15 并发 goroutine 验证每条链接的 `ParseShare`，结果通过现有 `Broadcaster` 以 `[EVENT:search_validate|...]` 格式推送，前端 SSE 监听逐条更新 UI。

**Tech Stack:** Go (goroutine + semaphore) / Gin / SSE / Vue 3 / Element Plus

---

## 文件结构

| 操作 | 文件 | 职责 |
|------|------|------|
| 修改 | `internal/utils/events.go` | 新增 `BroadcastSearchValidate` 函数 |
| 修改 | `internal/api/search.go` | Search handler 生成 `search_id`，异步启动验证 |
| 修改 | `web/src/utils/sse.js` | switch 新增 `search_validate` 事件分发 |
| 修改 | `web/src/views/Search.vue` | SSE 监听、三态 UI、验证进度条 |

---

### Task 1: 后端 SSE 事件广播函数

**Files:**
- Modify: `internal/utils/events.go`

- [ ] **Step 1: 新增 `SearchValidateEvent` 结构体和广播函数**

在 `internal/utils/events.go` 末尾添加：

```go
// SearchValidateEvent 搜索链接验证结果事件
type SearchValidateEvent struct {
	SearchID string `json:"search_id"`
	Index    int    `json:"index"`
	Valid    bool   `json:"valid"`
	Message  string `json:"message,omitempty"`
}

// BroadcastSearchValidate 推送搜索链接验证结果
func BroadcastSearchValidate(evt SearchValidateEvent) {
	b, _ := json.Marshal(evt)
	slog.Info("[EVENT:search_validate|" + string(b) + "]")
}
```

注意事件格式是 `[EVENT:search_validate|{...}]`，与现有 `sse.js` 的 `/\[EVENT:(\w+)\|(.*)\]/` 正则匹配。`\w+` 能匹配 `search_validate`（下划线是 `\w`），`(.*)` 捕获 JSON payload。

- [ ] **Step 2: 编译验证**

Run: `go build ./internal/utils/...`
Expected: 无错误

- [ ] **Step 3: 提交**

```bash
git add internal/utils/events.go
git commit -m "feat(search): 新增 SearchValidateEvent SSE 事件广播函数"
```

---

### Task 2: 后端搜索 API 集成异步验证

**Files:**
- Modify: `internal/api/search.go`

- [ ] **Step 1: 修改 Search handler，生成 search_id 并异步启动验证**

将 `internal/api/search.go` 的 `Search` 方法替换为：

```go
// Search 搜索资源
func (h *SearchHandler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.PureJSON(http.StatusBadRequest, gin.H{"error": "请提供搜索关键词"})
		return
	}

	sources := c.QueryArray("source")
	platforms := c.QueryArray("platform")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	result, err := h.client.Search(query, sources, platforms, page)
	if err != nil {
		c.PureJSON(http.StatusInternalServerError, gin.H{"error": "搜索失败"})
		return
	}

	// 生成 search_id 用于关联 SSE 验证事件
	searchID := "srch_" + generateHexID(8)

	// 异步启动链接验证
	if len(result.Items) > 0 {
		go validateSearchResults(searchID, result.Items)
	}

	c.PureJSON(http.StatusOK, gin.H{
		"total":     result.Total,
		"page":      result.Page,
		"items":     result.Items,
		"search_id": searchID,
	})
}
```

- [ ] **Step 2: 在文件末尾添加 `generateHexID` 和 `validateSearchResults` 函数**

在 `internal/api/search.go` 末尾（`ValidateLink` 方法之后）添加：

```go
// generateHexID 生成指定长度的随机 hex 字符串
func generateHexID(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)[:n]
}

// validateSearchResults 并发验证搜索结果中的链接有效性
// 通过 SSE 推送每条结果的验证状态
func validateSearchResults(searchID string, items []search.SearchItem) {
	sem := make(chan struct{}, 15) // 15 并发
	var wg sync.WaitGroup

	for i, item := range items {
		wg.Add(1)
		sem <- struct{}{} // 获取信号量
		go func(idx int, url string) {
			defer wg.Done()
			defer func() { <-sem }() // 释放信号量

			valid, message := validateSingleLink(url)
			utils.BroadcastSearchValidate(utils.SearchValidateEvent{
				SearchID: searchID,
				Index:    idx,
				Valid:    valid,
				Message:  message,
			})
		}(i, item.URL)
	}
	wg.Wait()
}

// validateSingleLink 验证单个分享链接有效性
func validateSingleLink(rawURL string) (bool, string) {
	if !isSafeURL(rawURL) {
		return false, "链接地址不合法"
	}

	driver := core.GetDriverByURL(rawURL)
	if driver == nil {
		return false, "不支持的链接格式"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := driver.ParseShare(ctx, rawURL, "", "")
	if err != nil {
		return false, err.Error()
	}
	return true, ""
}
```

- [ ] **Step 3: 更新 import 块**

`internal/api/search.go` 的 import 需要新增以下包：

```go
import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zcq/clouddrive-auto-save/internal/core"
	"github.com/zcq/clouddrive-auto-save/internal/core/search"
	"github.com/zcq/clouddrive-auto-save/internal/utils"
)
```

确认 `crypto/rand`、`encoding/hex`、`context`、`sync`、`time` 都已添加。

- [ ] **Step 4: 编译验证**

Run: `go build ./internal/api/...`
Expected: 无错误

- [ ] **Step 5: 提交**

```bash
git add internal/api/search.go
git commit -m "feat(search): 搜索 API 生成 search_id 并异步并发验证链接有效性"
```

---

### Task 3: 前端 SSE 事件解析支持

**Files:**
- Modify: `web/src/utils/sse.js`

- [ ] **Step 1: 在 switch 中新增 `search_validate` 事件分发**

在 `web/src/utils/sse.js` 的 `switch (eventType)` 块中（`case 'stats_update'` 之后），添加：

```javascript
            case 'search_validate':
              handlers.onSearchValidate?.(payload)
              break
```

同时在文件顶部 JSDoc 的 `@param` 中新增：

```javascript
 * @param {function(Object)} handlers.onSearchValidate 接收到搜索验证结果事件
```

- [ ] **Step 2: 构建验证**

Run: `cd web && npx vite build 2>&1 | tail -3`
Expected: 构建成功

- [ ] **Step 3: 提交**

```bash
git add web/src/utils/sse.js
git commit -m "feat(sse): 新增 search_validate 事件类型解析"
```

---

### Task 4: 前端 Search.vue 集成验证状态与进度

**Files:**
- Modify: `web/src/views/Search.vue`

这是最大的一个 Task，包含 4 个子步骤。

- [ ] **Step 4a: 添加验证相关状态变量**

在 `<script setup>` 中，`// 分页` 注释之前，添加：

```javascript
// 搜索验证状态
const currentSearchId = ref('')
const validateProgress = ref({ total: 0, valid: 0, invalid: 0, done: 0 })
```

- [ ] **Step 4b: 修改 `handleSearch`，保存 search_id 并初始化进度**

将 `handleSearch` 中的：

```javascript
    const data = await searchResources(params)
    results.value = data.items || []
```

替换为：

```javascript
    const data = await searchResources(params)
    results.value = (data.items || []).map(item => ({ ...item, valid: null, validMessage: '' }))
    currentSearchId.value = data.search_id || ''
    validateProgress.value = { total: results.value.length, valid: 0, invalid: 0, done: 0 }
```

- [ ] **Step 4c: 添加 SSE 监听逻辑**

在 `onMounted` 中初始化 SSE 监听，`onUnmounted` 中清理。需要修改 import 行和添加逻辑：

修改 import：

```javascript
import { ref, computed, onMounted, onUnmounted } from 'vue'
```

在 `onMounted` 内部末尾添加：

```javascript
  // 监听 SSE 验证事件
  if (window.__dashboardSSE) return // 复用 Dashboard 的 SSE 连接，不重复创建
```

实际上，Search 页面不是 Dashboard，它没有自己的 SSE 连接。需要通过 Dashboard 的全局 SSE 或新建独立监听。

由于 Dashboard 的 SSE 连接 (`EventSource('/api/dashboard/logs')`) 是页面级的，Search 页面需要建立自己的 SSE 监听。

在 `onMounted` 内部末尾、`onUnmounted` 在 `onMounted` 之后添加：

```javascript
// SSE 验证监听（连接到仪表盘日志流以接收验证事件）
let validateEventSource = null

onMounted(async () => {
  try {
    const data = await listSearchSources()
    sources.value = data || []
  } catch (error) {
    console.error('获取搜索源失败:', error)
  }

  // 建立 SSE 连接监听验证事件
  validateEventSource = new EventSource('/api/dashboard/logs')
  validateEventSource.onmessage = (event) => {
    const msg = event.data
    if (!msg || !msg.includes('[EVENT:search_validate|')) return

    const match = msg.match(/\[EVENT:search_validate\|(.+)\]/)
    if (!match) return
    try {
      const payload = JSON.parse(match[1])
      // 只处理当前搜索会话的事件
      if (payload.search_id !== currentSearchId.value) return

      const idx = payload.index
      if (idx >= 0 && idx < results.value.length) {
        results.value[idx].valid = payload.valid
        results.value[idx].validMessage = payload.message || ''
      }
      // 更新进度
      validateProgress.value.done++
      if (payload.valid) {
        validateProgress.value.valid++
      } else {
        validateProgress.value.invalid++
      }
    } catch (e) {
      // 解析失败忽略
    }
  }
})

onUnmounted(() => {
  if (validateEventSource) {
    validateEventSource.close()
    validateEventSource = null
  }
})
```

注意：需要把原来的 `onMounted` 内容合并到新的 `onMounted` 中。完整替换如下：

将现有的：

```javascript
onMounted(async () => {
  try {
    const data = await listSearchSources()
    sources.value = data || []
  } catch (error) {
    console.error('获取搜索源失败:', error)
  }
})
```

替换为上面包含 SSE 监听的完整 `onMounted` + `onUnmounted`。

- [ ] **Step 4d: 更新模板，增加验证状态图标和进度条**

1. 在搜索栏区域（`</div>` 闭合 `filter-section` 之后），添加验证进度条：

```html
      <!-- 验证进度 -->
      <div v-if="validateProgress.total > 0 && validateProgress.done < validateProgress.total" class="validate-progress">
        <el-icon class="is-loading"><SearchIcon /></el-icon>
        <span>验证进度：{{ validateProgress.done }}/{{ validateProgress.total }} ✅ {{ validateProgress.valid }} 条有效 | ❌ {{ validateProgress.invalid }} 条失效</span>
      </div>
      <div v-else-if="validateProgress.total > 0 && validateProgress.done === validateProgress.total" class="validate-progress done">
        <span>验证完成：✅ {{ validateProgress.valid }} 条有效 | ❌ {{ validateProgress.invalid }} 条失效</span>
      </div>
```

2. 更新 result-item 的点击和样式。将模板中现有的：

```html
        <div class="result-header">
          <div class="result-title">
            <span v-if="item.valid === true" class="valid-icon">✅</span>
            <span v-else-if="item.valid === false" class="valid-icon invalid" :title="item.validMessage">❌</span>
            {{ item.title }}
          </div>
          <el-button
            type="primary"
            size="small"
            @click.stop="handleCreateTask(item)"
          >
            创建任务
          </el-button>
        </div>
```

替换为：

```html
        <div class="result-header">
          <div class="result-title">
            <span v-if="item.valid === true" class="valid-icon">✅</span>
            <span v-else-if="item.valid === false" class="valid-icon invalid" :title="item.validMessage">❌</span>
            <span v-else class="valid-icon pending">⏳</span>
            {{ item.title }}
          </div>
          <el-button
            type="primary"
            size="small"
            :disabled="item.valid === false"
            @click.stop="handleCreateTask(item)"
          >
            创建任务
          </el-button>
        </div>
```

3. 更新 result-item 的点击行为（失效链接禁用点击）：

将：

```html
      <div
        v-for="item in paginatedResults"
        :key="item.url"
        class="result-item clickable"
        @click="handleResultClick(item)"
      >
```

替换为：

```html
      <div
        v-for="item in paginatedResults"
        :key="item.url"
        class="result-item clickable"
        :class="{ 'is-disabled': item.valid === false }"
        @click="item.valid !== false && handleResultClick(item)"
      >
```

4. 添加对应 CSS。在 `<style scoped>` 末尾添加：

```css
.validate-progress {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.6rem;
  margin-top: 0.5rem;
  color: var(--text-secondary);
  font-size: 0.85rem;
  background: var(--bg-secondary);
  border-radius: 8px;
}

.validate-progress.done {
  color: var(--el-color-success);
}

.valid-icon.pending {
  opacity: 0.5;
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 0.5; }
  50% { opacity: 1; }
}

.result-item.is-disabled {
  opacity: 0.5;
  pointer-events: none;
}
```

- [ ] **Step 4e: 构建验证**

Run: `cd web && npx vite build 2>&1 | tail -3`
Expected: 构建成功

- [ ] **Step 4f: 提交**

```bash
git add web/src/views/Search.vue
git commit -m "feat(search): 前端集成 SSE 验证状态监听、三态图标和进度条"
```

---

### Task 5: 全量测试与最终提交

- [ ] **Step 1: 后端全量测试**

Run: `cd /home/zcq/Github/clouddrive-auto-save && go test ./... 2>&1 | tail -20`
Expected: 所有 PASS

- [ ] **Step 2: 前端构建**

Run: `cd web && npx vite build 2>&1 | tail -3`
Expected: 构建成功

- [ ] **Step 3: 检查 SSE 事件格式一致性**

确认链路完整：
1. `events.go` 发送格式：`[EVENT:search_validate|{"search_id":"...","index":0,"valid":true}]`
2. `sse.js` 正则 `/\[EVENT:(\w+)\|(.*)\]/` 匹配 `search_validate` 类型
3. `Search.vue` 正则 `/\[EVENT:search_validate\|(.+)\]/` 提取 payload

- [ ] **Step 4: 最终提交（如有构建产物更新）**

```bash
git add -A
git commit -m "chore: 搜索渐进式验证功能全量构建"
```
