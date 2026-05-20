# 分享链接子目录重置为根目录 — 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在任务编辑中增加子目录状态提示和一键重置为根目录的功能

**Architecture:** 纯前端改动，仅修改 `web/src/views/Tasks.vue`。新增 `selectedDirName` ref 记录目录名、`isSubDirMode` computed 判断子目录模式、`resetToShareRoot()` 统一重置函数。模板中增加提示条和对话框重置按钮。

**Tech Stack:** Vue 3 Composition API, Element Plus (el-tag, el-button)

---

### Task 1: 新增状态变量和 computed 属性

**Files:**
- Modify: `web/src/views/Tasks.vue:531-543`（子目录浏览相关变量区域）

- [ ] **Step 1: 新增 `selectedDirName` ref**

在 `isInitialDir` ref 之后（第 535 行后）添加：

```js
const selectedDirName = ref('') // 记录用户选择的子目录名称，用于提示条显示
```

- [ ] **Step 2: 新增 `isSubDirMode` computed 属性**

在 `currentDirName` computed 之后（第 543 行后）添加：

```js
// 判断当前是否处于子目录模式
const isSubDirMode = computed(() => {
  const account = accounts.value.find(acc => acc.id === form.value.account_id)
  if (!account) return false

  if (account.platform === 'quark') {
    // Quark：从 URL 中解析 pdirFID，非 '0' 且非空则为子目录模式
    const match = form.value.share_url.match(/\/s\/(\w+)#\/list\/share\/(\w+)/)
    return match && match[2] && match[2] !== '0'
  } else {
    // 139：share_parent_id 非空则为子目录模式
    return !!form.value.share_parent_id
  }
})
```

- [ ] **Step 3: 提交**

```bash
git add web/src/views/Tasks.vue
git commit -m "feat(task): 新增子目录模式状态变量和 computed 属性"
```

---

### Task 2: 新增 `resetToShareRoot()` 重置函数

**Files:**
- Modify: `web/src/views/Tasks.vue:845`（`confirmSelectShareUrl` 函数之后）

- [ ] **Step 1: 添加重置函数**

在 `confirmSelectShareUrl` 函数的结束花括号之后（第 845 行后）添加：

```js
// 重置为根目录
const resetToShareRoot = () => {
  const account = accounts.value.find(acc => acc.id === form.value.account_id)
  if (!account) return

  if (account.platform === 'quark') {
    // 从 URL 中提取 pwdID，重建根目录 URL
    const match = form.value.share_url.match(/\/s\/([^#]+)/)
    if (match) {
      form.value.share_url = `https://pan.quark.cn/s/${match[1]}#/list/share/0`
    }
  } else {
    // 139 平台：清空 share_parent_id
    form.value.share_parent_id = ''
  }
  // 清空起始文件选择
  form.value.start_file_id = ''
  form.value.start_file_name = ''
  selectedStartFileName.value = ''
  // 清空目录名称记录
  selectedDirName.value = ''
  ElMessage.success('已重置为根目录')
}
```

- [ ] **Step 2: 提交**

```bash
git add web/src/views/Tasks.vue
git commit -m "feat(task): 新增 resetToShareRoot 重置函数"
```

---

### Task 3: 更新 `confirmSelectShareUrl()` 记录目录名称

**Files:**
- Modify: `web/src/views/Tasks.vue:843`（`confirmSelectShareUrl` 函数内）

- [ ] **Step 1: 在确认选择时记录目录名称**

在 `confirmSelectShareUrl` 函数中，`ElMessage.success` 之前（第 843 行前）添加：

```js
  // 记录选中的目录名称
  selectedDirName.value = currentDirName.value
```

完整上下文（第 837-844 行区域）应变为：

```js
  form.value.share_url = newUrl
  // 清空起始文件选择
  form.value.start_file_id = ''
  form.value.start_file_name = ''
  selectedStartFileName.value = ''
  // 记录选中的目录名称
  selectedDirName.value = currentDirName.value

  ElMessage.success(`已选择目录：${currentDirName.value}`)
```

- [ ] **Step 2: 提交**

```bash
git add web/src/views/Tasks.vue
git commit -m "feat(task): 选择子目录时记录目录名称"
```

---

### Task 4: 修复 `handleUrlChange()` 未清除 `share_parent_id` 的 bug

**Files:**
- Modify: `web/src/views/Tasks.vue:613-617`（`handleUrlChange` 函数）

- [ ] **Step 1: 增加清除逻辑**

将 `handleUrlChange` 函数从：

```js
const handleUrlChange = () => {
  form.value.start_file_id = ''
  form.value.start_file_name = ''
  selectedStartFileName.value = ''
}
```

改为：

```js
const handleUrlChange = () => {
  form.value.start_file_id = ''
  form.value.start_file_name = ''
  selectedStartFileName.value = ''
  form.value.share_parent_id = ''
  selectedDirName.value = ''
}
```

- [ ] **Step 2: 提交**

```bash
git add web/src/views/Tasks.vue
git commit -m "fix(task): URL 变更时自动清除 share_parent_id 防止子目录残留"
```

---

### Task 5: 在模板中添加子目录提示条

**Files:**
- Modify: `web/src/views/Tasks.vue:163-164`（分享链接 `el-form-item` 结束标签之后）

- [ ] **Step 1: 添加提示条模板**

在 `</el-form-item>`（分享链接的结束标签，第 164 行）之后添加：

```html
        <div v-if="isSubDirMode" class="subdir-hint">
          <el-tag type="warning" effect="light" closable @close="resetToShareRoot">
            <el-icon style="margin-right: 4px; vertical-align: middle;"><FolderOpen /></el-icon>
            当前目录：{{ selectedDirName || '子目录' }}
          </el-tag>
        </div>
