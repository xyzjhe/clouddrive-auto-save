# UCAS 全面升级 - 阶段一：UI/UX 改进实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 优化 UCAS 的 UI/UX 设计，包括侧边栏导航优化、仪表盘增强、卡片式组件、PWA 支持。

**Architecture:** 基于 Vue 3 + Element Plus 技术栈，采用组件化设计，支持响应式布局和暗黑模式。

**Tech Stack:** Vue 3.5, Element Plus 2.13, ECharts 5.x, Vue Router 5, Pinia 3, Vite 8

---

## 文件结构

### 新增文件
- `web/src/config/navigation.ts` - 导航配置
- `web/src/components/charts/TrendChart.vue` - 趋势图组件
- `web/src/components/charts/StorageChart.vue` - 存储分布图组件
- `web/src/components/cards/AccountCard.vue` - 账号卡片组件
- `web/src/components/cards/TaskCard.vue` - 任务卡片组件
- `web/public/manifest.json` - PWA 配置
- `web/public/sw.js` - Service Worker
- `web/src/components/PWAInstall.vue` - 安装引导组件

### 修改文件
- `web/src/layout/MainLayout.vue` - 主布局组件
- `web/src/views/Dashboard.vue` - 仪表盘页面
- `web/src/views/Accounts.vue` - 账号管理页面
- `web/src/views/Tasks.vue` - 任务管理页面
- `web/index.html` - 添加 PWA meta 标签
- `web/package.json` - 添加 ECharts 依赖

---

## Task 1: 安装 ECharts 依赖

**Files:**
- Modify: `web/package.json`

- [ ] **Step 1: 添加 ECharts 依赖**

```bash
cd web && npm install echarts vue-echarts
```

- [ ] **Step 2: 验证安装**

```bash
cd web && npm list echarts vue-echarts
```

Expected: 显示 echarts 和 vue-echarts 版本

- [ ] **Step 3: 提交**

```bash
git add web/package.json web/package-lock.json
git commit -m "chore: 添加 ECharts 图表库依赖"
```

---

## Task 2: 创建导航配置

**Files:**
- Create: `web/src/config/navigation.ts`

- [ ] **Step 1: 创建导航配置文件**

```typescript
// web/src/config/navigation.ts
export interface NavItem {
  name: string
  path: string
  icon: string
  description?: string
}

export interface NavGroup {
  name: string
  icon: string
  items: NavItem[]
  collapsible?: boolean
  defaultCollapsed?: boolean
}

export const navigationConfig: NavGroup[] = [
  {
    name: '概览',
    icon: '📊',
    items: [
      {
        name: '仪表盘',
        path: '/dashboard',
        icon: 'LayoutDashboard',
        description: '实时统计和任务监控'
      }
    ]
  },
  {
    name: '管理',
    icon: '🔧',
    items: [
      {
        name: '账号管理',
        path: '/accounts',
        icon: 'Users',
        description: '管理云盘账号'
      },
      {
        name: '任务列表',
        path: '/tasks',
        icon: 'ListTodo',
        description: '管理转存任务'
      }
    ]
  },
  {
    name: '工具',
    icon: '🛠️',
    items: [
      {
        name: '资源搜索',
        path: '/search',
        icon: 'Search',
        description: '搜索云盘资源'
      },
      {
        name: '插件管理',
        path: '/plugins',
        icon: 'Puzzle',
        description: '管理系统插件'
      }
    ]
  },
  {
    name: '系统',
    icon: '⚙️',
    items: [
      {
        name: '系统设置',
        path: '/settings',
        icon: 'Settings',
        description: '全局配置'
      },
      {
        name: '消息推送',
        path: '/notify',
        icon: 'Bell',
        description: '通知渠道配置'
      }
    ]
  }
]
```

- [ ] **Step 2: 提交**

```bash
git add web/src/config/navigation.ts
git commit -m "feat: 创建导航配置文件，支持分类分组"
```

---

## Task 3: 重构侧边栏组件

**Files:**
- Modify: `web/src/layout/MainLayout.vue`

- [ ] **Step 1: 读取当前 MainLayout.vue**

读取现有代码，了解当前实现。

- [ ] **Step 2: 重构侧边栏为分类分组+可折叠设计**

