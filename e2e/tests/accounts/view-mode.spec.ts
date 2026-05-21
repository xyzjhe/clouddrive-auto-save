import { test, expect } from '@playwright/test';

test.describe('账号管理：视图切换测试', () => {
  test('切换表格/卡片视图模式', async ({ page }) => {
    await page.goto('/accounts');
    await page.waitForLoadState('networkidle');

    // 默认应为表格视图
    const table = page.locator('.el-table');
    await expect(table).toBeVisible();

    // 切换到卡片视图
    const cardRadio = page.getByText('卡片');
    if (await cardRadio.isVisible()) {
      await cardRadio.click();
      await page.waitForTimeout(500);

      // 切换回表格视图
      const tableRadio = page.getByText('表格');
      await tableRadio.click();
      await page.waitForTimeout(500);
      await expect(table).toBeVisible();
    }
  });
});
