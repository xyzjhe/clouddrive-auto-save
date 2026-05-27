<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search as SearchIcon, Link as LinkIcon, Clock as ClockIcon, FileText as FileTextIcon } from 'lucide-vue-next'
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
  color: var(--neutral-800);
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
</style>
