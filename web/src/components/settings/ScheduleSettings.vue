<template>
  <el-row :gutter="24">
    <!-- 全局调度 -->
    <el-col :xs="24" :lg="12">
      <el-card class="inner-settings-card">
        <template #header>
          <div class="card-header">
            <div class="header-title">
              <el-icon><PhCalendarBlank /></el-icon>
              <span>全局定时任务</span>
            </div>
            <el-switch
              v-model="settings.global_schedule_enabled"
              active-value="true"
              inactive-value="false"
              @change="() => saveGlobalSchedule(false)"
            />
          </div>
        </template>

        <div class="schedule-summary" :class="{ 'is-disabled': settings.global_schedule_enabled === 'false' }">
          <el-icon><PhInfo /></el-icon>
          <span class="summary-text">
            当前设定：{{ settings.global_schedule_enabled === 'true' ? getCronDescription(settings.global_schedule_cron) : '未开启全局调度' }}
          </span>
        </div>

        <el-form label-position="top">
          <el-form-item label="配置模式">
            <el-radio-group v-model="cronMode" size="small">
              <el-radio-button label="daily">简易定时</el-radio-button>
              <el-radio-button label="advanced">高级 Cron</el-radio-button>
            </el-radio-group>
          </el-form-item>

          <el-form-item v-if="cronMode === 'daily'" label="每天运行时间">
            <div class="daily-picker-container">
              <el-time-picker
                v-model="dailyTime"
                format="HH:mm"
                placeholder="选择时间"
                @change="handleTimeChange"
                style="width: 100%"
              />
              <div class="presets-container">
                <el-button-group size="small">
                  <el-button @click="setPreset('00:00')">凌晨</el-button>
                  <el-button @click="setPreset('08:00')">早晨</el-button>
                  <el-button @click="setPreset('12:00')">中午</el-button>
                </el-button-group>
              </div>
            </div>
          </el-form-item>

          <el-form-item v-else label="全局 Cron 表达式">
            <el-input v-model="settings.global_schedule_cron" placeholder="e.g. 0 0 0 * * *">
              <template #append>
                <el-tooltip content="Cron 帮助" placement="top">
                  <el-button :icon="PhInfo" @click="showCronHelp" />
                </el-tooltip>
              </template>
            </el-input>
          </el-form-item>

          <div class="form-tip">
            设置全局默认运行时间，个别任务可单独重写此设置。
          </div>

          <div class="form-actions">
            <el-button type="primary" :loading="savingSchedule" @click="saveGlobalSchedule(true)">
              保存配置
            </el-button>
          </div>
        </el-form>
      </el-card>
    </el-col>

    <!-- OpenList 扫描 -->
    <el-col :xs="24" :lg="12">
      <el-card class="inner-settings-card">
        <template #header>
          <div class="card-header">
            <div class="header-title">
              <el-icon><PhArrowsClockwise /></el-icon>
              <span>OpenList 扫描</span>
            </div>
            <el-switch
              v-model="settings.openlist_enabled"
              active-value="true"
              inactive-value="false"
              @change="() => saveOpenListSettings(false)"
            />
          </div>
        </template>

        <el-form label-position="top">
          <el-form-item label="API 地址">
            <el-input v-model="settings.openlist_api_url" placeholder="http://127.0.0.1:23541" />
          </el-form-item>
          <el-form-item label="API Token">
            <el-input
              v-model="settings.openlist_api_token"
              placeholder="openlist-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
              type="password"
              show-password
            />
          </el-form-item>

          <div class="form-tip">
            配置 OpenList API 信息后，转存任务完成时将自动触发扫描。也可手动点击按钮触发。
          </div>

          <div class="form-actions">
            <el-button
              type="primary"
              plain
              :loading="openlistScanning"
              @click="handleOpenListScan"
              style="margin-right: 12px"
            >
              手动扫描
            </el-button>
            <el-button type="primary" :loading="savingOpenlist" @click="saveOpenListSettings(true)">
              保存配置
            </el-button>
          </div>
        </el-form>
      </el-card>
    </el-col>
  </el-row>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { PhCalendarBlank, PhInfo, PhArrowsClockwise } from '@phosphor-icons/vue'
import { getGlobalSettings, updateGlobalSettings, triggerOpenListScan } from '../../api/task'
import { ElMessage, ElMessageBox } from 'element-plus'

const settings = ref({
  global_schedule_enabled: 'false',
  global_schedule_cron: '0 0 0 * * *',
  global_schedule_ui_mode: 'daily',
  openlist_enabled: 'false',
  openlist_api_url: '',
  openlist_api_token: ''
})

const cronMode = ref('daily')
const dailyTime = ref(new Date(new Date().setHours(0, 0, 0, 0)))
const savingSchedule = ref(false)
const isProcessing = ref(false)
const openlistScanning = ref(false)
const savingOpenlist = ref(false)

