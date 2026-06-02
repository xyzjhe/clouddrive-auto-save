<template>
  <div class="dashboard-container">
    <el-row :gutter="20" class="console-row">
      <!-- 1. 左栏：系统遥测 (Telemetry) -->
      <el-col :xs="24" :md="6" class="telemetry-column">
        <el-card class="telemetry-card glass-card">
          <template #header>
            <div class="card-header-simple">
              <span class="pulse-dot"></span>
              <span class="panel-title">SYSTEM TELEMETRY</span>
            </div>
          </template>

          <div class="telemetry-body">
            <!-- CPU 负载 -->
            <div class="telemetry-item">
              <div class="telemetry-info">
                <span>CPU 负载</span>
                <span class="value-highlight">{{ cpuUsage > 0 ? cpuUsage + '%' : '--' }}</span>
                <span v-if="numCPU > 0" class="value-sub">{{ numCPU }} 核</span>
              </div>
              <el-progress
                :percentage="cpuUsage"
                :stroke-width="6"
                :show-text="false"
                color="var(--neon-teal)"
              />
            </div>

            <!-- RAM 负载 -->
            <div class="telemetry-item">
              <div class="telemetry-info">
                <span>RAM 负载</span>
                <span class="value-highlight">{{ ramUsage > 0 ? ramUsage + '%' : '--' }}</span>
                <span v-if="ramTotalGB > 0" class="value-sub">{{ ramUsedGB.toFixed(1) }} / {{ ramTotalGB.toFixed(1) }} GB</span>
              </div>
              <el-progress
                :percentage="ramUsage"
                :stroke-width="6"
                :show-text="false"
                color="var(--neon-purple)"
              />
            </div>

            <!-- 存储池配额进度圆环 -->
            <div class="telemetry-item circle-progress-item">
              <div class="telemetry-label-center">存储池容量比例</div>
              <el-progress 
                type="circle" 
                :percentage="Math.min(100, Math.round((stats.capacity_used / (10 * 1024 * 1024 * 1024 * 1024)) * 100))" 
                :stroke-width="8"
                :width="120"
                color="var(--neon-green)"
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

            <!-- 自动转存健康状态 -->
            <div class="telemetry-status-box breath-glow">
              <div class="status-indicator">
                <span class="status-indicator-dot"></span>
                <span>AUTO-SAVE ACTIVE</span>
              </div>
              <div class="status-sub">引擎正常监听中</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 2. 中栏：控制台核心任务与进度 (Console Core) -->
      <el-col :xs="24" :md="isLogCollapsed ? 17 : 12" class="console-core-column">
        <!-- 四个精简指标汇总磁贴 -->
        <el-row :gutter="12" class="stat-mini-grids">
          <el-col :span="6">
            <div class="mini-tile glass-card">
              <div class="tile-label">已规划任务</div>
              <div class="tile-value cyan">{{ stats.scheduled_tasks }}</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="mini-tile glass-card">
              <div class="tile-label">已转存容量</div>
              <div class="tile-value purple">{{ formatSize(stats.capacity_used) }}</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="mini-tile glass-card">
              <div class="tile-label">今日完成</div>
              <div class="tile-value green">{{ stats.today_completed }}</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="mini-tile glass-card">
              <div class="tile-label">活跃账号</div>
              <div class="tile-value orange">{{ stats.active_accounts }}</div>
            </div>
          </el-col>
        </el-row>

        <!-- 活跃转存卡片区 -->
        <el-card class="core-jobs-card glass-card">
          <template #header>
            <div class="card-header">
              <div class="header-title">
                <el-icon class="icon-spin-slow"><RefreshCw /></el-icon>
                <span>活跃执行队列</span>
              </div>
              <el-tag size="small" type="primary" effect="dark">{{ runningTasks.length }} 活跃中</el-tag>
            </div>
          </template>

          <div class="monitor-scroll-area">
            <div v-if="runningTasks.length > 0" class="running-tasks-list" style="margin-bottom: 24px;">
              <div v-for="task in runningTasks" :key="task.id" class="task-progress-card breath-glow">
                <div class="task-info">
                  <span class="task-name">{{ task.name }}</span>
                  <div class="task-actions">
                    <el-icon v-if="task.percent < 100" class="is-loading"><Loader2 /></el-icon>
                    <el-icon v-else-if="task.stage === 'Success'" color="var(--neon-green)"><CheckCircle2 /></el-icon>
                    <el-icon v-else-if="task.stage === 'Failed'" color="var(--color-danger)"><AlertCircle /></el-icon>
                    <el-button v-if="task.percent === 100" type="info" link @click="dismissTask(task.id)" class="close-btn">
                      <el-icon><X /></el-icon>
                    </el-button>
                  </div>
                </div>
                <div class="task-stage">
                  <el-tag size="small" :type="getStageTagType(task.stage)" effect="dark">{{ task.stage }}</el-tag>
                  <span class="stage-msg">{{ task.message }}</span>
                </div>
                <el-progress
                  :percentage="task.percent"
                  :status="task.stage === 'Failed' ? 'exception' : (task.percent === 100 ? 'success' : '')"
                  :stroke-width="6"
                  striped
                  :striped-flow="task.percent < 100"
                />
              </div>
            </div>

            <!-- 近期时间线，始终渲染以展示历史记录 -->
            <div class="recent-activities-section">
              <div class="section-title-simple" style="margin-bottom: 12px; font-size: 12px; color: var(--text-muted); font-weight: 600; letter-spacing: 0.05em;">RECENT ACTIVITIES</div>
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

        <!-- 浮动快捷悬浮球整合到控制台底部 -->
        <div class="console-actions-bar glass-card">
          <el-button type="primary" size="default" @click="$router.push('/tasks')">创建新任务</el-button>
          <el-button type="primary" plain size="default" @click="$router.push('/accounts')">管理账号</el-button>
          <el-button type="info" plain size="default" @click="clearLogs">清理结束任务</el-button>
        </div>
      </el-col>

      <!-- 3. 右栏：日志流 (Terminal Log) -->
      <el-col :xs="24" :md="isLogCollapsed ? 1 : 6" class="terminal-column">
        <el-card class="terminal-main-card glass-card" body-style="padding: 0">
          <template #header>
            <div class="card-header">
              <div class="header-title" v-if="!isLogCollapsed">
                <el-icon><Terminal /></el-icon>
                <span>TERMINAL LOG</span>
              </div>
              <div class="header-actions">
                <el-button link type="primary" :icon="isLogCollapsed ? Terminal : X" @click="isLogCollapsed = !isLogCollapsed" />
                <el-button v-if="!isLogCollapsed" link type="danger" :icon="Trash2" @click="clearLogs" />
              </div>
            </div>
          </template>

          <div v-if="!isLogCollapsed" class="terminal-window" ref="terminalRef">
            <div v-for="(log, index) in logs" :key="index" class="log-line" :class="getLogClass(log)">
              <span class="log-timestamp">{{ new Date().toLocaleTimeString() }}</span>
              <span class="log-content">{{ log }}</span>
            </div>
            <div v-if="logs.length === 0" class="terminal-empty">
              等待系统日志流中...<span class="terminal-cursor">_</span>
            </div>
          </div>
          <div v-else class="terminal-collapsed-indicator" @click="isLogCollapsed = false">
            <div class="indicator-text">展开日志面板</div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { onMounted, onUnmounted, reactive, ref, nextTick } from 'vue'
