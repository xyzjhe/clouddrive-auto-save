import { test, expect } from '@playwright/test';

test.describe('仪表盘：浮动操作按钮 (FAB) 测试', () => {
  test('FAB 下拉菜单包含三个操作项且可点击', async ({ page }) => {
    await page.goto('/');
    // Dashboard 有 SSE 长连接，不能用 waitForLoadState('networkidle')
    await expect(page.getByText('云端转存概览')).toBeVisible({ timeout: 10000 });

    const fab = page.locator('.el-dropdown').locator('button').first();
    await expect(fab).toBeVisible();
    await fab.click();

    const menu = page.locator('.el-dropdown-menu');
    await expect(menu).toBeVisible();
    await expect(menu.getByText('添加账号')).toBeVisible();
    await expect(menu.getByText('创建任务')).toBeVisible();
    await expect(menu.getByText('清空日志')).toBeVisible();
  });

  test('FAB "添加账号" 跳转到账号页面', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByText('云端转存概览')).toBeVisible({ timeout: 10000 });

    const fab = page.locator('.el-dropdown').locator('button').first();
    await fab.click();
    await page.getByText('添加账号', { exact: true }).click();

    await expect(page).toHaveURL(/\/accounts/);
  });

  test('FAB "创建任务" 跳转到任务页面', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByText('云端转存概览')).toBeVisible({ timeout: 10000 });

    const fab = page.locator('.el-dropdown').locator('button').first();
    await fab.click();
    await page.getByText('创建任务', { exact: true }).click();

    await expect(page).toHaveURL(/\/tasks/);
  });

  test('FAB "清空日志" 触发日志清空', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByText('云端转存概览')).toBeVisible({ timeout: 10000 });

    const fab = page.locator('.el-dropdown').locator('button').first();
    await fab.click();
    await page.getByText('清空日志').click();

    await expect(page.getByText(/已清空|清空成功/)).toBeVisible({ timeout: 5000 });
  });
});
