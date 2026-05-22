// e2e/tests/layout/sidebar.spec.ts
import { test, expect } from '@playwright/test';

test.describe('侧边栏导航', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('应支持分类分组折叠', async ({ page }) => {
    // 找到"工具"分组
    const toolGroup = page.locator('.nav-group-header:has-text("工具")');

    // 点击折叠
    await toolGroup.click();

    // 验证子菜单隐藏
    const toolItems = page.locator('.nav-group:has-text("工具") .nav-item');
    await expect(toolItems).not.toBeVisible();

    // 再次点击展开
    await toolGroup.click();

    // 验证子菜单显示
    await expect(toolItems).toBeVisible();
  });

  test('应支持搜索功能', async ({ page }) => {
    // 输入搜索关键词
    const searchInput = page.locator('input[placeholder="搜索功能..."]');
    await searchInput.fill('插件');

    // 验证只显示匹配的菜单项
    const menuItems = page.locator('.nav-item');
    await expect(menuItems).toHaveCount(1);
    await expect(menuItems.first()).toContainText('插件管理');
  });

  test('应正确导航到插件管理页面', async ({ page }) => {
    // 点击插件管理菜单
    const pluginMenu = page.locator('.nav-item:has-text("插件管理")');
    await pluginMenu.click();

    // 验证页面跳转
    await expect(page).toHaveURL(/.*\/plugins/);
  });

  test('应正确导航到资源搜索页面', async ({ page }) => {
    // 点击资源搜索菜单
    const searchMenu = page.locator('.nav-item:has-text("资源搜索")');
    await searchMenu.click();

    // 验证页面跳转
    await expect(page).toHaveURL(/.*\/search/);
  });

  test('应正确导航到通知配置页面', async ({ page }) => {
    // 点击通知配置菜单
    const notifyMenu = page.locator('.nav-item:has-text("消息推送")');
    await notifyMenu.click();

    // 验证页面跳转
    await expect(page).toHaveURL(/.*\/notify/);
  });
});
