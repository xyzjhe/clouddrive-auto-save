<script setup>
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { parseShareLink } from '../api/task'
import { formatSize } from '../utils/format'
import { PhFolder, PhFile, PhCaretRight, PhArrowLeft } from '@phosphor-icons/vue'

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
const fileListRef = ref(null)
// 面包屑：[{ id: '', name: '根目录' }, { id: 'xxx', name: '子目录' }]
const breadcrumbs = ref([])

watch(() => props.visible, async (val) => {
  if (val && props.url) {
    breadcrumbs.value = []
    await loadFiles('')
  } else {
    files.value = []
    breadcrumbs.value = []
  }
})

const currentParentId = () => {
  if (breadcrumbs.value.length === 0) return ''
  return breadcrumbs.value[breadcrumbs.value.length - 1].id
}

const loadFiles = async (parentId) => {
  loading.value = true
  // 滚动到列表顶部，确保 loading 转圈可见
  if (fileListRef.value) {
    fileListRef.value.scrollTop = 0
  }
  try {
    const res = await parseShareLink({
      share_url: props.url,
      extract_code: props.extractCode || '',
      parent_id: parentId || ''
    })
    files.value = Array.isArray(res) ? res : (res.items || [])
  } catch (e) {
    ElMessage.error('获取分享内容失败')
    files.value = []
  } finally {
    loading.value = false
  }
}

// 进入子目录（防重入）
const enterFolder = async (folder) => {
  if (loading.value) return
  breadcrumbs.value.push({ id: folder.id, name: folder.name })
  await loadFiles(folder.id)
}

// 点击面包屑跳转（防重入）
const navigateToBreadcrumb = async (index) => {
  if (loading.value) return
  // index = -1 表示根目录
  if (index === -1) {
    breadcrumbs.value = []
    await loadFiles('')
  } else {
    breadcrumbs.value = breadcrumbs.value.slice(0, index + 1)
    const target = breadcrumbs.value[breadcrumbs.value.length - 1]
    await loadFiles(target.id)
  }
}

// 返回上一级（防重入）
const goUp = async () => {
  if (loading.value) return
  if (breadcrumbs.value.length === 0) return
  breadcrumbs.value.pop()
  await loadFiles(currentParentId())
}

const handleCreateTask = () => {
  // 把当前所在目录 ID 一并传给任务创建页（可作为 share_parent_id）
  emit('create-task', {
    url: props.url,
    extractCode: props.extractCode,
    parentId: currentParentId()
  })
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
    width="640px"
  >
    <!-- 面包屑导航 -->
    <div v-if="breadcrumbs.length > 0" class="dialog-breadcrumb">
      <el-button link :icon="PhArrowLeft" size="small" @click="goUp">上一级</el-button>
      <el-divider direction="vertical" />
      <el-breadcrumb separator-icon="PhCaretRight">
        <el-breadcrumb-item>
          <a href="#" @click.prevent="navigateToBreadcrumb(-1)">根目录</a>
        </el-breadcrumb-item>
        <el-breadcrumb-item v-for="(crumb, idx) in breadcrumbs" :key="crumb.id">
          <a v-if="idx < breadcrumbs.length - 1" href="#" @click.prevent="navigateToBreadcrumb(idx)">{{ crumb.name }}</a>
          <span v-else>{{ crumb.name }}</span>
        </el-breadcrumb-item>
      </el-breadcrumb>
    </div>

    <div ref="fileListRef" v-loading="loading" class="file-list">
      <div v-if="files.length === 0 && !loading" class="empty-tip">
        暂无文件信息
      </div>
      <div
        v-for="file in files"
        :key="file.id"
        class="file-item"
        :class="{ 'is-folder': file.is_folder, 'is-disabled': loading }"
        @click="file.is_folder && !loading && enterFolder(file)"
      >
        <component
          :is="file.is_folder ? PhFolder : PhFile"
          class="file-icon"
          :size="16"
        />
        <span class="file-name">{{ file.name }}</span>
        <span v-if="!file.is_folder" class="file-size">{{ formatSize(file.size) }}</span>
        <PhCaretRight v-if="file.is_folder" class="enter-icon" :size="16" />
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
.dialog-breadcrumb {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  padding-bottom: 10px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.file-list {
  min-height: 240px;
  max-height: 480px;
  overflow-y: auto;
}

.file-item {
  display: flex;
  align-items: center;
  padding: 10px 8px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  cursor: default;
  transition: background-color 0.15s;
}

.file-item.is-folder {
  cursor: pointer;
}

.file-item.is-folder:hover {
  background-color: var(--hover-bg);
}

.file-item.is-disabled {
  pointer-events: none;
  opacity: 0.5;
}

.file-icon {
  margin-right: 10px;
  color: var(--text-muted);
  flex-shrink: 0;
}

.file-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-size {
  color: var(--text-muted);
  font-size: 12px;
  margin-left: 8px;
  font-family: var(--font-mono, monospace);
}

.enter-icon {
  margin-left: 8px;
  color: var(--text-muted);
}

.empty-tip {
  text-align: center;
  color: var(--text-muted);
  padding: 40px 0;
}
</style>