const getCronDescription = (cron) => {
  if (!cron) return '未设置'
  const parts = cron.split(' ')
  if (parts.length < 5) return '格式不完整'
  const p = parts.length === 5 ? ['0', ...parts] : parts
  if (p[3] === '*' && p[4] === '*' && p[5] === '*') {
    return `每天 ${p[2].padStart(2, '0')}:${p[1].padStart(2, '0')}:${p[0].padStart(2, '0')}`
  }
  return cron
}

const fetchScheduleSettings = async () => {
  try {
    const data = await getGlobalSettings()
    // 仅合并本组件关心的 key，避免引入 bark_* 等无关 key 导致保存时白名单校验失败
    const allowedKeys = ['global_schedule_enabled', 'global_schedule_cron', 'global_schedule_ui_mode', 'openlist_enabled', 'openlist_api_url', 'openlist_api_token']
    for (const key of allowedKeys) {
      if (data[key] !== undefined) {
        settings.value[key] = data[key]
      }
    }

    if (settings.value.global_schedule_ui_mode) {
      cronMode.value = settings.value.global_schedule_ui_mode
    } else {
      const cron = settings.value.global_schedule_cron || '0 0 0 * * *'
      const parts = cron.split(' ')
      if (parts.length >= 3 && parts[0] === '0' && parts[3] === '*' && parts[4] === '*' && parts[5] === '*') {
        cronMode.value = 'daily'
      } else {
        cronMode.value = 'advanced'
      }
    }

    const cron = settings.value.global_schedule_cron || '0 0 0 * * *'
    const parts = cron.split(' ')
    const p = parts.length === 5 ? ['0', ...parts] : parts
    if (p.length >= 3) {
      const d = new Date()
      d.setHours(parseInt(p[2]), parseInt(p[1]), parseInt(p[0]), 0)
      dailyTime.value = d
    }
  } catch (error) {
    console.error('加载系统调度设置失败:', error)
  }
}

const handleTimeChange = (val) => {
  if (!val) return
  const h = val.getHours()
  const m = val.getMinutes()
  const s = val.getSeconds()
  settings.value.global_schedule_cron = `${s} ${m} ${h} * * *`
}

watch(cronMode, (newMode) => {
  if (newMode === 'daily') {
    handleTimeChange(dailyTime.value)
  }
})

const setPreset = (timeStr) => {
  const [h, m] = timeStr.split(':').map(Number)
  const d = new Date()
  d.setHours(h, m, 0, 0)
  dailyTime.value = d
  handleTimeChange(d)
}

const saveGlobalSchedule = async (manual = false) => {
  if (isProcessing.value) return
  isProcessing.value = true
  if (manual) savingSchedule.value = true
  settings.value.global_schedule_ui_mode = cronMode.value
  try {
    await updateGlobalSettings(settings.value)
    if (manual) ElMessage.success('全局调度设置已保存')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '保存失败')
  } finally {
    isProcessing.value = false
    savingSchedule.value = false
  }
}

const showCronHelp = () => {
  ElMessageBox.alert(
    'Cron 表达式由 5 或 6 个字段组成：<br/>秒 分 时 日 月 周<br/>例如：<br/><b>0 0 0 * * *</b> (每天凌晨)<br/><b>0 30 15 * * *</b> (每天 15:30:00)',
    'Cron 帮助',
    { dangerouslyUseHTMLString: true }
  )
}

const handleOpenListScan = async () => {
  openlistScanning.value = true
  try {
    await triggerOpenListScan()
    ElMessage.success('OpenList 扫描已触发')
  } catch {
    // 错误统一由拦截器处理
  } finally {
    openlistScanning.value = false
  }
}

const saveOpenListSettings = async (manual = false) => {
  if (isProcessing.value) return
  isProcessing.value = true
  if (manual) savingOpenlist.value = true
  try {
    await updateGlobalSettings(settings.value)
    if (manual) ElMessage.success('OpenList 扫描设置已保存')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '保存失败')
  } finally {
    isProcessing.value = false
    savingOpenlist.value = false
  }
}

onMounted(() => {
  fetchScheduleSettings()
})
</script>

<style scoped>
.inner-settings-card {
  background: var(--surface-bg) !important;
  box-shadow: none !important;
  border: 1px solid var(--border-color) !important;
  margin-bottom: 16px;
  border-radius: 12px !important;
  flex: 1;
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
  font-size: 16px;
  color: var(--text-primary);
}

.schedule-summary {
  display: flex;
  align-items: center;
  gap: 8px;
  background-color: var(--accent-light);
  color: var(--accent);
  padding: 10px 14px;
  border-radius: 10px;
  margin-bottom: 20px;
  font-size: 13px;
  font-weight: 500;
  border: 1px solid var(--border-color);
}

.schedule-summary.is-disabled {
  background-color: rgba(255, 255, 255, 0.04);
  color: var(--text-muted);
  border-color: var(--border-color);
}

.summary-text {
  flex: 1;
}

.form-tip {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 8px;
}

.daily-picker-container {
  width: 100%;
}

.presets-container {
  margin-top: 12px;
}

.form-actions {
  margin-top: auto;
  padding-top: 24px;
  display: flex;
  justify-content: flex-end;
}
</style>
