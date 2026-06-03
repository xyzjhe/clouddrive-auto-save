import { createRouter, createWebHistory } from 'vue-router'
import MainLayout from '../layout/MainLayout.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      component: MainLayout,
      redirect: '/console',
      children: [
        {
          path: 'console',
          name: 'Console',
          component: () => import('../views/Dashboard.vue')
        },
        {
          path: 'accounts',
          name: 'Accounts',
          component: () => import('../views/Accounts.vue')
        },
        {
          path: 'tasks',
          name: 'Tasks',
          component: () => import('../views/Tasks.vue')
        },
        {
          path: 'settings',
          name: 'Settings',
          component: () => import('../views/Settings.vue')
        },
        {
          path: 'search',
          name: 'Search',
          component: () => import('../views/Search.vue')
        }
      ]
    }
  ]
})

// 捕获懒加载组件（chunk）加载失败的情况，自动刷新页面以获取最新的构建资源
router.onError((error, to) => {
  if (error.message.includes('Failed to fetch dynamically imported module') || error.message.includes('broken build')) {
    window.location.replace(to.fullPath)
  }
})

export default router
