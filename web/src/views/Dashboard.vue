<template>
  <div class="dashboard-container">
    <el-row :gutter="20" class="console-row">
      <!-- 左栏：统计磁贴 + 活跃任务 + 近期活动 (70%) -->
      <el-col :xs="24" :md="17" class="main-column">
        <!-- 四个统计磁贴 -->
        <el-row :gutter="12" class="stat-mini-grids">
          <el-col :span="6">
            <div class="stat-tile">
              <div class="stat-value">{{ stats.scheduled_tasks }}</div>
              <div class="stat-label">已规划任务</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-tile">
              <div class="stat-value">{{ formatSize(stats.capacity_used) }}</div>
              <div class="stat-label">已转存容量</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-tile">
              <div class="stat-value">{{ stats.today_completed }}</div>
              <div class="stat-label">今日完成</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-tile">
              <div class="stat-value">{{ stats.active_accounts }}</div>
              <div class="stat-label">活跃账号</div>
            </div>
          </el-col>
        </el-row>

        <!-- 活跃执行队列 -->
        <el-card class="section-card core-jobs-card">
          <template #header>
            <div class="card-header">
              <div class="header-title">
                <PhArrowsClockwise :size="18" weight="regular" class="spin-icon" />
                <span>活跃执行队列</span>
              </div>
              <el-tag size="small" type="primary" effect="plain">{{ runningTasks.length }} 活跃中</el-tag>
            </div>
          </template>

          <div class="monitor-scroll-area">
            <div v-if="runningTasks.length > 0" class="running-tasks-list" style="margin-bottom: 24px;">
              <div v-for="task in runningTasks" :key="task.id" class="task-progress-card">
                <div class="task-info">
                  <span class="task-name">{{ task.name }}</span>
                  <div class="task-actions">
                    <PhSpinner v-if="task.percent < 100" :size="16" class="spin-icon" />
                    <PhCheckCircle v-else-if="task.stage === 'Success'" :size="16" weight="fill" color="var(--color-success)" />
                    <PhWarningCircle v-else-if="task.stage === 'Failed'" :size="16" weight="fill" color="var(--color-danger)" />
                    <el-button v-if="task.percent === 100" type="info" link @click="dismissTask(task.id)" class="close-btn">
                      <PhX :size="14" />
                    </el-button>
                  </div>
                </div>
                <div class="task-stage">
                  <el-tag size="small" :type="getStageTagType(task.stage)" effect="plain">{{ task.stage }}</el-tag>
                  <span class="stage-msg">{{ task.message }}</span>
                </div>
                <el-progress
                  :percentage="task.percent"
                  :status="task.stage === 'Failed' ? 'exception' : (task.percent === 100 ? 'success' : '')"
                  :stroke-width="6"
                />
              </div>
            </div>

            <!-- 近期活动时间线 -->
            <div class="recent-activities-section">
              <div class="section-title-simple">近期活动</div>
              <el-timeline class="compact-timeline">
                <el-timeline-item
                  v-for="activity in stats.recent_activities"
                  :key="activity.id"
                  :timestamp="formatRelativeTime(activity.last_run)"
                  :type="getStatusType(activity.status)"
                >
                  <div class="timeline-content">
                    <span class="activity-name">{{ activity.name }}</span>
                    <span class="activity-desc">{{ getStatusText(activity.status) }}</span>
                    <el-button v-if="activity.status === 'failed'" size="small" link type="primary" @click="handleRetry(activity.id)" style="margin-left: 8px;">重试</el-button>
                  </div>
                </el-timeline-item>
              </el-timeline>
              <div v-if="stats.recent_activities.length === 0" class="monitor-empty">
                <el-empty :image-size="40" description="近期暂无转存动态" />
              </div>
            </div>
          </div>
        </el-card>

        <!-- 快捷操作栏 -->
        <div class="console-actions-bar">
          <el-button type="primary" size="default" @click="$router.push('/tasks')">创建新任务</el-button>
          <el-button type="primary" plain size="default" @click="$router.push('/accounts')">管理账号</el-button>
          <el-button type="info" plain size="default" @click="clearLogs">清理结束任务</el-button>
        </div>
      </el-col>

      <!-- 右栏：系统状态 + 日志面板 (30%) -->
      <el-col :xs="24" :md="7" class="sidebar-column">
        <!-- 系统状态卡片 -->
        <el-card class="section-card">
          <template #header>
            <div class="card-header-simple">
              <PhInfo :size="16" weight="regular" />
              <span class="panel-title">系统状态</span>
            </div>
          </template>

          <div class="system-status-body">
            <!-- CPU 负载 -->
            <div class="status-item">
              <div class="status-info">
                <span>CPU 负载</span>
                <span class="status-value">{{ cpuUsage > 0 ? cpuUsage + '%' : '--' }}</span>
                <span v-if="numCPU > 0" class="status-sub">{{ numCPU }} 核</span>
              </div>
              <el-progress
                :percentage="cpuUsage"
                :stroke-width="6"
                :show-text="false"
                color="var(--accent)"
              />
            </div>

            <!-- RAM 负载 -->
            <div class="status-item">
              <div class="status-info">
                <span>RAM 负载</span>
                <span class="status-value">{{ ramUsage > 0 ? ramUsage + '%' : '--' }}</span>
                <span v-if="ramTotalGB > 0" class="status-sub">{{ ramUsedGB.toFixed(1) }} / {{ ramTotalGB.toFixed(1) }} GB</span>
              </div>
              <el-progress
                :percentage="ramUsage"
                :stroke-width="6"
                :show-text="false"
                color="#8B5CF6"
              />
            </div>

            <!-- 存储池容量 -->
            <div class="status-item circle-progress-item">
              <div class="status-label-center">存储池容量比例</div>
              <el-progress
                type="circle"
                :percentage="Math.min(100, Math.round((stats.capacity_used / (10 * 1024 * 1024 * 1024 * 1024)) * 100))"
                :stroke-width="8"
                :width="120"
                color="var(--color-success)"
              >
                <template #default="{ percentage }">
                  <div class="progress-inner-value">{{ percentage }}%</div>
                  <div class="progress-inner-sub">已用</div>
                </template>
              </el-progress>
              <div class="storage-text-detail">
                {{ formatSize(stats.capacity_used) }} / 10 TB
              </div>
            </div>
          </div>
        </el-card>

        <!-- 系统日志面板 -->
        <el-card class="section-card log-panel">
          <template #header>
            <div class="card-header-simple">
              <span class="panel-title">系统日志</span>
              <el-button link size="small" @click="clearLogs">
                <PhTrash :size="14" weight="regular" />
              </el-button>
            </div>
          </template>
          <div class="log-list" ref="terminalRef">
            <div v-for="(log, index) in logs" :key="index" class="log-line" :class="getLogClass(log)">
              <span class="log-time">{{ new Date().toLocaleTimeString() }}</span>
              <span class="log-content">{{ log }}</span>
            </div>
            <div v-if="logs.length === 0" class="log-empty">等待系统日志流中...</div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { onMounted, onUnmounted, reactive, ref, nextTick } from 'vue'
