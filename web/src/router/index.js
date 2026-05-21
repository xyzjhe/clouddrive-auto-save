import { createRouter, createWebHistory } from 'vue-router'
import MainLayout from '../layout/MainLayout.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      component: MainLayout,
      children: [
        {
          path: '',
          name: 'dashboard',
          component: () => import('../views/Dashboard.vue')
        },
        {
          path: 'accounts',
          name: 'accounts',
          component: () => import('../views/Accounts.vue')
        },
        {
          path: 'tasks',
          name: 'tasks',
          component: () => import('../views/Tasks.vue')
        },
        {
          path: 'settings',
          name: 'settings',
          component: () => import('../views/Settings.vue')
        },
        {
          path: 'plugins',
          name: 'plugins',
          component: () => import('../views/Plugins.vue')
        },
        {
          path: 'search',
          name: 'search',
          component: () => import('../views/Search.vue')
        },
        {
          path: 'notify',
          name: 'notify',
          component: () => import('../views/Notify.vue')
        }
      ]
    }
  ]
})

export default router
