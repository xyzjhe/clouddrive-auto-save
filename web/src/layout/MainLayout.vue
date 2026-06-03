<template>
  <el-container class="app-wrapper">
    <!-- 侧边栏 -->
    <el-aside width="220px" class="sidebar">
      <div class="logo">
        <span class="logo-text">UCAS</span>
      </div>

      <div class="search-wrapper">
        <el-input
          v-model="searchQuery"
          placeholder="搜索功能..."
          clearable
          size="small"
        >
          <template #prefix>
            <el-icon><PhMagnifyingGlass :size="16" weight="regular" /></el-icon>
          </template>
        </el-input>
      </div>

      <el-scrollbar class="nav-scrollbar">
        <div class="nav-groups">
          <div
            v-for="group in filteredNavigation"
            :key="group.name"
            class="nav-group"
          >
            <div class="nav-group-header">
              <span class="nav-group-name">{{ group.name }}</span>
            </div>

            <div class="nav-items">
              <div
                v-for="item in group.items"
                :key="item.path"
                class="nav-item"
                :class="{ active: isActive(item.path) }"
                @click="navigateTo(item.path)"
              >
                <el-icon class="nav-item-icon">
                  <component :is="iconMap[item.icon]" :size="20" weight="regular" />
                </el-icon>
                <span class="nav-item-name">{{ item.name }}</span>
              </div>
            </div>
          </div>
        </div>
      </el-scrollbar>

      <SidebarFooter />
    </el-aside>

    <el-container>
      <!-- 顶栏 -->
      <el-header height="56px" class="navbar">
        <div class="header-left">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item>首页</el-breadcrumb-item>
            <el-breadcrumb-item>{{ currentPageTitle }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="header-right">
          <el-button
            circle
            @click="toggleDark()"
            class="theme-toggle"
          >
            <template #icon>
              <component :is="isDark ? PhSun : PhMoon" :size="18" weight="regular" />
            </template>
          </el-button>
          <el-divider direction="vertical" />
          <el-avatar :size="32" src="https://github.com/identicons/user.png" />
        </div>
      </el-header>

      <!-- 内容区 -->
      <el-main class="main-content">
        <router-view v-slot="{ Component }">
          <transition name="fade-page" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import SidebarFooter from '../components/SidebarFooter.vue'
import { useDark, useToggle } from '@vueuse/core'
import { navigationConfig } from '../config/navigation'
import {
  PhMagnifyingGlass,
  PhMoon,
  PhSun,
  PhSquaresFour,
  PhUsers,
  PhListChecks,
  PhGearSix
} from '@phosphor-icons/vue'

const iconMap = {
  SquaresFour: PhSquaresFour,
  Users: PhUsers,
  ListChecks: PhListChecks,
  MagnifyingGlass: PhMagnifyingGlass,
  GearSix: PhGearSix
}

const route = useRoute()
const router = useRouter()

// 首次打开无配置时默认采用明亮模式
if (!localStorage.getItem('vueuse-color-scheme')) {
  localStorage.setItem('vueuse-color-scheme', 'light')
}
const isDark = useDark()
const toggleDark = useToggle(isDark)

const searchQuery = ref('')

// 过滤导航项
const filteredNavigation = computed(() => {
  if (!searchQuery.value) return navigationConfig

  const query = searchQuery.value.toLowerCase()
  return navigationConfig
    .map(group => ({
      ...group,
      items: group.items.filter(item =>
        item.name.toLowerCase().includes(query) ||
        item.description?.toLowerCase().includes(query)
      )
    }))
    .filter(group => group.items.length > 0)
})

const isActive = (path) => route.path === path

const navigateTo = (path) => {
  router.push(path)
}

const activeMenu = computed(() => route.path)
const currentPageTitle = computed(() => {
  const titles = {
    '/': '控制台',
    '/console': '控制台',
    '/accounts': '账号管理',
    '/tasks': '任务管理',
    '/settings': '系统设置',
    '/search': '资源发现'
  }
  return titles[route.path] || '控制台'
})
</script>

<style scoped>
.app-wrapper {
  height: 100vh;
}

.sidebar {
  background: var(--surface-bg);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  position: relative;
  z-index: 20;
}

.logo {
  height: 56px;
  display: flex;
  align-items: center;
  padding: 0 24px;
  gap: 12px;
}

.logo-text {
  font-size: 20px;
  font-weight: 700;
  color: var(--accent);
  letter-spacing: -0.02em;
}

.search-wrapper {
  padding: 0.75rem 1rem;
}

.nav-scrollbar {
  flex: 1;
}

.nav-groups {
  padding: 0.5rem;
}

.nav-group {
  margin-bottom: 0.5rem;
}

.nav-group-header {
  padding: 0.75rem 1rem 0.25rem;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.nav-items {
  overflow: hidden;
}

.nav-item {
  display: flex;
  align-items: center;
  padding: 0.6rem 1rem;
  margin: 0.125rem 0;
  border-radius: 6px;
  cursor: pointer;
  transition: all var(--transition-fast);
  border-left: 3px solid transparent;
}

.nav-item:hover {
  background: var(--hover-bg);
  color: var(--text-primary);
}

.nav-item.active {
  background: var(--accent-light);
  color: var(--accent);
  font-weight: 600;
  border-left-color: var(--accent);
}

.nav-item-icon {
  margin-right: 0.6rem;
  font-size: 1.1rem;
  display: flex;
  align-items: center;
}

.nav-item-name {
  font-size: 0.875rem;
}

.navbar {
  background: var(--surface-bg);
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  z-index: 10;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.theme-toggle {
  border: 1px solid var(--border-color) !important;
  background: transparent !important;
  color: var(--text-secondary) !important;
  transition: all var(--transition-fast);
}

.theme-toggle:hover {
  border-color: var(--accent) !important;
  color: var(--accent) !important;
}

.fade-page-enter-active,
.fade-page-leave-active {
  transition: opacity 0.15s cubic-bezier(0.4, 0, 0.2, 1);
}

.fade-page-enter-from,
.fade-page-leave-to {
  opacity: 0;
}
</style>
