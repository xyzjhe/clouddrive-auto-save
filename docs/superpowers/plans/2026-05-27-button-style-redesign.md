# 按钮样式重新设计实施计划

> **致智能体工作者：** 必须使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 技能来逐任务实施此计划。步骤使用复选框（`- [ ]`）语法进行跟踪。

**目标：** 重构整个项目的按钮样式系统，从 Element Plus 默认样式 + 霓虹渐变覆盖迁移到现代极简 + 微妙科技感的统一设计语言

**架构：** 在 `variables.css` 中定义按钮专用 CSS 变量，在 `style.css` 中移除旧的按钮覆盖样式并添加新的全局样式，然后逐个更新 Vue 组件中的按钮实现，将表格操作列的 `el-button-group` + `link` 模式替换为自定义图标按钮组

**技术栈：** Vue 3 + Element Plus + CSS 变量 + lucide-vue-next 图标

---

## 文件结构

### 核心样式文件
- `web/src/variables.css` — 添加按钮专用 CSS 变量（亮/暗主题）
- `web/src/style.css` — 移除旧按钮覆盖，添加新全局按钮样式

### 组件文件（按优先级排序）
- `web/src/views/Tasks.vue` — 任务页面，按钮最密集，包含表格操作列、页面顶部、抽屉底部
- `web/src/views/Accounts.vue` — 账号页面，表格操作列和对话框底部
- `web/src/views/Dashboard.vue` — 仪表盘，底部快捷操作栏
- `web/src/views/Settings.vue` — 设置页，预设时间按钮组和通知渠道按钮
- `web/src/components/cards/TaskCard.vue` — 任务卡片，操作按钮
- `web/src/layout/MainLayout.vue` — 主布局，主题切换按钮

---

## Task 1: 添加按钮 CSS 变量到 variables.css

**Files:**
- Modify: `web/src/variables.css:1-62`（明亮模式部分）
- Modify: `web/src/variables.css:64-121`（暗黑模式部分）

- [ ] **Step 1: 在明亮模式 `:root` 块末尾添加按钮变量**

在 `variables.css` 的 `:root` 块最后（`--font-mono` 行之后）添加：

```css
  /* 按钮设计系统 */
  --btn-radius: 6px;
  --btn-font-weight: 500;
  --btn-transition: all 0.2s ease;

  /* 主要按钮 */
  --btn-primary-bg: #3b82f6;
  --btn-primary-text: #ffffff;
  --btn-primary-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);
  --btn-primary-hover-shadow: 0 4px 8px rgba(59, 130, 246, 0.4);

  /* 次要按钮 */
  --btn-secondary-bg: #ffffff;
  --btn-secondary-text: #3b82f6;
  --btn-secondary-border: #e5e7eb;
  --btn-secondary-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);

  /* 幽灵按钮 */
  --btn-ghost-bg: transparent;
  --btn-ghost-text: #6b7280;
  --btn-ghost-border: #d1d5db;

  /* 图标按钮 */
  --btn-icon-size: 32px;
  --btn-icon-radius: 6px;

  /* 语义色按钮 */
  --btn-success: #10b981;
  --btn-success-bg: rgba(16, 185, 129, 0.1);
  --btn-success-shadow: 0 1px 2px rgba(16, 185, 129, 0.2);

  --btn-danger: #ef4444;
  --btn-danger-bg: rgba(239, 68, 68, 0.1);
  --btn-danger-shadow: 0 1px 2px rgba(239, 68, 68, 0.2);
```

- [ ] **Step 2: 在暗黑模式 `html.dark` 块末尾添加按钮变量覆盖**

在 `variables.css` 的 `html.dark` 块最后（`--neutral-700` 行之后）添加：

```css
  /* 按钮设计系统 - 暗黑模式 */
  --btn-secondary-bg: #1f2937;
  --btn-secondary-border: #374151;
  --btn-ghost-border: #4b5563;
```

