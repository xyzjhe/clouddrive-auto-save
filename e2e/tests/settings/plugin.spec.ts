import { test, expect } from '@playwright/test';

test.describe('功能扩展插件：列表展示', () => {
  test('插件 Tab 展示已注册插件卡片', async ({ page }) => {
    // Mock 插件列表 API
    await page.route('**/api/plugins', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([
          {
            name: '示例插件',
            version: '1.0.0',
            description: '用于 E2E 测试的 Mock 插件',
            enabled: true,
            hooks: ['task_before', 'task_after'],
          },
        ]),
      });
    });

    await page.goto('/settings');
    await page.waitForLoadState('networkidle');

    // 切换到插件 Tab
    await page.click('#tab-plugins');

    // 验证插件卡片
    await expect(page.getByText('示例插件')).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('v1.0.0')).toBeVisible();
    await expect(page.getByText('用于 E2E 测试的 Mock 插件')).toBeVisible();

    // 验证钩子标签
    await expect(page.locator('.el-tag').filter({ hasText: 'task_before' })).toBeVisible();
    await expect(page.locator('.el-tag').filter({ hasText: 'task_after' })).toBeVisible();

    // 验证操作按钮
    await expect(page.getByRole('button', { name: '配置' })).toBeVisible();
  });

  test('无插件时显示安装新插件卡片', async ({ page }) => {
    await page.route('**/api/plugins', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([]),
      });
    });

    await page.goto('/settings');
    await page.waitForLoadState('networkidle');
    await page.click('#tab-plugins');

    // 验证安装新插件卡片存在
    await expect(page.getByText('安装新插件')).toBeVisible({ timeout: 5000 });
  });

  test('点击插件配置按钮显示提示', async ({ page }) => {
    await page.route('**/api/plugins', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([
          {
            name: '测试插件',
            version: '1.0.0',
            description: '测试描述',
            enabled: true,
            hooks: ['run'],
          },
        ]),
      });
    });

    await page.goto('/settings');
    await page.waitForLoadState('networkidle');
    await page.click('#tab-plugins');

    await expect(page.getByText('测试插件')).toBeVisible({ timeout: 5000 });
    await page.getByRole('button', { name: '配置' }).click();

    // 验证提示消息（当前为功能开发中提示）
    await expect(page.getByText(/配置插件.*功能开发中|开发中/)).toBeVisible({ timeout: 3000 });
  });
});
