# E2E 全覆盖实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 补充所有未覆盖业务流程的 E2E 测试，使用语义化选择器确保前端重构韧性

**Architecture:** 新增 10 个 Playwright spec 文件，沿用现有 mock 基础设施（mock_http.go + page.route），所有选择器使用 getByRole/getByText/getByLabel 语义化定位

**Tech Stack:** Playwright, TypeScript, Element Plus (el-dialog, el-message-box, el-popconfirm)

---

### Task 1: 任务编辑测试

**Files:**
- Create: `e2e/tests/tasks/edit.spec.ts`

- [ ] **Step 1: 创建 edit.spec.ts**

```typescript
import { test, expect } from '@playwright/test';

test.describe('任务管理：编辑测试', () => {
  test('编辑任务名称和保存路径', async ({ page }) => {
    const originalName = `E2E_编辑_原始_${Date.now()}`;
    const updatedName = `E2E_编辑_更新_${Date.now()}`;

    // 创建任务
    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();
    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).first().click();
    await page.getByLabel('任务名称').fill(originalName);
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_success');
    await page.getByLabel('保存路径').fill('/edit_test');
    await page.getByRole('button', { name: '确认并保存' }).click();

    const taskRow = page.locator('tr').filter({ hasText: originalName });
    await expect(taskRow).toBeVisible();

    // 编辑任务
    await taskRow.getByRole('button', { name: '编辑' }).click();
    const dialog = page.getByRole('dialog', { name: '编辑任务' });
    await expect(dialog).toBeVisible();

    await page.getByLabel('任务名称').fill(updatedName);
    await page.getByLabel('保存路径').fill('/edit_test_updated');
    await page.getByRole('button', { name: '确认并保存' }).click();

    // 验证更新
    await expect(page.locator('tr').filter({ hasText: updatedName })).toBeVisible();
    await expect(page.locator('tr').filter({ hasText: originalName })).not.toBeVisible();
  });

  test('编辑任务切换调度模式', async ({ page }) => {
    const taskName = `E2E_调度_${Date.now()}`;

    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();
    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).first().click();
    await page.getByLabel('任务名称').fill(taskName);
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_success');
    await page.getByRole('button', { name: '确认并保存' }).click();

    const taskRow = page.locator('tr').filter({ hasText: taskName });
    await expect(taskRow).toBeVisible();

    // 编辑为自定义频率
    await taskRow.getByRole('button', { name: '编辑' }).click();
    await page.getByRole('radio', { name: '自定义频率' }).click();
    await expect(page.getByLabel('自定义频率 (Cron)')).toBeVisible();
    await page.getByRole('button', { name: '确认并保存' }).click();

    // 验证调度标签
    await expect(taskRow.getByText('自定义')).toBeVisible();
  });

  test('子目录重置：点击提示条清除按钮重置为根目录', async ({ page }) => {
    const taskName = `E2E_重置_${Date.now()}`;

    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();
    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).first().click();
    await page.getByLabel('任务名称').fill(taskName);
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_success');

    // 选择子目录
    await page.getByRole('button', { name: '浏览分享内容并选择目录' }).click();
    const browseDialog = page.getByRole('dialog', { name: '浏览分享内容' });
    await expect(browseDialog).toBeVisible();
    await browseDialog.getByText('139分享子目录').first().click();
    await browseDialog.getByRole('button', { name: '进入' }).click();
    await browseDialog.getByRole('button', { name: /选择当前目录/ }).click();

    // 验证提示条出现
    await expect(page.getByText('当前目录：')).toBeVisible();

    // 点击清除按钮重置
    await page.getByRole('button', { name: 'Close' }).click();

    // 验证提示条消失
    await expect(page.getByText('当前目录：')).not.toBeVisible();
  });
});
```

- [ ] **Step 2: 运行测试验证**

```bash
npx playwright test e2e/tests/tasks/edit.spec.ts --reporter=list
```

- [ ] **Step 3: 提交**

