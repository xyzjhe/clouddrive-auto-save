# UI 全面重构设计文档 — Apple/Linear 风格

> 日期：2026-06-03
> 状态：已批准
> 范围：全部页面统一重构

## 1. 背景与目标

当前 UI 采用「暗黑霓虹科技风 + 毛玻璃玻璃拟态」设计语言，存在以下问题：

- 霓虹发光效果（neon glow）视觉噪音大，缺乏品质感
- 毛玻璃（glassmorphism）性能开销高且在某些浏览器表现不一致
- Emoji 图标与 lucide 图标混用，风格不统一
- 配色过于花哨，缺乏层次和克制
- 卡片/按钮/标签样式粗糙，缺少精致度

**目标**：将整体视觉风格从「霓虹科技风」迁移至 **Apple / Linear 风格**——极简克制、大量留白、精致圆角卡片、柔和阴影、灰阶为主 + 天空蓝单一强调色。

**核心原则**：
1. 去除所有霓虹发光、毛玻璃效果、脉冲动画
2. 纯白不透明卡片 + 柔和阴影
3. 暖灰阶 + 单一强调色（Sky Blue）
4. 图标从 lucide 换为 Phosphor Icons
5. 亮色模式为主，暗色模式次之

## 2. 设计 Token

### 2.1 配色体系

#### 亮色模式（主）

| 变量 | 值 | 用途 |
|------|-----|------|
| `--body-bg` | `#FAFAFA` | 页面背景，纯净浅灰无渐变 |
| `--surface-bg` | `#FFFFFF` | 卡片/表面，纯白不透明 |
| `--border-color` | `#E5E7EB` (gray-200) | 边框 |
| `--text-primary` | `#111827` (gray-900) | 主文字 |
| `--text-secondary` | `#6B7280` (gray-500) | 次文字 |
| `--text-muted` | `#9CA3AF` (gray-400) | 弱化文字 |

#### 强调色

| 变量 | 值 | 用途 |
|------|-----|------|
| `--accent` | `#0EA5E9` | 主强调，按钮/链接/选中态 |
| `--accent-light` | `#E0F2FE` | 强调色浅底，hover/标签背景 |
| `--accent-dark` | `#0284C7` | 强调色深色，按下态 |

#### 语义色

| 变量 | 值 | 用途 |
|------|-----|------|
| `--color-success` | `#10B981` | 成功 |
| `--color-warning` | `#F59E0B` | 警告 |
| `--color-danger` | `#EF4444` | 错误 |

#### 平台色

| 变量 | 值 | 用途 |
|------|-----|------|
| `--color-139` | `#EA580C` | 139 平台标识 |
| `--color-quark` | `#10B981` | 夸克平台标识 |

#### 暗色模式（次）

| 变量 | 值 | 用途 |
|------|-----|------|
| `--body-bg` | `#0F172A` (slate-900) | 页面背景 |
| `--surface-bg` | `#1E293B` (slate-800) | 卡片表面 |
| `--border-color` | `#334155` (slate-700) | 边框 |
| `--text-primary` | `#F1F5F9` (slate-100) | 主文字 |
| `--text-secondary` | `#94A3B8` (slate-400) | 次文字 |
| `--text-muted` | `#64748B` (slate-500) | 弱化文字 |
| `--accent` | `#38BDF8` (sky-400) | 强调色（比亮色稍亮） |
| `--accent-light` | `rgba(56,189,248,0.15)` | 强调色浅底 |

### 2.2 圆角

| 变量 | 值 | 用途 |
|------|-----|------|
| `--radius-sm` | `6px` | 标签、小按钮、输入框 |
| `--radius-md` | `10px` | 搜索栏、中等面板 |
| `--radius-lg` | `14px` | 卡片、对话框、抽屉 |
| `--radius-full` | `9999px` | 头像、圆形按钮 |

### 2.3 阴影

| 变量 | 值 | 用途 |
|------|-----|------|
| `--shadow-xs` | `0 1px 2px rgba(0,0,0,0.04)` | 标签、小元素 |
| `--shadow-sm` | `0 1px 3px rgba(0,0,0,0.06), 0 1px 2px rgba(0,0,0,0.04)` | 输入框、按钮 |
| `--shadow-md` | `0 4px 6px rgba(0,0,0,0.04), 0 2px 4px rgba(0,0,0,0.03)` | 卡片 |
| `--shadow-lg` | `0 10px 15px rgba(0,0,0,0.05), 0 4px 6px rgba(0,0,0,0.03)` | 弹窗、下拉 |

**关键**：去除所有 `backdrop-filter: blur()` 毛玻璃效果、去除所有 `neon-glow` 发光阴影。

