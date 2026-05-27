<template>
  <div class="settings-container" v-loading="pageLoading">
    <div class="welcome-section">
      <h2>系统设置 ⚙️</h2>
      <p>集中管理全局任务调度、OpenList 触发、多渠道消息通知及功能扩展插件</p>
    </div>

    <el-tabs v-model="activeTab" type="border-card" class="settings-tabs glass-card">
      <!-- Tab 1: 任务调度与扫描 -->
      <el-tab-pane name="schedule">
        <template #label>
          <div class="tab-label-inner">
            <el-icon><Calendar /></el-icon>
            <span>系统调度与扫描</span>
          </div>
        </template>

        <el-row :gutter="24">
          <!-- 全局调度 -->
          <el-col :xs="24" :lg="12">
            <el-card class="inner-settings-card">
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
                  <el-input v-model="settings.global_schedule_cron" placeholder="e.g. 0 0 0 * * *">
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
            </el-card>
          </el-col>

          <!-- OpenList 扫描 -->
          <el-col :xs="24" :lg="12">
            <el-card class="inner-settings-card">
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
      </el-tab-pane>

      <!-- Tab 2: 消息通知通道 -->
      <el-tab-pane name="notify">
        <template #label>
          <div class="tab-label-inner">
            <el-icon><Bell /></el-icon>
            <span>消息推送通道</span>
          </div>
        </template>

        <div class="notify-tabs-wrapper">
          <el-tabs v-model="activeNotifyTab" type="border-card" class="nested-tabs">
            <!-- 企业微信 -->
            <el-tab-pane label="企业微信" name="wechat">
              <el-form :model="wechatConfig" label-width="120px" class="wechat-form">
                <el-form-item label="启用">
                  <el-switch v-model="wechatConfig.enabled" />
                </el-form-item>
                <el-form-item label="Webhook URL">
                  <el-input
                    v-model="wechatConfig.config.webhook_url"
                    placeholder="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=..."
                  />
                </el-form-item>
                <el-form-item label="通知设置">
                  <el-checkbox v-model="wechatConfig.notify_on_success">成功通知</el-checkbox>
                  <el-checkbox v-model="wechatConfig.notify_on_failure">失败通知</el-checkbox>
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" class="save-wechat-btn" @click="handleSaveNotify('wechat')">保存</el-button>
                  <el-button class="test-wechat-btn" @click="handleTestNotify('wechat')">测试</el-button>
                </el-form-item>
              </el-form>
            </el-tab-pane>

            <!-- Telegram -->
            <el-tab-pane label="Telegram" name="telegram">
              <el-form :model="telegramConfig" label-width="120px" class="telegram-form">
                <el-form-item label="启用">
                  <el-switch v-model="telegramConfig.enabled" />
                </el-form-item>
                <el-form-item label="Bot Token">
                  <el-input
                    v-model="telegramConfig.config.bot_token"
                    placeholder="123456789:ABCdefGHIjklMNOpqrsTUVwxyz"
                  />
                </el-form-item>
                <el-form-item label="Chat ID">
                  <el-input v-model="telegramConfig.config.chat_id" placeholder="123456789" />
                </el-form-item>
                <el-form-item label="通知设置">
                  <el-checkbox v-model="telegramConfig.notify_on_success">成功通知</el-checkbox>
                  <el-checkbox v-model="telegramConfig.notify_on_failure">失败通知</el-checkbox>
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" class="save-telegram-btn" @click="handleSaveNotify('telegram')">保存</el-button>
                  <el-button class="test-telegram-btn" @click="handleTestNotify('telegram')">测试</el-button>
                </el-form-item>
              </el-form>
            </el-tab-pane>

            <!-- WxPusher -->
            <el-tab-pane label="WxPusher" name="wxpusher">
              <el-form :model="wxpusherConfig" label-width="120px" class="wxpusher-form">
                <el-form-item label="启用">
                  <el-switch v-model="wxpusherConfig.enabled" />
                </el-form-item>
                <el-form-item label="App Token">
                  <el-input v-model="wxpusherConfig.config.app_token" placeholder="AT_xxx" />
                </el-form-item>
                <el-form-item label="UID">
                  <el-input v-model="wxpusherConfig.config.uid" placeholder="UID_xxx" />
                </el-form-item>
                <el-form-item label="通知设置">
                  <el-checkbox v-model="wxpusherConfig.notify_on_success">成功通知</el-checkbox>
                  <el-checkbox v-model="wxpusherConfig.notify_on_failure">失败通知</el-checkbox>
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" class="save-wxpusher-btn" @click="handleSaveNotify('wxpusher')">保存</el-button>
                  <el-button class="test-wxpusher-btn" @click="handleTestNotify('wxpusher')">测试</el-button>
                </el-form-item>
              </el-form>
            </el-tab-pane>

            <!-- Bark -->
            <el-tab-pane label="Bark" name="bark">
              <el-form :model="barkConfig" label-width="120px" class="bark-form">
                <el-form-item label="启用">
                  <el-switch v-model="barkConfig.bark_enabled" active-value="true" inactive-value="false" />
                </el-form-item>
                <el-form-item label="服务器地址">
                  <el-input v-model="barkConfig.bark_server" placeholder="https://api.day.app" />
                </el-form-item>
                <el-form-item label="Device Key">
                  <el-input
                    v-model="barkConfig.bark_device_key"
                    placeholder="您的 Bark 设备 Key"
                    type="password"
                    show-password
                  />
                </el-form-item>

                <!-- Bark 高级配置折叠 -->
                <el-collapse class="advanced-collapse" style="margin-bottom: 22px; margin-left: 120px; max-width: 600px;">
                  <el-collapse-item name="1">
                    <template #title>
                      <span class="collapse-title">高级推送设置</span>
                    </template>
                    <el-form-item label="自定义图标 URL" label-width="140px">
                      <el-input v-model="barkConfig.bark_icon" placeholder="https://example.com/icon.png" />
                    </el-form-item>
                    <el-form-item label="自动保存历史" label-width="140px">
                      <el-switch v-model="barkConfig.bark_archive" active-value="true" inactive-value="false" />
                    </el-form-item>
                    <el-divider content-position="left">成功通知配置</el-divider>
                    <el-row :gutter="12" style="margin-left: 20px;">
                      <el-col :span="12">
                        <el-form-item label="通知级别" label-width="80px">
                          <el-select v-model="barkConfig.bark_success_level" placeholder="选择级别" style="width: 100%">
                            <el-option v-for="l in barkLevels" :key="l.value" :label="l.label" :value="l.value" />
                          </el-select>
                        </el-form-item>
                      </el-col>
                      <el-col :span="12">
                        <el-form-item label="提醒铃声" label-width="80px">
                          <el-select v-model="barkConfig.bark_success_sound" placeholder="选择铃声" style="width: 100%">
                            <el-option v-for="s in barkSounds" :key="s.value" :label="s.label" :value="s.value" />
                          </el-select>
                        </el-form-item>
                      </el-col>
                    </el-row>
                    <el-divider content-position="left">失败通知配置</el-divider>
                    <el-row :gutter="12" style="margin-left: 20px;">
                      <el-col :span="12">
                        <el-form-item label="通知级别" label-width="80px">
                          <el-select v-model="barkConfig.bark_failure_level" placeholder="选择级别" style="width: 100%">
                            <el-option v-for="l in barkLevels" :key="l.value" :label="l.label" :value="l.value" />
                          </el-select>
                        </el-form-item>
                      </el-col>
                      <el-col :span="12">
                        <el-form-item label="提醒铃声" label-width="80px">
                          <el-select v-model="barkConfig.bark_failure_sound" placeholder="选择铃声" style="width: 100%">
                            <el-option v-for="s in barkSounds" :key="s.value" :label="s.label" :value="s.value" />
                          </el-select>
                        </el-form-item>
                      </el-col>
                    </el-row>
                  </el-collapse-item>
                </el-collapse>

                <el-form-item>
                  <el-button type="primary" class="save-bark-btn" :loading="savingBark" @click="handleSaveNotify('bark')">保存</el-button>
                  <el-button class="test-bark-btn" @click="handleTestNotify('bark')">测试</el-button>
                </el-form-item>
              </el-form>
            </el-tab-pane>
          </el-tabs>
        </div>
      </el-tab-pane>

      <!-- Tab 3: 系统扩展插件 -->
      <el-tab-pane name="plugins">
        <template #label>
          <div class="tab-label-inner">
            <el-icon><Puzzle /></el-icon>
            <span>功能扩展插件</span>
          </div>
        </template>

        <div v-loading="pluginsLoading" class="plugins-grid">
          <div v-for="plugin in plugins" :key="plugin.name" class="plugin-card glass-card">
            <div class="plugin-header">
              <div class="plugin-icon">🧩</div>
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
              <el-icon size="36"><Plus /></el-icon>
              <div class="add-text">安装新插件</div>
            </div>
          </div>
        </div>
      </el-tab-pane>

      <!-- Tab: 搜索源配置 -->
      <el-tab-pane name="search">
        <template #label>
          <div class="tab-label-inner">
            <el-icon><Search /></el-icon>
            <span>搜索源</span>
          </div>
        </template>

        <el-row :gutter="24">
          <el-col :xs="24" :lg="12">
            <el-card class="inner-settings-card">
              <template #header>
                <div class="card-header">
                  <div class="header-title">
                    <span>CloudSaver 配置</span>
                  </div>
                </div>
              </template>
              <el-form label-position="top">
                <el-form-item label="服务地址">
                  <el-input v-model="searchConfig.cloudsaver.server" placeholder="http://localhost:8080" />
                </el-form-item>
                <el-form-item label="用户名">
                  <el-input v-model="searchConfig.cloudsaver.username" placeholder="用户名" />
                </el-form-item>
                <el-form-item label="密码">
                  <el-input v-model="searchConfig.cloudsaver.password" type="password" show-password placeholder="密码" />
                </el-form-item>
                <el-form-item label="Token 状态">
                  <el-tag :type="searchConfig.cloudsaver.token ? 'success' : 'info'">
                    {{ searchConfig.cloudsaver.token ? '已获取' : '未获取' }}
                  </el-tag>
                </el-form-item>
              </el-form>
            </el-card>
          </el-col>

          <el-col :xs="24" :lg="12">
            <el-card class="inner-settings-card">
              <template #header>
                <div class="card-header">
                  <div class="header-title">
                    <span>PanSou 配置</span>
                  </div>
                </div>
              </template>
              <el-form label-position="top">
                <el-form-item label="服务地址">
                  <el-input v-model="searchConfig.pansou.server" placeholder="https://so.252035.xyz" />
                </el-form-item>
              </el-form>
            </el-card>
          </el-col>
        </el-row>

        <div style="text-align: right; margin-top: 16px;">
          <el-button type="primary" @click="saveSearchConfig" :loading="searchConfigSaving">
            保存配置
          </el-button>
        </div>
      </el-tab-pane>
    </el-tabs>

    <!-- Bark 测试对话框 -->
    <el-dialog v-model="barkTestDialogVisible" title="发送测试推送" width="400px" append-to-body>
      <el-form :model="barkTestForm" label-position="top">
        <el-form-item label="推送标题">
          <el-input v-model="barkTestForm.title" />
        </el-form-item>
        <el-form-item label="推送内容">
          <el-input v-model="barkTestForm.body" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="barkTestDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="barkTesting" @click="handleSendBarkTest">立即发送</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { Calendar, Info, Scan, Bell, Puzzle, Plus, Search } from 'lucide-vue-next'
