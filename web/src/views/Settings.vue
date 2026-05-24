<template>
  <div class="settings-container" v-loading="pageLoading">
    <div class="welcome-section">
      <h2>系统设置 ⚙️</h2>
      <p>管理全局调度任务与消息推送配置</p>
    </div>

    <el-row v-if="!pageLoading" :gutter="24">
      <!-- 全局调度设置 -->
      <el-col :xs="24" :lg="12">
        <el-card class="settings-card">
          <template #header>
            <div class="card-header">
              <div class="header-title">
                <el-icon><Calendar /></el-icon>
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
          <div class="card-content">
            <!-- 当前规则摘要 -->
            <div class="schedule-summary" :class="{ 'is-disabled': settings.global_schedule_enabled === 'false' }">
              <el-icon><Info /></el-icon>
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
                <el-input
                  v-model="settings.global_schedule_cron"
                  placeholder="e.g. 0 0 0 * * *"
                >
                  <template #append>
                    <el-tooltip content="Cron 帮助" placement="top">
                      <el-button :icon="Info" @click="showCronHelp" />
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
          </div>
        </el-card>
      </el-col>

      <!-- OpenList 扫描配置 -->
      <el-col :xs="24" :lg="12">
        <el-card class="settings-card">
          <template #header>
            <div class="card-header">
              <div class="header-title">
                <el-icon><Scan /></el-icon>
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
          <div class="card-content">
            <el-form label-position="top">
              <el-form-item label="API 地址">
                <el-input
                  v-model="settings.openlist_api_url"
                  placeholder="http://127.0.0.1:23541"
                />
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
          </div>
        </el-card>
      </el-col>
    </el-row>

  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { Calendar, Info, Scan } from 'lucide-vue-next'
import { getGlobalSettings, updateGlobalSettings, triggerOpenListScan } from '../api/task'
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
const pageLoading = ref(true)

// OpenList 相关状态
const openlistScanning = ref(false)
const savingOpenlist = ref(false)

// 简单的 Cron 转中文描述
const getCronDescription = (cron) => {
  if (!cron) return '未设置'
  const parts = cron.split(' ')
  if (parts.length < 5) return '格式不完整'
  
  // 补齐到 6 位 (秒 分 时 日 月 周)
  const p = parts.length === 5 ? ['0', ...parts] : parts
  
  if (p[3] === '*' && p[4] === '*' && p[5] === '*') {
    return `每天 ${p[2].padStart(2, '0')}:${p[1].padStart(2, '0')}:${p[0].padStart(2, '0')}`
  }
  return cron // 复杂格式直接显示原始字符串
}

const fetchSettings = async () => {
  pageLoading.value = true
  try {
    const data = await getGlobalSettings()
    // 合并默认值
    settings.value = { ...settings.value, ...data }

    // 优先使用持久化的 UI 模式
    if (settings.value.global_schedule_ui_mode) {
      cronMode.value = settings.value.global_schedule_ui_mode
    } else {
      // 降级：通过 Cron 自动推断模式
      const cron = settings.value.global_schedule_cron || '0 0 0 * * *'
      const parts = cron.split(' ')
      if (parts.length >= 3 && parts[0] === '0' && parts[3] === '*' && parts[4] === '*' && parts[5] === '*') {
        cronMode.value = 'daily'
      } else {
        cronMode.value = 'advanced'
      }
    }

    // 初始化时间选择器
    const cron = settings.value.global_schedule_cron || '0 0 0 * * *'
    const parts = cron.split(' ')
    const p = parts.length === 5 ? ['0', ...parts] : parts
    if (p.length >= 3) {
      const d = new Date()
      d.setHours(parseInt(p[2]), parseInt(p[1]), parseInt(p[0]), 0)
      dailyTime.value = d
    }
  } catch (error) {
    ElMessage.error({ message: '加载设置失败', grouping: true })
  } finally {
    pageLoading.value = false
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
    // 错误由响应拦截器统一处理
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
  fetchSettings()
})
</script>

<style scoped>
.settings-container {
  padding: 24px;
}

.welcome-section {
  margin-bottom: 32px;
}

.welcome-section h2 {
  font-size: 26px;
  font-weight: 800;
  margin-bottom: 8px;
  color: var(--neutral-800);
  letter-spacing: -0.02em;
}

.welcome-section p {
  color: var(--neutral-500);
  font-size: 15px;
}

.settings-card {
  margin-bottom: 24px;
  border-radius: 14px;
  border: none;
  box-shadow: var(--shadow-md);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.settings-card:hover {
  box-shadow: var(--shadow-lg);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  min-height: 32px;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 700;
  font-size: 16px;
}

.schedule-summary {
  display: flex;
  align-items: center;
  gap: 8px;
  background-color: var(--brand-50);
  color: var(--brand-600);
  padding: 10px 14px;
  border-radius: 10px;
  margin-bottom: 20px;
  font-size: 13px;
  font-weight: 500;
  border: 1px solid var(--brand-100);
}

.schedule-summary.is-disabled {
  background-color: var(--neutral-100);
  color: var(--neutral-500);
  border-color: var(--neutral-200);
}

html.dark .schedule-summary.is-disabled {
  background-color: rgba(255, 255, 255, 0.03);
  border-color: rgba(255, 255, 255, 0.06);
}

.summary-text {
  flex: 1;
}

.form-tip {
  font-size: 12px;
  color: var(--neutral-500);
  margin-top: 8px;
}

.daily-picker-container {
  width: 100%;
}

.presets-container {
  margin-top: 12px;
}

.form-actions {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
}
</style>
