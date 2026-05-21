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