### 2.4 字体

| 变量 | 值 | 说明 |
|------|-----|------|
| `--font-sans` | `'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif` | UI 字体，从 Plus Jakarta Sans 换为 Inter |
| `--font-mono` | `'JetBrains Mono', 'Fira Code', monospace` | 代码/日志字体，保持不变 |

Inter 通过 Google Fonts CDN 在 `index.html` 中加载（400/500/600/700 字重）。

### 2.5 间距

采用 4px 网格体系：`4/8/12/16/20/24/32/40/48px`

### 2.6 过渡

| 变量 | 值 | 用途 |
|------|-----|------|
| `--transition-fast` | `150ms ease` | 按钮、图标 hover |
| `--transition-base` | `200ms ease` | 卡片、面板 |
| `--transition-slow` | `300ms ease` | 页面过渡、展开收起 |

## 3. 组件样式规范

### 3.1 卡片 (Card)

- 背景：`var(--surface-bg)` = `#FFFFFF`，不透明，无 `backdrop-filter`
- 边框：`1px solid var(--border-color)`
- 圆角：`var(--radius-lg)` = `14px`
- 阴影：`var(--shadow-md)`
- Hover：阴影加深至 `--shadow-lg` + `translateY(-1px)`
- 内边距：`20px 24px`

### 3.2 按钮 (Button)

| 类型 | 背景 | 文字 | 边框 |
|------|------|------|------|
| Primary | `var(--accent)` | `#FFFFFF` | 无 |
| Secondary | `#FFFFFF` | `var(--text-secondary)` | `1px solid var(--border-color)` |
| Ghost | 透明 | `var(--text-secondary)` | 无 |
| Danger | 透明 | `var(--color-danger)` | 无 |

- 圆角：`var(--radius-sm)` = `6px`
- 高度：`32px`（small）/ `36px`（default）
- 字重：`500`
- 图标按钮：圆形（`border-radius: 9999px`），`32px` 直径，Ghost 风格，hover `#F3F4F6` 底

### 3.3 输入框 (Input)

- 背景：`#FFFFFF`
- 边框：`1px solid var(--border-color)`
- 圆角：`var(--radius-sm)` = `6px`
- Focus：边框 `var(--accent)` + `box-shadow: 0 0 0 3px var(--accent-light)`

### 3.4 标签 (Tag)

- 圆角：`var(--radius-sm)` = `6px`
- 统一浅底 + 深色文字风格：

| 类型 | 背景 | 文字 |
|------|------|------|
| 信息/accent | `#E0F2FE` | `#0284C7` |
| 成功 | `#D1FAE5` | `#059669` |
| 警告 | `#FEF3C7` | `#D97706` |
| 错误 | `#FEE2E2` | `#DC2626` |
| 中性 | `#F3F4F6` | `#4B5563` |

### 3.5 表格 (Table)

- 透明背景，行间 `var(--border-color)` 分隔
- 表头：`var(--text-secondary)` 色，`500` 字重，`13px`
- Hover 行：`#F9FAFB` (gray-50)
- 操作列：圆形 Ghost 图标按钮

### 3.6 弹窗/抽屉 (Dialog/Drawer)

- 背景：`#FFFFFF`
- 圆角：`var(--radius-lg)` = `14px`
- 阴影：`var(--shadow-lg)`
- Header 底部 `1px solid var(--border-color)` 边框
- 去除毛玻璃效果

### 3.7 进度条 (Progress)

- 轨道：`#E5E7EB`
- 填充：语义色
- 圆角：`9999px`（胶囊形）
- 高度：`6px`（线性）/ `8px stroke`（圆环）
- 去除条纹流动动画，改用平滑宽度过渡

### 3.8 导航侧边栏

- 宽度：`220px`（从 240px 收窄）
- 背景：`#FFFFFF`，右边框分隔
- Logo：纯文字 "UCAS"，`var(--accent)` 色，`20px/700`，去除 CloudLogo SVG 呼吸动画
- 分组标题：纯文字大写，`var(--text-muted)` 色，`11px/600`，去除 emoji
- 导航项：圆角 `6px`，hover `#F3F4F6` 底，选中态 `var(--accent-light)` 底 + `var(--accent)` 文字 + 左侧 `3px` 蓝色竖条
- 去除 `<ChevronRight>` 折叠箭头，分组固定展示

## 4. 页面设计

### 4.1 MainLayout（全局骨架）

**顶栏**：
- 高度 `56px`（从 64px 收窄）
- 背景 `#FFFFFF`，底部 `1px solid var(--border-color)` 边框
- 左侧面包屑，右侧圆形暗色模式切换 + 圆形头像
- 去除毛玻璃效果

