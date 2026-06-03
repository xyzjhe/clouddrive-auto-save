# UI 全面重构实施计划 — Apple/Linear 风格

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将整体 UI 视觉风格从「霓虹科技风」迁移至 Apple/Linear 极简风格——灰阶 + Sky Blue 强调色、纯白不透明卡片、柔和阴影、Phosphor 图标。

**Architecture:** 自底向上重构。先重建设计系统（Design Tokens + 全局组件覆盖），再迁移图标库，最后逐页重构布局和样式。每个 Task 产出可独立验证的成果。

**Tech Stack:** Vue 3 + Element Plus + @phosphor-icons/vue + Inter 字体

---

## 文件变更总览

| 文件 | 操作 | 职责 |
|------|------|------|
| `web/src/variables.css` | **重写** | 全新 Apple/Linear Design Tokens |
| `web/src/style.css` | **重写** | 全局 Element Plus 组件覆盖（去除霓虹/毛玻璃） |
| `web/index.html` | **修改** | Google Fonts CDN：Plus Jakarta Sans → Inter |
| `web/src/config/navigation.ts` | **修改** | 去除 emoji 分组图标，更新图标名为 Phosphor 名 |
| `web/src/layout/MainLayout.vue` | **重构** | 侧边栏精简（220px、去 CloudLogo、去折叠箭头、新选中态）+ 顶栏（56px、去毛玻璃） |
| `web/src/components/CloudLogo.vue` | **删除** | 被 "UCAS" 纯文字替代 |
| `web/src/components/SidebarFooter.vue` | **修改** | 图标 lucide → Phosphor，去除 `html.dark` 硬编码 |
| `web/src/components/cards/TaskCard.vue` | **修改** | 新卡片规范 + 圆形 Ghost 图标按钮 + Phosphor 图标 |
| `web/src/components/cards/AccountCard.vue` | **修改** | 新卡片规范（白底 + 14px 圆角，去除渐变头部）+ Phosphor |
| `web/src/components/PWAInstall.vue` | **修改** | emoji ☁️ → Phosphor `CloudDuotone` 图标 |
| `web/src/components/ShareContentDialog.vue` | **修改** | Phosphor 图标 + 新弹窗规范样式 |
| `web/src/views/Dashboard.vue` | **重构** | 三栏 → 两栏布局，去除终端面板/脉冲动画，白底日志列表 |
| `web/src/views/Tasks.vue` | **重构** | 筛选栏文字按钮组 + 新表格/卡片/抽屉规范 + Phosphor |
| `web/src/views/Accounts.vue` | **重构** | 新表格/卡片规范 + Phosphor |
| `web/src/views/Settings.vue` | **重构** | 下划线 Tab + 新表单/插件卡片规范 + Phosphor |
| `web/src/views/Search.vue` | **重构** | 去除 emoji 验证状态 → Phosphor 图标 + 新结果卡片规范 |
| `web/package.json` | **修改** | 安装 `@phosphor-icons/vue`，卸载 `lucide-vue-next` |

---

## Task 1: 重建设计 Token 系统

**Files:**
- Rewrite: `web/src/variables.css`

- [ ] **Step 1: 用全新 Apple/Linear Design Token 重写 variables.css**

将 `web/src/variables.css` 全部内容替换为：

```css
:root {
  /* ==========================================
   * 亮色模式 (Light Mode) — Apple/Linear 风格
   * ========================================== */

  /* 页面与表面 */
  --body-bg: #FAFAFA;
  --surface-bg: #FFFFFF;
  --border-color: #E5E7EB;
  --bg-secondary: #F9FAFB;

  /* 强调色 (Sky Blue) */
  --accent: #0EA5E9;
  --accent-light: #E0F2FE;
  --accent-dark: #0284C7;

  /* 文字色 */
  --text-primary: #111827;
  --text-secondary: #6B7280;
  --text-muted: #9CA3AF;

  /* 语义色 */
  --color-success: #10B981;
  --color-success-light: #D1FAE5;
  --color-success-text: #059669;
  --color-warning: #F59E0B;
  --color-warning-light: #FEF3C7;
  --color-warning-text: #D97706;
  --color-danger: #EF4444;
  --color-danger-light: #FEE2E2;
  --color-danger-text: #DC2626;

  /* 平台色 */
  --color-139: #EA580C;
  --color-quark: #10B981;

  /* 圆角 */
  --radius-sm: 6px;
  --radius-md: 10px;
  --radius-lg: 14px;
  --radius-full: 9999px;

  /* 阴影 */
  --shadow-xs: 0 1px 2px rgba(0, 0, 0, 0.04);
  --shadow-sm: 0 1px 3px rgba(0, 0, 0, 0.06), 0 1px 2px rgba(0, 0, 0, 0.04);
  --shadow-md: 0 4px 6px rgba(0, 0, 0, 0.04), 0 2px 4px rgba(0, 0, 0, 0.03);
  --shadow-lg: 0 10px 15px rgba(0, 0, 0, 0.05), 0 4px 6px rgba(0, 0, 0, 0.03);

  /* 过渡 */
  --transition-fast: 150ms ease;
  --transition-base: 200ms ease;
  --transition-slow: 300ms ease;

  /* 字体 */
  --font-sans: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  --font-mono: 'JetBrains Mono', 'Fira Code', 'Courier New', monospace;

  /* 导航及组件背景 */
  --bg-sidebar: #FFFFFF;
  --bg-navbar: #FFFFFF;
  --bg-content: #FFFFFF;
  --dialog-bg: #FFFFFF;
  --input-bg: #FFFFFF;
  --hover-bg: #F3F4F6;
  --tag-info-bg: #F3F4F6;
  --switch-bg: #E5E7EB;
  --scrollbar-track-bg: #F3F4F6;
  --border: #E5E7EB;

  /* 灰色色阶 */
  --neutral-50: #F9FAFB;
  --neutral-100: #F3F4F6;
  --neutral-200: #E5E7EB;
  --neutral-300: #D1D5DB;
  --neutral-400: #9CA3AF;
  --neutral-500: #6B7280;
  --neutral-600: #4B5563;
  --neutral-700: #374151;
  --neutral-800: #1F2937;

  /* 按钮系统 */
  --btn-radius: 6px;
  --btn-font-weight: 500;
  --btn-transition: all 150ms ease;
  --btn-primary-bg: var(--accent);
  --btn-primary-text: #FFFFFF;
  --btn-icon-size: 32px;
}

html.dark {
  /* ==========================================
   * 暗色模式 (Dark Mode) — 精致深灰
   * ========================================== */

  --body-bg: #0F172A;
  --surface-bg: #1E293B;
  --border-color: #334155;
  --bg-secondary: #162032;

  --accent: #38BDF8;
  --accent-light: rgba(56, 189, 248, 0.15);
  --accent-dark: #0EA5E9;

  --text-primary: #F1F5F9;
  --text-secondary: #94A3B8;
  --text-muted: #64748B;

  --color-danger: #F87171;
  --color-success: #34D399;
  --color-warning: #FBBF24;

  --shadow-sm: 0 1px 3px rgba(0, 0, 0, 0.3), 0 1px 2px rgba(0, 0, 0, 0.2);
  --shadow-md: 0 4px 6px rgba(0, 0, 0, 0.3), 0 2px 4px rgba(0, 0, 0, 0.2);
  --shadow-lg: 0 10px 15px rgba(0, 0, 0, 0.4), 0 4px 6px rgba(0, 0, 0, 0.2);

  --bg-sidebar: #0F172A;
  --bg-navbar: #0F172A;
  --dialog-bg: #1E293B;
  --input-bg: #1E293B;
  --hover-bg: rgba(255, 255, 255, 0.05);
  --tag-info-bg: rgba(255, 255, 255, 0.06);
  --switch-bg: rgba(255, 255, 255, 0.1);
  --scrollbar-track-bg: #1E293B;
  --border: #334155;

  --neutral-50: rgba(255, 255, 255, 0.03);
  --neutral-100: rgba(255, 255, 255, 0.05);
  --neutral-200: rgba(255, 255, 255, 0.08);
  --neutral-500: #94A3B8;
  --neutral-600: #CBD5E1;
  --neutral-700: #E2E8F0;
  --neutral-800: #F1F5F9;
}
```