import {
  Calendar,
  Info,
  Scan,
  RefreshCw,
  Terminal,
  Trash2,
  CheckCircle2,
  AlertCircle,
  Loader2,
  X,
  Bell
} from 'lucide-vue-next'
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
const isLogCollapsed = ref(false)
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

.welcome-section {
  margin-bottom: 24px;
}

.welcome-section h2 {
  font-size: 26px;
  font-weight: 800;
  margin-bottom: 6px;
  color: var(--text-primary);
  letter-spacing: -0.02em;
}

.welcome-section p {
  color: var(--text-secondary);
  font-size: 15px;
}

/* 1. 左栏：系统遥测样式 */
.telemetry-card {
  height: 100%;
}

.card-header-simple {
  display: flex;
  align-items: center;
  gap: 8px;
}

.pulse-dot {
  width: 8px;
  height: 8px;
  background-color: var(--neon-teal);
  border-radius: 50%;
  box-shadow: 0 0 8px var(--neon-teal);
  animation: blink 2s infinite ease-in-out;
}

@keyframes blink {
  0%, 100% { opacity: 0.3; }
  50% { opacity: 1; }
}

.panel-title {
  font-size: 13px;
  font-weight: 700;
  color: var(--text-secondary);
  letter-spacing: 0.1em;
}