import {
  PhInfo,
  PhArrowsClockwise,
  PhTrash,
  PhCheckCircle,
  PhWarningCircle,
  PhSpinner,
  PhX
} from '@phosphor-icons/vue'
import { getStats, clearLogsAPI } from '../api/dashboard'
import { runTask, dismissTask as runDismissTask } from '../api/task'
import { ElMessage } from 'element-plus'
import { formatSize, formatTime as formatRelativeTime } from '../utils/format'

// 系统遥测
const cpuUsage = ref(0)
const ramUsage = ref(0)
const ramUsedGB = ref(0)
const ramTotalGB = ref(0)
const numCPU = ref(0)

const activeTab = ref('schedule')
const stats = reactive({
  scheduled_tasks: 0,
  capacity_used: 0,
  today_completed: 0,
  active_accounts: 0,
  recent_activities: [],
  running_tasks_list: []
})

const logs = ref([])
const terminalRef = ref(null)
const runningTasks = ref([])
let eventSource = null
let pollTimer = null

const fetchStats = async (isPoll = false) => {
  try {
    const data = await getStats()
    Object.assign(stats, data)

    // 更新系统遥测
    if (data.sys_info) {
      cpuUsage.value = data.sys_info.cpu_percent
      ramUsage.value = data.sys_info.ram_percent
      ramUsedGB.value = data.sys_info.ram_used_gb
      ramTotalGB.value = data.sys_info.ram_total_gb
      numCPU.value = data.sys_info.num_cpu
    }

    const apiTasks = data.running_tasks_list || []

    apiTasks.forEach(task => {
      const existing = runningTasks.value.find(t => String(t.id) === String(task.id))
      if (existing) {
        if (task.percent >= existing.percent) {
          existing.name = task.name
          existing.percent = task.percent
          existing.stage = task.stage
          existing.message = task.message
        }
      } else {
        runningTasks.value.push({
          id: task.id,
          name: task.name,
          percent: task.percent,
          stage: task.stage,
          message: task.message,
          timeoutId: null
        })
      }
    })

    runningTasks.value = runningTasks.value.filter(t =>
      apiTasks.some(at => String(at.id) === String(t.id))
    )
  } catch (error) {
    console.error('获取统计数据失败:', error)
  }
}