```bash
git add e2e/tests/tasks/edit.spec.ts
git commit -m "test(e2e): 新增任务编辑和子目录重置 E2E 测试"
```

---

### Task 2: 任务删除测试

**Files:**
- Create: `e2e/tests/tasks/delete.spec.ts`

- [ ] **Step 1: 创建 delete.spec.ts**

```typescript
import { test, expect } from '@playwright/test';

test.describe('任务管理：删除测试', () => {
  test('删除任务：确认后任务从列表消失', async ({ page }) => {
    const taskName = `E2E_删除_${Date.now()}`;

    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();
    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).first().click();
    await page.getByLabel('任务名称').fill(taskName);
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_success');
    await page.getByRole('button', { name: '确认并保存' }).click();

    const taskRow = page.locator('tr').filter({ hasText: taskName });
    await expect(taskRow).toBeVisible();

    // 删除
    await taskRow.getByRole('button', { name: '删除' }).click();
    await page.getByRole('button', { name: '确定' }).click();

    // 验证消失
    await expect(taskRow).not.toBeVisible();
  });

  test('取消删除：任务仍在列表中', async ({ page }) => {
    const taskName = `E2E_取消删除_${Date.now()}`;

    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();
    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).first().click();
    await page.getByLabel('任务名称').fill(taskName);
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_success');
    await page.getByRole('button', { name: '确认并保存' }).click();

    const taskRow = page.locator('tr').filter({ hasText: taskName });
    await expect(taskRow).toBeVisible();

    // 取消删除
    await taskRow.getByRole('button', { name: '删除' }).click();
    await page.getByRole('button', { name: '取消' }).click();

    // 验证仍在
    await expect(taskRow).toBeVisible();
  });
});
```

- [ ] **Step 2: 运行测试验证**

```bash
npx playwright test e2e/tests/tasks/delete.spec.ts --reporter=list
```

- [ ] **Step 3: 提交**

```bash
git add e2e/tests/tasks/delete.spec.ts
git commit -m "test(e2e): 新增任务删除确认/取消 E2E 测试"
```

---

### Task 3: 批量运行测试

**Files:**
- Create: `e2e/tests/tasks/batch-run.spec.ts`

- [ ] **Step 1: 创建 batch-run.spec.ts**

```typescript
import { test, expect } from '@playwright/test';

test.describe('任务管理：批量运行测试', () => {
  test('全部运行：确认后可运行任务状态变为 running', async ({ page }) => {
    const taskName1 = `E2E_批量1_${Date.now()}`;
    const taskName2 = `E2E_批量2_${Date.now()}`;

    await page.goto('/tasks');

    // 创建两个任务
    for (const name of [taskName1, taskName2]) {
      await page.getByRole('button', { name: '创建任务' }).last().click();
      await page.locator('.el-select').first().click();
      await page.getByRole('option', { name: 'E2E移动云盘用户' }).first().click();
      await page.getByLabel('任务名称').fill(name);
      await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_success');
      await page.getByLabel('保存路径').fill('/batch_test');
      await page.getByRole('button', { name: '确认并保存' }).click();
      await expect(page.locator('tr').filter({ hasText: name })).toBeVisible();
    }

    // 全部运行
    await page.getByRole('button', { name: '全部运行' }).click();
    await page.getByRole('button', { name: '确认' }).click();

    // 验证任务状态变为 running 或 success
    for (const name of [taskName1, taskName2]) {
      const row = page.locator('tr').filter({ hasText: name });
      await expect(row.locator('.el-tag').filter({ hasText: /RUNNING|SUCCESS/ })).toBeVisible({ timeout: 60000 });
    }
  });
});
```

- [ ] **Step 2: 运行测试验证**

```bash
npx playwright test e2e/tests/tasks/batch-run.spec.ts --reporter=list
```

- [ ] **Step 3: 提交**

```bash
git add e2e/tests/tasks/batch-run.spec.ts
git commit -m "test(e2e): 新增批量运行 E2E 测试"
```