- [ ] **Step 2: 验证变量文件无语法错误**

Run: `cd /home/zcq/Github/clouddrive-auto-save/web && npx vite build --mode development 2>&1 | head -5`
Expected: 无 CSS 相关错误

- [ ] **Step 3: 提交**

```bash
git add web/src/variables.css
git commit -m "refactor(ui): 重建设计 Token 系统，Apple/Linear 风格

以暖灰阶 + Sky Blue #0EA5E9 强调色替代霓虹配色体系。
去除所有 --neon-* 变量、毛玻璃背景和径向渐变。
新增圆角/阴影/过渡/语义色完整 Token 体系。"
```

---

## Task 2: 重写全局组件样式覆盖

**Files:**
- Rewrite: `web/src/style.css`

- [ ] **Step 1: 用 Apple/Linear 风格重写 style.css**

将 `web/src/style.css` 全部内容替换为：

```css
@import './variables.css';

body {
  margin: 0;
  font-family: var(--font-sans);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  background: var(--body-bg);
  min-height: 100vh;
  color: var(--text-primary);
  letter-spacing: -0.01em;
}

/* ==========================================
 * 全局卡片 — 纯白不透明 + 柔和阴影
 * ========================================== */
.el-card {
  border: 1px solid var(--border-color) !important;
  background: var(--surface-bg) !important;
  border-radius: var(--radius-lg) !important;
  box-shadow: var(--shadow-md) !important;
  transition: box-shadow var(--transition-base), transform var(--transition-base);
}

.el-card:hover {
  box-shadow: var(--shadow-lg) !important;
  transform: translateY(-1px);
}

.el-main {
  padding: 24px !important;
}

/* ==========================================
 * 滚动条 — 简洁灰色
 * ========================================== */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: var(--neutral-300);
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background: var(--neutral-400);
}

/* ==========================================
 * 按钮 — Apple/Linear 风格
 * ========================================== */
.el-button {
  border-radius: var(--btn-radius) !important;
  font-weight: var(--btn-font-weight) !important;
  transition: all var(--transition-fast) !important;
}

.el-button--primary {
  background: var(--accent) !important;
  color: #FFFFFF !important;
  border: none !important;
}

.el-button--primary:hover {
  background: var(--accent-dark) !important;
  transform: translateY(-1px);
}

.el-button--primary:active {
  transform: translateY(0);
}

.el-button--primary.is-plain {
  background: #FFFFFF !important;
  color: var(--accent) !important;
  border: 1px solid var(--border-color) !important;
}

.el-button--primary.is-plain:hover {
  border-color: var(--accent) !important;
  color: var(--accent) !important;
  background: var(--accent-light) !important;
}

.el-button--info.is-plain {
  background: transparent !important;
  color: var(--text-secondary) !important;
  border: 1px solid var(--border-color) !important;
}

.el-button--info.is-plain:hover {
  border-color: var(--text-secondary) !important;
  color: var(--text-primary) !important;
}

.el-button--success.is-plain {
  color: var(--color-success) !important;
  border-color: var(--color-success) !important;
}

.el-button--danger.is-plain {
  color: var(--color-danger) !important;
  border-color: var(--color-danger) !important;
}

.el-button.is-link {
  padding: 4px 8px !important;
  height: auto !important;
}

.el-button.is-link:hover {
  opacity: 0.8;
}

/* 圆形 Ghost 图标按钮 */
.action-buttons {
  display: flex;
  gap: 4px;
  align-items: center;
}

.btn-icon {
  width: var(--btn-icon-size);
  height: var(--btn-icon-size);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-full);
  cursor: pointer;
  transition: all var(--transition-fast);
  border: none;
  font-size: 16px;
  padding: 0;
  background: transparent;
  color: var(--text-secondary);
}

.btn-icon:hover {
  background: var(--hover-bg);
  color: var(--text-primary);
}

.btn-icon--success:hover {
  background: var(--color-success-light);
  color: var(--color-success);
}

.btn-icon--primary:hover {
  background: var(--accent-light);
  color: var(--accent);
}

.btn-icon--danger:hover {
  background: var(--color-danger-light);
  color: var(--color-danger);
}

.btn-icon:disabled {
  opacity: 0.4;
  cursor: not-allowed;
  transform: none !important;
}

.btn-icon:disabled:hover {
  background: transparent;
  color: var(--text-secondary);
}

.btn-icon:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

/* ==========================================
 * 弹窗/抽屉 — 纯白 + 柔和阴影
 * ========================================== */
.el-dialog, .el-drawer {
  background: var(--dialog-bg) !important;
  border: 1px solid var(--border-color) !important;
  box-shadow: var(--shadow-lg) !important;
}

.el-dialog {
  border-radius: var(--radius-lg) !important;
}

.el-dialog__header, .el-drawer__header {
  padding: 20px 24px 16px !important;
  border-bottom: 1px solid var(--border-color) !important;
  margin-right: 0 !important;
}

.el-dialog__body, .el-drawer__body {
  padding: 24px !important;
}

.el-dialog__footer {
  padding: 16px 24px 20px !important;
  border-top: 1px solid var(--border-color) !important;
}

/* ==========================================
 * 表单输入框
 * ========================================== */
.el-input__wrapper, .el-textarea__inner {
  background: var(--input-bg) !important;
  border: 1px solid var(--border-color) !important;
  border-radius: var(--radius-sm) !important;
  box-shadow: none !important;
  color: var(--text-primary) !important;
}

.el-input__wrapper:hover, .el-textarea__inner:hover {
  border-color: var(--neutral-300) !important;
}

.el-input__wrapper:focus-within, .el-textarea__inner:focus {
  border-color: var(--accent) !important;
  box-shadow: 0 0 0 3px var(--accent-light) !important;
}

.el-input__inner, .el-textarea__inner {
  color: var(--text-primary) !important;
}

.el-input__inner::placeholder, .el-textarea__inner::placeholder {
  color: var(--text-muted) !important;
}

/* ==========================================
 * 下拉框
 * ========================================== */
.el-select-dropdown {
  background: var(--dialog-bg) !important;
  border: 1px solid var(--border-color) !important;
  box-shadow: var(--shadow-lg) !important;
  border-radius: var(--radius-md) !important;
}

.el-select-dropdown__item {
  color: var(--text-secondary) !important;
}

.el-select-dropdown__item.hover, .el-select-dropdown__item:hover {
  background-color: var(--hover-bg) !important;
  color: var(--text-primary) !important;
}

.el-select-dropdown__item.selected {
  color: var(--accent) !important;
  font-weight: 600 !important;
  background-color: var(--accent-light) !important;
}

/* ==========================================
 * 表格
 * ========================================== */
.el-table {
  background: transparent !important;
  color: var(--text-primary) !important;
}

.el-table th, .el-table tr {
  background: transparent !important;
}

.el-table th.el-table__cell {
  color: var(--text-secondary) !important;
  font-weight: 500 !important;
  font-size: 13px !important;
}

.el-table td, .el-table th.is-leaf {
  border-bottom: 1px solid var(--border-color) !important;
}

.el-table .el-table__row:hover > td {
  background-color: var(--bg-secondary) !important;
}

/* ==========================================
 * 标签 — 统一浅底 + 深色文字
 * ========================================== */
.el-tag {
  border-radius: var(--radius-sm) !important;
  font-weight: 500 !important;
  border: none !important;
}

.el-tag--success {
  background: #D1FAE5 !important;
  color: #059669 !important;
}

.el-tag--warning {
  background: #FEF3C7 !important;
  color: #D97706 !important;
}

.el-tag--danger {
  background: #FEE2E2 !important;
  color: #DC2626 !important;
}

.el-tag--primary {
  background: #E0F2FE !important;
  color: #0284C7 !important;
}

.el-tag--info {
  background: #F3F4F6 !important;
  color: #4B5563 !important;
}

/* ==========================================
 * 开关
 * ========================================== */
.el-switch__core {
  background: var(--switch-bg) !important;
  border: none !important;
}

.el-switch.is-checked .el-switch__core {
  background-color: var(--accent) !important;
}

/* ==========================================
 * 折叠面板
 * ========================================== */
.el-collapse, .el-collapse-item__wrap, .el-collapse-item__header {
  background: transparent !important;
  border-color: var(--border-color) !important;
  color: var(--text-primary) !important;
}

/* ==========================================
 * 进度条 — 胶囊形 + 语义色
 * ========================================== */
.el-progress-bar__outer {
  border-radius: 9999px !important;
  background: #E5E7EB !important;
}

.el-progress-bar__inner {
  border-radius: 9999px !important;
  transition: width 0.4s ease !important;
}

/* 去除条纹流动动画 */
.el-progress-bar__inner.is-stripes {
  background-image: none !important;
}
```