```

- [ ] **Step 2: 添加提示条样式**

在 `<style scoped>` 区域的 `.share-url-actions` 样式之后（第 1580 行后）添加：

```css
.subdir-hint {
  margin-top: 6px;
}

.subdir-hint .el-tag {
  cursor: default;
}
```

- [ ] **Step 3: 提交**

```bash
git add web/src/views/Tasks.vue
git commit -m "feat(task): 分享链接下方增加子目录状态提示条"
```

---

### Task 6: 在浏览对话框中添加"重置为根目录"按钮

**Files:**
- Modify: `web/src/views/Tasks.vue:334-343`（面包屑导航区域）

- [ ] **Step 1: 修改面包屑区域布局**

将面包屑导航区域（第 334-343 行）从：

```html
        <div class="breadcrumb-nav" style="margin-bottom: 12px;">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item>
              <a href="#" @click.prevent="navigateToBreadcrumb(-1)" class="breadcrumb-link">根目录</a>
            </el-breadcrumb-item>
            <el-breadcrumb-item v-for="(crumb, index) in breadcrumbs" :key="crumb.id">
              <a href="#" @click.prevent="navigateToBreadcrumb(index)" class="breadcrumb-link">{{ crumb.name }}</a>
            </el-breadcrumb-item>
          </el-breadcrumb>
        </div>
```

改为：

```html
        <div class="breadcrumb-nav" style="margin-bottom: 12px;">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item>
              <a href="#" @click.prevent="navigateToBreadcrumb(-1)" class="breadcrumb-link">根目录</a>
            </el-breadcrumb-item>
            <el-breadcrumb-item v-for="(crumb, index) in breadcrumbs" :key="crumb.id">
              <a href="#" @click.prevent="navigateToBreadcrumb(index)" class="breadcrumb-link">{{ crumb.name }}</a>
            </el-breadcrumb-item>
          </el-breadcrumb>
          <el-button
            v-if="isSubDirMode"
            type="warning"
            link
            size="small"
            @click="resetToShareRoot(); startFileDialogVisible = false"
            style="margin-left: auto;"
          >
            重置为根目录
          </el-button>
        </div>
```

- [ ] **Step 2: 更新面包屑导航区域样式**

将 `.breadcrumb-nav` 样式从：

```css
.breadcrumb-nav {
  padding: 8px 0;
}
```

改为：

```css
.breadcrumb-nav {
  padding: 8px 0;
  display: flex;
  align-items: center;
}
```

- [ ] **Step 3: 提交**

```bash
git add web/src/views/Tasks.vue
git commit -m "feat(task): 浏览对话框面包屑区域增加重置为根目录按钮"
```

---

### Task 7: 同步编辑/创建时的 `selectedDirName` 状态

**Files:**
- Modify: `web/src/views/Tasks.vue:1011-1031`（`openAddDialog` 函数）
- Modify: `web/src/views/Tasks.vue:1033-1067`（`handleEdit` 函数）

- [ ] **Step 1: `openAddDialog` 中清空 `selectedDirName`**

在 `openAddDialog` 函数中，`shareFiles.value = []` 之前（第 1027 行前）添加：

```js
  selectedDirName.value = ''
```

- [ ] **Step 2: `handleEdit` 中根据现有状态初始化 `selectedDirName`**

在 `handleEdit` 函数中，`dialogVisible.value = true` 之前（第 1066 行前）添加：

```js
  // 如果已有子目录选择，初始化目录名称（无法从 ID 反推，显示占位文本）
  if (row.share_parent_id) {
    selectedDirName.value = '已选子目录'
  } else {
    // Quark 平台检查 URL 中的 pdirFID
    const match = (row.share_url || '').match(/\/s\/(\w+)#\/list\/share\/(\w+)/)
    selectedDirName.value = (match && match[2] && match[2] !== '0') ? '已选子目录' : ''
  }
```

- [ ] **Step 3: 提交**

```bash
git add web/src/views/Tasks.vue
git commit -m "feat(task): 编辑/创建对话框打开时同步 selectedDirName 状态"
```

---

### Task 8: 端到端验证

- [ ] **Step 1: 启动开发服务器**

```bash
cd /home/zcq/Github/clouddrive-auto-save && make dev-web
```

- [ ] **Step 2: 验证 Quark 平台子目录重置**

1. 创建任务 → 选择 Quark 账号 → 输入分享链接
2. 点击浏览按钮 → 进入子目录 → 点击"选择当前目录"
3. 验证分享链接下方出现提示条 `📁 当前目录：xxx`
4. 点击提示条的 `×` → 验证 URL 恢复为根目录（pdirFID=0）
5. 验证提示条消失

- [ ] **Step 3: 验证 139 平台子目录重置**

1. 创建任务 → 选择 139 账号 → 输入分享链接
2. 点击浏览按钮 → 进入子目录 → 点击"选择当前目录"
3. 验证提示条出现
4. 点击提示条的 `×` → 验证 `share_parent_id` 清空
5. 验证提示条消失

- [ ] **Step 4: 验证浏览对话框内重置按钮**

1. 在子目录模式下打开浏览对话框
2. 验证面包屑右侧出现"重置为根目录"按钮
3. 点击按钮 → 验证对话框关闭且状态重置

- [ ] **Step 5: 验证 URL 变更自动清除**

1. 在子目录模式下手动修改分享链接 URL
2. 验证 `share_parent_id` 和起始文件自动清除

- [ ] **Step 6: 验证根目录模式下不显示**

1. 未选择子目录时打开任务编辑
2. 验证提示条和重置按钮均不显示
