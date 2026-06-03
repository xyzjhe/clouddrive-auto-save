<script setup>
import { ref, onMounted } from 'vue'
import { PhCloud } from '@phosphor-icons/vue'

const deferredPrompt = ref(null)
const showInstallPrompt = ref(false)

onMounted(() => {
  window.addEventListener('beforeinstallprompt', (e) => {
    // 阻止 Chrome 自动弹出安装提示
    e.preventDefault()
    // 保存事件，稍后使用
    deferredPrompt.value = e
    // 显示自定义安装提示
    showInstallPrompt.value = true
  })

  window.addEventListener('appinstalled', () => {
    showInstallPrompt.value = false
    deferredPrompt.value = null
    console.log('PWA 已安装')
  })
})

const install = async () => {
  if (!deferredPrompt.value) return

  // 显示安装提示
  deferredPrompt.value.prompt()

  // 等待用户响应
  const { outcome } = await deferredPrompt.value.userChoice
  console.log(`用户选择: ${outcome}`)

  // 清理
  deferredPrompt.value = null
  showInstallPrompt.value = false
}

const dismiss = () => {
  showInstallPrompt.value = false
}
</script>

<template>
  <transition name="slide-up">
    <div v-if="showInstallPrompt" class="install-prompt">
      <div class="install-content">
        <div class="install-icon">
          <PhCloud :size="40" weight="duotone" />
        </div>
        <div class="install-text">
          <div class="install-title">安装 UCAS</div>
          <div class="install-desc">添加到主屏幕，获得更好的使用体验</div>
        </div>
        <div class="install-actions">
          <el-button type="primary" @click="install">安装</el-button>
          <el-button @click="dismiss">稍后</el-button>
        </div>
      </div>
    </div>
  </transition>
</template>

<style scoped>
.install-prompt {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: var(--bg-secondary);
  border-top: 1px solid var(--border);
  padding: 1rem;
  z-index: 9999;
}

.install-content {
  max-width: 600px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  gap: 1rem;
}

.install-icon {
  color: var(--accent);
  display: flex;
  align-items: center;
}

.install-text {
  flex: 1;
}

.install-title {
  font-weight: 600;
  margin-bottom: 0.25rem;
}

.install-desc {
  font-size: 0.85rem;
  color: var(--text-secondary);
}

.install-actions {
  display: flex;
  gap: 0.5rem;
}

.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.3s ease;
}

.slide-up-enter-from,
.slide-up-leave-to {
  transform: translateY(100%);
  opacity: 0;
}
</style>
