import { test, expect } from '@playwright/test';

test.describe('仪表盘：SSE 实时更新测试', () => {
  test('任务执行时仪表盘显示实时状态', async ({ page }) => {
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

    await taskRow.getByRole('button', { name: '运行' }).click();

    await page.goto('/');

    await expect(page.getByText(taskName)).toBeVisible({ timeout: 15000 });

    await expect(page.locator('.el-tag').filter({ hasText: 'SUCCESS' })).toBeVisible({ timeout: 60000 });
  });
});
