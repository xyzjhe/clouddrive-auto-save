# 搜索增强功能实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现搜索结果链接校验、分享内容预览弹窗、任务搜索替换功能

**Architecture:** 前端添加校验逻辑和弹窗组件，复用现有后端 API

**Tech Stack:** Vue 3, Element Plus, Go (Gin)

---

## 文件结构

| 文件 | 操作 | 说明 |
|------|------|------|
| `web/src/api/search.js` | 修改 | 添加 validateLink API |
| `web/src/views/Search.vue` | 修改 | 添加校验逻辑、点击事件 |
| `web/src/components/ShareContentDialog.vue` | 新增 | 分享内容弹窗组件 |
| `web/src/views/Tasks.vue` | 修改 | 添加搜索替换按钮和弹窗 |

---

## Task 1: 添加 validateLink API

**Files:**
- Modify: `web/src/api/search.js`

- [ ] **Step 1: 添加 validateLink 函数**

```javascript
// 在文件末尾添加
export function validateLink(url) {
  return request({
    url: '/search/validate',
    method: 'get',
    params: { url }
  })
}
```

- [ ] **Step 2: 验证前端编译**

Run: `cd /home/zcq/Github/clouddrive-auto-save/web && npm run build`
Expected: 构建成功

- [ ] **Step 3: 提交**

```bash
git add web/src/api/search.js
git commit -m "feat(search): 添加 validateLink API"
```

---

## Task 2: 搜索结果链接有效性校验

**Files:**
- Modify: `web/src/views/Search.vue`

- [ ] **Step 1: 添加校验状态字段和校验函数**

在 `<script setup>` 中添加：

```javascript
// 校验状态
const validating = ref(false)

// 批量校验链接有效性
const validateLinks = async (items) => {
  validating.value = true
  const promises = items.map(async (item) => {
    try {
      const res = await validateLink(item.url)
      item.valid = res.valid
      item.validMessage = res.message || ''
    } catch (e) {
      item.valid = false
      item.validMessage = '校验失败'
    }
  })
  await Promise.allSettled(promises)
  validating.value = false
}
```

- [ ] **Step 2: 修改 handleSearch 调用校验**

在 `handleSearch` 函数中，获取搜索结果后调用校验：

```javascript
const handleSearch = async () => {
  // ... 现有代码 ...
  const data = await searchResources(params)
  results.value = data.items || []
  
  // 自动校验链接有效性
  if (results.value.length > 0) {
    validateLinks(results.value)
  }
}
```

- [ ] **Step 3: 添加校验状态显示**

在模板的结果项中添加校验状态图标：

```vue
<div class="result-header">
  <div class="result-title">
    <span v-if="item.valid === true" class="valid-icon">✅</span>
    <span v-else-if="item.valid === false" class="valid-icon invalid" :title="item.validMessage">❌</span>
    <span v-else-if="validating" class="valid-icon">⏳</span>
    {{ item.title }}
  </div>
  <!-- ... 现有代码 ... -->
</div>
```

- [ ] **Step 4: 添加样式**

```vue
<style scoped>
.valid-icon {
  margin-right: 4px;
  font-size: 14px;
}

.valid-icon.invalid {
  cursor: help;
}
</style>
```

- [ ] **Step 5: 验证前端编译**

Run: `cd /home/zcq/Github/clouddrive-auto-save/web && npm run build`
Expected: 构建成功

- [ ] **Step 6: 提交**

```bash
git add web/src/views/Search.vue
git commit -m "feat(search): 搜索结果自动校验链接有效性"
```

---

## Task 3: 创建 ShareContentDialog 组件

**Files:**
- Create: `web/src/components/ShareContentDialog.vue`

- [ ] **Step 1: 创建组件文件**

