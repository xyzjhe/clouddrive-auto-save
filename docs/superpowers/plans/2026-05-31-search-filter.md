# 资源搜索筛选功能实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现资源搜索的网盘类型筛选功能，动态获取搜索源，解除硬编码限制

**Architecture:** 后端 Source 接口增加 platforms 参数，前端从 API 获取可用搜索源并新增网盘类型筛选器

**Tech Stack:** Go (Gin), Vue 3, Element Plus

---

## 文件结构

| 文件 | 操作 | 说明 |
|------|------|------|
| `internal/core/search/sources.go` | 修改 | Source 接口增加 platforms 参数 |
| `internal/core/search/pansou.go` | 修改 | 支持 platforms 参数，解除 quark 硬编码 |
| `internal/core/search/cloudsaver.go` | 修改 | 支持 platforms 参数，解除 quark 硬编码 |
| `internal/core/search/client.go` | 修改 | Search 方法传递 platforms |
| `internal/api/search.go` | 修改 | 解析 platform 查询参数 |
| `internal/core/search/*_test.go` | 修改 | 更新测试用例 |
| `web/src/views/Search.vue` | 修改 | 动态搜索源 + 网盘类型筛选 |
| `web/src/api/search.js` | 不变 | 已有 listSearchSources 接口 |

---

## Task 1: 修改 Source 接口定义

**Files:**
- Modify: `internal/core/search/sources.go:9-12`

- [ ] **Step 1: 修改 Source 接口**

```go
// Source 搜索源接口
type Source interface {
	Name() string
	Search(query string, platforms []string, page int) (*SearchResult, error)
}
```

- [ ] **Step 2: 验证编译**

Run: `cd /home/zcq/Github/clouddrive-auto-save && go build ./internal/core/search/`
Expected: 编译失败（因为实现类未更新）

- [ ] **Step 3: 提交**

```bash
git add internal/core/search/sources.go
git commit -m "refactor(search): Source 接口增加 platforms 参数"
```

---

## Task 2: 更新 PanSou 搜索源

**Files:**
- Modify: `internal/core/search/pansou.go:47-84`

- [ ] **Step 1: 修改 Search 方法签名和实现**

```go
// Search 搜索资源
func (s *PanSouSource) Search(query string, platforms []string, page int) (*SearchResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	params := url.Values{}
	params.Set("kw", query)
	params.Set("res", "merge")

	// 根据 platforms 动态构建 cloud_types
	if len(platforms) > 0 {
		params.Set("cloud_types", strings.Join(platforms, ","))
	} else {
		params.Set("cloud_types", "quark,139")
	}

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

	var allItems []SearchItem
	allItems = append(allItems, s.formatResults(result.Data.MergedByType.Quark, "quark")...)
	allItems = append(allItems, s.formatResults(result.Data.MergedByType.Cloud139, "139")...)

	return &SearchResult{
		Total: len(allItems),
		Page:  page,
		Items: allItems,
	}, nil
}
```

- [ ] **Step 2: 更新 psSearchResponse 结构体**

```go
// psSearchResponse PanSou 搜索响应
type psSearchResponse struct {
	Code int `json:"code"`
	Data struct {
		MergedByType struct {
			Quark    []psItem `json:"quark"`
			Cloud139 []psItem `json:"139"`
		} `json:"merged_by_type"`
	} `json:"data"`
}
```

- [ ] **Step 3: 更新 formatResults 方法增加 platform 参数**

```go
// formatResults 格式化搜索结果
func (s *PanSouSource) formatResults(data []psItem, platform string) []SearchItem {
	pattern := regexp.MustCompile(`^(.*?)(?:【(?:简介|介绍|描述)】|\[(?:简介|介绍|描述)\]|(?:简介|介绍|描述)[:：])(.*)$`)

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
			Platform:  platform,
			Summary:   content,
			UpdatedAt: toCST(item.DateTime),
			Channel:   item.Source,
		})
	}
	return items
}
```

- [ ] **Step 4: 验证编译**

Run: `cd /home/zcq/Github/clouddrive-auto-save && go build ./internal/core/search/`
Expected: 编译成功

- [ ] **Step 5: 提交**

```bash
git add internal/core/search/pansou.go
git commit -m "feat(search): PanSou 支持 platforms 参数和 139 类型"
```

---

## Task 3: 更新 CloudSaver 搜索源

**Files:**
- Modify: `internal/core/search/cloudsaver.go:80-106,149-218`

