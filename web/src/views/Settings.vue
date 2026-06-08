<template>
  <div class="settings-container">
    <div class="welcome-section">
      <h2>系统设置</h2>
      <p>集中管理全局任务调度、OpenList 触发、多渠道消息通知及功能扩展插件</p>
    </div>

    <el-tabs v-model="activeTab" class="settings-tabs">
      <!-- Tab 1: 任务调度与扫描 -->
      <el-tab-pane name="schedule">
        <template #label>
          <div class="tab-label-inner">
            <el-icon><PhCalendarBlank /></el-icon>
            <span>系统调度与扫描</span>
          </div>
        </template>
        <ScheduleSettings />
      </el-tab-pane>

      <!-- Tab 2: 消息通知通道 -->
      <el-tab-pane name="notify">
        <template #label>
          <div class="tab-label-inner">
            <el-icon><PhBell /></el-icon>
            <span>消息推送通道</span>
          </div>
        </template>
        <NotifySettings />
      </el-tab-pane>

      <!-- Tab 3: 系统扩展插件 -->
      <el-tab-pane name="plugins">
        <template #label>
          <div class="tab-label-inner">
            <el-icon><PhPuzzlePiece weight="duotone" /></el-icon>
            <span>功能扩展插件</span>
          </div>
        </template>
        <PluginSettings />
      </el-tab-pane>

      <!-- Tab 4: 搜索源配置 -->
      <el-tab-pane name="search">
        <template #label>
          <div class="tab-label-inner">
            <el-icon><PhMagnifyingGlass /></el-icon>
            <span>搜索源</span>
          </div>
        </template>
        <SearchSettings />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { PhCalendarBlank, PhBell, PhPuzzlePiece, PhMagnifyingGlass } from '@phosphor-icons/vue'
import ScheduleSettings from '../components/settings/ScheduleSettings.vue'
import NotifySettings from '../components/settings/NotifySettings.vue'
import PluginSettings from '../components/settings/PluginSettings.vue'
import SearchSettings from '../components/settings/SearchSettings.vue'

const activeTab = ref('schedule')
</script>

<style scoped>
.settings-container {
  padding: 8px 0;
}

.welcome-section {
  margin-bottom: 24px;
}

.welcome-section h2 {
  font-size: 26px;
  font-weight: 800;
  margin-bottom: 8px;
  color: var(--text-primary);
  letter-spacing: -0.02em;
}

.welcome-section p {
  color: var(--text-secondary);
  font-size: 15px;
}

/* 主 Tab 下划线风格 */
.settings-tabs {
  background: var(--surface-bg) !important;
  border: 1px solid var(--border-color) !important;
  border-radius: 16px !important;
  overflow: hidden;
}

.settings-tabs :deep(.el-tabs__header) {
  border-bottom: 1px solid var(--border-color) !important;
  background: transparent !important;
}

.settings-tabs :deep(.el-tabs__item) {
  color: var(--text-secondary) !important;
  font-weight: 500 !important;
  border: none !important;
  transition: all 0.3s;
  height: 52px;
}

.settings-tabs :deep(.el-tabs__item.is-active) {
  color: var(--accent) !important;
}

.settings-tabs :deep(.el-tabs__active-bar) {
  background-color: var(--accent) !important;
}

.settings-tabs :deep(.el-tabs__nav-wrap::after) {
  display: none !important;
}

.tab-label-inner {
  display: flex;
  align-items: center;
  gap: 8px;
}

/* 确保同一行内的卡片等高 */
.settings-tabs :deep(.el-row > .el-col) {
  display: flex;
}

/* 让卡片 body 和 form 纵向 flex，使 form-actions 吸附底部 */
.settings-tabs :deep(.inner-settings-card .el-card__body) {
  display: flex;
  flex-direction: column;
  flex: 1;
}

.settings-tabs :deep(.inner-settings-card .el-form) {
  flex: 1;
  display: flex;
  flex-direction: column;
}
</style>