```vue
<!-- web/src/layout/MainLayout.vue -->
<script setup>
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useDark, useToggle } from '@vueuse/core'
import { navigationConfig } from '../config/navigation'

const route = useRoute()
const router = useRouter()
const isDark = useDark()
const toggleDark = useToggle(isDark)

// 折叠状态管理
const collapsedGroups = ref(JSON.parse(localStorage.getItem('collapsedGroups') || '{}'))
const searchQuery = ref('')

const toggleGroup = (groupName) => {
  collapsedGroups.value[groupName] = !collapsedGroups.value[groupName]
  localStorage.setItem('collapsedGroups', JSON.stringify(collapsedGroups.value))
}

// 过滤导航项
const filteredNavigation = computed(() => {
  if (!searchQuery.value) return navigationConfig

  const query = searchQuery.value.toLowerCase()
  return navigationConfig
    .map(group => ({
      ...group,
      items: group.items.filter(item =>
        item.name.toLowerCase().includes(query) ||
        item.description?.toLowerCase().includes(query)
      )
    }))
    .filter(group => group.items.length > 0)
})

const isActive = (path) => route.path === path

const navigateTo = (path) => {
  router.push(path)
}
</script>

<template>
  <el-container class="main-layout">
    <el-aside width="240px" class="sidebar">
      <div class="sidebar-header">
        <div class="logo">
          <svg width="32" height="32" viewBox="0 0 32 32">
            <!-- SVG logo -->
          </svg>
          <span class="logo-text">UCAS</span>
        </div>
        <span class="version">v1.0.0</span>
      </div>

      <div class="search-wrapper">
        <el-input
          v-model="searchQuery"
          placeholder="搜索功能..."
          clearable
          size="small"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
      </div>

      <el-scrollbar class="nav-scrollbar">
        <div class="nav-groups">
          <div
            v-for="group in filteredNavigation"
            :key="group.name"
            class="nav-group"
          >
            <div
              class="nav-group-header"
              @click="toggleGroup(group.name)"
            >
              <span class="nav-group-icon">{{ group.icon }}</span>
              <span class="nav-group-name">{{ group.name }}</span>
              <el-icon
                class="nav-group-arrow"
                :class="{ collapsed: collapsedGroups[group.name] }"
              >
                <ArrowDown />
              </el-icon>
            </div>

            <transition name="slide">
              <div
                v-show="!collapsedGroups[group.name]"
                class="nav-items"
              >
                <div
                  v-for="item in group.items"
                  :key="item.path"
                  class="nav-item"
                  :class="{ active: isActive(item.path) }"
                  @click="navigateTo(item.path)"
                >
                  <el-icon class="nav-item-icon">
                    <component :is="item.icon" />
                  </el-icon>
                  <span class="nav-item-name">{{ item.name }}</span>
                </div>
              </div>
            </transition>
          </div>
        </div>
      </el-scrollbar>

      <div class="sidebar-footer">
        <el-button @click="toggleDark()" text>
          <el-icon>
            <Moon v-if="isDark" />
            <Sunny v-else />
          </el-icon>
        </el-button>
        <a href="https://github.com/zcq/clouddrive-auto-save" target="_blank">
          <el-button text>
            <el-icon><Github /></el-icon>
          </el-button>
        </a>
      </div>
    </el-aside>

    <el-container>
      <el-header class="main-header">
        <div class="breadcrumb">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item>{{ route.meta.title }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
      </el-header>

      <el-main class="main-content">
        <router-view v-slot="{ Component }">
          <transition name="fade-transform" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.main-layout {
  height: 100vh;
}

.sidebar {
  background: var(--bg-secondary);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  padding: 1rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--border);
}

.logo {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.logo-text {
  font-size: 1.25rem;
  font-weight: bold;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.version {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.search-wrapper {
  padding: 0.75rem 1rem;
}

.nav-scrollbar {
  flex: 1;
}

.nav-groups {
  padding: 0.5rem;
}

.nav-group {
  margin-bottom: 0.5rem;
}

.nav-group-header {
  display: flex;
  align-items: center;
  padding: 0.5rem 0.75rem;
  cursor: pointer;
  border-radius: 6px;
  transition: background-color 0.2s;
}

.nav-group-header:hover {
  background: var(--bg-tertiary);
}

.nav-group-icon {
  margin-right: 0.5rem;
  font-size: 1rem;
}

.nav-group-name {
  flex: 1;
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.nav-group-arrow {
  transition: transform 0.3s;
}

.nav-group-arrow.collapsed {
  transform: rotate(-90deg);
}

.nav-items {
  overflow: hidden;
}

.nav-item {
  display: flex;
  align-items: center;
  padding: 0.75rem 1rem;
  margin: 0.25rem 0;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.nav-item:hover {
  background: var(--bg-tertiary);
}

.nav-item.active {
  background: var(--brand-500);
  color: white;
}

.nav-item-icon {
  margin-right: 0.75rem;
  font-size: 1.1rem;
}

.nav-item-name {
  font-size: 0.9rem;
}

.sidebar-footer {
  padding: 1rem;
  border-top: 1px solid var(--border);
  display: flex;
  justify-content: center;
  gap: 0.5rem;
}

.main-header {
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border);
  display: flex;
  align-items: center;
  padding: 0 1.5rem;
}

.breadcrumb {
  flex: 1;
}

.main-content {
  background: var(--bg-primary);
  padding: 1.5rem;
  overflow-y: auto;
}

/* 过渡动画 */
.fade-transform-enter-active,
.fade-transform-leave-active {
  transition: all 0.3s;
}

.fade-transform-enter-from {
  opacity: 0;
  transform: translateX(-20px);
}

.fade-transform-leave-to {
  opacity: 0;
  transform: translateX(20px);
}

.slide-enter-active,
.slide-leave-active {
  transition: all 0.3s;
}

.slide-enter-from,
.slide-leave-to {
  opacity: 0;
  max-height: 0;
}

.slide-enter-to,
.slide-leave-from {
  max-height: 500px;
}

/* 响应式 */
@media (max-width: 768px) {
  .sidebar {
    position: fixed;
    z-index: 1000;
    height: 100vh;
    transform: translateX(-100%);
    transition: transform 0.3s;
  }

  .sidebar.open {
    transform: translateX(0);
  }
}
</style>
```

- [ ] **Step 3: 验证侧边栏功能**