import { getGlobalSettings, updateGlobalSettings, triggerOpenListScan, testBark } from '../api/task'
import { getSearchConfig, updateSearchConfig } from '../api/search'
import request from '../api/request'
import { ElMessage, ElMessageBox } from 'element-plus'

const activeTab = ref('schedule')
const pageLoading = ref(true)
const searchConfig = ref({
  cloudsaver: { server: '', username: '', password: '', token: '' },
  pansou: { server: '' }
})
const searchConfigSaving = ref(false)

// ==================== 1. 任务调度与扫描 ====================
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
    settings.value = { ...settings.value, ...data }

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

// ==================== 2. 消息通知通道 ====================
const activeNotifyTab = ref('wechat')
const savingBark = ref(false)
const barkTestDialogVisible = ref(false)
const barkTesting = ref(false)

const wechatConfig = ref({
  enabled: false,
  notify_on_success: true,
  notify_on_failure: true,
  config: { webhook_url: '' }
})

const telegramConfig = ref({
  enabled: false,
  notify_on_success: true,
  notify_on_failure: true,
  config: { bot_token: '', chat_id: '' }
})

const wxpusherConfig = ref({
  enabled: false,
  notify_on_success: true,
  notify_on_failure: true,
  config: { app_token: '', uid: '' }
})