const fetchRecentLogs = async () => {
  try {
    const response = await fetch('/api/dashboard/logs/recent')
    const data = await response.json()
    logs.value = data
    scrollToBottom()
  } catch (error) {
    console.error('获取历史日志失败:', error)
  }
}

const initSSE = () => {
  eventSource = new EventSource('/api/dashboard/logs')
  eventSource.onmessage = (event) => {
    const msg = event.data
    if (msg.includes('[PROGRESS:')) {
      handleProgressMessage(msg)
    } else if (msg.includes('[EVENT:')) {
      handleSystemEvent(msg)
    } else {
      logs.value.push(msg)
      if (logs.value.length > 200) logs.value.shift()
      scrollToBottom()
    }
  }
  eventSource.onerror = () => {
    console.error('SSE 连接异常')
  }
}

const handleSystemEvent = (msg) => {
  const match = msg.match(/\[EVENT:(.+)\]/)
  if (!match) return
  try {
    const ev = JSON.parse(match[1])
    if (ev.type === 'stats_update') {
      fetchStats(true)
    }
  } catch (e) {
    console.error('解析系统事件失败:', e)
  }
}

const handleProgressMessage = (msg) => {
  const match = msg.match(/\[PROGRESS:(.+)\]/)
  if (!match) return

  const parts = match[1].split(':')
  if (parts.length < 4) return

  const taskId = parts[0]
  const percent = parseInt(parts[1])
  const stage = parts[2]
  const info = parts.slice(3).join(':')

  const taskIdx = runningTasks.value.findIndex(t => String(t.id) === String(taskId))

  if (taskIdx > -1) {
    const task = runningTasks.value[taskIdx]
    task.percent = percent
    task.stage = stage
    task.message = info
  } else if (stage !== 'Success' && stage !== 'Failed' && stage !== 'Finished') {
    runningTasks.value.push({
      id: taskId,
      name: `任务 #${taskId}`,
      percent,
      stage,
      message: info,
      timeoutId: null
    })
    fetchStats()
  }
}

const dismissTask = async (taskId) => {
  try {
    await runDismissTask(taskId)
    const idx = runningTasks.value.findIndex(t => String(t.id) === String(taskId))
    if (idx > -1) {
      runningTasks.value.splice(idx, 1)
    }
  } catch (err) {
    console.error('忽略任务失败:', err)
  }
}

const getStageTagType = (stage) => {
  const map = {
    'Started': 'info',
    'Parsing': '',
    'Checking': 'warning',
    'Saving': 'primary',
    'Renaming': 'success',
    'Success': 'success',
    'Failed': 'danger'
  }
  return map[stage] || ''
}

const scrollToBottom = () => {
  nextTick(() => {
    if (terminalRef.value) {
      terminalRef.value.scrollTop = terminalRef.value.scrollHeight
    }
  })
}

const clearLogs = async () => {
  try {
    await clearLogsAPI()
    logs.value = []
    runningTasks.value = runningTasks.value.filter(task =>
      task.percent < 100 && task.stage !== 'Success' && task.stage !== 'Failed'
    )
    ElMessage.success('日志与已完成任务已清空')
  } catch (err) {
    console.error('清空日志失败:', err)
  }
}

const handleRetry = async (taskId) => {
  try {
    await runTask(taskId)
    ElMessage.success('已发起重试')
  } catch (err) {
    console.error(err)
  }
}

const getLogClass = (log) => {
  if (log.includes('ERROR')) return 'log-error'
  if (log.includes('WARN')) return 'log-warn'
  if (log.includes('SUCCESS')) return 'log-success'
  return ''
}



const getStatusType = (status) => {
  const types = {
    'success': 'success',
    'failed': 'danger',
    'running': 'primary'
  }
  return types[status] || 'info'
}

const getStatusText = (status) => {
  const texts = {
    'success': '转存成功',
    'failed': '转存失败',
    'running': '正在执行'
  }
  return texts[status] || '已准备'
}

onMounted(() => {
  fetchStats()
  initSSE()
  fetchRecentLogs()

  pollTimer = setInterval(() => {
    fetchStats(true)
  }, 5000)

})