启动开发服务器，验证：
1. 导航项正确显示
2. 分类分组可折叠
3. 搜索功能正常
4. 折叠状态持久化

```bash
cd web && npm run dev
```

- [ ] **Step 4: 提交**

```bash
git add web/src/layout/MainLayout.vue
git commit -m "feat: 重构侧边栏为分类分组+可折叠设计"
```

---

## Task 4: 创建图表组件

**Files:**
- Create: `web/src/components/charts/TrendChart.vue`
- Create: `web/src/components/charts/StorageChart.vue`

- [ ] **Step 1: 创建趋势图组件**

```vue
<!-- web/src/components/charts/TrendChart.vue -->
<script setup>
import { ref, onMounted, watch } from 'vue'
import * as echarts from 'echarts'

const props = defineProps({
  data: {
    type: Array,
    default: () => []
  }
})

const chartRef = ref(null)
let chart = null

const initChart = () => {
  if (!chartRef.value) return

  chart = echarts.init(chartRef.value)

  const option = {
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow'
      }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: [
      {
        type: 'category',
        data: props.data.map(item => item.date),
        axisTick: {
          alignWithLabel: true
        }
      }
    ],
    yAxis: [
      {
        type: 'value',
        name: '任务数'
      }
    ],
    series: [
      {
        name: '完成任务',
        type: 'bar',
        barWidth: '60%',
        data: props.data.map(item => item.count),
        itemStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: '#6366f1' },
            { offset: 1, color: '#8b5cf6' }
          ]),
          borderRadius: [4, 4, 0, 0]
        }
      }
    ]
  }

  chart.setOption(option)
}

watch(() => props.data, () => {
  if (chart) {
    chart.setOption({
      xAxis: [{
        data: props.data.map(item => item.date)
      }],
      series: [{
        data: props.data.map(item => item.count)
      }]
    })
  }
}, { deep: true })

onMounted(() => {
  initChart()
  window.addEventListener('resize', () => {
    chart?.resize()
  })
})
</script>

<template>
  <div ref="chartRef" class="trend-chart"></div>
</template>

<style scoped>
.trend-chart {
  width: 100%;
  height: 200px;
}
</style>
```

- [ ] **Step 2: 创建存储分布图组件**

```vue
<!-- web/src/components/charts/StorageChart.vue -->
<script setup>
import { ref, onMounted, watch } from 'vue'
import * as echarts from 'echarts'

const props = defineProps({
  data: {
    type: Array,
    default: () => []
  }
})

const chartRef = ref(null)
let chart = null

const initChart = () => {
  if (!chartRef.value) return

  chart = echarts.init(chartRef.value)

  const option = {
    tooltip: {
      trigger: 'item',
      formatter: '{a} <br/>{b}: {c} TB ({d}%)'
    },
    legend: {
      orient: 'vertical',
      right: 10,
      top: 'center'
    },
    series: [
      {
        name: '存储空间',
        type: 'pie',
        radius: ['40%', '70%'],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 10,
          borderColor: '#fff',
          borderWidth: 2
        },
        label: {
          show: false,
          position: 'center'
        },
        emphasis: {
          label: {
            show: true,
            fontSize: '14',
            fontWeight: 'bold'
          }
        },
        labelLine: {
          show: false
        },
        data: props.data.map(item => ({
          value: item.used,
          name: item.platform
        }))
      }
    ]
  }

  chart.setOption(option)
}

watch(() => props.data, () => {
  if (chart) {
    chart.setOption({
      series: [{
        data: props.data.map(item => ({
          value: item.used,
          name: item.platform
        }))
      }]
    })
  }
}, { deep: true })

onMounted(() => {
  initChart()
  window.addEventListener('resize', () => {
    chart?.resize()
  })
})
</script>

<template>
  <div ref="chartRef" class="storage-chart"></div>
</template>

<style scoped>
.storage-chart {
  width: 100%;
  height: 200px;
}
</style>
```

- [ ] **Step 3: 提交**

```bash
mkdir -p web/src/components/charts
git add web/src/components/charts/
git commit -m "feat: 创建 ECharts 图表组件（趋势图、存储分布图）"
```

---

## Task 5: 增强仪表盘页面

**Files:**
- Modify: `web/src/views/Dashboard.vue`

- [ ] **Step 1: 读取当前 Dashboard.vue**

读取现有代码，了解当前实现。

- [ ] **Step 2: 添加图表和任务监控面板**

