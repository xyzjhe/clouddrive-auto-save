<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Loading } from '@element-plus/icons-vue'
import { Search as SearchIcon, Link as LinkIcon, Clock as ClockIcon, FileText as FileTextIcon } from 'lucide-vue-next'
import { searchResources, listSearchSources, validateLink } from '../api/search'
import ShareContentDialog from '../components/ShareContentDialog.vue'

const router = useRouter()
const query = ref('')
const sources = ref([])
const selectedSources = ref([])
const results = ref([])
const loading = ref(false)
const page = ref(1)
const validating = ref(false)
const validatedCount = ref(0)
const totalCount = ref(0)

// 分页
const currentPage = ref(1)
const pageSize = ref(20)
const totalResults = computed(() => results.value.length)
const paginatedResults = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return results.value.slice(start, start + pageSize.value)
})
const handlePageChange = (newPage) => {
  currentPage.value = newPage
}
const handleSizeChange = (newSize) => {
  pageSize.value = newSize
  currentPage.value = 1
}

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

// 批量校验链接有效性，渐进式展示有效结果
// 使用并发限流（避免一次性发起几十个请求）+ 单请求超时（5s）+ 增量渲染
const validateLinks = async (items) => {
  validating.value = true
  validatedCount.value = 0
  totalCount.value = items.length
  // 保持原顺序的存放区
  const slotMap = new Map()
  let displayedCount = 0

  // 并发限流：最多同时 6 个请求，避免浏览器/服务端连接饱和
  const CONCURRENCY = 6
  let cursor = 0

  const runOne = async (idx) => {
    const item = items[idx]
    try {
      const res = await validateLink(item.url)
      if (res.valid) {
        slotMap.set(idx, item)
      }
    } catch {
      // 超时或失败，跳过
    }
    validatedCount.value++

    // 增量更新：按原顺序补齐到 results 中（保证显示顺序与搜索顺序一致）
    while (slotMap.has(displayedCount)) {
      results.value.push(slotMap.get(displayedCount))
      slotMap.delete(displayedCount)
      displayedCount++
    }
  }

  const workers = Array.from({ length: CONCURRENCY }, async () => {
    while (cursor < items.length) {
      const idx = cursor++
      await runOne(idx)
    }
  })
  await Promise.allSettled(workers)
  validating.value = false
}

const handleSearch = async () => {
  if (!query.value.trim()) {
    ElMessage.warning('请输入搜索关键词')
    return
  }

  loading.value = true
  currentPage.value = 1
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
    const items = data.items || []

    // 先校验链接有效性，验证完成后再展示过滤后的结果
    if (items.length > 0) {
      await validateLinks(items)
    } else {
      results.value = []
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

// 分享内容弹窗
const shareDialogVisible = ref(false)
const shareDialogUrl = ref('')
const shareDialogExtractCode = ref('')
const shareDialogTitle = ref('')

const handleResultClick = (item) => {
  shareDialogUrl.value = item.url
  shareDialogExtractCode.value = ''
  shareDialogTitle.value = item.title
  shareDialogVisible.value = true
}

const handleCreateTaskFromDialog = (data) => {
  shareDialogVisible.value = false
  const query = {
    share_url: data.url,
    extract_code: data.extractCode
  }
  // 若在子目录中创建任务，一并带上 parent_id（139 平台作为 share_parent_id）
  if (data.parentId) {
    query.share_parent_id = data.parentId
  }
  router.push({ name: 'Tasks', query })
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

    <div
      v-loading="loading"
      class="search-results"
    >
      <div
        v-for="item in paginatedResults"
        :key="item.url"
        class="result-item clickable"
        @click="handleResultClick(item)"
      >
        <div class="result-header">
          <div class="result-title">
            <span v-if="item.valid === true" class="valid-icon">✅</span>
            <span v-else-if="item.valid === false" class="valid-icon invalid" :title="item.validMessage">❌</span>
            <span v-else-if="validating" class="valid-icon">⏳</span>
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
          <span v-if="item.size" class="meta-item">
            <el-icon><FileTextIcon /></el-icon>
            {{ item.size }}
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
        v-if="!loading && totalResults === 0 && query && !validating"
        description="未找到相关资源"
      />

      <div v-if="validating && totalCount > 0" class="validation-progress">
        <el-icon class="is-loading"><Loading /></el-icon>
        <span>正在校验链接有效性（{{ validatedCount }} / {{ totalCount }}）</span>
      </div>

      <div v-if="totalResults > 0" class="pagination-wrapper">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :total="totalResults"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="handlePageChange"
          @size-change="handleSizeChange"
        />
      </div>
    </div>

    <ShareContentDialog
      v-model:visible="shareDialogVisible"
      :url="shareDialogUrl"
      :extract-code="shareDialogExtractCode"
      :title="shareDialogTitle"
      :show-replace="false"
      @create-task="handleCreateTaskFromDialog"
    />
  </div>
</template>

<style scoped>
.search-page {
  /* ... styles ... */
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.title-section h2 {
  margin: 0;
  font-size: 26px;
  font-weight: 800;
  color: var(--text-primary);
  letter-spacing: -0.02em;
}

.title-section p {
  color: var(--neutral-500);
  margin: 4px 0 0 0;
  font-size: 15px;
}

.search-bar {
  margin-bottom: 1.5rem;
}

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

.result-tags {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  margin-bottom: 0.5rem;
}

.tag-item {
  margin: 0;
}

.result-summary {
  color: var(--text-secondary);
  font-size: 0.9rem;
  line-height: 1.5;
}

.valid-icon {
  margin-right: 4px;
  font-size: 14px;
}

.valid-icon.invalid {
  cursor: help;
}

.result-item.clickable {
  cursor: pointer;
  transition: box-shadow 0.2s;
}

.result-item.clickable:hover {
  box-shadow: var(--shadow-md);
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 1.5rem;
  padding: 1rem 0;
}

.validation-progress {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.75rem;
  margin-top: 0.5rem;
  color: var(--text-secondary);
  font-size: 0.875rem;
}

.validation-progress .el-icon {
  color: var(--brand-500);
}
</style>
