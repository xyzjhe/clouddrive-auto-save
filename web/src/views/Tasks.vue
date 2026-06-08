<template>
  <div class="tasks-container">
    <div class="page-header">
      <div class="title-section">
        <h2>任务管理</h2>
        <p>监控并自动转存移动云盘和夸克网盘的分享资源</p>
      </div>
      <div class="header-actions">
        <el-radio-group v-model="viewMode" size="default" class="view-toggle" @change="toggleViewMode">
          <el-radio-button label="table">
            <el-icon><PhList /></el-icon>
          </el-radio-button>
          <el-radio-button label="card">
            <el-icon><PhGridFour /></el-icon>
          </el-radio-button>
        </el-radio-group>
        <el-popconfirm
          title="确定要一键启动所有可运行的任务吗？"
          confirm-button-text="确认"
          cancel-button-text="取消"
          width="240"
          @confirm="handleRunAll"
        >
          <template #reference>
            <el-button type="primary" plain :icon="PhPlay" :loading="runningAll">全部运行</el-button>
          </template>
        </el-popconfirm>
        <el-button type="primary" :icon="PhPlus" @click="openAddDialog">创建任务</el-button>
      </div>
    </div>

    <!-- 表格/卡片视图 -->
    <TaskTable
      :task-list="taskList"
      :loading="loading"
      :global-schedule="globalSchedule"
      :view-mode="viewMode"
      @run="handleRun"
      @edit="handleEdit"
      @delete="handleDelete"
      @add="openAddDialog"
    />

    <!-- 表单抽屉及所有子弹窗 -->
    <TaskForm
      ref="taskFormRef"
      :accounts="accounts"
      :submitting="submitting"
      :preview-loading="previewLoading"
      @submit="submitForm"
      @open-external="openExternalLink"
      @reset-share-root="resetToShareRoot"
      @preview="handlePreview"
      @view-share-content="handleViewShareContent"
    />
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import {
  PhPlus, PhPlay, PhList, PhGridFour
} from '@phosphor-icons/vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getTasks, createTask, updateTask, deleteTask, runTask, runAllTasks, previewTask, getScheduleSettings } from '../api/task'
import { getAccounts } from '../api/account'
import { useSSEStore } from '@/stores/sse'
import TaskTable from '../components/tasks/TaskTable.vue'
import TaskForm from '../components/tasks/TaskForm.vue'
import { getDefaultFormData } from '../components/tasks/utils'

const route = useRoute()
const sseStore = useSSEStore()

// ---- 核心状态 ----
const taskList = ref([])
const accounts = ref([])
const loading = ref(false)
const runningAll = ref(false)
const submitting = ref(false)
const previewLoading = ref(false)
const viewMode = ref(localStorage.getItem('taskViewMode') || 'table')

// ---- 表单子组件引用 ----
const taskFormRef = ref(null)

// ---- 全局调度设置 ----
const globalSchedule = ref({ enabled: false, cron: '' })

const fetchGlobalSettings = async () => {
  try {
    const data = await getScheduleSettings()
    globalSchedule.value = data
  } catch (err) {
    console.error('获取全局设置失败:', err)
  }
}

// ---- 视图切换 ----
const toggleViewMode = (mode) => {
  viewMode.value = mode
  localStorage.setItem('taskViewMode', mode)
}

// ---- 数据获取 ----
const fetchList = async (silent = false) => {
  if (!silent) loading.value = true
  try {
    const [taskData, accountData] = await Promise.all([getTasks(), getAccounts()])
    taskList.value = taskData
    accounts.value = accountData
  } catch (err) {
    console.error(err)
  } finally {
    if (!silent) loading.value = false
  }
}

// ---- 在新标签页打开链接 ----
const openExternalLink = async (url, extractCode) => {
  if (!url) return
  let finalUrl = url
  if (extractCode) {
    try {
      const urlObj = new URL(url)
      urlObj.searchParams.set('pwd', extractCode)
      finalUrl = urlObj.toString()
    } catch (e) {
      finalUrl = url.includes('?')
        ? `${url}&pwd=${extractCode}`
        : `${url}?pwd=${extractCode}`
    }

    try {
      await navigator.clipboard.writeText(extractCode)
      ElMessage.success(`提取码 ${extractCode} 已复制，请在新页面粘贴`)
    } catch (err) {
      console.warn('复制提取码失败:', err)
    }
  }
  window.open(finalUrl, '_blank')
}

