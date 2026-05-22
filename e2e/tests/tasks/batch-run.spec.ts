import { test, expect } from '@playwright/test';

test.describe('任务管理：批量运行测试', () => {
  test('全部运行：确认后可运行任务状态变为 running', async ({ page }) => {
    const taskName1 = `E2E_批量1_${Date.now()}`;
    const taskName2 = `E2E_批量2_${Date.now()}`;

    await page.goto('/tasks');
    // 等待页面加载完成
    await page.waitForLoadState('domcontentloaded');
    await page.waitForTimeout(2000);

    // 等待任务页面的创建任务按钮出现
    const createBtn = page.locator('.header-actions button:has-text("创建任务")');
    await expect(createBtn).toBeVisible({ timeout: 20000 });

    for (const name of [taskName1, taskName2]) {
      await createBtn.click();

      // 等待对话框打开
      const dialog = page.locator('.el-dialog');
      await expect(dialog).toBeVisible({ timeout: 5000 });

      await expect(page.getByLabel('任务名称')).toBeVisible({ timeout: 5000 });
      await page.locator('.el-dialog .el-select').first().click();
      await page.getByRole('option', { name: 'E2E移动云盘用户' }).first().click();
      await page.getByLabel('任务名称').fill(name);
      await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_success');
      await page.getByLabel('保存路径').fill('/batch_test');
      await page.getByRole('button', { name: '确认并保存' }).click();
      await expect(page.locator('tr').filter({ hasText: name })).toBeVisible({ timeout: 10000 });
    }

    await page.getByRole('button', { name: '全部运行' }).click();
    await page.getByRole('button', { name: '确认' }).click();

    for (const name of [taskName1, taskName2]) {
      const row = page.locator('tr').filter({ hasText: name });
      await expect(row.locator('.el-tag').filter({ hasText: /RUNNING|SUCCESS/ })).toBeVisible({ timeout: 120000 });
    }
  });
});
