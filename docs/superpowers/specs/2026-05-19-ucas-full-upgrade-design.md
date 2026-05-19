# UCAS 全面升级设计方案

**日期**: 2026-05-19
**版本**: 1.0
**状态**: 已确认

## 1. 背景与目标

### 1.1 背景

UCAS (Unified CloudDrive Auto-Save) 是一个统一云盘自动转存系统，当前支持移动云盘 (139) 和夸克网盘 (Quark)。通过对标分析 quark-auto-save、cloudpan-auto-save、TgtoDrive 三个优秀项目，发现 UCAS 在以下方面存在提升空间：

- **UI/UX 设计**：侧边栏导航扩展性不足，仪表盘信息密度较低，缺少卡片式 UI 组件
- **功能模块**：缺少插件系统、Telegram 集成、资源搜索等高级功能
- **架构能力**：仅支持 Bark 推送，缺少多渠道消息推送、PWA 支持等

### 1.2 目标

本方案旨在对 UCAS 进行全面升级，使其成为一流的云盘自动化平台：

1. **UI/UX 现代化**：优化导航、增强仪表盘、引入卡片式组件、支持 PWA
2. **功能扩展**：引入插件系统、Telegram 集成、资源搜索集成
3. **架构增强**：多渠道消息推送、为未来平台扩展做好准备

## 2. 设计方案

### 2.1 UI/UX 改进

#### 2.1.1 侧边栏导航优化

**当前问题**：扁平菜单（4个选项），难以适应功能增长。

**改进方案**：采用分类分组 + 可折叠设计，参考 TgtoDrive。

**导航结构**：
```
UCAS (品牌标识 + 版本号)
├── 📊 概览
│   └── 仪表盘
├── 🔧 管理
│   ├── 账号管理
│   └── 任务列表
├── 🛠️ 工具
│   ├── 资源搜索
│   └── 插件管理
└── ⚙️ 系统
    ├── 系统设置
    └── 消息推送
```

**实现要点**：
- 导航数据配置化，方便扩展
- 折叠状态本地持久化 (localStorage)
- 支持搜索快速定位
- 移动端响应式适配

**关键文件**：
- `web/src/layout/MainLayout.vue` - 主布局组件
- `web/src/config/navigation.ts` - 导航配置（新增）

#### 2.1.2 仪表盘增强

**当前问题**：仅有4张统计卡片和日志终端，信息密度较低。

**改进方案**：增加统计维度和可视化图表。

**新增内容**：
- **统计卡片**（保留现有4张）：已规划任务、已保存容量、今日完成、活跃账号
- **任务执行趋势图**：最近7天的任务完成数量柱状图
- **存储空间分布**：按平台统计的存储使用情况
- **实时任务监控面板**：运行中任务的进度条、阶段标签、预计剩余时间

**实现要点**：
- 使用 Chart.js 或 ECharts 绘制图表
- SSE 实时更新任务进度
- 响应式布局适配不同屏幕

**关键文件**：
- `web/src/views/Dashboard.vue` - 仪表盘页面
- `web/src/components/charts/` - 图表组件（新增）

#### 2.1.3 卡片式 UI 组件

**当前问题**：主要使用表格视图，视觉层次不够丰富。

**改进方案**：引入卡片式 UI 组件，参考 TgtoDrive。

**组件设计**：

**账号卡片**：
- 渐变色头部（按平台区分颜色）
- 存储空间进度条（梯度着色：绿/黄/红）
- 快捷操作按钮（校验、编辑）

**任务卡片**：
- 实时进度条（带 striped 动画）
- 状态标签（运行中/等待中/已完成/Fatal）
- 调度信息、上次执行时间

**实现要点**：
- 支持表格/卡片视图切换（已有基础）
- 卡片悬浮上移动画效果
- 响应式网格布局

**关键文件**：
- `web/src/views/Accounts.vue` - 账号管理页面
- `web/src/views/Tasks.vue` - 任务管理页面
- `web/src/components/cards/` - 卡片组件（新增）