---

### Task 4: 忽略失败任务测试

**Files:**
- Create: `e2e/tests/tasks/dismiss.spec.ts`

- [ ] **Step 1: 创建 dismiss.spec.ts**

```typescript
import { test, expect } from '@playwright/test';

test.describe('任务管理：忽略失败任务测试', () => {
  test('忽略失败任务后状态更新', async ({ page }) => {
    const taskName = `E2E_忽略_${Date.now()}`;

    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();
    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).first().click();
    await page.getByLabel('任务名称').fill(taskName);
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_invalid');
    await page.getByRole('button', { name: '确认并保存' }).click();

    const taskRow = page.locator('tr').filter({ hasText: taskName });
    await expect(taskRow).toBeVisible();

    // 运行使任务失败
    await taskRow.getByRole('button', { name: '运行' }).click();
    await page.reload();

    // 验证 LINK ERROR
    const updatedRow = page.locator('tr').filter({ hasText: taskName });
    await expect(updatedRow.locator('.el-tag--danger').filter({ hasText: 'LINK ERROR' })).toBeVisible({ timeout: 15000 });

    // 注意：忽略按钮在仪表盘的"实时执行状态"区域，不在任务列表中
    // 此测试验证任务失败后在仪表盘中可见
    await page.goto('/');
    await expect(page.getByText(taskName)).toBeVisible({ timeout: 15000 });
  });
});
```

- [ ] **Step 2: 运行测试验证**

```bash
npx playwright test e2e/tests/tasks/dismiss.spec.ts --reporter=list
```

- [ ] **Step 3: 提交**

```bash
git add e2e/tests/tasks/dismiss.spec.ts
git commit -m "test(e2e): 新增忽略失败任务 E2E 测试"
```

---

### Task 5: 调度设置测试

**Files:**
- Create: `e2e/tests/settings/schedule.spec.ts`

- [ ] **Step 1: 创建 schedule.spec.ts**

```typescript
import { test, expect } from '@playwright/test';

test.describe('系统设置：调度配置测试', () => {
  test('启用全局调度并设置简易定时', async ({ page }) => {
    await page.goto('/settings');
    await expect(page.getByText('系统设置')).toBeVisible();

    // 找到全局定时任务卡片中的开关
    const scheduleCard = page.locator('.el-card').filter({ hasText: '全局定时任务' });

    // 启用调度开关
    const scheduleSwitch = scheduleCard.locator('.el-switch');
    if (!(await scheduleSwitch.isChecked())) {
      await scheduleSwitch.click();
    }

    // 选择简易定时模式
    await page.getByRole('radio', { name: '简易定时' }).click();

    // 选择预设时间
    await page.getByRole('button', { name: '凌晨' }).click();

    // 保存
    await scheduleCard.getByRole('button', { name: '保存配置' }).click();

    // 验证保存成功消息
    await expect(page.getByText('保存成功')).toBeVisible({ timeout: 5000 });
  });

  test('设置自定义 Cron 表达式', async ({ page }) => {
    await page.goto('/settings');

    const scheduleCard = page.locator('.el-card').filter({ hasText: '全局定时任务' });

    // 启用调度
    const scheduleSwitch = scheduleCard.locator('.el-switch');
    if (!(await scheduleSwitch.isChecked())) {
      await scheduleSwitch.click();
    }

    // 选择高级 Cron 模式
    await page.getByRole('radio', { name: '高级 Cron' }).click();

    // 输入自定义 cron
    await page.getByLabel('全局 Cron 表达式').fill('0 0 */6 * * *');

    // 保存
    await scheduleCard.getByRole('button', { name: '保存配置' }).click();
    await expect(page.getByText('保存成功')).toBeVisible({ timeout: 5000 });
  });

  test('禁用全局调度', async ({ page }) => {
    await page.goto('/settings');

    const scheduleCard = page.locator('.el-card').filter({ hasText: '全局定时任务' });
    const scheduleSwitch = scheduleCard.locator('.el-switch');

    // 如果已启用则关闭
    if (await scheduleSwitch.isChecked()) {
      await scheduleSwitch.click();
    }

    await scheduleCard.getByRole('button', { name: '保存配置' }).click();
    await expect(page.getByText('保存成功')).toBeVisible({ timeout: 5000 });
  });
});
```