- [ ] **Step 1: 修改 Search 方法签名**

```go
// Search 搜索资源
func (s *CloudSaverSource) Search(query string, platforms []string, page int) (*SearchResult, error) {
	result, err := s.doSearch(query, "")
	if err != nil {
		return nil, err
	}

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

	items := s.cleanResults(result.Data, platforms)
	return &SearchResult{
		Total: len(items),
		Page:  page,
		Items: items,
	}, nil
}
```

- [ ] **Step 2: 修改 cleanResults 方法**

```go
// cleanResults 清洗搜索结果
func (s *CloudSaverSource) cleanResults(data []map[string]interface{}, platforms []string) []SearchItem {
	var items []SearchItem
	seen := make(map[string]bool)

	patternTitle := regexp.MustCompile(`(?:名称|标题)[：:]?\s*(.*)`)
	patternContent := regexp.MustCompile(`(?:描述|简介)[：:]?\s*(.*?)(?:链接|标签|$)`)
	patternHTML := regexp.MustCompile(`<[^>]+>`)

	for _, channel := range data {
		list, _ := channel["list"].([]interface{})
		for _, item := range list {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			title, _ := itemMap["title"].(string)
			content, _ := itemMap["content"].(string)
			pubDate, _ := itemMap["pubDate"].(string)
			channelID, _ := itemMap["channelId"].(string)

			var tags []string
			if tagsRaw, ok := itemMap["tags"].([]interface{}); ok {
				for _, t := range tagsRaw {
					if s, ok := t.(string); ok {
						tags = append(tags, s)
					}
				}
			}

			if m := patternTitle.FindStringSubmatch(title); len(m) > 1 {
				title = strings.TrimSpace(m[1])
			}

			if m := patternContent.FindStringSubmatch(content); len(m) > 1 {
				content = strings.TrimSpace(m[1])
			}
			content = patternHTML.ReplaceAllString(content, "")
			content = strings.TrimSpace(content)

			cloudLinks, _ := itemMap["cloudLinks"].([]interface{})
			for _, link := range cloudLinks {
				linkMap, ok := link.(map[string]interface{})
				if !ok {
					continue
				}
				cloudType, _ := linkMap["cloudType"].(string)

				// 根据 platforms 过滤（如果指定了）
				if len(platforms) > 0 && !contains(platforms, cloudType) {
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
					Platform:  cloudType,
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

- [ ] **Step 3: 验证编译**

Run: `cd /home/zcq/Github/clouddrive-auto-save && go build ./internal/core/search/`
Expected: 编译成功

- [ ] **Step 4: 提交**

```bash
git add internal/core/search/cloudsaver.go
git commit -m "feat(search): CloudSaver 支持 platforms 参数，解除 quark 硬编码"
```

---

## Task 4: 更新 Client 调用层

**Files:**
- Modify: `internal/core/search/client.go:90-143`

- [ ] **Step 1: 修改 Search 方法**

```go
// Search 搜索资源（并发 + 去重 + 排序）
func (c *Client) Search(query string, sources []string, platforms []string, page int) (*SearchResult, error) {
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
			result, err := src.Search(query, platforms, page)
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
```

- [ ] **Step 2: 验证编译**

Run: `cd /home/zcq/Github/clouddrive-auto-save && go build ./internal/core/search/`
Expected: 编译成功

- [ ] **Step 3: 提交**

```bash
git add internal/core/search/client.go
git commit -m "feat(search): Client.Search 传递 platforms 参数"
```

---

## Task 5: 更新 API Handler

**Files:**
- Modify: `internal/api/search.go:24-41`

- [ ] **Step 1: 修改 Search handler**

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

	c.PureJSON(http.StatusOK, result)
}
```

- [ ] **Step 2: 验证编译**

Run: `cd /home/zcq/Github/clouddrive-auto-save && go build ./internal/api/`
Expected: 编译成功

- [ ] **Step 3: 提交**

```bash
git add internal/api/search.go
git commit -m "feat(search): API 支持 platform 查询参数"
```

---

## Task 6: 更新测试用例

**Files:**
- Modify: `internal/core/search/pansou_test.go`
- Modify: `internal/core/search/cloudsaver_test.go`
- Modify: `internal/core/search/client_test.go`

- [ ] **Step 1: 更新 PanSou 测试**

更新所有调用 `Search` 方法的地方，增加 `platforms` 参数（可传 `nil` 或 `[]string{}`）

- [ ] **Step 2: 更新 CloudSaver 测试**

更新所有调用 `Search` 方法的地方，增加 `platforms` 参数

- [ ] **Step 3: 更新 Client 测试**

更新所有调用 `Search` 方法的地方，增加 `platforms` 参数

- [ ] **Step 4: 运行测试**

Run: `cd /home/zcq/Github/clouddrive-auto-save && go test ./internal/core/search/... -v`
Expected: 所有测试通过

- [ ] **Step 5: 提交**

```bash
git add internal/core/search/*_test.go
git commit -m "test(search): 更新测试用例适配 platforms 参数"
```

---

## Task 7: 更新前端搜索页面

**Files:**
- Modify: `web/src/views/Search.vue`

- [ ] **Step 1: 添加 import 和 onMounted**

```vue
<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search as SearchIcon, Link as LinkIcon, Clock as ClockIcon, FileText as FileTextIcon } from 'lucide-vue-next'
import { searchResources, listSearchSources } from '../api/search'

const router = useRouter()
const query = ref('')
const sources = ref([])
const selectedSources = ref([])
const results = ref([])
const loading = ref(false)
const page = ref(1)

// 网盘类型筛选
const platforms = [
  { label: '全部', value: '' },
  { label: '夸克网盘', value: 'quark' },
  { label: '移动云盘', value: '139' }
]
const selectedPlatforms = ref([])

onMounted(async () => {
  try {
    const data = await listSearchSources()
    sources.value = data || []
  } catch (error) {
    console.error('获取搜索源失败:', error)
  }
})
// ... 其他代码
</script>
```

- [ ] **Step 2: 更新 handleSearch 方法**

```vue
<script setup>
// ... 其他代码

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
    if (selectedPlatforms.value.length > 0) {
      params.platform = selectedPlatforms.value
    }
    const data = await searchResources(params)
    results.value = data.items || []
  } catch (error) {
    console.error('搜索失败:', error)
  } finally {
    loading.value = false
  }
}

// ... 其他代码
</script>
```

- [ ] **Step 3: 更新模板，添加网盘类型筛选**

```vue
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

      <div class="filter-section">
        <div v-if="sources.length > 0" class="source-filter">
          <span class="filter-label">搜索源：</span>
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

        <div class="platform-filter">
          <span class="filter-label">网盘类型：</span>
          <el-checkbox-group v-model="selectedPlatforms">
            <el-checkbox
              v-for="p in platforms"
              :key="p.value"
              :label="p.value"
            >
              {{ p.label }}
            </el-checkbox>
          </el-checkbox-group>
        </div>
      </div>
    </div>

    <!-- ... 其他模板代码 -->
  </div>
</template>
```

- [ ] **Step 4: 添加样式**

```vue
<style scoped>
/* ... 其他样式 */

.filter-section {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  margin-top: 1rem;
}

.source-filter,
.platform-filter {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.filter-label {
  font-size: 0.9rem;
  color: var(--text-secondary);
  white-space: nowrap;
}
</style>
```

- [ ] **Step 5: 验证前端编译**

Run: `cd /home/zcq/Github/clouddrive-auto-save/web && npm run build`
Expected: 构建成功

- [ ] **Step 6: 提交**

```bash
git add web/src/views/Search.vue
git commit -m "feat(search): 前端动态搜索源 + 网盘类型筛选"
```

---

## Task 8: 运行完整测试

- [ ] **Step 1: 运行后端测试**

Run: `cd /home/zcq/Github/clouddrive-auto-save && go test ./... -v`
Expected: 所有测试通过

- [ ] **Step 2: 运行 lint 检查**

Run: `cd /home/zcq/Github/clouddrive-auto-save && make check`
Expected: 所有检查通过

- [ ] **Step 3: 运行 E2E 测试**

Run: `cd /home/zcq/Github/clouddrive-auto-save && export PATH=$PATH:/usr/local/go/bin && export PLAYWRIGHT_HOST_PLATFORM_OVERRIDE=ubuntu24.04-x64 && make e2e-test`
Expected: 所有测试通过

- [ ] **Step 4: 最终提交**

```bash
git add -A
git commit -m "feat(search): 完成资源搜索筛选功能

- 搜索源选择框从后端动态获取
- 新增网盘类型筛选（夸克/移动云盘）
- 解除后端硬编码的网盘类型限制
- 更新测试用例"
```
