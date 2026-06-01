<script setup>
import { computed } from 'vue'
import { formatSize } from '../../utils/format'

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
  if (storagePercentage.value < 60) return 'var(--color-success)'
  if (storagePercentage.value < 80) return 'var(--color-warning)'
  return 'var(--color-danger)'
})


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