- [ ] **Step 3: 验证变量定义正确**

运行开发服务器确认无 CSS 语法错误：

```bash
cd /home/zcq/Github/clouddrive-auto-save && make dev-web
```

Expected: 开发服务器启动成功，无 CSS 错误

- [ ] **Step 4: 提交**

```bash
git add web/src/variables.css
git commit -m "feat(style): 添加按钮设计系统 CSS 变量"
```

---

## Task 2: 更新 style.css 全局按钮样式

**Files:**
- Modify: `web/src/style.css:82-99`

- [ ] **Step 1: 移除旧的按钮覆盖样式**

删除 `style.css` 中第 82-99 行的旧按钮样式：

```css
/* 删除以下内容 */
.el-button {
  border-radius: 8px !important;
  font-weight: 500;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1) !important;
}

.el-button--primary {
  background: linear-gradient(135deg, var(--neon-teal) 0%, var(--neon-blue) 100%) !important;
  border: none !important;
  color: var(--primary-btn-text) !important;
  font-weight: 600 !important;
  box-shadow: 0 0 12px rgba(0, 242, 254, 0.2) !important;
}

.el-button--primary:hover {
  box-shadow: 0 0 20px rgba(0, 242, 254, 0.4) !important;
  transform: translateY(-1px);
}
```

- [ ] **Step 2: 添加新的全局按钮样式**

在原位置（第 82 行开始）添加新的按钮样式：

```css
/* ==========================================
 * 按钮设计系统 - 现代极简 + 微妙科技感
 * ========================================== */

/* 基础按钮覆盖 */
.el-button {
  border-radius: var(--btn-radius) !important;
  font-weight: var(--btn-font-weight) !important;
  transition: var(--btn-transition) !important;
}

/* 主要按钮 */
.el-button--primary {
  background: var(--btn-primary-bg) !important;
  color: var(--btn-primary-text) !important;
  border: none !important;
  box-shadow: var(--btn-primary-shadow) !important;
}

.el-button--primary:hover {
  box-shadow: var(--btn-primary-hover-shadow) !important;
  transform: translateY(-1px);
}

.el-button--primary:active {
  transform: translateY(0);
  box-shadow: var(--btn-primary-shadow) !important;
}

/* 次要按钮（plain 模式） */
.el-button--primary.is-plain {
  background: var(--btn-secondary-bg) !important;
  color: var(--btn-secondary-text) !important;
  border: 1px solid var(--btn-secondary-border) !important;
  box-shadow: var(--btn-secondary-shadow) !important;
}

.el-button--primary.is-plain:hover {
  border-color: var(--btn-primary-bg) !important;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1) !important;
}

/* 信息按钮（幽灵风格） */
.el-button--info.is-plain {
  background: var(--btn-ghost-bg) !important;
  color: var(--btn-ghost-text) !important;
  border: 1px solid var(--btn-ghost-border) !important;
}

.el-button--info.is-plain:hover {
  border-color: var(--btn-primary-bg) !important;
  color: var(--btn-primary-bg) !important;
}

/* 成功按钮 */
.el-button--success.is-plain {
  color: var(--btn-success) !important;
  border-color: var(--btn-success) !important;
}

/* 危险按钮 */
.el-button--danger.is-plain {
  color: var(--btn-danger) !important;
  border-color: var(--btn-danger) !important;
}

/* 链接按钮样式优化 */
.el-button.is-link {
  padding: 4px 8px !important;
  height: auto !important;
}

.el-button.is-link:hover {
  opacity: 0.8;
}

/* 自定义图标按钮 */
.action-buttons {
  display: flex;
  gap: 8px;
  align-items: center;
}

.btn-icon {
  width: var(--btn-icon-size);
  height: var(--btn-icon-size);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--btn-icon-radius);
  cursor: pointer;
  transition: var(--btn-transition);
  border: none;
  font-size: 14px;
  padding: 0;
}

.btn-icon--success {
  color: var(--btn-success);
  background: var(--btn-success-bg);
  box-shadow: var(--btn-success-shadow);
}

.btn-icon--success:hover {
  box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);
}

.btn-icon--primary {
  color: var(--neon-blue);
  background: rgba(59, 130, 246, 0.1);
  box-shadow: 0 1px 2px rgba(59, 130, 246, 0.2);
}

.btn-icon--primary:hover {
  box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);
}

.btn-icon--danger {
  color: var(--btn-danger);
  background: var(--btn-danger-bg);
  box-shadow: var(--btn-danger-shadow);
}

.btn-icon--danger:hover {
  box-shadow: 0 2px 4px rgba(239, 68, 68, 0.3);
}

.btn-icon:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none !important;
}

.btn-icon:disabled:hover {
  box-shadow: inherit;
}
```