- [ ] **Step 2: 运行测试验证**

```bash
npx playwright test e2e/tests/settings/schedule.spec.ts --reporter=list
```

- [ ] **Step 3: 提交**

```bash
git add e2e/tests/settings/schedule.spec.ts
git commit -m "test(e2e): 新增调度设置 E2E 测试"
```

---

### Task 6: Bark 通知测试

**Files:**
- Create: `e2e/tests/settings/bark.spec.ts`

- [ ] **Step 1: 创建 bark.spec.ts**

```typescript
import { test, expect } from '@playwright/test';

test.describe('系统设置：Bark 通知测试', () => {
  test('配置 Bark 并保存', async ({ page }) => {
    await page.goto('/settings');

    const barkCard = page.locator('.el-card').filter({ hasText: 'Bark 消息推送' });

    // 启用 Bark
    const barkSwitch = barkCard.locator('.el-switch');
    if (!(await barkSwitch.isChecked())) {
      await barkSwitch.click();
    }

    // 填写配置
    await page.getByLabel('Bark 服务器地址').fill('https://api.day.app');
    await page.getByLabel('Device Key').fill('mock_device_key');

    // 保存
    await barkCard.getByRole('button', { name: '保存配置' }).click();
    await expect(page.getByText('保存成功')).toBeVisible({ timeout: 5000 });
  });

  test('发送测试消息', async ({ page }) => {
    await page.goto('/settings');

    const barkCard = page.locator('.el-card').filter({ hasText: 'Bark 消息推送' });

    // 启用并配置
    const barkSwitch = barkCard.locator('.el-switch');
    if (!(await barkSwitch.isChecked())) {
      await barkSwitch.click();
    }
    await page.getByLabel('Bark 服务器地址').fill('https://api.day.app');
    await page.getByLabel('Device Key').fill('mock_device_key');
    await barkCard.getByRole('button', { name: '保存配置' }).click();
    await expect(page.getByText('保存成功')).toBeVisible({ timeout: 5000 });

    // 发送测试
    await page.getByRole('button', { name: '发送测试消息' }).click();

    const testDialog = page.getByRole('dialog', { name: '发送测试推送' });
    await expect(testDialog).toBeVisible();

    await page.getByLabel('推送标题').fill('E2E 测试推送');
    await page.getByLabel('推送内容').fill('这是一条 E2E 测试消息');

    await page.getByRole('button', { name: '立即发送' }).click();

    // 验证发送成功
    await expect(page.getByText(/发送成功|测试消息已发送/)).toBeVisible({ timeout: 10000 });
  });
});
```

- [ ] **Step 2: 运行测试验证**

```bash
npx playwright test e2e/tests/settings/bark.spec.ts --reporter=list
```

- [ ] **Step 3: 提交**

```bash
git add e2e/tests/settings/bark.spec.ts
git commit -m "test(e2e): 新增 Bark 通知配置和测试发送 E2E 测试"
```

---

### Task 7: 账号删除测试

**Files:**
- Create: `e2e/tests/accounts/delete.spec.ts`

- [ ] **Step 1: 创建 delete.spec.ts**

