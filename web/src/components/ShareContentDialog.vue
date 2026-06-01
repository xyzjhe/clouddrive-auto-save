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
    // 后端直接返回数组，不是 { items: [...] } 格式
    files.value = Array.isArray(res) ? res : (res.items || [])
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