```vue
<!-- web/src/views/Dashboard.vue -->
<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import TrendChart from '../components/charts/TrendChart.vue'
import StorageChart from '../components/charts/StorageChart.vue'

const router = useRouter()

// 统计数据
const stats = ref({
  totalTasks: 0,
  savedCapacity: 0,
  todayCompleted: 0,
  activeAccounts: 0
})

// 趋势数据（最近7天）
const trendData = ref([
  { date: '05-13', count: 5 },
  { date: '05-14', count: 8 },
  { date: '05-15', count: 3 },
  { date: '05-16', count: 12 },
  { date: '05-17', count: 7 },
  { date: '05-18', count: 9 },
  { date: '05-19', count: 6 }
])

// 存储分布数据
const storageData = ref([
  { platform: '夸克网盘', used: 1.8, total: 5 },
  { platform: '移动云盘', used: 0.6, total: 2 }
])

// 运行中任务
const runningTasks = ref([])

// 日志数据
const logs = ref([])

// SSE 连接
let eventSource = null

const connectSSE = () => {
  eventSource = new EventSource('/api/dashboard/logs')

  eventSource.onmessage = (event) => {
    const data = event.data

    // 解析进度更新
    if (data.startsWith('[PROGRESS:')) {
      const match = data.match(/\[PROGRESS:(\d+):(\d+):(.+?):(.+?)\]/)
      if (match) {
        const [, taskId, percent, stage, message] = match
        updateRunningTask(taskId, parseInt(percent), stage, message)
      }
    }

    // 解析事件
    if (data.startsWith('[EVENT:')) {
      const match = data.match(/\[EVENT:(.+?)\]/)
      if (match) {
        const eventType = match[1]
        if (eventType === 'stats_update') {
          fetchStats()
        }
      }
    }

    // 添加到日志
    logs.value.unshift({
      time: new Date().toLocaleTimeString(),
      message: data
    })

    // 限制日志数量
    if (logs.value.length > 100) {
      logs.value.pop()
    }
  }
}

const updateRunningTask = (taskId, percent, stage, message) => {
  const index = runningTasks.value.findIndex(t => t.id === taskId)
  if (index !== -1) {
    runningTasks.value[index] = {
      ...runningTasks.value[index],
      percent,
      stage,
      message
    }
  } else {
    runningTasks.value.push({
      id: taskId,
      percent,
      stage,
      message
    })
  }
}

const fetchStats = async () => {
  try {
    const response = await fetch('/api/dashboard/stats')
    const data = await response.json()
    stats.value = data
  } catch (error) {
    console.error('获取统计信息失败:', error)
  }
}

const clearLogs = () => {
  logs.value = []
}

onMounted(() => {
  fetchStats()
  connectSSE()
})

onUnmounted(() => {
  eventSource?.close()
})
</script>

<template>
  <div class="dashboard">
    <div class="welcome">
      <h1>仪表盘</h1>
      <p>实时监控系统状态和任务执行情况</p>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon" style="background: #6366f1;">📋</div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.totalTasks }}</div>
          <div class="stat-label">已规划任务</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon" style="background: #10b981;">💾</div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.savedCapacity }} TB</div>
          <div class="stat-label">已保存容量</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon" style="background: #f59e0b;">✅</div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.todayCompleted }}</div>
          <div class="stat-label">今日完成</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon" style="background: #8b5cf6;">👤</div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.activeAccounts }}</div>
          <div class="stat-label">活跃账号</div>
        </div>
      </div>
    </div>

    <!-- 图表区域 -->
    <div class="charts-grid">
      <div class="chart-card">
        <h3>任务执行趋势（最近7天）</h3>
        <TrendChart :data="trendData" />
      </div>

      <div class="chart-card">
        <h3>存储空间分布</h3>
        <StorageChart :data="storageData" />
      </div>
    </div>

    <!-- 运行中任务 -->
    <div v-if="runningTasks.length > 0" class="running-tasks">
      <h3>运行中任务</h3>
      <div
        v-for="task in runningTasks"
        :key="task.id"
        class="task-item"
      >
        <div class="task-header">
          <span class="task-name">任务 #{{ task.id }}</span>
          <span class="task-stage">{{ task.stage }}</span>
        </div>
        <el-progress
          :percentage="task.percent"
          :stroke-width="8"
          striped
          striped-flow
        />
        <div class="task-message">{{ task.message }}</div>
      </div>
    </div>

    <!-- 日志终端 -->
    <div class="log-terminal">
      <div class="log-header">
        <span>实时日志</span>
        <el-button size="small" @click="clearLogs">清空</el-button>
      </div>
      <div class="log-content">
        <div
          v-for="(log, index) in logs"
          :key="index"
          class="log-line"
          :class="{
            'log-error': log.message.includes('[Fatal]') || log.message.includes('ERROR'),
            'log-warn': log.message.includes('WARN'),
            'log-success': log.message.includes('成功')
          }"
        >
          <span class="log-time">{{ log.time }}</span>
          <span class="log-message">{{ log.message }}</span>
        </div>
      </div>
    </div>

    <!-- 浮动操作按钮 -->
    <div class="fab">
      <el-button
        type="primary"
        circle
        size="large"
        @click="router.push('/accounts?action=add')"
      >
        <el-icon><Plus /></el-icon>
      </el-button>
    </div>
  </div>
</template>

<style scoped>
.dashboard {
  position: relative;
}

.welcome {
  margin-bottom: 1.5rem;
}

.welcome h1 {
  margin: 0 0 0.25rem 0;
  font-size: 1.75rem;
}

.welcome p {
  margin: 0;
  color: var(--text-secondary);
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.stat-card {
  background: var(--bg-secondary);
  border-radius: 12px;
  padding: 1.25rem;
  display: flex;
  align-items: center;
  gap: 1rem;
  box-shadow: var(--shadow-sm);
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
}

.stat-value {
  font-size: 1.75rem;
  font-weight: bold;
}

.stat-label {
  font-size: 0.85rem;
  color: var(--text-secondary);
}

.charts-grid {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.chart-card {
  background: var(--bg-secondary);
  border-radius: 12px;
  padding: 1.25rem;
  box-shadow: var(--shadow-sm);
}

.chart-card h3 {
  margin: 0 0 1rem 0;
  font-size: 1rem;
}

.running-tasks {
  background: var(--bg-secondary);
  border-radius: 12px;
  padding: 1.25rem;
  margin-bottom: 1.5rem;
  box-shadow: var(--shadow-sm);
}

.running-tasks h3 {
  margin: 0 0 1rem 0;
  font-size: 1rem;
}

.task-item {
  padding: 1rem;
  background: var(--bg-tertiary);
  border-radius: 8px;
  margin-bottom: 0.75rem;
}

.task-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 0.5rem;
}

.task-name {
  font-weight: 600;
}

.task-stage {
  color: var(--text-secondary);
  font-size: 0.85rem;
}

.task-message {
  margin-top: 0.5rem;
  font-size: 0.85rem;
  color: var(--text-secondary);
}

.log-terminal {
  background: #1e1e1e;
  border-radius: 12px;
  overflow: hidden;
  margin-bottom: 1.5rem;
}

.log-header {
  padding: 0.75rem 1rem;
  background: #2d2d2d;
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: #ccc;
}

.log-content {
  padding: 1rem;
  height: 300px;
  overflow-y: auto;
  font-family: 'JetBrains Mono', monospace;
  font-size: 0.85rem;
}

.log-line {
  padding: 0.25rem 0;
  color: #d4d4d4;
}

.log-time {
  color: #6a9955;
  margin-right: 1rem;
}

.log-error {
  color: #f44747;
}

.log-warn {
  color: #cca700;
}

.log-success {
  color: #6a9955;
}

.fab {
  position: fixed;
  bottom: 2rem;
  right: 2rem;
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .charts-grid {
    grid-template-columns: 1fr;
  }
}
</style>
```

