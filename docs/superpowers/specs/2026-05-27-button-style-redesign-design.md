# 按钮样式重新设计规范

**日期**: 2026-05-27
**状态**: 已批准
**设计风格**: 现代极简 + 微妙科技感

---

## 1. 设计背景

### 1.1 问题描述

当前界面的按钮样式缺乏设计感，主要表现为：
- Element Plus 默认样式过于普通
- 现有的渐变覆盖效果（霓虹发光）与整体设计风格不协调
- 表格操作按钮使用 `link` 模式，视觉层次不清晰
- 按钮交互反馈不够精致

### 1.2 设计目标

- 建立统一、精致的按钮视觉语言
- 保持极简风格，同时融入微妙的科技感
- 提升按钮的交互体验和视觉层次
- 确保亮/暗主题下的良好表现

---

## 2. 设计原则

### 2.1 极简为主

- 干净、扁平的设计，专注于内容和功能
- 避免过度装饰，保持视觉简洁
- 使用清晰的视觉层次引导用户操作

### 2.2 微妙科技感

- 通过精致的阴影效果体现科技感
- hover 时按钮微微上浮，增强交互反馈
- 保持专业感，避免过度花哨

### 2.3 一致性

- 所有按钮遵循统一的视觉语言
- 相同类型的按钮在不同页面保持一致
- 确保亮/暗主题下的视觉一致性

---

## 3. 按钮类型规范

### 3.1 主要按钮（Primary Button）

**使用场景**: 创建、保存、确认等核心操作

**样式定义**:
```css
.btn-primary {
  background: #3b82f6;
  color: #ffffff;
  border: none;
  padding: 10px 20px;
  border-radius: 6px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);
}

.btn-primary:hover {
  box-shadow: 0 4px 8px rgba(59, 130, 246, 0.4);
  transform: translateY(-1px);
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
  box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);
}
```

**交互效果**:
- hover: 阴影增强 + 微微上浮（translateY(-1px)）
- active: 恢复原位
- disabled: 降低透明度，移除上浮效果

**示例位置**:
- Tasks.vue: 创建任务、确认并保存
- Accounts.vue: 确认更新/确认添加
- Dashboard.vue: 创建新任务

### 3.2 次要按钮（Secondary Button）

**使用场景**: 次要操作，如全部运行、管理账号等

**样式定义**:
```css
.btn-secondary {
  background: #ffffff;
  color: #3b82f6;
  border: 1px solid #e5e7eb;
  padding: 10px 20px;
  border-radius: 6px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.btn-secondary:hover {
  border-color: #3b82f6;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.btn-secondary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
```

**交互效果**:
- hover: 边框变为 #3b82f6，阴影增强
- disabled: 降低透明度

**示例位置**:
- Tasks.vue: 全部运行
- Dashboard.vue: 管理账号

### 3.3 幽灵按钮（Ghost Button）

**使用场景**: 取消、关闭等低优先级操作

**样式定义**:
```css
.btn-ghost {
  background: transparent;
  color: #6b7280;
  border: 1px solid #d1d5db;
  padding: 10px 20px;
  border-radius: 6px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-ghost:hover {
  border-color: #3b82f6;
  color: #3b82f6;
}

.btn-ghost:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
```

**交互效果**:
- hover: 边框和文字变为 #3b82f6
- disabled: 降低透明度

**示例位置**:
- Tasks.vue: 取消
- Dashboard.vue: 清理结束任务

### 3.4 表格操作图标按钮（Icon Button）

**使用场景**: 表格行内操作（运行、编辑、删除）

**样式定义**:
```css
.btn-icon {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  border: none;
  font-size: 14px;
}

.btn-icon--success {
  color: #10b981;
  background: rgba(16, 185, 129, 0.1);
  box-shadow: 0 1px 2px rgba(16, 185, 129, 0.2);
}

.btn-icon--success:hover {
  box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);
}

.btn-icon--primary {
  color: #3b82f6;
  background: rgba(59, 130, 246, 0.1);
  box-shadow: 0 1px 2px rgba(59, 130, 246, 0.2);
}

.btn-icon--primary:hover {
  box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);
}

.btn-icon--danger {
  color: #ef4444;
  background: rgba(239, 68, 68, 0.1);
  box-shadow: 0 1px 2px rgba(239, 68, 68, 0.2);
}

.btn-icon--danger:hover {
  box-shadow: 0 2px 4px rgba(239, 68, 68, 0.3);
}
```

**交互效果**:
- hover: 阴影增强
- 使用 `title` 属性显示文字提示

**示例位置**:
- Tasks.vue: 表格操作列（运行、编辑、删除）
- Accounts.vue: 表格操作列（校验、编辑、删除）

---

## 4. 尺寸规范

### 4.1 小型按钮（Small）

```css
.btn--small {
  padding: 8px 16px;
  font-size: 12px;
}
```

**使用场景**: 空间紧凑的区域，如表格内、卡片内

### 4.2 默认按钮（Default）

```css
/* 默认尺寸，无需额外类 */
padding: 10px 20px;
font-size: 14px;
```

**使用场景**: 大多数场景的默认选择

