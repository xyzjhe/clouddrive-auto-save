<template>
  <div class="notify-tabs-wrapper">
    <el-tabs v-model="activeNotifyTab" class="nested-tabs">
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
            <el-switch v-model="barkConfig.enabled" />
          </el-form-item>
          <el-form-item label="通知设置">
            <el-checkbox v-model="barkConfig.notify_on_success">成功通知</el-checkbox>
            <el-checkbox v-model="barkConfig.notify_on_failure">失败通知</el-checkbox>
          </el-form-item>
          <el-form-item label="服务器地址">
            <el-input v-model="barkConfig.config.server" placeholder="https://api.day.app" />
          </el-form-item>
          <el-form-item label="Device Key">
            <el-input
              v-model="barkConfig.config.device_key"
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
                <el-input v-model="barkConfig.config.icon" placeholder="https://example.com/icon.png" />
              </el-form-item>
              <el-form-item label="自动保存历史" label-width="140px">
                <el-switch v-model="barkConfig.config.archive" active-value="true" inactive-value="false" />
              </el-form-item>
              <el-divider content-position="left">成功通知配置</el-divider>
              <el-row :gutter="12" style="margin-left: 20px;">
                <el-col :span="12">
                  <el-form-item label="通知级别" label-width="80px">
                    <el-select v-model="barkConfig.config.success_level" placeholder="选择级别" style="width: 100%">
                      <el-option v-for="l in barkLevels" :key="l.value" :label="l.label" :value="l.value" />
                    </el-select>
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="提醒铃声" label-width="80px">
                    <el-select v-model="barkConfig.config.success_sound" placeholder="选择铃声" style="width: 100%">
                      <el-option v-for="s in barkSounds" :key="s.value" :label="s.label" :value="s.value" />
                    </el-select>
                  </el-form-item>
                </el-col>
              </el-row>
              <el-divider content-position="left">失败通知配置</el-divider>
              <el-row :gutter="12" style="margin-left: 20px;">
                <el-col :span="12">
                  <el-form-item label="通知级别" label-width="80px">
                    <el-select v-model="barkConfig.config.failure_level" placeholder="选择级别" style="width: 100%">
                      <el-option v-for="l in barkLevels" :key="l.value" :label="l.label" :value="l.value" />
                    </el-select>
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="提醒铃声" label-width="80px">
                    <el-select v-model="barkConfig.config.failure_sound" placeholder="选择铃声" style="width: 100%">
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
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { testBark } from '../../api/task'
import request from '../../api/request'
import { ElMessage } from 'element-plus'

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
  enabled: false,
  notify_on_success: true,
  notify_on_failure: true,
  config: {
    server: 'https://api.day.app',
    device_key: '',
    success_sound: 'birdsong.caf',
    success_level: 'active',
    failure_sound: 'alarm.caf',
    failure_level: 'timeSensitive',
    archive: 'true',
    icon: ''
  }
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
    } else if (type === 'bark') {
      barkConfig.value = targetConfig
      // 补全默认值
      if (!barkConfig.value.config.server) barkConfig.value.config.server = 'https://api.day.app'
      if (!barkConfig.value.config.success_sound) barkConfig.value.config.success_sound = 'default'
      if (!barkConfig.value.config.failure_sound) barkConfig.value.config.failure_sound = 'default'
      if (!barkConfig.value.config.archive) barkConfig.value.config.archive = 'true'
    }
  } catch (error) {
    console.error(`加载 ${type} 配置失败:`, error)
  }
}

const handleSaveNotify = async (type) => {
  let config = {}
  if (type === 'wechat') config = wechatConfig.value
  else if (type === 'telegram') config = telegramConfig.value
  else if (type === 'wxpusher') config = wxpusherConfig.value
  else if (type === 'bark') {
    savingBark.value = true
    config = barkConfig.value
  }

  try {
    await request({ url: `/notify/${type}`, method: 'put', data: config })
    if (type === 'bark') {
      ElMessage.success('Bark 推送设置已保存')
    } else {
      ElMessage.success(`${type} 配置已保存`)
    }
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '保存失败')
  } finally {
    if (type === 'bark') savingBark.value = false
  }
}

const handleTestNotify = async (type) => {
  if (type === 'bark') {
    if (!barkConfig.value.config.device_key) {
      ElMessage.warning('请先填写 Device Key')
      return
    }
    barkTestForm.value.level = barkConfig.value.config.success_level || 'active'
    barkTestForm.value.sound = barkConfig.value.config.success_sound || 'birdsong.caf'
    barkTestForm.value.icon = barkConfig.value.config.icon || ''
    barkTestForm.value.isArchive = barkConfig.value.config.archive || 'true'
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
      bark_server: barkConfig.value.config.server,
      bark_device_key: barkConfig.value.config.device_key,
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

onMounted(async () => {
  await Promise.all([
    fetchNotifierConfig('wechat'),
    fetchNotifierConfig('telegram'),
    fetchNotifierConfig('wxpusher'),
    fetchNotifierConfig('bark')
  ])
})
</script>

<style scoped>
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

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