- [ ] **Step 3: 验证仪表盘功能**

启动开发服务器，验证：
1. 统计卡片正确显示
2. 图表正常渲染
3. 运行中任务实时更新
4. 日志终端正常工作

```bash
cd web && npm run dev
```

- [ ] **Step 4: 提交**

```bash
git add web/src/views/Dashboard.vue
git commit -m "feat: 增强仪表盘，添加图表和任务监控面板"
```

---

## Task 6: 创建卡片组件

**Files:**
- Create: `web/src/components/cards/AccountCard.vue`
- Create: `web/src/components/cards/TaskCard.vue`

- [ ] **Step 1: 创建账号卡片组件**

```vue
<!-- web/src/components/cards/AccountCard.vue -->
<script setup>
import { computed } from 'vue'

const props = defineProps({
  account: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['check', 'edit', 'delete'])

const platformColors = {
  '139': 'linear-gradient(135deg, #f59e0b, #f97316)',
  'quark': 'linear-gradient(135deg, #6366f1, #8b5cf6)'
}

const storagePercentage = computed(() => {
  if (!props.account.capacity) return 0
  return Math.round((props.account.usedSpace / props.account.capacity) * 100)
})

const storageColor = computed(() => {
  if (storagePercentage.value < 60) return '#10b981'
  if (storagePercentage.value < 80) return '#f59e0b'
  return '#ef4444'
})

const formatSize = (bytes) => {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = bytes
  let unitIndex = 0
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex++
  }
  return `${size.toFixed(1)} ${units[unitIndex]}"
}
</script>

<template>
  <div class="account-card">
    <div
      class="card-header"
      :style="{ background: platformColors[account.platform] || platformColors.quark }"
    >
      <div class="platform-name">{{ account.platform === '139' ? '移动云盘' : '夸克网盘' }}</div>
      <div class="nickname">{{ account.nickname }}</div>
    </div>

    <div class="card-body">
      <div class="storage-info">
        <div class="storage-header">
          <span class="storage-label">存储空间</span>
          <span class="storage-value">
            {{ formatSize(account.usedSpace) }} / {{ formatSize(account.capacity) }}
          </span>
        </div>
        <el-progress
          :percentage="storagePercentage"
          :color="storageColor"
          :stroke-width="8"
          :show-text="false"
        />
      </div>

      <div class="card-actions">
        <el-button size="small" @click="emit('check', account.id)">
          校验
        </el-button>
        <el-button size="small" @click="emit('edit', account.id)">
          编辑
        </el-button>
        <el-button
          size="small"
          type="danger"
          @click="emit('delete', account.id)"
        >
          删除
        </el-button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.account-card {
  background: var(--bg-secondary);
  border-radius: 12px;
  overflow: hidden;
  box-shadow: var(--shadow-sm);
  transition: transform 0.2s, box-shadow 0.2s;
}

.account-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-md);
}

.card-header {
  padding: 1.25rem;
  color: white;
}

.platform-name {
  font-size: 1.25rem;
  font-weight: bold;
  margin-bottom: 0.25rem;
}

.nickname {
  font-size: 0.9rem;
  opacity: 0.9;
}

.card-body {
  padding: 1.25rem;
}

.storage-info {
  margin-bottom: 1rem;
}

.storage-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 0.5rem;
}

.storage-label {
  font-size: 0.85rem;
  color: var(--text-secondary);
}

.storage-value {
  font-size: 0.85rem;
  font-weight: 600;
}

.card-actions {
  display: flex;
  gap: 0.5rem;
}