// ---- 重置为根目录 ----
const resetToShareRoot = () => {
  const form = taskFormRef.value?.getFormData()
  if (!form) return
  const account = accounts.value.find(acc => acc.id === form.account_id)
  if (!account) return

  if (account.platform === 'quark') {
    const match = form.share_url.match(/\/s\/([^#]+)/)
    if (match) {
      form.share_url = `https://pan.quark.cn/s/${match[1]}#/list/share/0`
    }
  } else {
    form.share_parent_id = ''
  }
  form.start_file_id = ''
  form.start_file_name = ''
  ElMessage.success('已重置为根目录')
}

// ---- 预览 ----
const handlePreview = async () => {
  const form = taskFormRef.value?.getFormData()
  if (!form?.account_id || !form?.share_url) {
    return ElMessage.warning('请先填写执行账号和分享链接')
  }

  previewLoading.value = true
  try {
    const data = await previewTask(form)
    taskFormRef.value?.setPreviewData(data)
  } catch (err) {
    console.error(err)
  } finally {
    previewLoading.value = false
  }
}

// ---- 查看分享内容弹窗 ----
const handleViewShareContent = (item) => {
  taskFormRef.value?.openShareContentDialog(item)
}

// ---- 打开新建对话框 ----
const openAddDialog = () => {
  taskFormRef.value?.open(getDefaultFormData())
}

// ---- 编辑任务 ----
const handleEdit = (row) => {
  // TaskCard 传递的是 id，需要找到完整对象
  const taskRow = (typeof row === 'object' && row.id) ? row : taskList.value.find(t => t.id === row)
  if (!taskRow) return

  const formData = {
    id: taskRow.id,
    name: taskRow.name,
    account_id: taskRow.account_id,
    share_url: taskRow.share_url,
    extract_code: taskRow.extract_code,
    save_path: taskRow.save_path,
    pattern: taskRow.pattern,
    replacement: taskRow.replacement,
    start_file_id: taskRow.start_file_id,
    start_file_name: taskRow.start_file_name,
    share_parent_id: taskRow.share_parent_id || '',
    cron: taskRow.cron,
    schedule_mode: taskRow.schedule_mode || 'global',
    max_retries: taskRow.max_retries ?? 3,
    ignore_extension: taskRow.ignore_extension ?? false
  }

  let startFileName = ''
  if (taskRow.start_file_id) {
    startFileName = taskRow.start_file_name || `ID: ${taskRow.start_file_id} (文件名未记录)`
  }

  let dirName = ''
  if (taskRow.share_parent_id) {
    dirName = '已选子目录'
  } else {
    const match = (taskRow.share_url || '').match(/\/s\/(\w+)#\/list\/share\/(\w+)/)
    dirName = (match && match[2] && match[2] !== '0') ? '已选子目录' : ''
  }

  taskFormRef.value?.openEdit(formData, startFileName, dirName)
}

// ---- 提交表单 ----
const submitForm = async () => {
  const form = taskFormRef.value?.getFormData()
  if (!form?.name || !form?.account_id || !form?.share_url) {
    return ElMessage.warning('请填写必要的信息')
  }

  submitting.value = true
  try {
    if (form.id) {
      const updatedTask = await updateTask(form.id, form)
      const idx = taskList.value.findIndex(t => t.id === form.id)
      if (idx > -1) {
        Object.assign(taskList.value[idx], updatedTask)
      }
      ElMessage.success('任务更新成功')
    } else {
      await createTask(form)
      ElMessage.success('任务保存成功')
      fetchList()
    }
    // 关闭抽屉
    taskFormRef.value?.close()
  } catch (err) {
    console.error(err)
    ElMessage.error('保存失败: ' + (err.response?.data?.error || err.message || '未知错误'))
  } finally {
    submitting.value = false
  }
}

// ---- 运行任务 ----
const handleRun = async (row) => {
  const taskId = (typeof row === 'object' && row.id) ? row.id : row
  const taskRow = (typeof row === 'object' && row.id) ? row : taskList.value.find(t => t.id === row)
  try {
    await runTask(taskId)
    if (taskRow && taskRow.status !== 'success') {
      taskRow.status = 'running'
    }
    ElMessage.success('任务已提交执行队列')
  } catch (err) {
    // 错误已由拦截器展示
  }
}

// ---- 全部运行 ----
const handleRunAll = async () => {
  runningAll.value = true
  try {
    const res = await runAllTasks()
    ElMessage.success(res.message || `批量执行已开启，已成功触发 ${res.count} 个任务`)

    taskList.value.forEach(task => {
      const isFatal = task.message && task.message.includes('[Fatal]')
      if (task.status !== 'running' && task.status !== 'success' && !isFatal) {
        task.status = 'running'
      }
    })
  } catch (err) {
    // 错误已由拦截器展示
  } finally {
    runningAll.value = false
  }
}

// ---- 删除任务 ----
const handleDelete = (row) => {
  const taskId = (typeof row === 'object' && row.id) ? row.id : row
  ElMessageBox.confirm('确定要删除此转存任务吗？', '确认', {
    type: 'warning'
  }).then(async () => {
    await deleteTask(taskId)
    ElMessage.success('任务已删除')
    fetchList()
  }).catch(() => {})
}

// ---- SSE 事件处理 ----
let unsubSSE = null
let offTaskUpdate = null
let offTaskDelete = null

onMounted(async () => {
  await fetchList()
  fetchGlobalSettings()

  unsubSSE = sseStore.subscribe()

  offTaskUpdate = sseStore.on('task_update', (ev) => {
    if (!ev.task) return
    const task = ev.task
    const idx = taskList.value.findIndex(t => t.id === task.id)
    if (idx > -1) {
      const row = taskList.value[idx]
      row.status = task.status
      row.message = task.message
      row.last_run = task.last_run
      row.percent = task.percent
      row.stage = task.stage
      if (task.max_retries !== undefined) row.max_retries = task.max_retries
      if (task.ignore_extension !== undefined) row.ignore_extension = task.ignore_extension

      if (task.status === 'success' || task.status === 'failed') {
        fetchList(true)
      }
    } else {
      fetchList()
    }
  })

  offTaskDelete = sseStore.on('task_delete', (ev) => {
    taskList.value = taskList.value.filter(t => t.id !== ev.taskId)
  })

  // 从搜索页跳转创建任务
  if (route.query.share_url) {
    openAddDialog()
    const form = taskFormRef.value?.getFormData()
    if (form) {
      form.share_url = route.query.share_url
      if (route.query.title) {
        form.name = `转存-${route.query.title}`
      }
      if (route.query.platform) {
        const match = accounts.value.find(a => a.platform === route.query.platform)
        if (match) {
          form.account_id = match.id
        }
      }
      if (route.query.share_parent_id) {
        form.share_parent_id = route.query.share_parent_id
      }
    }
  }

  // 全局快捷键
  document.addEventListener('keydown', handleKeydown)
  window.addEventListener('beforeunload', handleBeforeUnload)
})

onUnmounted(() => {
  if (offTaskDelete) offTaskDelete()
  if (offTaskUpdate) offTaskUpdate()
  if (unsubSSE) unsubSSE()
  document.removeEventListener('keydown', handleKeydown)
  window.removeEventListener('beforeunload', handleBeforeUnload)
})

const handleKeydown = (e) => {
  if ((e.ctrlKey || e.metaKey) && e.key === 's') {
    e.preventDefault()
    if (taskFormRef.value?.isOpen()) {
      submitForm()
    }
  }
  if ((e.ctrlKey || e.metaKey) && e.key === 'r') {
    e.preventDefault()
    if (!taskFormRef.value?.isOpen()) {
      handleRunAll()
    }
  }
}

const handleBeforeUnload = (e) => {
  if (taskFormRef.value?.isOpen()) {
    e.preventDefault()
    e.returnValue = ''
  }
}
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 16px;
}

.header-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.title-section h2 {
  margin: 0;
  font-size: 24px;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: -0.02em;
}

.title-section p {
  color: var(--neutral-500);
  margin: 4px 0 0 0;
  font-size: 15px;
}

.view-toggle :deep(.el-radio-button__inner) {
  padding: 8px 12px;
  display: flex;
  align-items: center;
}
</style>
