# 资源搜索筛选功能设计

**日期：** 2026-05-31
**状态：** 已批准

## 问题背景

当前资源搜索界面存在以下问题：

1. **搜索源选择框无效**：`sources` 硬编码为 `['CloudSaver', 'PanSou']`，不反映实际配置状态
2. **CloudSaver 结果缺失**：`cleanResults` 方法硬编码只保留 `quark` 类型，过滤掉其他网盘
3. **缺少网盘类型筛选**：无法按网盘类型（夸克/移动云盘）筛选搜索结果

## 设计目标

1. 搜索源选择框从后端动态获取，只显示已配置可用的源
2. 新增网盘类型筛选器，支持按平台过滤结果
3. 解除后端硬编码的网盘类型限制

## 详细设计

### 1. 后端改动

#### 1.1 Source 接口扩展

**文件：** `internal/core/search/sources.go`

```go
// Source 搜索源接口
type Source interface {
    Name() string
    Search(query string, platforms []string, page int) (*SearchResult, error)
}
```

#### 1.2 PanSou 搜索源

**文件：** `internal/core/search/pansou.go`

- `Search` 方法增加 `platforms []string` 参数
- 移除硬编码 `cloud_types=quark`
- 根据 platforms 动态构建 `cloud_types` 参数
- 解析响应支持 `quark` 和 `139` 两种类型

```go
func (s *PanSouSource) Search(query string, platforms []string, page int) (*SearchResult, error) {
    params := url.Values{}
    params.Set("kw", query)
    params.Set("res", "merge")

    // 根据 platforms 动态构建 cloud_types
    if len(platforms) > 0 {
        params.Set("cloud_types", strings.Join(platforms, ","))
    }

    // ... 解析响应时处理 quark 和 139 两种类型
}
```

#### 1.3 CloudSaver 搜索源

**文件：** `internal/core/search/cloudsaver.go`

- `Search` 方法增加 `platforms []string` 参数
- `cleanResults` 方法移除硬编码过滤
- 根据 platforms 参数过滤结果

```go
func (s *CloudSaverSource) cleanResults(data []map[string]interface{}, platforms []string) []SearchItem {
    // 移除: if cloudType != "quark" { continue }

    // 根据 platforms 过滤（如果指定了）
    if len(platforms) > 0 && !contains(platforms, cloudType) {
        continue
    }
}
```

#### 1.4 Client 调用层

**文件：** `internal/core/search/client.go`

```go
func (c *Client) Search(query string, sources []string, platforms []string, page int) (*SearchResult, error) {
    // ... 传递 platforms 给各搜索源
    result, err := src.Search(query, platforms, page)
}
```

#### 1.5 API Handler

**文件：** `internal/api/search.go`

```go
func (h *SearchHandler) Search(c *gin.Context) {
    query := c.Query("q")
    sources := c.QueryArray("source")
    platforms := c.QueryArray("platform")  // 新增
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

    result, err := h.client.Search(query, sources, platforms, page)
    // ...
}
```

### 2. 前端改动

**文件：** `web/src/views/Search.vue`

```vue
<script setup>
// 移除硬编码
// const sources = ref(['CloudSaver', 'PanSou'])

// 动态获取可用搜索源
const sources = ref([])
const selectedSources = ref([])

onMounted(async () => {
  const data = await listSearchSources()
  sources.value = data
})

// 新增网盘类型筛选
const platforms = [
  { label: '全部', value: '' },
  { label: '夸克网盘', value: 'quark' },
  { label: '移动云盘', value: '139' }
]
const selectedPlatforms = ref([])

const handleSearch = async () => {
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
  // ...
}
</script>

<template>
  <!-- 搜索源选择框 -->
  <div class="source-filter">
    <el-checkbox-group v-model="selectedSources">
      <el-checkbox v-for="source in sources" :key="source" :label="source">
        {{ source }}
      </el-checkbox>
    </el-checkbox-group>
  </div>

  <!-- 网盘类型筛选 -->
  <div class="platform-filter">
    <el-checkbox-group v-model="selectedPlatforms">
      <el-checkbox v-for="p in platforms" :key="p.value" :label="p.value">
        {{ p.label }}
      </el-checkbox>
    </el-checkbox-group>
  </div>
</template>
```

### 3. API 接口

**请求：**
```
GET /api/search?q=关键词&source=CloudSaver&source=PanSou&platform=quark&platform=139&page=1
```

**参数说明：**
- `q`：搜索关键词（必填）
- `source`：搜索源（可选，多选）
- `platform`：网盘类型（可选，多选：`quark`、`139`）
- `page`：页码（可选，默认 1）

**响应：**
```json
{
  "total": 10,
  "page": 1,
  "items": [
    {
      "title": "资源标题",
      "source": "CloudSaver",
      "platform": "quark",
      "url": "https://pan.quark.cn/s/xxx",
      "summary": "资源描述",
      "updated_at": "2026-05-31 10:00:00",
      "tags": ["标签1"],
      "channel": "频道名"
    }
  ]
}
```

## 测试要点

1. 搜索源选择框只显示已配置的源
2. 网盘类型筛选正确过滤结果
3. CloudSaver 能返回移动云盘结果（如果配置了）
4. PanSou 能根据 platform 参数调整请求
5. 多选组合（source + platform）正确工作