- [ ] **Step 3: 验证样式生效**

运行开发服务器并访问页面：

```bash
make dev-web
```

Expected: 页面按钮样式更新，主要按钮显示蓝色实心背景 + 阴影

- [ ] **Step 4: 提交**

```bash
git add web/src/style.css
git commit -m "feat(style): 替换按钮全局样式为现代极简设计"
```

---

## Task 3: 更新 Tasks.vue 表格操作列

**Files:**
- Modify: `web/src/views/Tasks.vue:86-100`

- [ ] **Step 1: 替换表格操作列按钮**

将 Tasks.vue 第 86-100 行的操作列模板：

```html
<el-table-column label="操作" width="220" fixed="right">
  <template #default="{ row }">
    <el-button-group>
      <el-button 
        link 
        type="success" 
        :icon="Play" 
        :disabled="row.status === 'running' || !!(row.message && row.message.includes('[Fatal]'))" 
        @click="handleRun(row)"
      >
        运行
      </el-button>
      <el-button link type="primary" :icon="Edit" @click="handleEdit(row)">编辑</el-button>
      <el-button link type="danger" :icon="Trash2" @click="handleDelete(row)">删除</el-button>
    </el-button-group>
```

替换为：

```html
<el-table-column label="操作" width="140" fixed="right">
  <template #default="{ row }">
    <div class="action-buttons">
      <button
        class="btn-icon btn-icon--success"
        title="运行"
        :disabled="row.status === 'running' || !!(row.message && row.message.includes('[Fatal]'))"
        @click="handleRun(row)"
      >
        <Play :size="14" />
      </button>
      <button
        class="btn-icon btn-icon--primary"
        title="编辑"
        @click="handleEdit(row)"
      >
        <Edit :size="14" />
      </button>
      <button
        class="btn-icon btn-icon--danger"
        title="删除"
        @click="handleDelete(row)"
      >
        <Trash2 :size="14" />
      </button>
    </div>
```

- [ ] **Step 2: 验证表格操作按钮**

运行开发服务器，访问任务页面：

```bash
make dev-web
```

Expected: 表格操作列显示为图标按钮，hover 时有阴影效果

- [ ] **Step 3: 提交**

```bash
git add web/src/views/Tasks.vue
git commit -m "feat(tasks): 表格操作列改为图标按钮组"
```

---

## Task 4: 更新 Tasks.vue 页面顶部按钮

**Files:**
- Modify: `web/src/views/Tasks.vue:17-29`

- [ ] **Step 1: 更新页面顶部按钮样式**

将 Tasks.vue 第 25 行的"全部运行"按钮：

```html
<el-button :icon="Play" :loading="runningAll">全部运行</el-button>
```

替换为：

```html
<el-button type="primary" plain :icon="Play" :loading="runningAll">全部运行</el-button>
```

第 28 行的"创建任务"按钮保持不变（已经是 `type="primary"`）

- [ ] **Step 2: 验证顶部按钮**

Expected: "全部运行"按钮显示为次要按钮样式（白色背景 + 蓝色边框），"创建任务"按钮显示为主要按钮样式（蓝色背景 + 阴影）

