export interface NavItem {
  name: string
  path: string
  icon: string
  description?: string
}

export interface NavGroup {
  name: string
  icon: string
  items: NavItem[]
  collapsible?: boolean
  defaultCollapsed?: boolean
}

export const navigationConfig: NavGroup[] = [
  {
    name: '概览',
    icon: '📊',
    items: [
      {
        name: '控制台',
        path: '/console',
        icon: 'LayoutDashboard',
        description: '系统状态与实时转存监控'
      }
    ]
  },
  {
    name: '管理',
    icon: '🔧',
    items: [
      {
        name: '账号管理',
        path: '/accounts',
        icon: 'Users',
        description: '管理云盘账号'
      },
      {
        name: '任务列表',
        path: '/tasks',
        icon: 'ListTodo',
        description: '管理转存任务'
      }
    ]
  },
  {
    name: '工具',
    icon: '🛠️',
    items: [
      {
        name: '资源发现',
        path: '/search',
        icon: 'Search',
        description: '搜索并发现云盘资源'
      }
    ]
  },
  {
    name: '系统',
    icon: '⚙️',
    items: [
      {
        name: '系统设置',
        path: '/settings',
        icon: 'Settings',
        description: '全局参数、推送与插件管理'
      }
    ]
  }
]
