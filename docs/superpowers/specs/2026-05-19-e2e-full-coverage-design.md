# E2E 全覆盖设计（阶段 1）

## 背景

当前项目有 8 个 E2E spec 文件，覆盖账号展示、任务创建执行、预览、分享浏览等核心流程。但存在大量业务场景未覆盖：任务编辑/删除/批量运行、设置页配置、仪表盘交互、账号管理操作、SSE 实时更新等。用户计划后期重构前端，E2E 测试需要具备重构韧性。

## 目标

1. 补充所有未覆盖的用户可见业务流程的 E2E 测试
2. 使用语义化选择器确保测试在前端重构时不需要修改
3. 复用现有 mock 基础设施，按需扩展 mock 响应

## 选择器策略

遵循 Playwright 推荐的语义化选择器优先级：

1. `getByRole('button', { name: '...' })` — 按钮、链接、输入框
2. `getByText('...')` — 文本内容
3. `getByLabel('...')` — 表单标签
4. `getByPlaceholder('...')` — 占位符

**禁止使用**：CSS class、`data-testid`、XPath、DOM 结构选择器（如 `locator('div.el-dialog')`）。

**例外**：当多个元素的可访问名称有包含关系时（如"选择目录" vs "浏览分享内容并选择目录"），使用 `getByRole('button', { name: '...', exact: true })` 避免严格模式冲突。

## Mock 策略

复用现有基础设施：
- `internal/core/mock_driver.go` — MockDriver 拦截云盘 API
- `internal/core/mock_http.go` — mockTransport 返回预设 JSON 响应

需要扩展的 mock 场景：
- 删除操作：返回成功响应
- 批量运行：返回触发成功
- Bark 测试通知：mock Bark API 返回成功/失败
- 账号健康检查：mock 检查接口返回不同状态

## 新增测试文件

### 1. `e2e/tests/tasks/edit.spec.ts` — 任务编辑

**场景：**
- 编辑任务名称和保存路径，验证更新成功
- 编辑任务切换调度模式（跟随全局 → 自定义 cron → 手动执行）
- 子目录重置：在已选择子目录的任务上，点击提示条 × 按钮重置为根目录
- 编辑对话框打开时，已有子目录选择的任务正确显示提示条

### 2. `e2e/tests/tasks/delete.spec.ts` — 任务删除

**场景：**
- 删除任务：点击删除按钮 → 确认弹窗 → 确认 → 验证任务从列表消失
- 取消删除：点击删除按钮 → 确认弹窗 → 取消 → 验证任务仍在列表中

### 3. `e2e/tests/tasks/batch-run.spec.ts` — 批量运行

**场景：**
- 全部运行：点击"全部运行"按钮 → 确认弹窗 → 确认 → 验证可运行任务状态变为 running
- 无可用任务：所有任务已运行或有 Fatal 错误时，全部运行仍可点击但不触发新任务

### 4. `e2e/tests/tasks/dismiss.spec.ts` — 忽略失败任务

**场景：**
- 忽略失败任务：任务执行失败后，点击忽略按钮 → 验证任务状态/消息更新

### 5. `e2e/tests/settings/schedule.spec.ts` — 调度设置

**场景：**
- 启用全局调度：打开设置页 → 启用全局调度 → 设置 cron → 保存 → 验证设置生效
- 禁用全局调度：关闭全局调度开关 → 验证任务不再按计划执行
- 自定义 cron 预设：选择预设频率（每小时、每 6 小时等）→ 验证 cron 值正确填入
- 无效 cron 校验：输入无效 cron 表达式 → 保存 → 验证错误提示

### 6. `e2e/tests/settings/bark.spec.ts` — Bark 通知

**场景：**
- 配置 Bark：填写 Bark URL 和 Key → 保存 → 验证配置生效
- 测试 Bark 发送：点击"测试 Bark"按钮 → 验证成功/失败反馈

### 7. `e2e/tests/accounts/delete.spec.ts` — 账号删除

**场景：**
- 删除账号：点击删除按钮 → 确认弹窗 → 确认 → 验证账号从列表消失
- 删除有任务关联的账号：验证关联任务的处理（应提示或阻止）

### 8. `e2e/tests/accounts/check.spec.ts` — 账号健康检查

**场景：**
- 检查有效账号：点击检查按钮 → 验证状态变为正常
- 检查失效账号：mock 返回失效 → 验证状态变为已失效

### 9. `e2e/tests/dashboard/logs.spec.ts` — 日志管理

**场景：**
- 查看历史日志：执行任务后 → 验证日志出现在历史记录中
- 清空日志：点击清空按钮 → 确认 → 验证日志列表为空

### 10. `e2e/tests/dashboard/sse.spec.ts` — SSE 实时更新

**场景：**
- 任务状态实时更新：触发任务执行 → 通过 SSE 验证状态从 pending → running → success 变化
- 任务完成后列表同步：验证任务完成后统计数据更新

## 文件组织

沿用现有目录结构：
```
e2e/tests/
├── accounts/
│   ├── cloud139.spec.ts      (已有)
│   ├── quark.spec.ts          (已有)
│   ├── delete.spec.ts         (新增)
│   └── check.spec.ts          (新增)
├── dashboard/
│   ├── overview.spec.ts       (已有)
│   ├── logs.spec.ts           (新增)
│   └── sse.spec.ts            (新增)
├── settings/
│   ├── global.spec.ts         (已有，需增强)
│   ├── schedule.spec.ts       (新增)
│   └── bark.spec.ts           (新增)
└── tasks/
    ├── create.spec.ts         (已有)
    ├── execute.spec.ts        (已有)
    ├── preview.spec.ts        (已有)
    ├── share-browse.spec.ts   (已有)
    ├── edit.spec.ts           (新增)
    ├── delete.spec.ts         (新增)
    ├── batch-run.spec.ts      (新增)
    └── dismiss.spec.ts        (新增)
```

## 涉及文件

| 文件 | 改动类型 |
|------|---------|
| `e2e/tests/tasks/edit.spec.ts` | 新增 |
| `e2e/tests/tasks/delete.spec.ts` | 新增 |
| `e2e/tests/tasks/batch-run.spec.ts` | 新增 |
| `e2e/tests/tasks/dismiss.spec.ts` | 新增 |
| `e2e/tests/settings/schedule.spec.ts` | 新增 |
| `e2e/tests/settings/bark.spec.ts` | 新增 |
| `e2e/tests/accounts/delete.spec.ts` | 新增 |
| `e2e/tests/accounts/check.spec.ts` | 新增 |
| `e2e/tests/dashboard/logs.spec.ts` | 新增 |
| `e2e/tests/dashboard/sse.spec.ts` | 新增 |
| `e2e/tests/settings/global.spec.ts` | 增强（当前仅占位） |
| `internal/core/mock_http.go` | 可能需要扩展 mock 响应 |
| `internal/core/mock_driver.go` | 可能需要扩展 mock 方法 |

## 验证方式

```bash
# 运行全部 E2E 测试
make e2e-test

# 运行单个 spec
npx playwright test e2e/tests/tasks/edit.spec.ts

# 查看测试报告
npx playwright show-report
```