- [ ] **Step 2: 验证构建**

Run: `cd /home/zcq/Github/clouddrive-auto-save/web && npx vite build --mode development 2>&1 | tail -3`
Expected: 构建成功

- [ ] **Step 3: 提交**

```bash
git add web/src/style.css
git commit -m "refactor(ui): 重写全局组件样式，Apple/Linear 风格

去除所有毛玻璃(backdrop-filter)、霓虹发光、脉冲动画。
卡片改为纯白不透明+柔和阴影，按钮统一6px圆角，
标签改为浅底+深色文字，图标按钮改为圆形Ghost风格。"
```

---

## Task 3: 更新字体与 HTML 入口

**Files:**
- Modify: `web/index.html`

- [ ] **Step 1: 将 Google Fonts CDN 从 Plus Jakarta Sans 换为 Inter**

在 `web/index.html` 中，将第 21 行的 Google Fonts 链接：

```html
<link href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;500;600;700;800&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet" />
```

替换为：

```html
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet" />
```

同时将第 11 行的 theme-color 从 `#6366f1` 改为 `#0EA5E9`：

```html
<meta name="theme-color" content="#0EA5E9" />
```

- [ ] **Step 2: 提交**

```bash
git add web/index.html
git commit -m "refactor(ui): 字体从 Plus Jakarta Sans 换为 Inter

Inter 是 Apple/Linear 风格标配字体。同步更新 theme-color 为 Sky Blue。"
```