#### 2.1.4 PWA 支持

**目标**：让 UCAS 可以像原生 App 一样安装到手机主屏幕。

**实现内容**：

**Web App Manifest** (`manifest.json`)：
```json
{
  "name": "UCAS - 统一云盘自动转存系统",
  "short_name": "UCAS",
  "description": "自动化云盘转存管理工具",
  "start_url": "/",
  "display": "standalone",
  "background_color": "#ffffff",
  "theme_color": "#6366f1",
  "icons": [
    { "src": "/icon-192.png", "sizes": "192x192", "type": "image/png" },
    { "src": "/icon-512.png", "sizes": "512x512", "type": "image/png" }
  ]
}
```

**Service Worker**：
- 缓存静态资源（JS/CSS/图片）
- 支持离线访问（部分功能）

**iOS 适配**：
- `<meta name="apple-mobile-web-app-capable" content="yes">`
- `<meta name="apple-mobile-web-app-status-bar-style" content="default">`

**安装引导**：
- 监听 `beforeinstallprompt` 事件
- 自定义安装提示界面

**关键文件**：
- `web/public/manifest.json` - PWA 配置（新增）
- `web/public/sw.js` - Service Worker（新增）
- `web/src/components/PWAInstall.vue` - 安装引导组件（新增）
- `web/index.html` - 添加 PWA meta 标签

### 2.2 功能扩展

#### 2.2.1 插件系统架构

**目标**：支持模块化扩展，参考 quark-auto-save 的插件架构。

**生命周期钩子**：
- `task_before`：任务执行前（预处理、参数校验）
- `run`：执行转存（核心业务逻辑）
- `task_after`：任务执行后（通知、刷新、清理）

**插件目录结构**：
```
plugins/
├── emby/
│   ├── plugin.json    # 插件元数据
│   └── main.go        # 插件实现
├── alist/
│   ├── plugin.json
│   └── main.go
└── README.md          # 插件开发指南
```

**plugin.json 格式**：
```json
{
  "name": "emby",
  "version": "1.0.0",
  "description": "Emby 媒体库自动刷新",
  "author": "UCAS",
  "hooks": ["task_after"],
  "config": {
    "server_url": "",
    "api_key": ""
  }
}
```

**插件管理界面**：
- 卡片式展示已安装插件
- 支持启用/禁用/配置
- 安装新插件入口

**关键文件**：
- `internal/core/plugin/manager.go` - 插件管理器（新增）
- `internal/core/plugin/interface.go` - 插件接口定义（新增）
- `internal/api/plugin.go` - 插件 API（新增）
- `web/src/views/Plugins.vue` - 插件管理页面（新增）

#### 2.2.2 Telegram 机器人集成

**目标**：支持通过 Telegram 远程管理任务，参考 cloudpan-auto-save。

**支持的命令**：
- `/start` - 启动机器人，显示帮助
- `/tasks` - 查看所有任务列表
- `/run <任务ID>` - 执行指定任务
- `/run_all` - 批量执行所有任务
- `/add` - 交互式创建新任务
- `/status` - 查看系统状态
- `/logs` - 查看最近日志
- `/search <关键词>` - 搜索资源

**配置界面**：
- Bot Token 输入
- 允许的用户 ID（白名单）
- 通知设置（成功/失败分别配置）

**关键文件**：
- `internal/core/telegram/bot.go` - Telegram 机器人（新增）
- `internal/core/telegram/handler.go` - 命令处理器（新增）
- `internal/api/telegram.go` - Telegram 配置 API（新增）
- `web/src/views/Settings.vue` - 添加 Telegram 配置区域

#### 2.2.3 资源搜索集成

**目标**：集成 CloudSaver/PanSou 等资源搜索引擎，支持搜索后一键创建任务。

**功能设计**：
- 搜索输入框 + 搜索源选择（CloudSaver/PanSou/115/夸克）
- 搜索结果展示：来源、平台、更新时间、内容摘要
- 一键创建任务：搜索结果直接转为转存任务

