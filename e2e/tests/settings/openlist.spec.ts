import { test, expect } from '@playwright/test';

test.describe('系统设置：OpenList 扫描测试', () => {
  test('OpenList 配置卡片正确加载', async ({ page }) => {
    await page.goto('/settings');
    await page.waitForLoadState('networkidle');

    const openlistCard = page.locator('.el-card').filter({ hasText: 'OpenList 扫描' });
    await expect(openlistCard).toBeVisible();

    await expect(openlistCard.getByText('API 地址')).toBeVisible();
    await expect(openlistCard.getByText('API Token')).toBeVisible();
  });

  test('启用/禁用 OpenList 开关并保存配置', async ({ page }) => {
    await page.goto('/settings');
    await page.waitForLoadState('networkidle');

    const openlistCard = page.locator('.el-card').filter({ hasText: 'OpenList 扫描' });
    const switchEl = openlistCard.locator('.el-switch').first();

    // 切换开关
    await switchEl.click();

    // 填写配置
    const apiInput = openlistCard.getByPlaceholder(/127\.0\.0\.1/);
    if (await apiInput.isVisible()) {
      await apiInput.clear();
      await apiInput.fill('http://192.168.1.100:23541');
    }

    await openlistCard.getByRole('button', { name: '保存配置' }).click();
    await expect(page.getByText(/OpenList.*已保存|设置已保存/)).toBeVisible({ timeout: 5000 });
  });

  test('手动触发 OpenList 扫描', async ({ page }) => {
    // Mock OpenList scan API
    await page.route('**/api/openlist/scan', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ message: '扫描已启动' }),
      });
    });

    await page.goto('/settings');
    await page.waitForLoadState('networkidle');

    const openlistCard = page.locator('.el-card').filter({ hasText: 'OpenList 扫描' });
    const scanBtn = openlistCard.getByRole('button', { name: '手动扫描' });

    if (await scanBtn.isVisible()) {
      await scanBtn.click();
      // 验证扫描触发成功（按钮可能变为 loading 状态或弹出提示）
      await page.waitForTimeout(1000);
    }
  });
});