---

## Task 4: 安装 Phosphor Icons 并迁移导航配置

**Files:**
- Modify: `web/package.json`
- Modify: `web/src/config/navigation.ts`

- [ ] **Step 1: 安装 @phosphor-icons/vue**

Run: `cd /home/zcq/Github/clouddrive-auto-save/web && npm install @phosphor-icons/vue`

- [ ] **Step 2: 更新 navigation.ts — 去除 emoji，更新图标名为 Phosphor 格式**

将 `web/src/config/navigation.ts` 全部内容替换为：

```typescript
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
```

关键变更：
- `NavGroup` 接口去除 `icon`、`collapsible`、`defaultCollapsed` 字段（分组不再需要 emoji 图标和折叠功能）
- 所有导航项的 `icon` 值更新为 Phosphor 组件名

- [ ] **Step 3: 验证构建**

Run: `cd /home/zcq/Github/clouddrive-auto-save/web && npx vite build --mode development 2>&1 | tail -3`
Expected: 构建成功（此时页面会因图标名变更暂时显示空白，待 Task 5 修复）

- [ ] **Step 4: 提交**

```bash
git add web/package.json web/package-lock.json web/src/config/navigation.ts
git commit -m "feat(ui): 安装 Phosphor Icons 并更新导航配置

去除导航分组 emoji 图标和折叠功能，图标名从 lucide 格式
更新为 Phosphor 格式（SquaresFour/GearSix/MagnifyingGlass 等）。"
```

---

## Task 5: 重构 MainLayout — 侧边栏 + 顶栏

**Files:**
- Modify: `web/src/layout/MainLayout.vue`

- [ ] **Step 1: 重写 MainLayout.vue 的 template、script 和 style**

**template 变更要点：**
- `<el-aside width="240px">` → `<el-aside width="220px">`
- Logo 区：去除 `<CloudLogo>` 组件，改为纯文字 `<span class="logo-text">UCAS</span>`
- 去除搜索框的 `<Search />` lucide 导入，改用 Phosphor `MagnifyingGlass`
- 导航分组：去除 `.nav-group-icon`（emoji 显示）、去除 `.nav-group-arrow`（ChevronRight 折叠箭头）
- 导航项：`<component :is="item.icon" />` 改为从 Phosphor 导入
- `<el-header height="64px">` → `<el-header height="56px">`
- 顶栏主题切换按钮图标：`Moon`/`Sun` 改为 Phosphor 版本

**script 变更要点：**
- 去除 `import CloudLogo` 和 `import { ... } from 'lucide-vue-next'`
- 改为 `import { MagnifyingGlass, Moon, Sun, SquaresFour, Users, ListChecks, GearSix } from '@phosphor-icons/vue'`
- 去除 `collapsedGroups` 折叠状态管理（分组固定展示）
- 创建图标映射对象 `iconMap`，将 navigation.ts 中的图标名映射到 Phosphor 组件

**style 变更要点：**
- `.sidebar` 宽度 220px，背景 `#FFFFFF`，无 `--bg-sidebar` 变量
- `.logo` 区高度 56px（与顶栏对齐），文字色 `var(--accent)`
- `.nav-group-header` 去除 `border-bottom`，不去除分组间视觉分隔
- `.nav-item.active` 样式：`var(--accent-light)` 底 + `var(--accent)` 文字 + 左侧 3px 竖条
- `.navbar` 去除 `backdrop-filter`，背景 `#FFFFFF`
- `.theme-toggle` 去除霓虹发光 hover

关键代码片段：

```vue
<template>
  <el-container class="app-wrapper">
    <el-aside width="220px" class="sidebar">
      <div class="logo">
        <span class="logo-text">UCAS</span>
      </div>

      <div class="search-wrapper">
        <el-input v-model="searchQuery" placeholder="搜索功能..." clearable size="small">
          <template #prefix>
            <MagnifyingGlass :size="16" weight="regular" />
          </template>
        </el-input>
      </div>

      <el-scrollbar class="nav-scrollbar">
        <div class="nav-groups">
          <div v-for="group in filteredNavigation" :key="group.name" class="nav-group">
            <div class="nav-group-header">{{ group.name }}</div>
            <div class="nav-items">
              <div
                v-for="item in group.items"
                :key="item.path"
                class="nav-item"
                :class="{ active: isActive(item.path) }"
                @click="navigateTo(item.path)"
              >
                <component :is="iconMap[item.icon]" :size="20" weight="regular" class="nav-item-icon" />
                <span class="nav-item-name">{{ item.name }}</span>
              </div>
            </div>
          </div>
        </div>
      </el-scrollbar>

      <SidebarFooter />
    </el-aside>

    <el-container>
      <el-header height="56px" class="navbar">
        <div class="header-left">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item>首页</el-breadcrumb-item>
            <el-breadcrumb-item>{{ currentPageTitle }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="header-right">
          <el-button circle class="theme-toggle" @click="toggleDark()">
            <component :is="isDark ? Sun : Moon" :size="18" weight="regular" />
          </el-button>
          <el-divider direction="vertical" />
          <el-avatar :size="32" src="https://github.com/identicons/user.png" />
        </div>
      </el-header>

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
```

```vue
<script setup>
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import SidebarFooter from '../components/SidebarFooter.vue'
import { useDark, useToggle } from '@vueuse/core'
import { navigationConfig } from '../config/navigation'
import {
  MagnifyingGlass,
  Moon,
  Sun,
  SquaresFour,
  Users,
  ListChecks,
  GearSix
} from '@phosphor-icons/vue'

const iconMap = {
  SquaresFour,
  Users,
  ListChecks,
  MagnifyingGlass,
  GearSix,
}

const route = useRoute()
const router = useRouter()

if (!localStorage.getItem('vueuse-color-scheme')) {
  localStorage.setItem('vueuse-color-scheme', 'light')
}
const isDark = useDark()
const toggleDark = useToggle(isDark)

const searchQuery = ref('')

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
const navigateTo = (path) => { router.push(path) }

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
```

