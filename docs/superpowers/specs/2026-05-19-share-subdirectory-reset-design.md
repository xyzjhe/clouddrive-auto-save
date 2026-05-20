# 分享链接子目录重置为根目录功能设计

## 背景

当前任务编辑中，用户可以通过浏览对话框选择分享链接的子目录作为新的"根目录"。但选择后没有便捷的方式回到原始根目录，只能重新打开浏览对话框、通过面包屑导航回根目录再选择。此外，手动修改分享链接 URL 时，旧的 `share_parent_id` 不会被清除，可能导致 139 平台的子目录选择残留到新链接上。

## 目标

1. 在分享链接输入框下方增加子目录状态提示条，显示当前选中的目录名称，并提供一键清除按钮
2. 在浏览分享内容对话框的面包屑导航区域增加"重置为根目录"按钮
3. 修复手动修改 URL 时 `share_parent_id` 未清除的潜在 bug

## 涉及文件

| 文件 | 改动类型 |
|------|---------|
| `web/src/views/Tasks.vue` | 前端 UI 和逻辑 |

无后端改动。

## 设计详情

### 1. 子目录提示条

在分享链接 `el-input` 下方增加一个条件显示的 `el-tag` 组件：

- **显示条件**：处于子目录模式时显示
  - Quark 平台：从 `share_url` 中解析 `pdirFID`，值非 `0` 且非空
  - 139 平台：`share_parent_id` 非空
- **内容**：图标 + 当前目录名称（如 `📁 当前目录：电视剧/国产剧`）
- **操作**：`closable` 属性，点击 `×` 触发 `resetToShareRoot()`
- **目录名称来源**：新增 `selectedDirName` ref，在 `confirmSelectShareUrl()` 时记录当前目录名称（`currentDirName.value`），重置时清空。无法从 URL 或 ID 反推目录名，必须在用户选择时记录。

### 2. 浏览对话框中的重置按钮

在对话框的面包屑导航区域（`el-breadcrumb` 右侧）增加一个文字按钮：

- **显示条件**：当前处于子目录模式（与提示条共用判断逻辑）
- **文本**：`重置为根目录`
- **样式**：`el-button` type="warning" link，与面包屑同行右对齐
- **行为**：调用 `resetToShareRoot()` 后关闭对话框

### 3. 重置函数 `resetToShareRoot()`

新增统一的重置逻辑：

```js
const resetToShareRoot = () => {
  const platform = getSelectedPlatform()
  if (platform === 'quark') {
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
  // 清空目录名称记录
  selectedDirName.value = ''
  ElMessage.success('已重置为根目录')
}
```

### 4. URL 变更自动清除

修改 `handleUrlChange()` 函数，在 URL 变化时增加清除逻辑：

```js
const handleUrlChange = () => {
  form.value.start_file_id = ''
  form.value.start_file_name = ''
  form.value.share_parent_id = ''  // 新增：清除子目录选择
  selectedDirName.value = ''        // 新增：清除目录名称记录
}
```

### 5. 状态管理

新增以下 ref 变量：

- `selectedDirName`：`ref('')` — 记录用户选择的子目录名称，用于提示条显示

在 `confirmSelectShareUrl()` 中赋值：

```js
selectedDirName.value = currentDirName.value
```

在 `resetToShareRoot()` 和 `handleUrlChange()` 中清空。

## 验证方式

1. **Quark 平台**：创建任务 → 浏览分享链接 → 选择子目录 → 确认 → 验证提示条显示目录名称 → 点击 `×` → 验证 URL 恢复为根目录
2. **139 平台**：创建任务 → 浏览分享链接 → 选择子目录 → 确认 → 验证提示条显示 → 点击 `×` → 验证 `share_parent_id` 清空
3. **浏览对话框重置**：在子目录模式下打开浏览对话框 → 点击"重置为根目录" → 验证对话框关闭且状态重置
4. **URL 变更清除**：手动修改分享链接 URL → 验证 `share_parent_id` 和起始文件自动清除
5. **根目录模式下不显示**：未选择子目录时，提示条和重置按钮均不显示
