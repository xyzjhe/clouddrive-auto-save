<template>
  <div class="sidebar-footer">
    <a
      class="github-link"
      href="https://github.com/zhaocongqi/clouddrive-auto-save"
      target="_blank"
      rel="noopener noreferrer"
    >
      <PhGithubLogo :size="16" />
      <span>GitHub 仓库</span>
      <PhArrowSquareOut :size="12" />
    </a>
    <div
      class="version-info"
      :class="{ 'has-update': hasUpdate }"
      @click="hasUpdate && openReleases()"
    >
      <span class="version-text">v{{ currentVersion }}</span>
      <span v-if="hasUpdate" class="update-badge">NEW</span>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { PhGithubLogo, PhArrowSquareOut } from '@phosphor-icons/vue'
import { getVersion } from '../api/version'

const currentVersion = ref('...')
const hasUpdate = ref(false)

const GITHUB_RELEASES_URL = 'https://github.com/zhaocongqi/clouddrive-auto-save/releases'
const GITHUB_API_URL = 'https://api.github.com/repos/zhaocongqi/clouddrive-auto-save/releases/latest'

function parseVersion(v) {
  const cleaned = v.replace(/^v/, '')
  const parts = cleaned.split('.').map(Number)
  return parts.length === 3 && parts.every(n => !isNaN(n)) ? parts : null
}

function compareVersions(current, latest) {
  const c = parseVersion(current)
  const l = parseVersion(latest)
  if (!c || !l) return false
  for (let i = 0; i < 3; i++) {
    if (l[i] > c[i]) return true
    if (l[i] < c[i]) return false
  }
  return false
}

function openReleases() {
  window.open(GITHUB_RELEASES_URL, '_blank', 'noopener,noreferrer')
}

onMounted(async () => {
  try {
    const res = await getVersion()
    currentVersion.value = res.version || 'dev'
  } catch {
    currentVersion.value = 'dev'
  }

  if (currentVersion.value === 'dev') return

  try {
    const resp = await fetch(GITHUB_API_URL)
    if (!resp.ok) return
    const data = await resp.json()
    const latestTag = data.tag_name || ''
    hasUpdate.value = compareVersions(currentVersion.value, latestTag)
  } catch {
    // 静默失败
  }
})
</script>

<style scoped>
.sidebar-footer {
  margin-top: auto;
  padding: 12px 16px;
  border-top: 1px solid var(--border-color);
}

.github-link {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 10px;
  color: var(--neutral-500);
  text-decoration: none;
  font-size: 13px;
  font-weight: 500;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.github-link:hover {
  background: var(--neutral-100);
  color: var(--neutral-700);
}

.version-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 12px;
  font-size: 12px;
  color: var(--neutral-400);
  font-family: var(--font-mono);
  font-weight: 500;
}

.version-info.has-update {
  cursor: pointer;
  color: var(--neutral-600);
  border-radius: 10px;
  padding: 6px 12px;
  margin: 2px 0;
  transition: background 0.2s;
}

.version-info.has-update:hover {
  background: var(--neutral-100);
}

.update-badge {
  display: inline-flex;
  align-items: center;
  padding: 1px 6px;
  border-radius: 10px;
  background: var(--color-danger);
  color: white;
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.5px;
  font-family: var(--font-sans);
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.7; }
}
</style>
