// e2e/tests/plugins/list.spec.ts
import { test, expect } from '@playwright/test';

test.describe('插件管理页面', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/plugins');
  });

  test('应正确展示插件列表', async ({ page }) => {
    // 等待页面加载
    await page.waitForSelector('.plugins-grid');

    // 验证插件卡片存在
    const pluginCards = page.locator('.plugin-card');
    await expect(pluginCards).toHaveCount(3); // emby, alist, add-card
  });

  test('应支持启用/禁用插件', async ({ page }) => {
    // 找到第一个插件的开关
    const switchEl = page.locator('.plugin-card').first().locator('.el-switch');

    // 点击开关
    await switchEl.click();

    // 验证状态变化
    await expect(switchEl).toHaveClass(/is-checked/);
  });

  test('应支持配置插件', async ({ page }) => {
    // 点击配置按钮
    const configBtn = page.locator('.plugin-card').first().locator('button:has-text("配置")');
    await configBtn.click();

    // 验证配置对话框打开
    await expect(page.locator('.el-dialog')).toBeVisible();
  });
});
