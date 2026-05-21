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
        name: '仪表盘',
        path: '/dashboard',
        icon: 'LayoutDashboard',
        description: '实时统计和任务监控'
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
        name: '资源搜索',
        path: '/search',
        icon: 'Search',
        description: '搜索云盘资源'
      },
      {
        name: '插件管理',
        path: '/plugins',
        icon: 'Puzzle',
        description: '管理系统插件'
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
        description: '全局配置'
      },
      {
        name: '消息推送',
        path: '/notify',
        icon: 'Bell',
        description: '通知渠道配置'
      }
    ]
  }
]
