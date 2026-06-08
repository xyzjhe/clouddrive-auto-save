<template>
  <div v-loading="pluginsLoading" class="plugins-grid">
    <div v-for="plugin in plugins" :key="plugin.name" class="plugin-card">
      <div class="plugin-header">
        <div class="plugin-icon"><PhPuzzlePiece :size="20" /></div>
        <div class="plugin-info">
          <div class="plugin-name">{{ plugin.name }}</div>
          <div class="plugin-version">v{{ plugin.version }}</div>
        </div>
        <el-switch :model-value="plugin.enabled" @change="handleTogglePlugin(plugin)" />
      </div>
      <div class="plugin-description">{{ plugin.description }}</div>
      <div class="plugin-hooks">
        <el-tag v-for="hook in plugin.hooks" :key="hook" size="small" type="info">{{ hook }}</el-tag>
      </div>
      <div class="plugin-actions">
        <el-button size="small" @click="handleConfigurePlugin(plugin)">配置</el-button>
      </div>
    </div>
    <!-- 安装新插件占位卡片 -->
    <div class="plugin-card add-card" @click="handleInstallPlugin">
      <div class="add-content">
        <el-icon size="36"><PhPlus /></el-icon>
        <div class="add-text">安装新插件</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { PhPuzzlePiece, PhPlus } from '@phosphor-icons/vue'
import request from '../../api/request'
import { ElMessage } from 'element-plus'

const plugins = ref([])
const pluginsLoading = ref(false)

const fetchPlugins = async () => {
  pluginsLoading.value = true
  try {
    const data = await request({ url: '/plugins', method: 'get' })
    plugins.value = data || []
  } catch (error) {
    console.error('获取插件列表失败:', error)
  } finally {
    pluginsLoading.value = false
  }
}

const handleTogglePlugin = async (plugin) => {
  // 仅作 UI 演示，目前为 mock 逻辑
  ElMessage.info(`切换插件 ${plugin.name} 状态功能开发中`)
}

const handleConfigurePlugin = (plugin) => {
  ElMessage.info(`配置插件 ${plugin.name} 功能开发中`)
}

const handleInstallPlugin = () => {
  ElMessage.info('本地插件上传/安装功能开发中')
}

onMounted(() => {
  fetchPlugins()
})
</script>

<style scoped>
/* 插件卡片网格 */
.plugins-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
  padding: 12px;
}

.plugin-card {
  background: var(--surface-bg) !important;
  border: 1px solid var(--border-color) !important;
  border-radius: 12px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  min-height: 180px;
  transition: all 0.3s;
}

.plugin-card:hover {
  border-color: var(--border-color) !important;
  box-shadow: var(--shadow-md) !important;
  transform: translateY(-2px);
}

.plugin-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.plugin-icon {
  width: 40px;
  height: 40px;
  background: rgba(0, 242, 254, 0.1);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.4rem;
}

.plugin-info {
  flex: 1;
}

.plugin-name {
  font-weight: 700;
  font-size: 1.05rem;
  color: var(--text-primary);
}

.plugin-version {
  font-size: 0.8rem;
  color: var(--text-secondary);
}

.plugin-description {
  color: var(--text-secondary);
  font-size: 0.9rem;
  line-height: 1.4;
  margin-bottom: 12px;
  flex: 1;
}

.plugin-hooks {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  margin-bottom: 12px;
}

.plugin-actions {
  display: flex;
  justify-content: flex-end;
}

.add-card {
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px dashed var(--border-color) !important;
  cursor: pointer;
  background: transparent !important;
  box-shadow: none !important;
}

.add-card:hover {
  border-color: var(--accent) !important;
}

.add-content {
  text-align: center;
  color: var(--text-secondary);
}

.add-text {
  margin-top: 8px;
  font-weight: 600;
}
</style>
