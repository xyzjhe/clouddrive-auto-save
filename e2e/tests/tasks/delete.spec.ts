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

    await taskRow.getByRole('button', { name: '删除' }).click();
    await expect(page.getByText('确定要删除此转存任务吗？')).toBeVisible({ timeout: 5000 });
    await page.getByRole('button', { name: '确定' }).click();

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

    await taskRow.getByRole('button', { name: '删除' }).click();
    await expect(page.getByText('确定要删除此转存任务吗？')).toBeVisible({ timeout: 5000 });
    await page.getByRole('button', { name: '取消' }).click();

    await expect(taskRow).toBeVisible();
  });
});