```vue
<script setup>
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { parseShareLink } from '../api/task'

const props = defineProps({
  visible: Boolean,
  url: String,
  extractCode: String,
  title: String,
  showReplace: Boolean
})

const emit = defineEmits(['update:visible', 'create-task', 'replace-link'])

const loading = ref(false)
const files = ref([])

watch(() => props.visible, async (val) => {
  if (val && props.url) {
    await loadFiles()
  } else {
    files.value = []
  }
})

const loadFiles = async () => {
  loading.value = true
  try {
    const res = await parseShareLink({
      share_url: props.url,
      extract_code: props.extractCode || ''
    })
    files.value = res.items || []
  } catch (e) {
    ElMessage.error('获取分享内容失败')
    files.value = []
  } finally {
    loading.value = false
  }
}

const formatSize = (bytes) => {
  if (!bytes) return ''
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let i = 0
  while (bytes >= 1024 && i < units.length - 1) {
    bytes /= 1024
    i++
  }
  return `${bytes.toFixed(1)} ${units[i]}`
}

const handleCreateTask = () => {
  emit('create-task', { url: props.url, extractCode: props.extractCode })
}

const handleReplaceLink = () => {
  emit('replace-link', { url: props.url, extractCode: props.extractCode })
}

const handleClose = () => {
  emit('update:visible', false)
}
</script>

<template>
  <el-dialog
    :model-value="visible"
    @update:model-value="handleClose"
    :title="`📁 分享内容：${title || '未知'}`"
    width="500px"
  >
    <div v-loading="loading" class="file-list">
      <div v-if="files.length === 0 && !loading" class="empty-tip">
        暂无文件信息
      </div>
      <div v-for="file in files" :key="file.id" class="file-item">
        <span class="file-icon">{{ file.is_folder ? '📁' : '📄' }}</span>
        <span class="file-name">{{ file.name }}</span>
        <span v-if="!file.is_folder" class="file-size">{{ formatSize(file.size) }}</span>
      </div>
    </div>
    <template #footer>
      <el-button type="primary" @click="handleCreateTask">创建任务</el-button>
      <el-button v-if="showReplace" type="warning" @click="handleReplaceLink">替换任务链接</el-button>
      <el-button @click="handleClose">关闭</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.file-list {
  min-height: 200px;
  max-height: 400px;
  overflow-y: auto;
}

.file-item {
  display: flex;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.file-icon {
  margin-right: 8px;
  font-size: 16px;
}

.file-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-size {
  color: var(--el-text-color-secondary);
  font-size: 12px;
  margin-left: 8px;
}

.empty-tip {
  text-align: center;
  color: var(--el-text-color-secondary);
  padding: 40px 0;
}
</style>
```

- [ ] **Step 2: 验证前端编译**

Run: `cd /home/zcq/Github/clouddrive-auto-save/web && npm run build`
Expected: 构建成功

- [ ] **Step 3: 提交**

```bash
git add web/src/components/ShareContentDialog.vue
git commit -m "feat(search): 添加分享内容弹窗组件"
```

---

## Task 4: 搜索结果点击展示分享内容弹窗

**Files:**
- Modify: `web/src/views/Search.vue`

- [ ] **Step 1: 导入组件和添加状态**

在 `<script setup>` 中添加：

```javascript
import ShareContentDialog from '../components/ShareContentDialog.vue'

// 分享内容弹窗
const shareDialogVisible = ref(false)
const shareDialogUrl = ref('')
const shareDialogExtractCode = ref('')
const shareDialogTitle = ref('')
```

- [ ] **Step 2: 添加点击事件处理**

```javascript
const handleResultClick = (item) => {
  shareDialogUrl.value = item.url
  shareDialogExtractCode.value = ''
  shareDialogTitle.value = item.title
  shareDialogVisible.value = true
}

const handleCreateTaskFromDialog = (data) => {
  shareDialogVisible.value = false
  router.push({
    name: 'Tasks',
    query: {
      share_url: data.url,
      extract_code: data.extractCode
    }
  })
}
```

- [ ] **Step 3: 修改模板添加点击事件**

在搜索结果项上添加点击事件：

```vue
<div
  v-for="item in results"
  :key="item.url"
  class="result-item clickable"
  @click="handleResultClick(item)"
>
```

- [ ] **Step 4: 添加弹窗组件**

在模板末尾添加：

```vue
<ShareContentDialog
  v-model:visible="shareDialogVisible"
  :url="shareDialogUrl"
  :extract-code="shareDialogExtractCode"
  :title="shareDialogTitle"
  :show-replace="false"
  @create-task="handleCreateTaskFromDialog"
/>
```

- [ ] **Step 5: 添加样式**

```vue
<style scoped>
.result-item.clickable {
  cursor: pointer;
  transition: box-shadow 0.2s;
}

.result-item.clickable:hover {
  box-shadow: var(--shadow-md);
}
</style>
```

- [ ] **Step 6: 验证前端编译**

Run: `cd /home/zcq/Github/clouddrive-auto-save/web && npm run build`
Expected: 构建成功

- [ ] **Step 7: 提交**

```bash
git add web/src/views/Search.vue
git commit -m "feat(search): 搜索结果点击展示分享内容弹窗"
```

---

## Task 5: 任务管理中搜索替换功能

**Files:**
- Modify: `web/src/views/Tasks.vue`

- [ ] **Step 1: 导入搜索 API 和组件**

在 `<script setup>` 中添加：

```javascript
import { searchResources } from '../api/search'
import ShareContentDialog from '../components/ShareContentDialog.vue'

// 搜索替换相关
const searchReplaceVisible = ref(false)
const searchReplaceQuery = ref('')
const searchReplaceResults = ref([])
const searchReplaceLoading = ref(false)

// 分享内容弹窗
const shareDialogVisible = ref(false)
const shareDialogUrl = ref('')
const shareDialogExtractCode = ref('')
const shareDialogTitle = ref('')
```

- [ ] **Step 2: 添加搜索替换函数**