.card-actions .el-button {
  flex: 1;
}
</style>
```

- [ ] **Step 2: 创建任务卡片组件**

```vue
<!-- web/src/components/cards/TaskCard.vue -->
<script setup>
import { computed } from 'vue'

const props = defineProps({
  task: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['run', 'edit', 'delete'])

const statusConfig = {
  'pending': { label: '等待中', color: '#909399' },
  'running': { label: '运行中', color: '#409eff' },
  'completed': { label: '已完成', color: '#67c23a' },
  'failed': { label: '失败', color: '#f56c6c' },
  'fatal': { label: 'Fatal', color: '#f56c6c' }
}

const currentStatus = computed(() => {
  return statusConfig[props.task.status] || statusConfig.pending
})

const scheduleText = computed(() => {
  if (props.task.scheduleMode === 'global') return '跟随全局'
  if (props.task.scheduleMode === 'custom') return props.task.cron
  return '手动执行'
})
</script>

<template>
  <div class="task-card">
    <div class="card-header">
      <div class="task-name">{{ task.name }}</div>
      <el-tag
        :color="currentStatus.color"
        effect="dark"
        size="small"
      >
        {{ currentStatus.label }}
      </el-tag>
    </div>

    <div class="card-info">
      <div class="info-item">
        <span class="info-label">平台</span>
        <span class="info-value">{{ task.accountName }}</span>
      </div>
      <div class="info-item">
        <span class="info-label">保存路径</span>
        <span class="info-value">{{ task.savePath }}</span>
      </div>
      <div class="info-item">
        <span class="info-label">调度</span>
        <span class="info-value">{{ scheduleText }}</span>
      </div>
    </div>

    <div v-if="task.status === 'running'" class="progress-section">
      <el-progress
        :percentage="task.progress || 0"
        :stroke-width="8"
        striped
        striped-flow
      />
      <div class="progress-text">{{ task.progressMessage }}</div>
    </div>

    <div class="card-actions">
      <el-button
        size="small"
        type="primary"
        :disabled="task.status === 'running'"
        @click="emit('run', task.id)"
      >
        执行
      </el-button>
      <el-button size="small" @click="emit('edit', task.id)">
        编辑
      </el-button>
      <el-button
        size="small"
        type="danger"
        @click="emit('delete', task.id)"
      >
        删除
      </el-button>
    </div>
  </div>
</template>

<style scoped>
.task-card {
  background: var(--bg-secondary);
  border-radius: 12px;
  padding: 1.25rem;
  box-shadow: var(--shadow-sm);
  transition: transform 0.2s, box-shadow 0.2s;
}

.task-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-md);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.task-name {
  font-size: 1.1rem;
  font-weight: 600;
}

.card-info {
  margin-bottom: 1rem;
}

.info-item {
  display: flex;
  justify-content: space-between;
  padding: 0.5rem 0;
  border-bottom: 1px solid var(--border);
}

.info-item:last-child {
  border-bottom: none;
}

.info-label {
  color: var(--text-secondary);
  font-size: 0.85rem;
}

.info-value {
  font-size: 0.85rem;
  font-weight: 500;
}

.progress-section {
  margin-bottom: 1rem;
}

.progress-text {
  margin-top: 0.5rem;
  font-size: 0.85rem;
  color: var(--text-secondary);
  text-align: center;
}

.card-actions {
  display: flex;
  gap: 0.5rem;
}

.card-actions .el-button {
  flex: 1;
}
</style>
```

- [ ] **Step 3: 提交**

```bash
mkdir -p web/src/components/cards
git add web/src/components/cards/
git commit -m "feat: 创建卡片组件（账号卡片、任务卡片）"
```

---

## Task 7: 集成卡片组件到页面

**Files:**
- Modify: `web/src/views/Accounts.vue`
- Modify: `web/src/views/Tasks.vue`

- [ ] **Step 1: 修改账号管理页面支持卡片视图**

```vue
<!-- web/src/views/Accounts.vue -->
<script setup>
import { ref, onMounted } from 'vue'
import AccountCard from '../components/cards/AccountCard.vue'

// ... existing code ...

const viewMode = ref(localStorage.getItem('accountViewMode') || 'table')

const toggleViewMode = () => {
  viewMode.value = viewMode.value === 'table' ? 'card' : 'table'
  localStorage.setItem('accountViewMode', viewMode.value)
}

// ... existing code ...
</script>

<template>
  <div class="accounts-page">
    <div class="page-header">
      <h1>账号管理</h1>
      <div class="header-actions">
        <el-button @click="toggleViewMode">
          <el-icon>
            <Grid v-if="viewMode === 'table'" />
            <List v-else />
          </el-icon>
          {{ viewMode === 'table' ? '卡片视图' : '表格视图' }}
        </el-button>
        <el-button type="primary" @click="showAddDialog">
          添加账号
        </el-button>
      </div>
    </div>

    <!-- 表格视图 -->
    <el-table
      v-if="viewMode === 'table'"
      :data="accounts"
      style="width: 100%"
    >
      <!-- ... existing table columns ... -->
    </el-table>

    <!-- 卡片视图 -->
    <div
      v-else
      class="cards-grid"
    >
      <AccountCard
        v-for="account in accounts"
        :key="account.id"
        :account="account"
        @check="handleCheck"
        @edit="handleEdit"
        @delete="handleDelete"
      />
    </div>

    <!-- ... existing dialogs ... -->
  </div>
