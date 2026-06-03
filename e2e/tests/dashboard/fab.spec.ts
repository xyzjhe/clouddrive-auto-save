import { test, expect } from '@playwright/test';

test.describe('控制台：底部动作栏测试', () => {
  test('控制台底部动作栏包含三个按钮且可见', async ({ page }) => {
    await page.goto('/');
    // Dashboard 有 SSE 长连接，不能用 waitForLoadState('networkidle')
    await expect(page.locator('.stat-tile').first()).toBeVisible({ timeout: 10000 });

    const actionsBar = page.locator('.console-actions-bar');
    await expect(actionsBar).toBeVisible();
    await expect(actionsBar.getByText('创建新任务')).toBeVisible();
    await expect(actionsBar.getByText('管理账号')).toBeVisible();
    await expect(actionsBar.getByText('清理结束任务')).toBeVisible();
  });

  test('点击 "管理账号" 跳转到账号页面', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('.stat-tile').first()).toBeVisible({ timeout: 10000 });

    await page.locator('.console-actions-bar').getByText('管理账号').click();
    await expect(page).toHaveURL(/\/accounts/);
  });

  test('点击 "创建新任务" 跳转到任务页面', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('.stat-tile').first()).toBeVisible({ timeout: 10000 });

    await page.locator('.console-actions-bar').getByText('创建新任务').click();
    await expect(page).toHaveURL(/\/tasks/);
  });

  test('点击 "清理结束任务" 触发清空行为', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('.stat-tile').first()).toBeVisible({ timeout: 10000 });

    await page.locator('.console-actions-bar').getByText('清理结束任务').click();
    await expect(page.getByText(/已清空|清空成功/)).toBeVisible({ timeout: 5000 });
  });
});