- [ ] **Step 3: 提交**

```bash
git add web/src/views/Tasks.vue
git commit -m "feat(tasks): 更新页面顶部按钮样式"
```

---

## Task 5: 更新 Tasks.vue 抽屉底部按钮

**Files:**
- Modify: `web/src/views/Tasks.vue:321-322`（需要先定位确切行号）

- [ ] **Step 1: 定位抽屉底部按钮**

在 Tasks.vue 中搜索"确认并保存"或"取消"按钮的位置。

- [ ] **Step 2: 更新抽屉底部按钮样式**

确保"取消"按钮使用幽灵样式，"确认并保存"按钮使用主要样式：

```html
<el-button @click="dialogVisible = false">取消</el-button>
<el-button type="primary" :loading="submitting" @click="submitForm">确认并保存</el-button>
```

这两个按钮已经使用了正确的类型，只需要确认全局样式生效。

- [ ] **Step 3: 验证抽屉按钮**

Expected: "取消"按钮显示为幽灵样式，"确认并保存"显示为主要按钮样式

- [ ] **Step 4: 提交**

```bash
git add web/src/views/Tasks.vue
git commit -m "feat(tasks): 确认抽屉底部按钮样式"
```

---

## Task 6: 更新 Accounts.vue 按钮

**Files:**
- Modify: `web/src/views/Accounts.vue:80-84`（表格操作列）
- Modify: `web/src/views/Accounts.vue:229-230`（对话框底部）

- [ ] **Step 1: 替换 Accounts.vue 表格操作列**

将 Accounts.vue 第 80-84 行：

```html
<el-button-group>
  <el-button link type="primary" :icon="RefreshCcw" @click="handleCheck(row)">校验</el-button>
  <el-button link type="primary" :icon="Edit" @click="handleEdit(row)">编辑</el-button>
  <el-button link type="danger" :icon="Trash2" @click="handleDelete(row)">删除</el-button>
</el-button-group>
```

替换为：

```html
<div class="action-buttons">
  <button
    class="btn-icon btn-icon--primary"
    title="校验"
    @click="handleCheck(row)"
  >
    <RefreshCcw :size="14" />
  </button>
  <button
    class="btn-icon btn-icon--primary"
    title="编辑"
    @click="handleEdit(row)"
  >
    <Edit :size="14" />
  </button>
  <button
    class="btn-icon btn-icon--danger"
    title="删除"
    @click="handleDelete(row)"
  >
    <Trash2 :size="14" />
  </button>
</div>
```

- [ ] **Step 2: 验证 Accounts 页面**

Expected: 账号表格操作列显示为图标按钮

- [ ] **Step 3: 提交**

```bash
git add web/src/views/Accounts.vue
git commit -m "feat(accounts): 表格操作列改为图标按钮组"
```

---

## Task 7: 更新 Dashboard.vue 按钮

**Files:**
- Modify: `web/src/views/Dashboard.vue:171-173`（底部快捷操作栏）

- [ ] **Step 1: 更新 Dashboard 底部按钮**

将 Dashboard.vue 第 171-173 行：

```html
<el-button type="primary" size="default" @click="$router.push('/tasks')">创建新任务</el-button>
<el-button type="primary" plain size="default" @click="$router.push('/accounts')">管理账号</el-button>
<el-button type="info" plain size="default" @click="clearLogs">清理结束任务</el-button>
```

这些按钮已经使用了正确的类型，全局样式会自动生效。无需修改。

- [ ] **Step 2: 验证 Dashboard 按钮**

Expected: "创建新任务"显示为主要按钮，"管理账号"显示为次要按钮，"清理结束任务"显示为幽灵按钮

- [ ] **Step 3: 提交（如果无需修改则跳过）**

```bash
git add web/src/views/Dashboard.vue
git commit -m "feat(dashboard): 确认按钮样式正确应用"
```

---

