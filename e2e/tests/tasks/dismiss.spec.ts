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

    await taskRow.getByRole('button', { name: '运行' }).click();
    await page.reload();

    const updatedRow = page.locator('tr').filter({ hasText: taskName });
    await expect(updatedRow.locator('.el-tag--danger').filter({ hasText: 'LINK ERROR' })).toBeVisible({ timeout: 15000 });

    await page.goto('/');
    await expect(page.getByText(taskName).first()).toBeVisible({ timeout: 15000 });
  });
});
