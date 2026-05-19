import { test, expect } from '@playwright/test';

test.describe('账号管理：健康检查测试', () => {
  test('校验有效账号', async ({ page }) => {
    await page.goto('/accounts');

    const firstRow = page.locator('.el-table__row').first();
    await firstRow.getByRole('button', { name: '校验' }).click();

    await expect(page.getByText(/校验成功|验证成功|正常/).first()).toBeVisible({ timeout: 15000 });
  });
});