.telemetry-body {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.telemetry-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.telemetry-info {
  display: flex;
  justify-content: space-between;
  font-size: 0.85rem;
  color: var(--text-secondary);
  font-weight: 600;
}

.value-highlight {
  color: var(--neon-teal);
}

.value-sub {
  font-size: 0.75rem;
  color: var(--text-secondary);
  margin-left: 4px;
  font-family: var(--font-mono, monospace);
}

.telemetry-info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.85rem;
  color: var(--text-secondary);
  padding-top: 4px;
}

.circle-progress-item {
  align-items: center;
  margin-top: 8px;
  gap: 12px;
}

.telemetry-label-center {
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

.telemetry-status-box {
  background: rgba(57, 255, 20, 0.06);
  border: 1px solid rgba(57, 255, 20, 0.12);
  border-radius: 12px;
  padding: 14px;
  margin-top: 8px;
  text-align: center;
}

.status-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-weight: 700;
  font-size: 0.9rem;
  color: var(--neon-green);
}

.status-indicator-dot {
  width: 6px;
  height: 6px;
  background-color: var(--neon-green);
  border-radius: 50%;
  box-shadow: 0 0 6px var(--neon-green);
}

.status-sub {
  font-size: 0.75rem;
  color: var(--text-secondary);
  margin-top: 4px;
}

/* 2. 中栏：指标与活跃任务样式 */
.stat-mini-grids {
  margin-bottom: 16px;
}

.mini-tile {
  padding: 12px;
  text-align: center;
  border-radius: 12px;
}

.tile-label {
  font-size: 0.75rem;
  color: var(--text-secondary);
  margin-bottom: 4px;
  font-weight: 600;
}

.tile-value {
  font-size: 1.2rem;
  font-weight: 800;
}

.tile-value.cyan { color: var(--neon-teal); }
.tile-value.purple { color: var(--neon-purple); }
.tile-value.green { color: var(--neon-green); }
.tile-value.orange { color: var(--neon-orange); }

.core-jobs-card {
  height: calc(100% - 130px);
  min-height: 400px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 700;
  color: var(--text-primary);
}

.icon-spin-slow {
  animation: spin 8s infinite linear;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.monitor-scroll-area {
  max-height: 380px;
  overflow-y: auto;
  padding-right: 4px;
}

.task-progress-card {
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid var(--border-color);
  border-radius: 12px;
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

.empty-active-placeholder {
  padding-top: 10px;
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
  border-radius: 12px;
  justify-content: flex-end;
}

/* 3. 右栏：日志流样式 */
.terminal-main-card {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.terminal-window {
  background: #04060b !important;
  font-family: var(--font-mono);
  padding: 16px;
  height: 480px;
  overflow-y: auto;
  font-size: 0.8rem;
  line-height: 1.6;
}

.log-line {
  margin-bottom: 6px;
  color: var(--text-secondary);
}

.log-timestamp {
  color: var(--text-muted);
  margin-right: 8px;
}

.log-error { color: #f56c6c; }
.log-warn { color: var(--neon-orange); }
.log-success { color: var(--neon-green); }

.terminal-empty {
  color: var(--text-muted);
  display: flex;
  align-items: center;
  gap: 4px;
}

.terminal-cursor {
  animation: flash 1s infinite steps(2);
}

@keyframes flash {
  0%, 100% { opacity: 0; }
  50% { opacity: 1; }
}

.terminal-collapsed-indicator {
  height: 520px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  background: rgba(255, 255, 255, 0.01);
  border-radius: 8px;
}

.indicator-text {
  writing-mode: vertical-rl;
  text-orientation: mixed;
  letter-spacing: 0.2em;
  font-weight: 700;
  font-size: 0.9rem;
  color: var(--neon-teal);
}
</style>