**关键文件**：
- `internal/core/search/client.go` - 搜索客户端（新增）
- `internal/api/search.go` - 搜索 API（新增）
- `web/src/views/Search.vue` - 资源搜索页面（新增）

#### 2.2.4 多渠道消息推送

**目标**：支持企业微信、Telegram、WxPusher、Bark 四种推送渠道。

**架构设计**：
- 统一的 NotifyManager 调度器
- 每个渠道独立实现 Notifier 接口
- 支持批量执行时智能汇总推送

**配置界面**：
- 每个渠道独立配置区域
- 支持启用/禁用
- 成功/失败通知级别分别配置
- 测试发送按钮

**关键文件**：
- `internal/core/notify/manager.go` - 通知管理器（新增）
- `internal/core/notify/wechat.go` - 企业微信推送（新增）
- `internal/core/notify/telegram.go` - Telegram 推送（新增）
- `internal/core/notify/wxpusher.go` - WxPusher 推送（新增）
- `internal/api/notify.go` - 通知配置 API（新增）

### 2.3 架构增强

#### 2.3.1 平台扩展准备

**目标**：为未来支持 115、天翼、百度网盘做好架构准备。

**现有架构**：
- CloudDrive 接口定义清晰
- 驱动工厂模式已就绪：`RegisterDriver(platform, factory)`
- 数据库 Account 模型支持多平台

**扩展检查清单**：
- 实现 CloudDrive 接口的所有方法
- 支持 Cookie / Token / OAuth 等认证方式
- 映射平台特定错误码到统一错误类型
- 遵守平台 API 调用频率限制
- 编写完整的单元测试和 E2E 测试
- 更新 README 和 API 文档

**关键文件**：
- `internal/core/drive.go` - CloudDrive 接口（已有）
- `internal/core/driver_115/` - 115 驱动（待实现）
- `internal/core/driver_189/` - 天翼驱动（待实现）
- `internal/core/driver_baidu/` - 百度驱动（待实现）

## 3. 实施计划

### 3.1 阶段划分

本方案分三个阶段实施，每个阶段独立可交付：

**阶段一：UI/UX 改进（2-3 周）**
- 侧边栏导航优化
- 仪表盘增强
- 卡片式 UI 组件
- PWA 支持

**阶段二：功能扩展（3-4 周）**
- 插件系统架构
- Telegram 机器人集成
- 资源搜索集成

**阶段三：架构增强（1-2 周）**
- 多渠道消息推送
- 平台扩展准备（文档和接口优化）

### 3.2 详细任务清单

#### 阶段一：UI/UX 改进

| 任务 | 优先级 | 预计工时 | 依赖 |
|------|--------|----------|------|
| 导航配置化 | P0 | 2天 | 无 |
| 侧边栏组件重构 | P0 | 3天 | 导航配置化 |
| 统计图表组件 | P1 | 2天 | 无 |
| 仪表盘布局重构 | P1 | 2天 | 统计图表组件 |
| 账号卡片组件 | P1 | 2天 | 无 |
| 任务卡片组件 | P1 | 2天 | 无 |
| PWA 配置 | P2 | 1天 | 无 |
| Service Worker | P2 | 2天 | PWA 配置 |
| iOS 适配 | P2 | 1天 | PWA 配置 |
| 安装引导组件 | P2 | 1天 | Service Worker |

#### 阶段二：功能扩展

| 任务 | 优先级 | 预计工时 | 依赖 |
|------|--------|----------|------|
| 插件接口定义 | P0 | 1天 | 无 |
| 插件管理器 | P0 | 3天 | 插件接口定义 |
| 插件 API | P0 | 2天 | 插件管理器 |
| 插件管理页面 | P1 | 2天 | 插件 API |
| Telegram Bot 核心 | P0 | 3天 | 无 |
| Telegram 命令处理 | P0 | 2天 | Telegram Bot 核心 |
| Telegram 配置 API | P1 | 1天 | Telegram Bot 核心 |
| 搜索客户端 | P1 | 2天 | 无 |
| 搜索 API | P1 | 1天 | 搜索客户端 |
| 资源搜索页面 | P1 | 2天 | 搜索 API |