scoped style 的关键覆盖：

```css
.sidebar { width: 220px; background: var(--surface-bg); border-right: 1px solid var(--border-color); display: flex; flex-direction: column; }
.logo { height: 56px; display: flex; align-items: center; padding: 0 24px; }
.logo-text { font-size: 20px; font-weight: 700; color: var(--accent); letter-spacing: -0.02em; }
.nav-group-header { padding: 0.75rem 1rem 0.25rem; font-size: 11px; font-weight: 600; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.05em; }
.nav-item.active { background: var(--accent-light); color: var(--accent); font-weight: 600; border-left: 3px solid var(--accent); }
.navbar { background: var(--surface-bg); border-bottom: 1px solid var(--border-color); height: 56px; }
.theme-toggle { border: 1px solid var(--border-color) !important; background: transparent !important; color: var(--text-secondary) !important; }
.theme-toggle:hover { border-color: var(--accent) !important; color: var(--accent) !important; }
```

- [ ] **Step 2: 验证开发服务器启动**

Run: `cd /home/zcq/Github/clouddrive-auto-save && timeout 10 make dev-web 2>&1 | tail -5 || true`
Expected: Vite 开发服务器成功启动

- [ ] **Step 3: 提交**

```bash
git add web/src/layout/MainLayout.vue
git commit -m "refactor(ui): 重构 MainLayout 侧边栏和顶栏

侧边栏收窄至220px，Logo改为纯文字UCAS，去除CloudLogo SVG。
去除导航分组折叠功能和emoji图标，选中态改为左侧蓝色竖条。
顶栏高度降至56px，去除毛玻璃效果。图标从lucide换为Phosphor。"
```

---

## Task 6: 删除 CloudLogo 并迁移子组件图标

**Files:**
- Delete: `web/src/components/CloudLogo.vue`
- Modify: `web/src/components/SidebarFooter.vue`
- Modify: `web/src/components/PWAInstall.vue`
- Modify: `web/src/components/cards/TaskCard.vue`
- Modify: `web/src/components/cards/AccountCard.vue`
- Modify: `web/src/components/ShareContentDialog.vue`

- [ ] **Step 1: 删除 CloudLogo.vue**

Run: `rm web/src/components/CloudLogo.vue`

- [ ] **Step 2: 更新 SidebarFooter.vue — lucide → Phosphor，去除 html.dark 硬编码**

在 script 中将：
```js
import { Github, ExternalLink } from 'lucide-vue-next'
```
替换为：
```js
import { GithubLogo, ArrowSquareOut } from '@phosphor-icons/vue'
```

在 template 中将 `<Github :size="16" />` 替换为 `<GithubLogo :size="16" weight="regular" />`，将 `<ExternalLink :size="12" />` 替换为 `<ArrowSquareOut :size="12" weight="regular" />`。

在 style 中删除所有 `html.dark` 选择器块（3 处），让组件通过 CSS 变量自动适配暗色。

- [ ] **Step 3: 更新 PWAInstall.vue — emoji ☁️ → Phosphor 图标**

在 script 中添加：
```js
import { Cloud } from '@phosphor-icons/vue'
```

在 template 中将 `<div class="install-icon">☁️</div>` 替换为：
```html
<div class="install-icon">
  <Cloud :size="40" weight="duotone" />
</div>
```

在 style 中将 `.install-icon { font-size: 2.5rem; }` 替换为 `.install-icon { color: var(--accent); display: flex; align-items: center; }`。

- [ ] **Step 4: 更新 TaskCard.vue — lucide → Phosphor + 新卡片/按钮样式**

在 script 中将：
```js
import { Play, Edit, Trash2 } from 'lucide-vue-next'
```
替换为：
```js
import { Play, PencilSimple, Trash } from '@phosphor-icons/vue'
```

在 template 中：
- `<Play :size="14" />` → `<Play :size="16" weight="fill" />`
- `<Edit :size="14" />` → `<PencilSimple :size="16" weight="regular" />`
- `<Trash2 :size="14" />` → `<Trash :size="16" weight="regular" />`
- `<el-tag>` 去除 `effect="dark"` 属性（使用全局的浅底风格）

在 scoped style 中：
- `.task-card` 背景改为 `var(--surface-bg)`，圆角改为 `var(--radius-lg)` = `14px`，阴影改为 `var(--shadow-sm)`，hover 的 `translateY(-4px)` 改为 `translateY(-1px)`，hover 阴影改为 `var(--shadow-md)`
- `.info-item` 的 `border-bottom: 1px solid var(--border)` 改为 `border-bottom: 1px solid var(--border-color)`
- `<el-progress>` 去除 `striped striped-flow` 属性

- [ ] **Step 5: 更新 AccountCard.vue — 新卡片规范（去除渐变头部）**

在 template 中：
- 去除 `.card-header` 的 `:style="{ background: platformColors[...] }"` 内联渐变背景
- `.card-header` 改为简洁的白底 + 平台名文字样式（无渐变色背景）
- 添加 Phosphor `HardDrives` 图标作为平台图标
- 操作按钮改用文字按钮（Ghost 风格）

在 script 中添加：
```js
import { HardDrives } from '@phosphor-icons/vue'
```

在 scoped style 中：
- `.account-card` 背景改为 `var(--surface-bg)`，圆角 `var(--radius-lg)` = `14px`，阴影 `var(--shadow-sm)`
- `.card-header` 去除 `color: white`，改为 `color: var(--text-primary)` + `padding: 1rem 1.25rem` + `border-bottom: 1px solid var(--border-color)`
- 去除 `platformColors` 渐变定义

- [ ] **Step 6: 更新 ShareContentDialog.vue — Phosphor 图标**

在 script 中将所有 `lucide-vue-next` 导入替换为对应 Phosphor 组件：
- `Folder` → `Folder` from `@phosphor-icons/vue`
- `File` → `File` from `@phosphor-icons/vue`
- `ChevronRight` → `CaretRight` from `@phosphor-icons/vue`
- `ArrowLeft` → `ArrowLeft` from `@phosphor-icons/vue`