**内容区**：
- 内边距 `24px`
- 页面过渡保持 `fade-page`

### 4.2 Dashboard（控制台）

**布局变更**：三栏 → 两栏

- **左栏（主区，约 70%）**：
  - 顶部：4 个统计磁贴，白卡片 + 彩色数值 + 灰色标签。数值 `24px/700` 语义色，标签 `13px` + `var(--text-muted)`
  - 中部：活跃任务队列白卡片列表，每行轻底色区分。运行中用 Phosphor `Spinner` + `var(--accent)`，成功 `CheckCircle` 绿色，失败 `WarningCircle` 红色
  - 底部：近期活动时间线，`el-timeline`，简洁圆点 + 文字描述，无脉冲动画

- **右栏（约 30%）**：
  - 系统状态卡片：CPU/RAM 线性进度条，存储小号环形图（`80px`）
  - 日志面板：白底卡片 + 等宽字体列表（非纯黑终端）。日志行按级别用极淡背景色区分（ERROR 极淡红底、WARN 极淡黄底、SUCCESS 极淡绿底）。每页 50 条，底部"查看全部"链接

- **去除**：`breath-glow` 动画、脉冲圆点、`AUTO-SAVE ACTIVE` 状态框、纯黑背景、条纹流动进度条、CloudLogo 呼吸动画

### 4.3 Tasks（任务管理）

- 页面标题：`24px/700`，右侧操作按钮组保持
- 筛选栏：文字按钮组样式（无底色边框，选中态下划线 + `var(--accent)` 色）
- 表格视图：新表格规范。任务名粗体 + 平台小标签。操作列圆形 Ghost 图标按钮
- 卡片视图：白卡片 + `14px` 圆角，无顶部渐变条，内容区间 `var(--border-color)` 细线分隔
- 创建/编辑抽屉：`560px` 宽，表单标签 `var(--text-secondary)` + `13px`
- 所有弹窗应用新卡片/表格/按钮规范

### 4.4 Accounts（账号管理）

- 表格视图：平台用 Phosphor `HardDrives` 统一图标 + 文字标签区分。存储进度条新规范
- 卡片视图：白卡片 + `14px` 圆角。顶部平台名+昵称+状态标签，中部存储环形图 `80px` + 数据，底部操作链接
- 添加账号弹窗：`480px` 宽，Alert 改为浅底标签样式

### 4.5 Settings（系统设置）

- Tab 栏：去除 `border-card`，改为底部下划线 Tab（选中态 `var(--accent)` 下划线）
- 消息推送嵌套 Tab 也用下划线风格
- 插件卡片白底 + `14px` 圆角，图标用 Phosphor `PuzzlePiece` duotone
- "安装新插件"虚线边框占位卡保持

### 4.6 Search（资源搜索）

- 搜索栏：大号输入框 `10px` 圆角，Primary 搜索按钮
- 筛选区：简洁灰色文字标签
- 验证状态：去除 emoji，改用 Phosphor `CheckCircle`/`XCircle`/`Spinner` `16px` 语义色图标
- 结果卡片：白底 + `14px` 圆角，标题 `16px/600`，元数据 `var(--text-muted)` + 小号 Phosphor 图标 `14px`

## 5. 图标迁移计划

### 5.1 图标库替换

从 `lucide-vue-next` 迁移至 `@phosphor-icons/vue`。

主要使用的 Phosphor 变体：
- **Regular**：导航图标、操作图标（默认变体）
- **Duotone**：空状态图标、品牌标识、Dashboard 统计图标
- **Fill**：选中态图标

### 5.2 图标映射表

| 当前 lucide | Phosphor 替代 | 变体 |
|------------|--------------|------|
| `LayoutDashboard` | `SquaresFour` | Regular |
| `User` | `User` | Regular |
| `ListChecks` | `ListChecks` | Regular |
| `Settings` | `GearSix` | Regular |
| `Search` | `MagnifyingGlass` | Regular |
| `Puzzle` | `PuzzlePiece` | Duotone |
| `Bell` | `Bell` | Regular |
| `Moon`/`Sun` | `Moon`/`Sun` | Regular |
| `Play` | `Play` | Fill |
| `Edit` | `PencilSimple` | Regular |
| `Trash2` | `Trash` | Regular |
| `RefreshCw` | `ArrowsClockwise` | Regular |
| `Plus` | `Plus` | Regular |
| `Folder`/`File` | `Folder`/`File` | Regular |
| `Info` | `Info` | Regular |
| `Cloud` | `Cloud` | Duotone |
| `CheckCircle2` | `CheckCircle` | Fill |
| `AlertCircle`/`AlertTriangle` | `WarningCircle`/`Warning` | Fill |
| `Loader2` | `Spinner` | Regular |
| `Terminal` | `Terminal` | Regular |
| `Github` | `GithubLogo` | Regular |
| `ExternalLink` | `ArrowSquareOut` | Regular |
| `HardDrive` | `HardDrives` | Regular |
| `Calendar` | `CalendarBlank` | Regular |
| `Clock` | `Clock` | Regular |
| `Link` | `Link` | Regular |
| `List`/`LayoutGrid` | `List`/`GridFour` | Regular |

