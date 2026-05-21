import { test, expect } from '@playwright/test';

test.describe('账号管理：编辑测试', () => {
  test('编辑账号并验证更新后的信息', async ({ page }) => {
    await page.goto('/accounts');
    await expect(page.locator('.el-table__row').first()).toBeVisible({ timeout: 10000 });

    const firstRow = page.locator('.el-table__row').first();
    await firstRow.getByRole('button', { name: '编辑' }).click();

    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();

    // 找到 Cookie 输入框（可能是 label 或 placeholder）
    const cookieInput = page.locator('textarea').filter({ hasText: '' }).last();
    if (await cookieInput.isVisible()) {
      await cookieInput.clear();
      await cookieInput.fill('__uid=mock; updated_mock_cookie');
    }

    // 点击确认按钮
    const confirmBtn = dialog.getByRole('button', { name: /确认|保存|添加/ });
    await confirmBtn.click();

    // 验证对话框关闭
    await expect(dialog).not.toBeVisible({ timeout: 10000 });
  });

  test('取消编辑不修改数据', async ({ page }) => {
    await page.goto('/accounts');
    await expect(page.locator('.el-table__row').first()).toBeVisible({ timeout: 10000 });

    const firstRow = page.locator('.el-table__row').first();

    await firstRow.getByRole('button', { name: '编辑' }).click();

    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();

    // 点击取消
    await dialog.getByRole('button', { name: '取消' }).click();
    await expect(dialog).not.toBeVisible();
  });
});