## Task 8: 更新 Settings.vue 按钮

**Files:**
- Modify: `web/src/views/Settings.vue:62-66`（预设时间按钮组）
- Modify: `web/src/views/Settings.vue:177-178`（通知渠道按钮）

- [ ] **Step 1: 验证 Settings 页面按钮**

Settings 页面的按钮已经使用了正确的类型（`type="primary"` 和默认类型），全局样式会自动生效。需要确认：

1. 预设时间按钮组（凌晨、早晨、中午）使用默认样式
2. 通知渠道的"保存"按钮使用 `type="primary"`，"测试"按钮使用默认样式

- [ ] **Step 2: 验证 Settings 按钮**

Expected: 按钮样式符合设计规范

- [ ] **Step 3: 提交（如果无需修改则跳过）**

```bash
git add web/src/views/Settings.vue
git commit -m "feat(settings): 确认按钮样式正确应用"
```

---

## Task 9: 更新 TaskCard.vue 按钮

**Files:**
- Modify: `web/src/components/cards/TaskCard.vue:70-88`

- [ ] **Step 1: 更新 TaskCard 操作按钮**

将 TaskCard.vue 中的操作按钮：

```html
<el-button size="small" type="primary" :disabled="task.status==='running'" @click="emit('run', task.id)">执行</el-button>
<el-button size="small" @click="emit('edit', task.id)">编辑</el-button>
<el-button size="small" type="danger" @click="emit('delete', task.id)">删除</el-button>
```

替换为图标按钮组：

```html
<div class="action-buttons">
  <button
    class="btn-icon btn-icon--success"
    title="执行"
    :disabled="task.status==='running'"
    @click="emit('run', task.id)"
  >
    <Play :size="14" />
  </button>
  <button
    class="btn-icon btn-icon--primary"
    title="编辑"
    @click="emit('edit', task.id)"
  >
    <Edit :size="14" />
  </button>
  <button
    class="btn-icon btn-icon--danger"
    title="删除"
    @click="emit('delete', task.id)"
  >
    <Trash2 :size="14" />
  </button>
</div>
```

- [ ] **Step 2: 验证 TaskCard 按钮**

Expected: 卡片操作按钮显示为图标按钮组

- [ ] **Step 3: 提交**

```bash
git add web/src/components/cards/TaskCard.vue
git commit -m "feat(task-card): 操作按钮改为图标按钮组"
```

---

## Task 10: 整体验证和收尾

**Files:**
- 无文件修改

- [ ] **Step 1: 运行完整构建**

```bash
cd /home/zcq/Github/clouddrive-auto-save && make build-web
```

Expected: 构建成功，无错误

- [ ] **Step 2: 运行 E2E 测试**

```bash
make e2e-test
```

Expected: 所有 E2E 测试通过

- [ ] **Step 3: 视觉回归检查**

启动开发服务器，手动检查以下页面：

1. Dashboard 页面 - 底部按钮样式
2. Tasks 页面 - 表格操作图标按钮、页面顶部按钮、抽屉按钮
3. Accounts 页面 - 表格操作图标按钮
4. Settings 页面 - 预设时间按钮、通知渠道按钮

检查亮/暗主题切换是否正常。

- [ ] **Step 4: 最终提交**

```bash
git add -A
git commit -m "feat(style): 按钮样式系统全面重构完成"
```

---

## 设计决策参考

### 为什么选择 6px 圆角？
- 比 Element Plus 默认的 4px 更现代
- 比 8px 或 12px 更紧凑、专业
- 与整体设计风格协调

### 为什么使用阴影而非渐变？
- 阴影效果更精致、克制
- 渐变容易显得花哨
- 阴影在亮/暗主题下表现一致

### 为什么表格操作使用图标按钮？
- 节省表格空间
- 视觉更简洁
- 通过 title 属性保持可访问性

---

**计划版本:** v1.0
**创建日期:** 2026-05-27
**预计工时:** 30-45 分钟