</template>

<style scoped>
.accounts-page {
  /* ... existing styles ... */
}

.cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1rem;
}
</style>
```

- [ ] **Step 2: 修改任务管理页面支持卡片视图**

```vue
<!-- web/src/views/Tasks.vue -->
<script setup>
import { ref, onMounted } from 'vue'
import TaskCard from '../components/cards/TaskCard.vue'

// ... existing code ...

const viewMode = ref(localStorage.getItem('taskViewMode') || 'table')

const toggleViewMode = () => {
  viewMode.value = viewMode.value === 'table' ? 'card' : 'table'
  localStorage.setItem('taskViewMode', viewMode.value)
}

// ... existing code ...
</script>

<template>
  <div class="tasks-page">
    <div class="page-header">
      <h1>任务列表</h1>
      <div class="header-actions">
        <el-button @click="toggleViewMode">
          <el-icon>
            <Grid v-if="viewMode === 'table'" />
            <List v-else />
          </el-icon>
          {{ viewMode === 'table' ? '卡片视图' : '表格视图' }}
        </el-button>
        <el-button type="primary" @click="showAddDialog">
          创建任务
        </el-button>
        <el-button @click="handleRunAll">
          全部执行
        </el-button>
      </div>
    </div>

    <!-- 表格视图 -->
    <el-table
      v-if="viewMode === 'table'"
      :data="tasks"
      style="width: 100%"
    >
      <!-- ... existing table columns ... -->
    </el-table>

    <!-- 卡片视图 -->
    <div
      v-else
      class="cards-grid"
    >
      <TaskCard
        v-for="task in tasks"
        :key="task.id"
        :task="task"
        @run="handleRun"
        @edit="handleEdit"
        @delete="handleDelete"
      />
    </div>

    <!-- ... existing dialogs ... -->
  </div>
</template>

<style scoped>
.tasks-page {
  /* ... existing styles ... */
}

.cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 1rem;
}
</style>
```

- [ ] **Step 3: 验证卡片视图功能**

启动开发服务器，验证：
1. 表格/卡片视图切换正常
2. 卡片组件正确显示数据
3. 视图模式持久化

```bash
cd web && npm run dev
```

- [ ] **Step 4: 提交**

```bash
git add web/src/views/Accounts.vue web/src/views/Tasks.vue
git commit -m "feat: 集成卡片组件到账号和任务管理页面"
```

---

## Task 8: 配置 PWA

**Files:**
- Create: `web/public/manifest.json`
- Modify: `web/index.html`

- [ ] **Step 1: 创建 manifest.json**

```json
{
  "name": "UCAS - 统一云盘自动转存系统",
  "short_name": "UCAS",
  "description": "自动化云盘转存管理工具",
  "start_url": "/",
  "display": "standalone",
  "background_color": "#ffffff",
  "theme_color": "#6366f1",
  "orientation": "any",
  "icons": [
    {
      "src": "/icon-192.png",
      "sizes": "192x192",
      "type": "image/png"
    },
    {
      "src": "/icon-512.png",
      "sizes": "512x512",
      "type": "image/png"
    }
  ]
}
```

- [ ] **Step 2: 添加 PWA meta 标签到 index.html**

```html
<!DOCTYPE html>
<html lang="zh-CN">
  <head>
    <meta charset="UTF-8" />
    <link rel="icon" type="image/svg+xml" href="/vite.svg" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />

    <!-- PWA Meta Tags -->
    <meta name="description" content="UCAS - 统一云盘自动转存系统" />
    <meta name="theme-color" content="#6366f1" />
    <link rel="manifest" href="/manifest.json" />

    <!-- iOS PWA Support -->
    <meta name="apple-mobile-web-app-capable" content="yes" />
    <meta name="apple-mobile-web-app-status-bar-style" content="default" />
    <meta name="apple-mobile-web-app-title" content="UCAS" />
    <link rel="apple-touch-icon" href="/icon-192.png" />

    <title>UCAS - 统一云盘自动转存系统</title>
  </head>
  <body>
    <div id="app"></div>
    <script type="module" src="/src/main.js"></script>
  </body>
</html>
```

- [ ] **Step 3: 创建占位图标文件**

```bash
# 创建简单的 SVG 图标作为占位符
cat > web/public/icon-192.svg << 'EOF'
<svg xmlns="http://www.w3.org/2000/svg" width="192" height="192" viewBox="0 0 192 192">
  <rect width="192" height="192" rx="32" fill="#6366f1"/>
  <text x="96" y="120" font-size="80" text-anchor="middle" fill="white" font-family="Arial">☁️</text>
</svg>
EOF
```

- [ ] **Step 4: 提交**

```bash
git add web/public/manifest.json web/public/icon-192.svg web/index.html
git commit -m "feat: 配置 PWA 支持，添加 manifest 和 meta 标签"
```

---

## Task 9: 创建 Service Worker

**Files:**
- Create: `web/public/sw.js`

- [ ] **Step 1: 创建 Service Worker 文件**

```javascript
// web/public/sw.js
const CACHE_NAME = 'ucas-v1'
const STATIC_ASSETS = [
  '/',
  '/index.html',
  '/manifest.json'
]

// 安装事件 - 缓存静态资源
self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => {
      return cache.addAll(STATIC_ASSETS)
    })
  )
  self.skipWaiting()
})