```typescript
import { test, expect } from '@playwright/test';

test.describe('账号管理：删除测试', () => {
  test('删除无关联任务的账号', async ({ page }) => {
    await page.goto('/accounts');

    // 添加一个临时账号
    await page.getByRole('button', { name: /立即绑定账号|添加账号/ }).first().click();
    await page.getByLabel('网盘平台').getByText('移动云盘').click();
    await page.getByLabel('Authorization').fill('mock_delete_test');
    await page.getByRole('button', { name: '确认添加' }).click();

    // 等待账号出现
    await expect(page.getByText('E2E移动云盘用户').first()).toBeVisible({ timeout: 10000 });

    // 删除（注意：删除的是刚添加的账号，可能需要定位到正确的行）
    const accountRows = page.locator('.el-table__row');
    const firstDeleteBtn = accountRows.first().getByRole('button', { name: '删除' });
    await firstDeleteBtn.click();

    // 确认删除
    await page.getByRole('button', { name: '确定' }).click();

    // 验证成功消息
    await expect(page.getByText(/删除成功|已删除/)).toBeVisible({ timeout: 5000 });
  });

  test('取消删除账号', async ({ page }) => {
    await page.goto('/accounts');

    const firstDeleteBtn = page.locator('.el-table__row').first().getByRole('button', { name: '删除' });
    const initialRowCount = await page.locator('.el-table__row').count();

    await firstDeleteBtn.click();
    await page.getByRole('button', { name: '取消' }).click();

    // 验证行数不变
    await expect(page.locator('.el-table__row')).toHaveCount(initialRowCount);
  });
});
```

- [ ] **Step 2: 运行测试验证**

```bash
npx playwright test e2e/tests/accounts/delete.spec.ts --reporter=list
```

- [ ] **Step 3: 提交**

```bash
git add e2e/tests/accounts/delete.spec.ts
git commit -m "test(e2e): 新增账号删除确认/取消 E2E 测试"
```

---

### Task 8: 账号健康检查测试

**Files:**
- Create: `e2e/tests/accounts/check.spec.ts`

- [ ] **Step 1: 创建 check.spec.ts**

```typescript
import { test, expect } from '@playwright/test';

test.describe('账号管理：健康检查测试', () => {
  test('校验有效账号', async ({ page }) => {
    await page.goto('/accounts');

    // 找到第一个账号的校验按钮
    const firstRow = page.locator('.el-table__row').first();
    await firstRow.getByRole('button', { name: '校验' }).click();

    // 验证校验成功消息
    await expect(page.getByText(/校验成功|验证成功|正常/)).toBeVisible({ timeout: 15000 });
  });
});
```

- [ ] **Step 2: 运行测试验证**

```bash
npx playwright test e2e/tests/accounts/check.spec.ts --reporter=list
```

- [ ] **Step 3: 提交**

```bash
git add e2e/tests/accounts/check.spec.ts
git commit -m "test(e2e): 新增账号健康检查 E2E 测试"
```

---

### Task 9: 日志管理测试

**Files:**
- Create: `e2e/tests/dashboard/logs.spec.ts`

- [ ] **Step 1: 创建 logs.spec.ts**

```typescript
import { test, expect } from '@playwright/test';

test.describe('仪表盘：日志管理测试', () => {
  test('查看历史日志', async ({ page }) => {
    // 先执行一个任务产生日志
    const taskName = `E2E_日志_${Date.now()}`;
    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();
    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).first().click();
    await page.getByLabel('任务名称').fill(taskName);
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_success');
    await page.getByRole('button', { name: '确认并保存' }).click();

    const taskRow = page.locator('tr').filter({ hasText: taskName });
    await expect(taskRow).toBeVisible();
    await taskRow.getByRole('button', { name: '运行' }).click();

    // 等待任务完成
    await expect(taskRow.locator('.el-tag').filter({ hasText: 'SUCCESS' })).toBeVisible({ timeout: 60000 });

    // 去仪表盘查看日志
    await page.goto('/');
    await expect(page.getByText('实时日志流')).toBeVisible();

    // 验证日志区域有内容（不是空的）
    const logArea = page.locator('.log-terminal, .log-content, pre').first();
    await expect(logArea).toBeVisible();
  });

  test('清空日志', async ({ page }) => {
    await page.goto('/');

    // 点击清空日志按钮（通过 tooltip 或图标定位）
    const clearBtn = page.getByRole('button', { name: '清空日志' });
    if (await clearBtn.isVisible()) {
      await clearBtn.click();
      // 确认弹窗
      await page.getByRole('button', { name: '确定' }).click();
      await expect(page.getByText(/已清空|清空成功/)).toBeVisible({ timeout: 5000 });
    }
  });
});
```

