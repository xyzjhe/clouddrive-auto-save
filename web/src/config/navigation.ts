export interface NavItem {
  name: string
  path: string
  icon: string
  description?: string
}

export interface NavGroup {
  name: string
  items: NavItem[]
}

export const navigationConfig: NavGroup[] = [
  {
    name: '概览',
    items: [
      {
        name: '控制台',
        path: '/console',
        icon: 'SquaresFour',
        description: '系统状态与实时转存监控'
      }
    ]
  },
  {
    name: '管理',
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
        icon: 'ListChecks',
        description: '管理转存任务'
      }
    ]
  },
  {
    name: '工具',
    items: [
      {
        name: '资源发现',
        path: '/search',
        icon: 'MagnifyingGlass',
        description: '搜索并发现云盘资源'
      }
    ]
  },
  {
    name: '系统',
    items: [
      {
        name: '系统设置',
        path: '/settings',
        icon: 'GearSix',
        description: '全局参数、推送与插件管理'
      }
    ]
  }
]
