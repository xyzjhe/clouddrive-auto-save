<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getGlobalSettings, updateGlobalSettings, testBark } from '../api/task'

const notifiers = ref([])
const loading = ref(false)
const activeTab = ref('wechat')

// 配置表单
const wechatConfig = ref({
  enabled: false,
  webhook_url: '',
  notify_on_success: true,
  notify_on_failure: true
})

const telegramConfig = ref({
  enabled: false,
  bot_token: '',
  chat_id: '',
  notify_on_success: true,
  notify_on_failure: true
})

const wxpusherConfig = ref({
  enabled: false,
  app_token: '',
  uid: '',
  notify_on_success: true,
  notify_on_failure: true
})

// Bark 配置状态
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

// 测试对话框状态
const testDialogVisible = ref(false)
const testForm = ref({
  title: 'UCAS 测试通知',
  body: '这是一条自定义参数的测试推送消息。',
  level: 'active',
  sound: 'birdsong.caf',
  icon: '',
  isArchive: 'true'
})

const testing = ref(false)
const savingBark = ref(false)

const fetchBarkSettings = async () => {
  try {
    const data = await getGlobalSettings()
    // 合并默认值
    Object.keys(barkConfig.value).forEach(key => {
      if (data[key] !== undefined) {
        barkConfig.value[key] = data[key]
      }
    })
    
    // 铃声值容错处理：如果为空字符串（Bark 默认），映射为 UI 的 'default'
    if (barkConfig.value.bark_success_sound === '') barkConfig.value.bark_success_sound = 'default'
    if (barkConfig.value.bark_failure_sound === '') barkConfig.value.bark_failure_sound = 'default'
    // 确保 archive 有值
    if (barkConfig.value.bark_archive === undefined) barkConfig.value.bark_archive = 'true'
  } catch (error) {
    console.error('加载 Bark 设置失败:', error)
  }
}

const fetchNotifiers = async () => {
  loading.value = true
  try {
    const response = await fetch('/api/notify')
    const data = await response.json()
    notifiers.value = data.data || []
  } catch (error) {
    console.error('获取通知渠道失败:', error)
  } finally {
    loading.value = false
  }
}

