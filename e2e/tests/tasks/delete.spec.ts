import { test, expect } from '@playwright/test';

test.describe('任务管理：删除测试', () => {
  test('删除任务：确认后任务从列表消失', async ({ page }) => {
    const taskName = `E2E_删除_${Date.now()}`;

    await page.goto('/tasks');
    // 等待页面加载完成
    await page.waitForLoadState('domcontentloaded');
    await page.waitForTimeout(2000);

    // 等待任务页面的创建任务按钮出现
    const createBtn = page.locator('.header-actions button:has-text("创建任务")');
    await expect(createBtn).toBeVisible({ timeout: 20000 });
    await createBtn.click();

    // 等待对话框打开
    const dialog = page.locator('.el-dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });

    await expect(page.getByLabel('任务名称')).toBeVisible({ timeout: 5000 });

    // 点击下拉选择框并等待选项出现
    await page.locator('.el-dialog .el-select').first().click();
    await page.waitForTimeout(1000); // 等待下拉动画完成

    // 使用 getByRole 选择器找到选项
    const option = page.getByRole('option', { name: 'E2E移动云盘用户' }).first();
    await expect(option).toBeVisible({ timeout: 10000 });
    await option.click();

    await page.getByLabel('任务名称').fill(taskName);
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_success');
    await page.getByRole('button', { name: '确认并保存' }).click();

    const taskRow = page.locator('tr').filter({ hasText: taskName });
    await expect(taskRow).toBeVisible({ timeout: 10000 });

    // 使用 CSS 选择器找到删除按钮（danger 类型的按钮）
    await taskRow.locator('.el-button--danger').click();
    const msgBox = page.locator('.el-message-box');
    await expect(msgBox).toBeVisible({ timeout: 5000 });
    await msgBox.locator('.el-button--primary').click();

    await expect(taskRow).not.toBeVisible({ timeout: 5000 });
  });

  test('取消删除：任务仍在列表中', async ({ page }) => {
    const taskName = `E2E_取消删除_${Date.now()}`;

    await page.goto('/tasks');
    // 等待页面加载完成
    await page.waitForLoadState('domcontentloaded');
    await page.waitForTimeout(2000);

    // 等待任务页面的创建任务按钮出现
    const createBtn = page.locator('.header-actions button:has-text("创建任务")');
    await expect(createBtn).toBeVisible({ timeout: 20000 });
    await createBtn.click();

    // 等待对话框打开
    const dialog = page.locator('.el-dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });

    await expect(page.getByLabel('任务名称')).toBeVisible({ timeout: 5000 });

    // 点击下拉选择框并等待选项出现
    await page.locator('.el-dialog .el-select').first().click();
    await page.waitForTimeout(1000); // 等待下拉动画完成

    // 使用 getByRole 选择器找到选项
    const option = page.getByRole('option', { name: 'E2E移动云盘用户' }).first();
    await expect(option).toBeVisible({ timeout: 10000 });
    await option.click();

    await page.getByLabel('任务名称').fill(taskName);
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_success');
    await page.getByRole('button', { name: '确认并保存' }).click();

    const taskRow = page.locator('tr').filter({ hasText: taskName });
    await expect(taskRow).toBeVisible({ timeout: 10000 });

    // 使用 CSS 选择器找到删除按钮（danger 类型的按钮）
    await taskRow.locator('.el-button--danger').click();
    const msgBox = page.locator('.el-message-box');
    await expect(msgBox).toBeVisible({ timeout: 5000 });
    await msgBox.locator('.el-button:not(.el-button--primary)').click();

    await expect(taskRow).toBeVisible();
  });
});