onUnmounted(() => {
  if (eventSource) eventSource.close()
  if (pollTimer) clearInterval(pollTimer)
})
</script>

<style scoped>
.dashboard-container {
  padding: 0;
}

.console-row {
  display: flex;
  flex-wrap: wrap;
}

/* 卡片通用样式 */
.section-card {
  margin-bottom: 16px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header-simple {
  display: flex;
  align-items: center;
  gap: 8px;
}

.panel-title {
  font-size: 13px;
  font-weight: 700;
  color: var(--text-secondary);
  letter-spacing: 0.1em;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 700;
  color: var(--text-primary);
}

/* 旋转动画 */
.spin-icon {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* 统计磁贴 */
.stat-mini-grids {
  margin-bottom: 16px;
}

.stat-tile {
  background: var(--surface-bg);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md, 8px);
  padding: 16px;
  box-shadow: var(--shadow-sm);
  text-align: center;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  line-height: 1.2;
  color: var(--text-primary);
}

.stat-label {
  font-size: 13px;
  color: var(--text-muted);
  margin-top: 4px;
}

/* 活跃任务卡片 */
.core-jobs-card {
  min-height: 400px;
}

.monitor-scroll-area {
  max-height: 380px;
  overflow-y: auto;
  padding-right: 4px;
}

.task-progress-card {
  background: var(--surface-bg);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md, 8px);
  padding: 16px;
  margin-bottom: 12px;
}

.task-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.task-name {
  font-weight: 700;
  font-size: 1rem;
  color: var(--text-primary);
}

.task-stage {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.stage-msg {
  font-size: 0.85rem;
  color: var(--text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
}

.section-title-simple {
  margin-bottom: 12px;
  font-size: 12px;
  color: var(--text-muted);
  font-weight: 600;
  letter-spacing: 0.05em;
}

.compact-timeline {
  padding-left: 8px;
}

.timeline-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.activity-name {
  font-weight: 600;
  font-size: 0.9rem;
  color: var(--text-primary);
}

.activity-desc {
  font-size: 0.8rem;
  color: var(--text-muted);
}

.console-actions-bar {
  display: flex;
  gap: 12px;
  margin-top: 16px;
  padding: 12px 16px;
  border-radius: var(--radius-md, 8px);
  justify-content: flex-end;
  background: var(--surface-bg);
  border: 1px solid var(--border-color);
  box-shadow: var(--shadow-sm);
}

/* 右栏：系统状态 */
.sidebar-column {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.system-status-body {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.status-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.status-info {
  display: flex;
  justify-content: space-between;
  font-size: 0.85rem;
  color: var(--text-secondary);
  font-weight: 600;
}

.status-value {
  color: var(--accent);
  font-family: var(--font-mono, monospace);
}

.status-sub {
  font-size: 0.75rem;
  color: var(--text-secondary);
  margin-left: 4px;
  font-family: var(--font-mono, monospace);
}

.circle-progress-item {
  align-items: center;
  margin-top: 8px;
  gap: 12px;
}

.status-label-center {
  font-size: 0.85rem;
  color: var(--text-secondary);
  font-weight: 600;
}

.progress-inner-value {
  font-size: 1.6rem;
  font-weight: 800;
  color: var(--text-primary);
  line-height: 1;
}

.progress-inner-sub {
  font-size: 0.75rem;
  color: var(--text-muted);
  margin-top: 4px;
}

.storage-text-detail {
  font-size: 0.8rem;
  color: var(--text-secondary);
  font-weight: 600;
}

/* 日志面板 */
.log-panel :deep(.el-card__body) {
  padding: 0;
}

.log-list {
  max-height: 400px;
  overflow-y: auto;
  font-family: var(--font-mono);
  font-size: 13px;
}

.log-line {
  padding: 6px 12px;
  border-radius: var(--radius-sm, 4px);
  margin-bottom: 4px;
  line-height: 1.5;
}

.log-time {
  color: var(--text-muted);
  margin-right: 8px;
}

.log-content {
  color: var(--text-secondary);
}

.log-error {
  background: #FEF2F2;
  color: #991B1B;
}

.log-warn {
  background: #FFFBEB;
  color: #92400E;
}

.log-success {
  background: #F0FDF4;
  color: #166534;
}

.log-info {
  color: var(--text-secondary);
}

.log-empty {
  text-align: center;
  color: var(--text-muted);
  padding: 2rem 0;
}

.monitor-empty {
  padding-top: 10px;
}
</style>
