<template>
  <div>
    <!-- 筛选栏 -->
    <div class="task-filter-bar">
      <el-radio-group v-model="statusFilter" size="default">
        <el-radio-button label="all">全部</el-radio-button>
        <el-radio-button label="pending">等待中</el-radio-button>
        <el-radio-button label="running">运行中</el-radio-button>
        <el-radio-button label="success">成功</el-radio-button>
        <el-radio-button label="failed">失败</el-radio-button>
      </el-radio-group>
      <el-input v-model="searchQuery" placeholder="搜索任务名称..." clearable style="width: 200px" :prefix-icon="PhMagnifyingGlass" />
    </div>

    <!-- 表格视图 -->
    <el-card v-if="viewMode === 'table'" class="table-card">
      <el-table v-if="filteredList.length > 0 || loading" :data="filteredList" v-loading="loading" style="width: 100%">
        <el-table-column label="任务名称" min-width="180">
          <template #default="{ row }">
            <div class="task-name-cell">
              <span class="name">{{ row.name }}</span>
              <div class="account-tag" v-if="row.account">
                <el-tag size="small" :type="row.account.platform === 'quark' ? 'success' : 'warning'">
                  {{ row.account.nickname || row.account.platform }}
                </el-tag>
              </div>
              <div class="account-tag" v-else>
                <el-tag size="small" type="danger">账号已移除</el-tag>
              </div>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="save_path" label="保存路径" min-width="150" show-overflow-tooltip />

        <el-table-column prop="schedule_mode" label="调度规则" width="140">
          <template #default="{ row }">
            <div v-if="row.schedule_mode === 'global'" class="schedule-tag">
              <el-tag size="small" type="primary">跟随全局</el-tag>
              <div class="schedule-sub" v-if="globalSchedule.enabled">{{ globalSchedule.cron }}</div>
              <div class="schedule-sub disabled" v-else>全局已关闭</div>
            </div>
            <div v-else-if="row.schedule_mode === 'custom'" class="schedule-tag">
              <el-tag size="small" type="warning">自定义</el-tag>
              <div class="schedule-sub">{{ row.cron }}</div>
            </div>
            <el-tag v-else size="small" type="info">手动执行</el-tag>
          </template>
        </el-table-column>

        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <div class="status-wrapper">
              <el-tooltip v-if="row.message && row.message.includes('[Fatal]')" :content="row.message" placement="top" effect="dark">
                <el-tag type="danger" style="cursor:help"><div class="status-inner"><el-icon><PhWarning weight="fill" /></el-icon>LINK ERROR</div></el-tag>
              </el-tooltip>
              <el-tooltip v-else-if="row.retry_count > 0 && row.status === 'pending'" :content="`重试 ${row.retry_count}/${row.max_retries} 次`" placement="top" effect="dark">
                <el-tag type="warning"><div class="status-inner"><el-icon class="icon-spin"><PhArrowsClockwise /></el-icon>RETRY {{ row.retry_count }}/{{ row.max_retries }}</div></el-tag>
              </el-tooltip>
              <el-tag v-else :type="getStatusType(row.status)">
                <div class="status-inner"><el-icon v-if="row.status === 'running'" class="icon-spin"><PhArrowsClockwise /></el-icon>{{ row.status.toUpperCase() }}</div>
              </el-tag>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="最后运行" width="180">
          <template #default="{ row }">
            {{ formatTime(row.last_run) }}
          </template>
        </el-table-column>

        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <button
                class="btn-icon btn-icon--success"
                title="运行"
                aria-label="运行"
                :disabled="row.status === 'running' || !!(row.message && row.message.includes('[Fatal]'))"
                @click="emit('run', row)"
              >
                <PhPlay :size="14" />
              </button>
              <button
                class="btn-icon btn-icon--primary"
                title="编辑"
                aria-label="编辑"
                @click="emit('edit', row)"
              >
                <PhPencilSimple :size="14" />
              </button>
              <button
                class="btn-icon btn-icon--danger"
                title="删除"
                aria-label="删除"
                @click="emit('delete', row)"
              >
                <PhTrash :size="14" />
              </button>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-else description="当前没有任何转存任务">
        <el-button type="primary" :icon="PhPlus" @click="emit('add')">创建新任务</el-button>
      </el-empty>
    </el-card>

    <!-- 卡片视图 -->
    <div v-else-if="viewMode === 'card'" class="card-view-container" v-loading="loading">
      <template v-if="filteredList.length > 0 || loading">
        <el-row :gutter="20">
          <el-col v-for="row in filteredList" :key="row.id" :xs="24" :sm="12" :md="8" :lg="6">
            <TaskCard
              :task="row"
              @run="emit('run', $event)"
              @edit="emit('edit', $event)"
              @delete="emit('delete', $event)"
            />
          </el-col>
        </el-row>
      </template>
      <el-empty v-else description="当前没有任何转存任务">
        <el-button type="primary" :icon="PhPlus" @click="emit('add')">创建新任务</el-button>
      </el-empty>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import {
  PhPlus, PhPlay, PhPencilSimple, PhTrash, PhArrowsClockwise,
  PhWarning, PhMagnifyingGlass
} from '@phosphor-icons/vue'
import TaskCard from '../cards/TaskCard.vue'
import { formatTime, getStatusTagType as getStatusType } from '../../utils/format'

const props = defineProps({
  taskList: { type: Array, required: true },
  loading: { type: Boolean, default: false },
  globalSchedule: { type: Object, default: () => ({ enabled: false, cron: '' }) },
  viewMode: { type: String, default: 'table' }
})

const emit = defineEmits(['run', 'edit', 'delete', 'add'])

// 筛选状态
const statusFilter = ref('all')
const searchQuery = ref('')

// 过滤后的任务列表
const filteredList = computed(() => {
  let list = props.taskList
  if (statusFilter.value !== 'all') {
    list = list.filter(t => t.status === statusFilter.value)
  }
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.trim().toLowerCase()
    list = list.filter(t => t.name.toLowerCase().includes(q))
  }
  return list
})
</script>

<style scoped>
.task-filter-bar {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
}

.schedule-tag {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.schedule-sub {
  font-size: 11px;
  color: var(--neutral-500);
  font-family: var(--font-mono);
}

.schedule-sub.disabled {
  color: var(--color-danger);
}

.task-name-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.task-name-cell .name {
  font-weight: 700;
  color: var(--text-primary);
}

.status-inner {
  display: flex;
  align-items: center;
  gap: 6px;
}

.icon-spin {
  animation: rotate 2s linear infinite;
}

@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.card-view-container {
  min-height: 300px;
}
</style>
