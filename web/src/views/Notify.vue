<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'

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

onMounted(() => {
  fetchNotifiers()
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
    </el-tabs>
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
</style>
