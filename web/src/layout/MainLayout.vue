<template>
  <el-container class="app-wrapper">
    <!-- 侧边栏 -->
    <el-aside width="240px" class="sidebar">
      <div class="logo">
        <CloudLogo :size="36" id="sidebar-grad" />
        <span>UCAS</span>
      </div>

      <el-menu
        :default-active="activeMenu"
        class="side-menu"
        router
      >
        <el-menu-item index="/">
          <el-icon><LayoutDashboard :size="20" /></el-icon>
          <span>仪表盘概览</span>
        </el-menu-item>
        <el-menu-item index="/accounts">
          <el-icon><User :size="20" /></el-icon>
          <span>账号管理</span>
        </el-menu-item>
        <el-menu-item index="/tasks">
          <el-icon><ListChecks :size="20" /></el-icon>
          <span>任务列表</span>
        </el-menu-item>
        <el-menu-item index="/settings">
          <el-icon><SettingsIcon :size="20" /></el-icon>
          <span>系统设置</span>
        </el-menu-item>
      </el-menu>
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
          <transition name="fade-transform" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import CloudLogo from '../components/CloudLogo.vue'
import SidebarFooter from '../components/SidebarFooter.vue'
import { useDark, useToggle } from '@vueuse/core'
import {
  LayoutDashboard,
  User,
  ListChecks,
  Settings as SettingsIcon,
  Moon,
  Sun
} from 'lucide-vue-next'

const route = useRoute()
const isDark = useDark()
const toggleDark = useToggle(isDark)

const activeMenu = computed(() => route.path)
const currentPageTitle = computed(() => {
  const titles = {
    '/': '仪表盘',
    '/accounts': '账号管理',
    '/tasks': '任务管理',
    '/settings': '系统设置'
  }
  return titles[route.path] || '概览'
})
</script>

<style scoped>
.app-wrapper {
  height: 100vh;
}

.sidebar {
  background: var(--bg-sidebar);
  border-right: 1px solid var(--neutral-200);
  display: flex;
  flex-direction: column;
  position: relative;
  z-index: 20;
}

html.dark .sidebar {
  border-right: 1px solid rgba(255, 255, 255, 0.04);
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

.side-menu {
  border-right: none;
  padding: 0 12px;
  flex: 1;
}

.el-menu-item {
  height: 44px;
  line-height: 44px;
  margin: 2px 0;
  border-radius: 10px;
  color: var(--neutral-500);
  font-weight: 500;
  font-size: 14px;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.el-menu-item:hover {
  background: var(--neutral-100);
  color: var(--neutral-700);
}

html.dark .el-menu-item:hover {
  background: rgba(255, 255, 255, 0.04);
  color: var(--neutral-300);
}

.el-menu-item.is-active {
  background: var(--brand-50);
  color: var(--brand-600);
  font-weight: 600;
}

html.dark .el-menu-item.is-active {
  background: rgba(99, 102, 241, 0.1);
  color: var(--brand-400);
}

.navbar {
  background: var(--bg-navbar);
  backdrop-filter: blur(16px) saturate(1.2);
  border-bottom: 1px solid var(--neutral-200);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 28px;
  z-index: 10;
}

html.dark .navbar {
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.theme-toggle {
  border: 1px solid var(--neutral-200) !important;
  background: var(--bg-content) !important;
  color: var(--neutral-600) !important;
  transition: all 0.2s;
}

.theme-toggle:hover {
  border-color: var(--brand-200) !important;
  color: var(--brand-600) !important;
  background: var(--brand-50) !important;
}

html.dark .theme-toggle {
  border-color: rgba(255, 255, 255, 0.08) !important;
  background: rgba(255, 255, 255, 0.04) !important;
  color: var(--neutral-400) !important;
}

html.dark .theme-toggle:hover {
  border-color: var(--brand-400) !important;
  color: var(--brand-400) !important;
  background: rgba(99, 102, 241, 0.1) !important;
}

.fade-transform-enter-active,
.fade-transform-leave-active {
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}

.fade-transform-enter-from {
  opacity: 0;
  transform: translateX(-12px);
}

.fade-transform-leave-to {
  opacity: 0;
  transform: translateX(12px);
}
</style>
