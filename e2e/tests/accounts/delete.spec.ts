import { test, expect } from '@playwright/test';

test.describe('账号管理：删除测试', () => {
  test('删除无关联任务的账号', async ({ page }) => {
    await page.goto('/accounts');

    await page.getByRole('button', { name: /立即绑定账号|添加账号/ }).first().click();
    await page.getByLabel('网盘平台').getByText('移动云盘').click();
    await page.getByLabel('Authorization').fill('mock_delete_test');
    await page.getByRole('button', { name: '确认添加' }).click();

    await expect(page.getByText('E2E移动云盘用户').first()).toBeVisible({ timeout: 10000 });

    const accountRows = page.locator('.el-table__row');
    const firstDeleteBtn = accountRows.first().getByRole('button', { name: '删除' });
    await firstDeleteBtn.click();

    await page.getByRole('button', { name: '确定' }).click();

    // 等待 Element Plus 消息提示出现
    await expect(page.locator('.el-message').first()).toBeVisible({ timeout: 5000 });
  });

  test('取消删除账号', async ({ page }) => {
    await page.goto('/accounts');

    const firstDeleteBtn = page.locator('.el-table__row').first().getByRole('button', { name: '删除' });
    const initialRowCount = await page.locator('.el-table__row').count();

    await firstDeleteBtn.click();
    await page.getByRole('button', { name: '取消' }).click();

    await expect(page.locator('.el-table__row')).toHaveCount(initialRowCount);
  });
});
