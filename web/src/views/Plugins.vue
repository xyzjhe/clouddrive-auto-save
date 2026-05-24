<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from 'lucide-vue-next'

const plugins = ref([])
const loading = ref(false)

const fetchPlugins = async () => {
  loading.value = true
  try {
    const response = await fetch('/api/plugins')
    const data = await response.json()
    plugins.value = data.data || []
  } catch (error) {
    console.error('获取插件列表失败:', error)
    ElMessage.error('获取插件列表失败')
  } finally {
    loading.value = false
  }
}

const handleToggle = async (plugin) => {
  // TODO: 实现启用/禁用逻辑
  ElMessage.info('功能开发中')
}

const handleConfig = (plugin) => {
  // TODO: 打开配置对话框
  ElMessage.info('功能开发中')
}

onMounted(() => {
  fetchPlugins()
})
</script>

<template>
  <div class="plugins-page">
    <div class="page-header">
      <div class="title-section">
        <h2>插件管理</h2>
        <p>管理系统插件，扩展 UCAS 功能</p>
      </div>
      <el-button type="primary">
        安装插件
      </el-button>
    </div>

    <div
      v-loading="loading"
      class="plugins-grid"
    >
      <div
        v-for="plugin in plugins"
        :key="plugin.name"
        class="plugin-card"
      >
        <div class="plugin-header">
          <div class="plugin-icon">🧩</div>
          <div class="plugin-info">
            <div class="plugin-name">{{ plugin.name }}</div>
            <div class="plugin-version">v{{ plugin.version }}</div>
          </div>
          <el-switch
            :model-value="plugin.enabled"
            @change="handleToggle(plugin)"
          />
        </div>

        <div class="plugin-description">
          {{ plugin.description }}
        </div>

        <div class="plugin-hooks">
          <el-tag
            v-for="hook in plugin.hooks"
            :key="hook"
            size="small"
            type="info"
          >
            {{ hook }}
          </el-tag>
        </div>

        <div class="plugin-actions">
          <el-button
            size="small"
            @click="handleConfig(plugin)"
          >
            配置
          </el-button>
        </div>
      </div>

      <!-- 安装新插件卡片 -->
      <div class="plugin-card add-card">
        <div class="add-content">
          <el-icon size="48"><Plus /></el-icon>
          <div class="add-text">安装新插件</div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.plugins-page {
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

.plugins-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1rem;
}

.plugin-card {
  background: var(--bg-secondary);
  border-radius: 12px;
  padding: 1.25rem;
  box-shadow: var(--shadow-sm);
}

.plugin-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1rem;
}

.plugin-icon {
  width: 48px;
  height: 48px;
  background: var(--brand-500);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
}

.plugin-info {
  flex: 1;
}

.plugin-name {
  font-weight: 600;
  font-size: 1.1rem;
}

.plugin-version {
  font-size: 0.85rem;
  color: var(--text-secondary);
}

.plugin-description {
  color: var(--text-secondary);
  font-size: 0.9rem;
  margin-bottom: 1rem;
}

.plugin-hooks {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  margin-bottom: 1rem;
}

.plugin-actions {
  display: flex;
  justify-content: flex-end;
}

.add-card {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 200px;
  border: 2px dashed var(--border);
  cursor: pointer;
  transition: border-color 0.2s;
}

.add-card:hover {
  border-color: var(--brand-500);
}

.add-content {
  text-align: center;
  color: var(--text-secondary);
}

.add-text {
  margin-top: 0.5rem;
}
</style>