const barkConfig = ref({
  bark_enabled: 'false',
  bark_server: 'https://api.day.app',
  bark_device_key: '',
  bark_success_sound: 'birdsong.caf',
  bark_success_level: 'active',
  bark_failure_sound: 'alarm.caf',
  bark_failure_level: 'timeSensitive',
  bark_archive: 'true',
  bark_icon: ''
})

const barkTestForm = ref({
  title: 'UCAS 测试通知',
  body: '这是一条来自系统设置的消息。',
  level: 'active',
  sound: 'birdsong.caf',
  icon: '',
  isArchive: 'true'
})

const barkLevels = [
  { label: '活跃 (默认)', value: 'active' },
  { label: '时效性 (专注模式可见)', value: 'timeSensitive' },
  { label: '静默', value: 'passive' },
  { label: '告警 (忽略静音)', value: 'critical' }
]

const barkSounds = [
  { label: '清脆鸟鸣 (birdsong.caf)', value: 'birdsong.caf' },
  { label: '警示音 (alarm.caf)', value: 'alarm.caf' },
  { label: '小步舞曲 (minuet.caf)', value: 'minuet.caf' },
  { label: '经典电铃 (bell.caf)', value: 'bell.caf' },
  { label: '默认 (系统)', value: 'default' }
]