图标使用时添加 `weight="regular"` 属性，尺寸保持不变。

- [ ] **Step 7: 验证构建**

Run: `cd /home/zcq/Github/clouddrive-auto-save/web && npx vite build --mode development 2>&1 | tail -3`
Expected: 构建成功，无 import 错误

- [ ] **Step 8: 提交**

```bash
git add -A web/src/components/
git commit -m "refactor(ui): 迁移子组件图标至 Phosphor，删除 CloudLogo

CloudLogo.vue 已删除，被纯文字UCAS替代。
SidebarFooter/TaskCard/AccountCard/PWAInstall/ShareContentDialog
全部从 lucide-vue-next 迁移至 @phosphor-icons/vue。
卡片改为白底+14px圆角+柔和阴影，去除渐变头部。
图标按钮改为圆形Ghost风格。"
```

---

## Task 7: 重构 Dashboard — 三栏改两栏

**Files:**
- Modify: `web/src/views/Dashboard.vue`

这是最大的重构任务。Dashboard 当前是三栏布局（遥测 | 控制台核心 | 终端日志），需改为两栏（主区 70% | 侧边 30%），并去除所有霓虹动画。

- [ ] **Step 1: 重写 Dashboard template 布局**

核心布局变更：

```html
<div class="dashboard-container">
  <el-row :gutter="20">
    <!-- 左栏主区（约 70%） -->
    <el-col :xs="24" :md="17">
      <!-- 4 个统计磁贴 -->
      <el-row :gutter="12" class="stat-mini-grids">
        <el-col v-for="tile in statTiles" :key="tile.key" :span="6">
          <div class="stat-tile">
            <div class="stat-value" :style="{ color: tile.color }">{{ tile.value }}</div>
            <div class="stat-label">{{ tile.label }}</div>
          </div>
        </el-col>
      </el-row>

      <!-- 活跃任务队列 -->
      <el-card class="section-card">
        <template #header>
          <div class="card-header-simple">
            <span class="panel-title">活跃任务</span>
            <el-tag v-if="activeTasks.length" size="small">{{ activeTasks.length }}</el-tag>
          </div>
        </template>
        <!-- 任务列表内容，使用 Phosphor 图标替代 lucide -->
      </el-card>

      <!-- 近期活动时间线 -->
      <el-card class="section-card">
        <template #header>
          <div class="card-header-simple">
            <span class="panel-title">近期活动</span>
          </div>
        </template>
        <el-timeline class="compact-timeline">
          <!-- 时间线内容保持，去除脉冲动画 -->
        </el-timeline>
      </el-card>
    </el-col>

    <!-- 右栏侧边（约 30%） -->
    <el-col :xs="24" :md="7">
      <!-- 系统状态 -->
      <el-card class="section-card">
        <template #header>
          <div class="card-header-simple">
            <span class="panel-title">系统状态</span>
          </div>
        </template>
        <!-- CPU/RAM 线性进度条 + 存储环形图 80px -->
      </el-card>

      <!-- 日志面板（白底卡片，非黑终端） -->
      <el-card class="section-card log-panel">
        <template #header>
          <div class="card-header-simple">
            <span class="panel-title">系统日志</span>
            <el-button size="small" text @click="clearLogs">清空</el-button>
          </div>
        </template>
        <div class="log-list">
          <div v-for="log in recentLogs" :key="log.id" class="log-line" :class="'log-' + log.level">
            <span class="log-time">{{ log.time }}</span>
            <span class="log-content">{{ log.content }}</span>
          </div>
          <div v-if="!recentLogs.length" class="log-empty">等待系统日志流中...</div>
        </div>
      </el-card>
    </el-col>
  </el-row>
</div>
```

**去除的元素：**
- `telemetry-column`（左栏独立遥测面板）→ 合并到右栏
- `console-core-column` 的三栏自适应宽度逻辑
- `.terminal-window` 纯黑终端区域 → 白底日志列表
- `breath-glow` 类名和脉冲动画
- `.pulse-dot` 和 `AUTO-SAVE ACTIVE` 状态框
- `glass-card` 类名
- 条纹流动进度条（`striped striped-flow`）

- [ ] **Step 2: 更新 Dashboard script — lucide → Phosphor**

将所有 lucide 导入替换：
```js
// 旧
import { Calendar, Info, Scan, RefreshCw, Terminal, Trash2, CheckCircle2, AlertCircle, Loader2, X, Bell } from 'lucide-vue-next'
// 新
import { CalendarBlank, Info, ArrowsClockwise, Trash, CheckCircle, WarningCircle, Spinner, X, Bell } from '@phosphor-icons/vue'
```

图标使用时统一添加 `weight` 属性：
- 状态图标（成功/失败）：`weight="fill"`
- 操作图标：`weight="regular"`
- Spinner：`<Spinner :size="16" class="spin-icon" />`

- [ ] **Step 3: 重写 Dashboard scoped style**

关键样式变更：
- 去除所有 `.glass-card`、`.breath-glow`、`.neon-border` 类
- `.stat-tile` 改为白卡片 + `var(--radius-md)` + `var(--shadow-sm)`，无霓虹发光
- `.task-progress-card` 去除 `breath-glow` 动画，改为简洁的 `border: 1px solid var(--border-color)`
- `.terminal-window` 改为 `.log-panel .log-list`：白底、`font-family: var(--font-mono)`、最大高度 400px、overflow-y auto
- `.log-line` 样式：
  - 基础：`padding: 6px 12px; border-radius: var(--radius-sm); font-size: 13px; line-height: 1.5;`
  - `.log-error { background: #FEF2F2; color: #991B1B; }` （极淡红底）
  - `.log-warn { background: #FFFBEB; color: #92400E; }` （极淡黄底）
  - `.log-success { background: #F0FDF4; color: #166534; }` （极淡绿底）
- 去除 `@keyframes neon-pulse`、`@keyframes blink`
- 进度条去除 `striped-flow`，颜色从 `var(--neon-*)` 改为语义色变量
- 去除 `#04060b` 硬编码背景