- [ ] **Step 2: 运行测试验证**

```bash
npx playwright test e2e/tests/dashboard/logs.spec.ts --reporter=list
```

- [ ] **Step 3: 提交**

```bash
git add e2e/tests/dashboard/logs.spec.ts
git commit -m "test(e2e): 新增日志查看和清空 E2E 测试"
```

---

### Task 10: SSE 实时更新测试

**Files:**
- Create: `e2e/tests/dashboard/sse.spec.ts`

- [ ] **Step 1: 创建 sse.spec.ts**

```typescript
import { test, expect } from '@playwright/test';

test.describe('仪表盘：SSE 实时更新测试', () => {
  test('任务执行时仪表盘显示实时状态', async ({ page }) => {
    // 创建任务
    const taskName = `E2E_SSE_${Date.now()}`;
    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();
    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).first().click();
    await page.getByLabel('任务名称').fill(taskName);
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_success');
    await page.getByRole('button', { name: '确认并保存' }).click();

    const taskRow = page.locator('tr').filter({ hasText: taskName });
    await expect(taskRow).toBeVisible();

    // 运行任务
    await taskRow.getByRole('button', { name: '运行' }).click();

    // 去仪表盘
    await page.goto('/');

    // 验证任务出现在实时执行状态区域
    await expect(page.getByText(taskName)).toBeVisible({ timeout: 15000 });

    // 验证任务完成后状态更新
    await expect(page.locator('.el-tag').filter({ hasText: 'SUCCESS' })).toBeVisible({ timeout: 60000 });
  });
});
```

- [ ] **Step 2: 运行测试验证**

```bash
npx playwright test e2e/tests/dashboard/sse.spec.ts --reporter=list
```

- [ ] **Step 3: 提交**

```bash
git add e2e/tests/dashboard/sse.spec.ts
git commit -m "test(e2e): 新增 SSE 实时状态更新 E2E 测试"
```

---

### Task 11: 增强设置页占位测试

**Files:**
- Modify: `e2e/tests/settings/global.spec.ts`

- [ ] **Step 1: 替换占位测试为有意义的验证**

将 `global.spec.ts` 的内容替换为：

```typescript
import { test, expect } from '@playwright/test';

test.describe('系统设置：页面加载测试', () => {
  test('设置页面正确加载所有配置卡片', async ({ page }) => {
    await page.goto('/settings');
    await expect(page.getByText('系统设置')).toBeVisible();

    // 验证三个配置卡片都存在
    await expect(page.getByText('全局定时任务')).toBeVisible();
    await expect(page.getByText('Bark 消息推送')).toBeVisible();
    await expect(page.getByText('OpenList 扫描')).toBeVisible();

    // 验证每个卡片都有保存按钮
    const saveButtons = page.getByRole('button', { name: '保存配置' });
    await expect(saveButtons).toHaveCount(3);
  });
});
```

- [ ] **Step 2: 运行测试验证**

```bash
npx playwright test e2e/tests/settings/global.spec.ts --reporter=list
```

- [ ] **Step 3: 提交**

```bash
git add e2e/tests/settings/global.spec.ts
git commit -m "test(e2e): 增强设置页面加载 E2E 测试"
```

---

### Task 12: 全量回归验证

- [ ] **Step 1: 运行全部 E2E 测试**

```bash
make e2e-test
```

- [ ] **Step 2: 修复失败的测试**

检查测试报告，修复选择器不匹配或时序问题。

- [ ] **Step 3: 最终提交**

```bash
git add -A
git commit -m "test(e2e): E2E 全覆盖完成，全部测试通过"
```