### 4.3 大型按钮（Large）

```css
.btn--large {
  padding: 12px 24px;
  font-size: 16px;
}
```

**使用场景**: 重要的主要操作，需要突出显示

---

## 5. 颜色规范

### 5.1 主要蓝（Primary Blue）

- **颜色值**: #3b82f6
- **使用场景**: 主要按钮背景、次要按钮文字、链接按钮 hover
- **阴影**: rgba(59, 130, 246, 0.3)

### 5.2 成功绿（Success Green）

- **颜色值**: #10b981
- **使用场景**: 运行按钮、成功状态指示
- **阴影**: rgba(16, 185, 129, 0.2)

### 5.3 危险红（Danger Red）

- **颜色值**: #ef4444
- **使用场景**: 删除按钮、错误状态指示
- **阴影**: rgba(239, 68, 68, 0.2)

### 5.4 中性灰（Neutral Gray）

- **颜色值**: #6b7280
- **使用场景**: 幽灵按钮、次要文字
- **边框**: #d1d5db

---

## 6. CSS 变量定义

```css
:root {
  /* 按钮基础变量 */
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

  /* 成功色 */
  --btn-success: #10b981;
  --btn-success-bg: rgba(16, 185, 129, 0.1);
  --btn-success-shadow: 0 1px 2px rgba(16, 185, 129, 0.2);

  /* 危险色 */
  --btn-danger: #ef4444;
  --btn-danger-bg: rgba(239, 68, 68, 0.1);
  --btn-danger-shadow: 0 1px 2px rgba(239, 68, 68, 0.2);
}
```

---

## 7. 暗黑主题适配

### 7.1 颜色调整

```css
html.dark {
  /* 次要按钮背景 */
  --btn-secondary-bg: #1f2937;
  --btn-secondary-border: #374151;

  /* 幽灵按钮边框 */
  --btn-ghost-border: #4b5563;
}
```

### 7.2 阴影调整

暗黑主题下，阴影颜色需要调整为更深的色调，以保持视觉效果：

```css
html.dark .btn-primary {
  box-shadow: 0 2px 4px rgba(59, 130, 246, 0.2);
}

html.dark .btn-primary:hover {
  box-shadow: 0 4px 8px rgba(59, 130, 246, 0.3);
}
```

---

## 8. 实现指南

### 8.1 全局样式覆盖

在 `web/src/style.css` 中移除现有的按钮覆盖样式，替换为新的设计规范。

### 8.2 组件更新

需要更新以下组件中的按钮实现：

1. **Tasks.vue**
   - 表格操作列：将 `el-button-group` + `link` 模式改为图标按钮组
   - 页面顶部：主要按钮和次要按钮应用新样式
   - 抽屉底部：主要按钮和幽灵按钮应用新样式

2. **Accounts.vue**
   - 表格操作列：将 `el-button-group` + `link` 模式改为图标按钮组
   - 对话框底部：主要按钮和幽灵按钮应用新样式

3. **Dashboard.vue**
   - 底部快捷操作栏：主要按钮、次要按钮、幽灵按钮应用新样式
   - 忽略/关闭按钮：图标按钮应用新样式

4. **Settings.vue**
   - 预设时间按钮组：应用新样式
   - 通知渠道按钮：主要按钮和次要按钮应用新样式

5. **TaskCard.vue**
   - 卡片操作按钮：应用新样式

### 8.3 图标按钮实现

将现有的 `el-button-group` + `link` 模式改为自定义图标按钮：

```vue
<!-- 之前 -->
<el-button-group>
  <el-button link type="success" :icon="Play" @click="handleRun(row)">运行</el-button>
  <el-button link type="primary" :icon="Edit" @click="handleEdit(row)">编辑</el-button>
  <el-button link type="danger" :icon="Trash2" @click="handleDelete(row)">删除</el-button>
</el-button-group>

<!-- 之后 -->
<div class="action-buttons">
  <button
    class="btn-icon btn-icon--success"
    title="运行"
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

---

## 9. 测试验证

### 9.1 视觉测试

- 在亮/暗主题下验证所有按钮类型
- 验证不同尺寸按钮的显示效果
- 验证按钮的交互状态（hover、active、disabled）

### 9.2 功能测试

- 验证所有按钮的点击事件正常触发
- 验证按钮的 loading 状态
- 验证按钮的禁用状态

### 9.3 响应式测试

- 在不同屏幕尺寸下验证按钮显示
- 验证表格操作按钮在移动端的可用性

---

## 10. 设计决策记录

### 10.1 为什么选择 6px 圆角？

- 比 Element Plus 默认的 4px 更现代
- 比 8px 或 12px 更紧凑、专业
- 与整体设计风格协调

### 10.2 为什么使用阴影而非渐变？

- 阴影效果更精致、克制
- 渐变容易显得花哨
- 阴影在亮/暗主题下表现一致

### 10.3 为什么表格操作使用图标按钮？

- 节省表格空间
- 视觉更简洁
- 通过 title 属性保持可访问性

---

**文档版本**: v1.0
**最后更新**: 2026-05-27
**作者**: Claude Code