- [ ] **Step 4: 验证构建**

Run: `cd /home/zcq/Github/clouddrive-auto-save/web && npx vite build --mode development 2>&1 | tail -3`

- [ ] **Step 5: 提交**

```bash
git add web/src/views/Dashboard.vue
git commit -m "refactor(ui): Dashboard 三栏改两栏，Apple/Linear 风格

去除纯黑终端面板，改为白底日志列表按级别分色。
去除遥测独立面板，合并至右栏系统状态卡片。
去除所有霓虹脉冲/呼吸动画和glass-card毛玻璃效果。
图标从lucide迁移至Phosphor。"
```

---

## Task 8: 重构 Tasks 页面

**Files:**
- Modify: `web/src/views/Tasks.vue`

- [ ] **Step 1: 更新 Tasks script — lucide → Phosphor**

将所有 lucide 导入替换为 Phosphor 等价物：
```js
import {
  Plus, Play, PencilSimple, Trash, ArrowsClockwise, Folder, File as FileIcon,
  Info, Cloud, ArrowSquareOut, Warning, Clock, FolderOpen, List, GridFour,
  MagnifyingGlass
} from '@phosphor-icons/vue'
```

- [ ] **Step 2: 更新 Tasks template**

关键变更：
- 筛选栏 `<el-radio-group>` 去除默认边框样式，改为文字按钮组（通过 CSS 实现下划线选中态）
- 表格操作列的 `.btn-icon` 按钮图标尺寸从 `:size="14"` 改为 `:size="16"`，添加 `weight="regular"`
- `<el-progress>` 去除 `striped striped-flow`
- `<el-drawer>` 宽度从 600px 改为 560px
- 所有 `<el-tag>` 去除 `effect="dark"` 和 `effect="plain"` 属性

- [ ] **Step 3: 更新 Tasks scoped style**

- `.page-header` 标题从 `26px/800` 改为 `24px/700`
- `.task-filter-bar .el-radio-button__inner` 改为文字按钮风格（无背景无边框，选中态下划线 `border-bottom: 2px solid var(--accent)` + `color: var(--accent)`）
- `.table-card` 去除霓虹样式，使用全局卡片规范
- `.btn-icon` 样式已在 style.css 中全局定义，scoped 中去除局部覆盖
- 去除所有 `html.dark` 硬编码选择器
- 去除 `.glass-card` 类名引用

- [ ] **Step 4: 验证构建并提交**

```bash
git add web/src/views/Tasks.vue
git commit -m "refactor(ui): Tasks 页面 Apple/Linear 风格重构

筛选栏改为文字按钮组+下划线选中态。去除霓虹发光和毛玻璃。
图标迁移至Phosphor，操作按钮改为圆形Ghost风格。
表格/卡片/抽屉统一应用新设计规范。"
```

---

## Task 9: 重构 Accounts 页面

**Files:**
- Modify: `web/src/views/Accounts.vue`

- [ ] **Step 1: 更新 Accounts script — lucide → Phosphor**

```js
import { Plus, ArrowsCounterClockwise, Trash, PencilSimple, HardDrives, Info, GridFour, List, ArrowSquareOut } from '@phosphor-icons/vue'
```

注意 `RefreshCcw` → `ArrowsCounterClockwise`，`Edit` → `PencilSimple`。

- [ ] **Step 2: 更新 Accounts template**

- 平台图标从 lucide `HardDrive` 改为 Phosphor `HardDrives`
- 表格中 `.platform-icon` 的彩色背景圆角块改为简洁的文字 + 小标签
- 卡片视图的渐变顶部条去除
- `<el-progress>` 去除 `striped` 相关属性

- [ ] **Step 3: 更新 Accounts scoped style**

- `.account-card` 去除 `transform: translateY(-5px) scale(1.02)` 夸张动效，改为 `translateY(-1px)`
- 去除 `--neon-glow-teal` 阴影
- `.platform-icon` 从彩色方块改为 Ghost 图标样式
- 去除所有 `html.dark` 硬编码选择器
- 添加账号弹窗宽度从 520px 改为 480px
- `<el-alert>` 样式适配新配色

- [ ] **Step 4: 验证构建并提交**

```bash
git add web/src/views/Accounts.vue
git commit -m "refactor(ui): Accounts 页面 Apple/Linear 风格重构

平台图标改为Phosphor HardDrives，去除彩色背景方块。
卡片去除渐变顶部条和夸张hover动效，应用新卡片规范。
图标迁移至Phosphor，去除html.dark硬编码。"
```

---

## Task 10: 重构 Settings 页面

**Files:**
- Modify: `web/src/views/Settings.vue`

- [ ] **Step 1: 更新 Settings script — lucide → Phosphor**

```js
import { CalendarBlank, Info, ArrowsClockwise, Bell, PuzzlePiece, Plus, MagnifyingGlass } from '@phosphor-icons/vue'
```

- [ ] **Step 2: 更新 Settings template**

- `<el-tabs type="border-card">` 改为 `<el-tabs>` （去除 border-card 样式，由 CSS 实现下划线风格）
- 嵌套 Tabs 同样去除 `type="border-card"`
- 插件图标改为 Phosphor `PuzzlePiece` duotone
- 所有 `<el-tag>` 统一去除 `effect` 属性

- [ ] **Step 3: 更新 Settings scoped style**

- `.settings-tabs` / `.el-tabs` 覆盖为下划线 Tab 风格：
  ```css
  .settings-tabs :deep(.el-tabs__header) { border-bottom: 1px solid var(--border-color); }
  .settings-tabs :deep(.el-tabs__item) { color: var(--text-secondary); font-weight: 500; border: none; }
  .settings-tabs :deep(.el-tabs__item.is-active) { color: var(--accent); border-bottom: 2px solid var(--accent); }
  ```
- `.plugin-card` 去除 hover 发光，改为阴影加深
- `.add-card` 虚线边框颜色改为 `var(--border-color)`
- 去除 `.glass-card` 类名和 `html.dark` 硬编码

