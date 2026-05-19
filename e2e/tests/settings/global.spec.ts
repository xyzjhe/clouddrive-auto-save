import { test, expect } from '@playwright/test';

test.describe('系统设置：页面加载测试', () => {
  test('设置页面正确加载所有配置卡片', async ({ page }) => {
    await page.goto('/settings');
    await expect(page.getByText('系统设置')).toBeVisible();

    await expect(page.getByText('全局定时任务')).toBeVisible();
    await expect(page.getByText('Bark 消息推送')).toBeVisible();
    await expect(page.getByText('OpenList 扫描')).toBeVisible();

    const saveButtons = page.getByRole('button', { name: '保存配置' });
    await expect(saveButtons).toHaveCount(3);
  });
});