#### 阶段三：架构增强

| 任务 | 优先级 | 预计工时 | 依赖 |
|------|--------|----------|------|
| 通知接口定义 | P0 | 1天 | 无 |
| 通知管理器 | P0 | 2天 | 通知接口定义 |
| 企业微信推送 | P1 | 1天 | 通知管理器 |
| Telegram 推送 | P1 | 1天 | 通知管理器 |
| WxPusher 推送 | P1 | 1天 | 通知管理器 |
| 通知配置 API | P1 | 1天 | 通知管理器 |
| 通知配置页面 | P1 | 1天 | 通知配置 API |
| 平台扩展文档 | P2 | 1天 | 无 |

## 4. 技术选型

### 4.1 前端

| 技术 | 用途 | 版本 |
|------|------|------|
| Vue 3 | 框架 | 3.5+ |
| Element Plus | UI 组件库 | 2.13+ |
| ECharts | 图表库 | 5.x |
| Vue Router | 路由 | 5.x |
| Pinia | 状态管理 | 3.x |
| Axios | HTTP 客户端 | 1.x |
| lucide-vue-next | 图标库 | latest |

### 4.2 后端

| 技术 | 用途 | 版本 |
|------|------|------|
| Go | 语言 | 1.25+ |
| Gin | HTTP 框架 | latest |
| GORM | ORM | latest |
| glebarez/sqlite | SQLite 驱动 | latest |
| robfig/cron | 定时调度 | v3 |
| slog | 日志 | 标准库 |

### 4.3 新增依赖

| 依赖 | 用途 | 风险评估 |
|------|------|----------|
| Chart.js / ECharts | 前端图表 | 低风险，纯前端库 |
| Telegram Bot API | Telegram 集成 | 低风险，HTTP API |
| CloudSaver SDK | 资源搜索 | 中风险，第三方服务 |

## 5. 验证方案

### 5.1 UI/UX 改进验证

- **侧边栏导航**：验证折叠/展开、搜索、移动端响应式
- **仪表盘**：验证图表渲染、实时更新、不同屏幕尺寸
- **卡片组件**：验证表格/卡片切换、动画效果、数据绑定
- **PWA**：验证安装流程、离线访问、iOS 兼容性

### 5.2 功能扩展验证

- **插件系统**：验证插件加载、生命周期钩子、配置管理
- **Telegram**：验证命令执行、通知推送、白名单控制
- **资源搜索**：验证搜索结果、一键创建任务、错误处理

### 5.3 架构增强验证

- **多渠道推送**：验证各渠道配置、测试发送、批量汇总
- **平台扩展**：验证接口文档、示例代码、测试覆盖

### 5.4 端到端测试

- 使用 Playwright 编写 E2E 测试
- 覆盖核心流程：创建任务、执行转存、查看结果
- 验证新功能：插件管理、Telegram 命令、资源搜索

## 6. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 插件系统设计复杂 | 延期 | 参考 quark-auto-save 成熟方案 |
| Telegram API 限制 | 功能受限 | 实现重试机制、错误处理 |
| 第三方搜索服务不稳定 | 搜索失败 | 多源备份、降级方案 |
| PWA iOS 兼容性 | 体验不一致 | 逐步适配、降级方案 |

## 7. 总结

本方案对 UCAS 进行全面升级，涵盖 UI/UX 改进、功能扩展、架构增强三个方面。通过参考 quark-auto-save、cloudpan-auto-save、TgtoDrive 三个优秀项目的成熟经验，结合 UCAS 自身的 Go + Vue 3 技术栈优势，打造一流的云盘自动化平台。

方案分三个阶段实施，每个阶段独立可交付，总预计工期 6-9 周。第一阶段聚焦 UI/UX 改进，快速见效；第二阶段引入核心功能扩展；第三阶段完善架构增强和文档。