const handleSave = async (type) => {
  if (type === 'bark') {
    savingBark.value = true
    try {
      await updateGlobalSettings(barkConfig.value)
      ElMessage.success('Bark 推送设置已保存')
    } catch (error) {
      console.error('保存 Bark 配置失败:', error)
      ElMessage.error(error.response?.data?.error || '保存失败')
    } finally {
      savingBark.value = false
    }
    return
  }

  let config = {}
  switch (type) {
    case 'wechat':
      config = wechatConfig.value
      break
    case 'telegram':
      config = telegramConfig.value
      break
    case 'wxpusher':
      config = wxpusherConfig.value
      break
  }

  try {
    const response = await fetch(`/api/notify/${type}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(config)
    })

    const data = await response.json()
    if (data.code === 0) {
      ElMessage.success('配置已保存')
    } else {
      ElMessage.error(data.message || '保存失败')
    }
  } catch (error) {
    console.error('保存配置失败:', error)
    ElMessage.error('保存失败')
  }
}

const handleTest = async (type) => {
  if (type === 'bark') {
    if (!barkConfig.value.bark_device_key) {
      ElMessage.warning('请先填写 Device Key')
      return
    }
    // 默认使用当前配置的值作为测试初始值
    testForm.value.level = barkConfig.value.bark_success_level || 'active'
    testForm.value.sound = barkConfig.value.bark_success_sound || 'birdsong.caf'
    testForm.value.icon = barkConfig.value.bark_icon || ''
    testForm.value.isArchive = barkConfig.value.bark_archive || 'true'
    testDialogVisible.value = true
    return
  }

  try {
    const response = await fetch(`/api/notify/${type}/test`, {
      method: 'POST'
    })

    const data = await response.json()
    if (data.code === 0) {
      ElMessage.success('测试消息已发送')
    } else {
      ElMessage.error(data.message || '测试失败')
    }
  } catch (error) {
    console.error('测试失败:', error)
    ElMessage.error('测试失败')
  }
}

const handleTestBark = async () => {
  testing.value = true
  try {
    await testBark({
      bark_server: barkConfig.value.bark_server,
      bark_device_key: barkConfig.value.bark_device_key,
      ...testForm.value
    })
    ElMessage.success('测试消息已发送，请检查手机')
    testDialogVisible.value = false
  } catch (error) {
    ElMessage.error('测试发送失败: ' + (error.response?.data?.error || error.message))
  } finally {
    testing.value = false
  }
}

onMounted(() => {
  fetchNotifiers()
  fetchBarkSettings()
})
</script>

<template>
  <div class="notify-page">
    <div class="page-header">
      <div class="title-section">
        <h2>消息推送</h2>
        <p>配置通知渠道，接收任务执行通知</p>
      </div>
    </div>

    <el-tabs v-model="activeTab" type="border-card">
      <!-- 企业微信 -->
      <el-tab-pane label="企业微信" name="wechat">
        <el-form :model="wechatConfig" label-width="120px">
          <el-form-item label="启用">
            <el-switch v-model="wechatConfig.enabled" />
          </el-form-item>

          <el-form-item label="Webhook URL">
            <el-input
              v-model="wechatConfig.webhook_url"
              placeholder="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=..."
            />
          </el-form-item>

          <el-form-item label="通知设置">
            <el-checkbox v-model="wechatConfig.notify_on_success">
              成功通知
            </el-checkbox>
            <el-checkbox v-model="wechatConfig.notify_on_failure">
              失败通知
            </el-checkbox>
          </el-form-item>

          <el-form-item>
            <el-button type="primary" @click="handleSave('wechat')">
              保存
            </el-button>
            <el-button @click="handleTest('wechat')">
              测试
            </el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <!-- Telegram -->
      <el-tab-pane label="Telegram" name="telegram">
        <el-form :model="telegramConfig" label-width="120px">
          <el-form-item label="启用">
            <el-switch v-model="telegramConfig.enabled" />
          </el-form-item>

          <el-form-item label="Bot Token">
            <el-input
              v-model="telegramConfig.bot_token"
              placeholder="123456789:ABCdefGHIjklMNOpqrsTUVwxyz"
            />
          </el-form-item>

          <el-form-item label="Chat ID">
            <el-input
              v-model="telegramConfig.chat_id"
              placeholder="123456789"
            />
          </el-form-item>

          <el-form-item label="通知设置">
            <el-checkbox v-model="telegramConfig.notify_on_success">
              成功通知
            </el-checkbox>
            <el-checkbox v-model="telegramConfig.notify_on_failure">
              失败通知
            </el-checkbox>
          </el-form-item>

          <el-form-item>
            <el-button type="primary" @click="handleSave('telegram')">
              保存
            </el-button>
            <el-button @click="handleTest('telegram')">
              测试
            </el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <!-- WxPusher -->
      <el-tab-pane label="WxPusher" name="wxpusher">
        <el-form :model="wxpusherConfig" label-width="120px">
          <el-form-item label="启用">
            <el-switch v-model="wxpusherConfig.enabled" />
          </el-form-item>

          <el-form-item label="App Token">
            <el-input
              v-model="wxpusherConfig.app_token"
              placeholder="AT_xxx"
            />
          </el-form-item>

          <el-form-item label="UID">
            <el-input
              v-model="wxpusherConfig.uid"
              placeholder="UID_xxx"
            />
          </el-form-item>

          <el-form-item label="通知设置">
            <el-checkbox v-model="wxpusherConfig.notify_on_success">
              成功通知
            </el-checkbox>
            <el-checkbox v-model="wxpusherConfig.notify_on_failure">
              失败通知
            </el-checkbox>
          </el-form-item>

          <el-form-item>
            <el-button type="primary" @click="handleSave('wxpusher')">
              保存
            </el-button>
            <el-button @click="handleTest('wxpusher')">
              测试
            </el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <!-- Bark -->
      <el-tab-pane label="Bark" name="bark">
        <el-form :model="barkConfig" label-width="120px">
          <el-form-item label="启用">
            <el-switch
              v-model="barkConfig.bark_enabled"
              active-value="true"
              inactive-value="false"
            />
          </el-form-item>

          <el-form-item label="服务器地址">
            <el-input
              v-model="barkConfig.bark_server"
              placeholder="https://api.day.app"
            />
          </el-form-item>

          <el-form-item label="Device Key">
            <el-input
              v-model="barkConfig.bark_device_key"
              placeholder="您的 Bark 设备 Key"
              type="password"
              show-password
            />
          </el-form-item>

          <!-- 高级设置折叠 -->
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
            <el-button type="primary" :loading="savingBark" @click="handleSave('bark')">
              保存
            </el-button>
            <el-button @click="handleTest('bark')">
              测试
            </el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>
    </el-tabs>

    <!-- Bark 测试对话框 -->
    <el-dialog
      v-model="testDialogVisible"
      title="发送测试推送"
      width="400px"
      append-to-body
      class="custom-dialog"
    >
      <el-form :model="testForm" label-position="top">
        <el-form-item label="推送标题">
          <el-input v-model="testForm.title" placeholder="输入推送标题" />
        </el-form-item>
        <el-form-item label="推送内容">
          <el-input v-model="testForm.body" type="textarea" :rows="3" placeholder="输入推送内容" />
        </el-form-item>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="通知级别">
              <el-select v-model="testForm.level" style="width: 100%">
                <el-option v-for="l in barkLevels" :key="l.value" :label="l.label" :value="l.value" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="提醒铃声">
              <el-select v-model="testForm.sound" style="width: 100%">
                <el-option v-for="s in barkSounds" :key="s.value" :label="s.label" :value="s.value" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="自定义图标 URL">
          <el-input v-model="testForm.icon" placeholder="https://example.com/icon.png" />
        </el-form-item>
        <el-form-item label="自动保存到历史记录">
          <el-switch v-model="testForm.isArchive" active-value="true" inactive-value="false" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="testDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="testing" @click="handleTestBark">
            立即发送
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.notify-page {
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

.advanced-collapse {
  margin-top: 20px;
  border: none !important;
}

:deep(.advanced-collapse .el-collapse-item__header) {
  height: 40px;
  border-bottom: none;
  background-color: var(--neutral-100);
  padding: 0 12px;
  border-radius: 8px;
}

html.dark :deep(.advanced-collapse .el-collapse-item__header) {
  background-color: rgba(255, 255, 255, 0.04);
}

:deep(.advanced-collapse .el-collapse-item__wrap) {
  border-bottom: none;
  padding: 12px;
}

.collapse-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--neutral-500);
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