// 激活事件 - 清理旧缓存
self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames
          .filter((name) => name !== CACHE_NAME)
          .map((name) => caches.delete(name))
      )
    })
  )
  self.clients.claim()
})

// 请求事件 - 网络优先策略
self.addEventListener('fetch', (event) => {
  // 跳过 API 请求
  if (event.request.url.includes('/api/')) {
    return
  }

  event.respondWith(
    fetch(event.request)
      .then((response) => {
        // 成功获取资源后，更新缓存
        const responseClone = response.clone()
        caches.open(CACHE_NAME).then((cache) => {
          cache.put(event.request, responseClone)
        })
        return response
      })
      .catch(() => {
        // 网络失败时，尝试从缓存获取
        return caches.match(event.request)
      })
  )
})
```

- [ ] **Step 2: 注册 Service Worker**

在 `web/src/main.js` 中添加注册代码：

```javascript
// web/src/main.js
import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import './style.css'

const app = createApp(App)

app.use(router)
app.use(ElementPlus)
app.mount('#app')

// 注册 Service Worker
if ('serviceWorker' in navigator) {
  window.addEventListener('load', () => {
    navigator.serviceWorker.register('/sw.js')
      .then((registration) => {
        console.log('SW registered:', registration)
      })
      .catch((error) => {
        console.log('SW registration failed:', error)
      })
  })
}
```

- [ ] **Step 3: 提交**

```bash
git add web/public/sw.js web/src/main.js
git commit -m "feat: 创建 Service Worker，支持离线缓存"
```

---

## Task 10: 创建 PWA 安装引导组件

**Files:**
- Create: `web/src/components/PWAInstall.vue`
- Modify: `web/src/App.vue`

- [ ] **Step 1: 创建 PWA 安装引导组件**

```vue
<!-- web/src/components/PWAInstall.vue -->
<script setup>
import { ref, onMounted } from 'vue'

const deferredPrompt = ref(null)
const showInstallPrompt = ref(false)

onMounted(() => {
  window.addEventListener('beforeinstallprompt', (e) => {
    // 阻止 Chrome 自动弹出安装提示
    e.preventDefault()
    // 保存事件，稍后使用
    deferredPrompt.value = e
    // 显示自定义安装提示
    showInstallPrompt.value = true
  })

  window.addEventListener('appinstalled', () => {
    showInstallPrompt.value = false
    deferredPrompt.value = null
    console.log('PWA 已安装')
  })
})

const install = async () => {
  if (!deferredPrompt.value) return

  // 显示安装提示
  deferredPrompt.value.prompt()

  // 等待用户响应
  const { outcome } = await deferredPrompt.value.userChoice
  console.log(`用户选择: ${outcome}`)

  // 清理
  deferredPrompt.value = null
  showInstallPrompt.value = false
}

const dismiss = () => {
  showInstallPrompt.value = false
}
</script>

<template>
  <transition name="slide-up">
    <div v-if="showInstallPrompt" class="install-prompt">
      <div class="install-content">
        <div class="install-icon">☁️</div>
        <div class="install-text">
          <div class="install-title">安装 UCAS</div>
          <div class="install-desc">添加到主屏幕，获得更好的使用体验</div>
        </div>
        <div class="install-actions">
          <el-button type="primary" @click="install">安装</el-button>
          <el-button @click="dismiss">稍后</el-button>
        </div>
      </div>
    </div>
  </transition>
</template>

<style scoped>
.install-prompt {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: var(--bg-secondary);
  border-top: 1px solid var(--border);
  padding: 1rem;
  z-index: 9999;
}

.install-content {
  max-width: 600px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  gap: 1rem;
}

.install-icon {
  font-size: 2.5rem;
}

.install-text {
  flex: 1;
}

.install-title {
  font-weight: 600;
  margin-bottom: 0.25rem;
}

.install-desc {
  font-size: 0.85rem;
  color: var(--text-secondary);
}

.install-actions {
  display: flex;
  gap: 0.5rem;
}

.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.3s ease;
}

.slide-up-enter-from,
.slide-up-leave-to {
  transform: translateY(100%);
  opacity: 0;
}
</style>
```

- [ ] **Step 2: 在 App.vue 中添加安装引导组件**

```vue
<!-- web/src/App.vue -->
<script setup>
import PWAInstall from './components/PWAInstall.vue'
</script>

<template>
  <router-view />
  <PWAInstall />
</template>
```

- [ ] **Step 3: 验证 PWA 功能**

启动开发服务器，验证：
1. manifest.json 正确加载
2. Service Worker 注册成功
3. 安装引导正常显示（需要 HTTPS 或 localhost）

```bash
cd web && npm run dev
```

- [ ] **Step 4: 提交**

```bash
git add web/src/components/PWAInstall.vue web/src/App.vue
git commit -m "feat: 创建 PWA 安装引导组件"
```

---

## 阶段一完成

所有 UI/UX 改进任务已完成，包括：

✅ 侧边栏导航优化（分类分组+可折叠）
✅ 仪表盘增强（图表+任务监控）
✅ 卡片式 UI 组件（账号卡片、任务卡片）
✅ PWA 支持（manifest、Service Worker、安装引导）

**下一步：** 进入阶段二 - 功能扩展（插件系统、Telegram 集成、资源搜索）