const fetchBarkSettings = async () => {
  try {
    const data = await getGlobalSettings()
    Object.keys(barkConfig.value).forEach(key => {
      if (data[key] !== undefined) {
        barkConfig.value[key] = data[key]
      }
    })
    if (barkConfig.value.bark_success_sound === '') barkConfig.value.bark_success_sound = 'default'
    if (barkConfig.value.bark_failure_sound === '') barkConfig.value.bark_failure_sound = 'default'
    if (barkConfig.value.bark_archive === undefined) barkConfig.value.bark_archive = 'true'
  } catch (error) {
    console.error('加载 Bark 设置失败:', error)
  }
}

const fetchNotifierConfig = async (type) => {
  try {
    const config = await request({ url: `/notify/${type}`, method: 'get' })
    const targetConfig = {
      enabled: config.enabled || false,
      notify_on_success: config.notify_on_success !== false,
      notify_on_failure: config.notify_on_failure !== false,
      config: config.config || {}
    }
    if (type === 'wechat') {
      wechatConfig.value = targetConfig
    } else if (type === 'telegram') {
      telegramConfig.value = targetConfig
    } else if (type === 'wxpusher') {
      wxpusherConfig.value = targetConfig
    }
  } catch (error) {
    console.error(`加载 ${type} 配置失败:`, error)
  }
}