### 5.3 图标使用规范

- 默认尺寸：`20px`（导航项）、`18px`（按钮内）、`16px`（表格行内、元数据旁）、`14px`（小标签旁）
- 默认颜色：`currentColor`（继承文字色），不使用独立彩色
- 间距：图标与文字间距 `6px`

## 6. 技术实施策略

### 6.1 自底向上：设计系统优先

**阶段 1：设计系统重建**
1. 重写 `web/src/variables.css` — 全新 Design Token
2. 重写 `web/src/style.css` — 全局组件样式覆盖（尽量减少 `!important`，优先使用 Element Plus CSS 变量覆盖；对确实需要穿透的深层样式保留 `!important`）
3. 更新 `web/src/main.js` — 字体引入（Inter 替换 Plus Jakarta Sans）
4. 更新 `web/index.html` — Google Fonts CDN 链接

**阶段 2：图标迁移**
1. 安装 `@phosphor-icons/vue`
2. 更新所有组件中的图标导入和引用
3. 去除 emoji 图标（导航分组、PWA、验证状态）
4. 卸载 `lucide-vue-next`

**阶段 3：页面逐个重构**
1. `MainLayout.vue` — 侧边栏 + 顶栏新规范
2. `Dashboard.vue` — 三栏改两栏，去除终端面板
3. `Tasks.vue` — 筛选栏、表格、卡片、抽屉新规范
4. `Accounts.vue` — 表格、卡片、弹窗新规范
5. `Settings.vue` — Tab 样式、表单、插件网格新规范
6. `Search.vue` — 搜索栏、结果卡片、验证图标新规范

**阶段 4：组件清理**
1. 删除 `CloudLogo.vue`（不再使用）
2. 删除 `variables.css` 中所有 `--neon-*` 变量
3. 删除 `style.css` 中所有 `breath-glow`/`neon-pulse`/`glass-card` 样式
4. 清理各组件 scoped 样式中的 `html.dark` 硬编码覆盖

### 6.2 关键文件清单

| 文件 | 变更类型 | 说明 |
|------|---------|------|
| `web/src/variables.css` | 重写 | 全新 Design Token |
| `web/src/style.css` | 重写 | 全局组件样式 |
| `web/index.html` | 修改 | 字体 CDN |
| `web/src/main.js` | 无变更 | — |
| `web/src/layout/MainLayout.vue` | 重构 | 侧边栏+顶栏 |
| `web/src/views/Dashboard.vue` | 重构 | 布局+组件 |
| `web/src/views/Tasks.vue` | 重构 | 筛选+表格+卡片+抽屉 |
| `web/src/views/Accounts.vue` | 重构 | 表格+卡片+弹窗 |
| `web/src/views/Settings.vue` | 重构 | Tab+表单+插件 |
| `web/src/views/Search.vue` | 重构 | 搜索栏+结果+验证 |
| `web/src/components/CloudLogo.vue` | 删除 | 不再使用 |
| `web/src/components/cards/TaskCard.vue` | 修改 | 新卡片规范 |
| `web/src/components/cards/AccountCard.vue` | 修改 | 新卡片规范 |
| `web/src/components/SidebarFooter.vue` | 修改 | 去除硬编码 `html.dark` |
| `web/src/components/ShareContentDialog.vue` | 修改 | 新弹窗规范 |
| `web/src/config/navigation.ts` | 修改 | 去除 emoji，更新图标名 |

## 7. 验证方案

1. `make dev-web` 启动前端开发服务器，逐页面检查：
   - 亮色模式下所有组件（卡片、按钮、输入框、标签、表格、弹窗）的视觉表现
   - 暗色模式下基础可用性（不花哨但可读）
2. 响应式检查：桌面（1440px）、笔记本（1280px）、平板（768px）
3. `make check` 运行完整 CI 流水线确保无编译错误
4. `make e2e-test` 运行 Playwright 测试确保功能无回归
5. 图标迁移验证：确认所有页面无遗漏的 lucide 引用