```javascript
const handleSearchReplace = () => {
  searchReplaceQuery.value = form.value.name
  searchReplaceResults.value = []
  searchReplaceVisible.value = true
  if (searchReplaceQuery.value) {
    doSearchReplace()
  }
}

const doSearchReplace = async () => {
  if (!searchReplaceQuery.value.trim()) return
  searchReplaceLoading.value = true
  try {
    const data = await searchResources({ q: searchReplaceQuery.value })
    searchReplaceResults.value = data.items || []
  } catch (e) {
    ElMessage.error('搜索失败')
  } finally {
    searchReplaceLoading.value = false
  }
}

const handleReplaceFromSearch = (item) => {
  form.value.share_url = item.url
  searchReplaceVisible.value = false
  ElMessage.success('链接已替换，请保存任务')
}

const handleViewShareContent = (item) => {
  shareDialogUrl.value = item.url
  shareDialogExtractCode.value = ''
  shareDialogTitle.value = item.title
  shareDialogVisible.value = true
}

const handleReplaceFromDialog = (data) => {
  form.value.share_url = data.url
  form.value.extract_code = data.extractCode || ''
  shareDialogVisible.value = false
  ElMessage.success('链接已替换，请保存任务')
}
```

- [ ] **Step 3: 在任务编辑抽屉中添加搜索替换按钮**

在分享链接输入框旁添加按钮：

```vue
<el-form-item label="分享链接">
  <div class="share-url-input">
    <el-input v-model="form.share_url" placeholder="请输入分享链接" />
    <el-button type="primary" @click="handleSearchReplace" style="margin-left: 8px">
      搜索替换
    </el-button>
  </div>
</el-form-item>
```

- [ ] **Step 4: 添加搜索替换弹窗**

在模板末尾添加：

```vue
<el-dialog
  v-model="searchReplaceVisible"
  title="🔍 搜索替换资源"
  width="600px"
>
  <div class="search-replace-bar">
    <el-input
      v-model="searchReplaceQuery"
      placeholder="输入搜索关键词"
      @keyup.enter="doSearchReplace"
    >
      <template #append>
        <el-button @click="doSearchReplace">搜索</el-button>
      </template>
    </el-input>
  </div>
  <div v-loading="searchReplaceLoading" class="search-replace-results">
    <div v-for="item in searchReplaceResults" :key="item.url" class="search-result-item">
      <div class="result-info">
        <span class="result-title">{{ item.title }}</span>
        <span class="result-source">{{ item.source }} - {{ item.platform }}</span>
      </div>
      <div class="result-actions">
        <el-button size="small" @click="handleViewShareContent(item)">查看内容</el-button>
        <el-button size="small" type="primary" @click="handleReplaceFromSearch(item)">替换</el-button>
      </div>
    </div>
    <el-empty v-if="!searchReplaceLoading && searchReplaceResults.length === 0" description="暂无搜索结果" />
  </div>
</el-dialog>

<ShareContentDialog
  v-model:visible="shareDialogVisible"
  :url="shareDialogUrl"
  :extract-code="shareDialogExtractCode"
  :title="shareDialogTitle"
  :show-replace="true"
  @replace-link="handleReplaceFromDialog"
/>
```

- [ ] **Step 5: 添加样式**

```vue
<style scoped>
.share-url-input {
  display: flex;
  width: 100%;
}

.search-replace-bar {
  margin-bottom: 16px;
}

.search-replace-results {
  min-height: 200px;
  max-height: 400px;
  overflow-y: auto;
}

.search-result-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.result-info {
  flex: 1;
  overflow: hidden;
}

.result-title {
  display: block;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.result-source {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.result-actions {
  display: flex;
  gap: 8px;
  margin-left: 16px;
}
</style>
```

- [ ] **Step 6: 验证前端编译**

Run: `cd /home/zcq/Github/clouddrive-auto-save/web && npm run build`
Expected: 构建成功

- [ ] **Step 7: 提交**

```bash
git add web/src/views/Tasks.vue
git commit -m "feat(tasks): 添加搜索替换功能"
```

---

## Task 6: 运行完整测试

- [ ] **Step 1: 运行后端测试**

Run: `cd /home/zcq/Github/clouddrive-auto-save && export PATH=$PATH:/usr/local/go/bin && go test ./... -v`
Expected: 所有测试通过

- [ ] **Step 2: 运行 lint 检查**

Run: `cd /home/zcq/Github/clouddrive-auto-save && export PATH=$PATH:/usr/local/go/bin && make check`
Expected: 所有检查通过

- [ ] **Step 3: 运行 E2E 测试**

Run: `cd /home/zcq/Github/clouddrive-auto-save && export PATH=$PATH:/usr/local/go/bin && export PLAYWRIGHT_HOST_PLATFORM_OVERRIDE=ubuntu24.04-x64 && make e2e-test`
Expected: 所有测试通过

- [ ] **Step 4: 最终提交**

```bash
git add -A
git commit -m "feat(search): 完成搜索增强功能

- 搜索结果自动校验链接有效性
- 点击展示分享链接内容弹窗
- 任务管理中搜索替换功能"
```
