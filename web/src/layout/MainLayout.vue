<template>
  <el-container class="app-wrapper">
    <!-- 侧边栏 -->
    <el-aside width="240px" class="sidebar">
      <div class="logo">
        <CloudLogo :size="36" id="sidebar-grad" />
        <span>UCAS</span>
      </div>

      <div class="search-wrapper">
        <el-input
          v-model="searchQuery"
          placeholder="搜索功能..."
          clearable
          size="small"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
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
            <div
              class="nav-group-header"
              @click="toggleGroup(group.name)"
            >
              <span class="nav-group-icon">{{ group.icon }}</span>
              <span class="nav-group-name">{{ group.name }}</span>
              <el-icon
                class="nav-group-arrow"
                :class="{ collapsed: collapsedGroups[group.name] }"
              >
                <ArrowDown />
              </el-icon>
            </div>

            <transition name="slide">
              <div
                v-show="!collapsedGroups[group.name]"
                class="nav-items"
              >
                <div
                  v-for="item in group.items"
                  :key="item.path"
                  class="nav-item"
                  :class="{ active: isActive(item.path) }"
                  @click="navigateTo(item.path)"
                >
                  <el-icon class="nav-item-icon">
                    <component :is="item.icon" />
                  </el-icon>
                  <span class="nav-item-name">{{ item.name }}</span>
                </div>
              </div>
            </transition>
          </div>
        </div>
      </el-scrollbar>

      <SidebarFooter />
    </el-aside>

    <el-container>
      <!-- 顶栏 -->
      <el-header height="64px" class="navbar">
        <div class="header-left">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item>首页</el-breadcrumb-item>
            <el-breadcrumb-item>{{ currentPageTitle }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="header-right">
          <el-button
            circle
            :icon="isDark ? Sun : Moon"
            @click="toggleDark()"
            class="theme-toggle"
          />
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
import CloudLogo from '../components/CloudLogo.vue'
import SidebarFooter from '../components/SidebarFooter.vue'
import { useDark, useToggle } from '@vueuse/core'
import { navigationConfig } from '../config/navigation'
import {
  LayoutDashboard,
  User,
  ListChecks,
  Settings as SettingsIcon,
  Search,
  Puzzle,
  Bell,
  Moon,
  Sun,
  ArrowDown
} from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()

// 首次打开无配置时默认采用明亮模式
if (!localStorage.getItem('vueuse-color-scheme')) {
  localStorage.setItem('vueuse-color-scheme', 'light')
}
const isDark = useDark()
const toggleDark = useToggle(isDark)

// 折叠状态管理
const collapsedGroups = ref(JSON.parse(localStorage.getItem('collapsedGroups') || '{}'))
const searchQuery = ref('')

const toggleGroup = (groupName) => {
  collapsedGroups.value[groupName] = !collapsedGroups.value[groupName]
  localStorage.setItem('collapsedGroups', JSON.stringify(collapsedGroups.value))
}

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
  background: var(--bg-sidebar);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  position: relative;
  z-index: 20;
}

.logo {
  height: 64px;
  display: flex;
  align-items: center;
  padding: 0 24px;
  gap: 12px;
  font-size: 20px;
  font-weight: 800;
  color: var(--brand-600);
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
  display: flex;
  align-items: center;
  padding: 0.5rem 0.75rem;
  cursor: pointer;
  border-radius: 6px;
  transition: background-color 0.2s;
}

.nav-group-header:hover {
  background: var(--neutral-100);
}

.nav-group-icon {
  margin-right: 0.5rem;
  font-size: 1rem;
}

.nav-group-name {
  flex: 1;
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--neutral-500);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.nav-group-arrow {
  transition: transform 0.3s;
}

.nav-group-arrow.collapsed {
  transform: rotate(-90deg);
}

.nav-items {
  overflow: hidden;
}

.nav-item {
  display: flex;
  align-items: center;
  padding: 0.75rem 1rem;
  margin: 0.25rem 0;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.nav-item:hover {
  background: var(--neutral-100);
  color: var(--neutral-700);
}

.nav-item.active {
  background: var(--brand-50);
  color: var(--brand-600);
  font-weight: 600;
}

.nav-item-icon {
  margin-right: 0.75rem;
  font-size: 1.1rem;
}

.nav-item-name {
  font-size: 0.9rem;
}

.navbar {
  background: var(--bg-navbar);
  backdrop-filter: blur(16px) saturate(1.2);
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 28px;
  z-index: 10;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.theme-toggle {
  border: 1px solid var(--border-color) !important;
  background: var(--input-bg) !important;
  color: var(--text-secondary) !important;
  transition: all 0.2s;
}

.theme-toggle:hover {
  border-color: var(--neon-teal) !important;
  color: var(--neon-teal) !important;
  background: var(--hover-bg) !important;
  box-shadow: var(--neon-glow-teal) !important;
}

.fade-page-enter-active,
.fade-page-leave-active {
  transition: opacity 0.15s cubic-bezier(0.4, 0, 0.2, 1);
}

.fade-page-enter-from,
.fade-page-leave-to {
  opacity: 0;
}

.slide-enter-active,
.slide-leave-active {
  transition: all 0.3s;
}

.slide-enter-from,
.slide-leave-to {
  opacity: 0;
  max-height: 0;
}

.slide-enter-to,
.slide-leave-from {
  max-height: 500px;
}
</style>