const handleSaveNotify = async (type) => {
  if (type === 'bark') {
    savingBark.value = true
    try {
      await updateGlobalSettings(barkConfig.value)
      ElMessage.success('Bark 推送设置已保存')
    } catch (error) {
      ElMessage.error(error.response?.data?.error || '保存失败')
    } finally {
      savingBark.value = false
    }
    return
  }

  let config = {}
  if (type === 'wechat') config = wechatConfig.value
  else if (type === 'telegram') config = telegramConfig.value
  else if (type === 'wxpusher') config = wxpusherConfig.value

  try {
    await request({ url: `/notify/${type}`, method: 'put', data: config })
    ElMessage.success(`${type} 配置已保存`)
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

const handleTestNotify = async (type) => {
  if (type === 'bark') {
    if (!barkConfig.value.bark_device_key) {
      ElMessage.warning('请先填写 Device Key')
      return
    }
    barkTestForm.value.level = barkConfig.value.bark_success_level || 'active'
    barkTestForm.value.sound = barkConfig.value.bark_success_sound || 'birdsong.caf'
    barkTestForm.value.icon = barkConfig.value.bark_icon || ''
    barkTestForm.value.isArchive = barkConfig.value.bark_archive || 'true'
    barkTestDialogVisible.value = true
    return
  }

  try {
    await request({ url: `/notify/${type}/test`, method: 'post' })
    ElMessage.success('测试消息已发送，请检查接收设备')
  } catch {
    ElMessage.error('测试发送失败')
  }
}

const handleSendBarkTest = async () => {
  barkTesting.value = true
  try {
    await testBark({
      bark_server: barkConfig.value.bark_server,
      bark_device_key: barkConfig.value.bark_device_key,
      ...barkTestForm.value
    })
    ElMessage.success('测试消息已发送，请检查手机')
    barkTestDialogVisible.value = false
  } catch (error) {
    ElMessage.error('测试发送失败: ' + (error.response?.data?.error || error.message))
  } finally {
    barkTesting.value = false
  }
}

// ==================== 3. 扩展插件管理 ====================
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

// ==================== 4. 搜索源配置 ====================
const loadSearchConfig = async () => {
  try {
    const data = await getSearchConfig()
    if (data) {
      searchConfig.value = data
    }
  } catch (e) {
    console.error('加载搜索配置失败:', e)
  }
}

const saveSearchConfig = async () => {
  searchConfigSaving.value = true
  try {
    await updateSearchConfig(searchConfig.value)
    ElMessage.success('搜索配置已保存')
  } catch (e) {
    console.error('保存搜索配置失败:', e)
  } finally {
    searchConfigSaving.value = false
  }
}

// ==================== 初始化挂载 ====================
onMounted(async () => {
  pageLoading.value = true
  await Promise.all([
    fetchScheduleSettings(),
    fetchBarkSettings(),
    fetchNotifierConfig('wechat'),
    fetchNotifierConfig('telegram'),
    fetchNotifierConfig('wxpusher'),
    fetchPlugins(),
    loadSearchConfig()
  ])
  pageLoading.value = false
})
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

.settings-tabs {
  background: var(--surface-bg) !important;
  border: 1px solid var(--border-color) !important;
  border-radius: 16px !important;
  overflow: hidden;
}

:deep(.el-tabs__header) {
  background: rgba(255, 255, 255, 0.02) !important;
  border-bottom: 1px solid var(--border-color) !important;
}

:deep(.el-tabs__item) {
  color: var(--text-secondary) !important;
  font-weight: 600;
  transition: all 0.3s;
  height: 52px;
}

:deep(.el-tabs__item.is-active) {
  color: var(--neon-teal) !important;
  background: rgba(255, 255, 255, 0.04) !important;
}

.tab-label-inner {
  display: flex;
  align-items: center;
  gap: 8px;
}

.inner-settings-card {
  background: transparent !important;
  box-shadow: none !important;
  border: 1px solid var(--border-color) !important;
  margin-bottom: 16px;
  border-radius: 12px !important;
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
  background-color: rgba(0, 242, 254, 0.08);
  color: var(--neon-teal);
  padding: 10px 14px;
  border-radius: 10px;
  margin-bottom: 20px;
  font-size: 13px;
  font-weight: 500;
  border: 1px solid rgba(0, 242, 254, 0.15);
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
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
}

/* 消息推送嵌套选项卡 */
.notify-tabs-wrapper {
  padding: 8px;
}

.nested-tabs {
  background: transparent !important;
  border: none !important;
  box-shadow: none !important;
}

:deep(.nested-tabs .el-tabs__header) {
  border-bottom: 1px solid var(--border-color) !important;
  background: transparent !important;
}

:deep(.nested-tabs .el-tabs__item) {
  height: 44px;
}

.advanced-collapse {
  margin-top: 20px;
  border: none !important;
}

:deep(.advanced-collapse .el-collapse-item__header) {
  height: 40px;
  border-bottom: none;
  background-color: rgba(255, 255, 255, 0.03);
  padding: 0 12px;
  border-radius: 8px;
}

:deep(.advanced-collapse .el-collapse-item__wrap) {
  border-bottom: none;
  padding: 12px;
}

.collapse-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
}

/* 插件卡片网络 */
.plugins-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
  padding: 12px;
}

.plugin-card {
  background: rgba(255, 255, 255, 0.02) !important;
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
  border-color: rgba(0, 242, 254, 0.25) !important;
  box-shadow: var(--neon-glow-teal) !important;
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
  border-color: var(--neon-teal) !important;
}

.add-content {
  text-align: center;
  color: var(--text-secondary);
}

.add-text {
  margin-top: 8px;
  font-weight: 600;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