- [ ] **Step 4: 验证构建并提交**

```bash
git add web/src/views/Settings.vue
git commit -m "refactor(ui): Settings 页面 Apple/Linear 风格重构

Tab栏从border-card改为底部下划线风格。插件卡片去除hover发光。
图标迁移至Phosphor PuzzlePiece duotone。去除毛玻璃和霓虹效果。"
```

---

## Task 11: 重构 Search 页面

**Files:**
- Modify: `web/src/views/Search.vue`

- [ ] **Step 1: 更新 Search script — lucide → Phosphor**

```js
import { MagnifyingGlass, Link, Clock, FileText } from '@phosphor-icons/vue'
```

- [ ] **Step 2: 更新 Search template**

- 验证状态 emoji 替换：
  - `✅` → `<CheckCircle :size="16" weight="fill" color="var(--color-success)" />`
  - `❌` → `<XCircle :size="16" weight="fill" color="var(--color-danger)" />`
  - `⏳` → `<Spinner :size="16" weight="regular" class="spin-icon" />`
- 需要额外导入 `CheckCircle, XCircle, Spinner` from `@phosphor-icons/vue`
- 搜索结果元数据图标更新为 Phosphor 版本
- "创建任务" 按钮保持 Primary 风格（已由全局覆盖处理）

- [ ] **Step 3: 更新 Search scoped style**

- `.result-item` 改为白底 + `var(--radius-lg)` + `var(--shadow-sm)`
- hover 改为 `var(--shadow-md)` + `translateY(-1px)`
- 验证进度的 emoji 动画改为 CSS `spin` 动画（用于 Spinner 图标）：
  ```css
  .spin-icon { animation: spin 1s linear infinite; }
  @keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
  ```
- 去除 `@keyframes pulse` 的 emoji 脉冲动画

- [ ] **Step 4: 验证构建并提交**

```bash
git add web/src/views/Search.vue
git commit -m "refactor(ui): Search 页面 Apple/Linear 风格重构

验证状态emoji替换为Phosphor图标(CheckCircle/XCircle/Spinner)。
结果卡片改为白底+14px圆角+柔和阴影。图标迁移至Phosphor。"
```

---

## Task 12: 卸载 lucide-vue-next 并清理残留

**Files:**
- Modify: `web/package.json`

- [ ] **Step 1: 全局搜索 lucide 残留引用**

Run: `cd /home/zcq/Github/clouddrive-auto-save/web && grep -rn "lucide-vue-next" src/ || echo "无残留"`

Expected: 无输出（所有引用已在前面 Tasks 中替换）

- [ ] **Step 2: 全局搜索 emoji 残留**

Run: `cd /home/zcq/Github/clouddrive-auto-save/web && grep -rn "📊\|🔧\|🛠️\|⚙️\|☁️\|✅\|❌\|⏳\|🧩" src/ || echo "无残留"`

Expected: 无输出

- [ ] **Step 3: 全局搜索霓虹残留变量/类名**

Run: `cd /home/zcq/Github/clouddrive-auto-save/web && grep -rn "neon-\|glass-card\|breath-glow\|neon-pulse\|neon-glow\|backdrop-filter" src/ || echo "无残留"`

Expected: 无输出

- [ ] **Step 4: 卸载 lucide-vue-next**

Run: `cd /home/zcq/Github/clouddrive-auto-save/web && npm uninstall lucide-vue-next`

- [ ] **Step 5: 验证构建**

Run: `cd /home/zcq/Github/clouddrive-auto-save/web && npx vite build 2>&1 | tail -5`
Expected: 构建成功，无 import 错误

- [ ] **Step 6: 提交**

```bash
git add -A web/
git commit -m "chore(ui): 卸载 lucide-vue-next，清理所有残留引用

确认所有组件已迁移至 @phosphor-icons/vue。
去除所有 emoji 图标残留和霓虹变量残留。"
```

---

## Task 13: 最终验证与 E2E 适配

**Files:**
- 可能需要修改: `e2e/tests/` 下的 Playwright 测试文件（如 CSS 选择器/类名变更导致测试失败）

- [ ] **Step 1: 运行前端构建**

Run: `cd /home/zcq/Github/clouddrive-auto-save && make build-web`
Expected: 构建成功

- [ ] **Step 2: 运行 Go 检查**

Run: `cd /home/zcq/Github/clouddrive-auto-save && make check`
Expected: 全部通过

- [ ] **Step 3: 启动开发服务器进行手动视觉检查**

Run: `cd /home/zcq/Github/clouddrive-auto-save && make dev`

在浏览器中检查：
1. 亮色模式：所有页面卡片/按钮/表格/标签/弹窗视觉正确
2. 暗色模式：基础可用性（不花哨但可读）
3. 响应式：桌面（1440px）、笔记本（1280px）
4. 无 lucide 图标残留（应显示 Phosphor 图标）
5. 无霓虹发光/毛玻璃效果残留

- [ ] **Step 4: 运行 E2E 测试（如有失败则修复选择器）**

Run: `cd /home/zcq/Github/clouddrive-auto-save && make e2e-test`

如果测试失败，检查失败原因是否为 CSS 选择器/类名/DOM 结构变更导致。修复策略：
- `.glass-card` 类名已被去除 → 更新 E2E 中使用该类名的选择器
- `effect="dark"` 已被去除 → 更新 Tag 相关断言
- 图标组件变化 → 更新图标相关定位器

- [ ] **Step 5: 最终提交**

```bash
git add -A
git commit -m "fix(e2e): 适配 UI 重构后的 E2E 测试选择器

更新因 DOM 结构和 CSS 类名变更导致的失败测试用例。"
```

---

## 自检清单

- [x] **Spec 覆盖**：设计文档中所有 Design Token（§2）、组件规范（§3）、页面设计（§4）、图标迁移（§5）均有对应 Task
- [x] **无占位符**：所有步骤包含具体代码、命令或明确指令，无 TBD/TODO
- [x] **类型一致性**：所有 Task 中使用的 Phosphor 组件名与 §5.2 图标映射表一致；CSS 变量名与 Task 1 定义的 Token 一致
