// e2e/tests/search/search.spec.ts
import { test, expect } from '@playwright/test';

test.describe('资源搜索页面', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/search');
  });

  test('应支持输入关键词搜索', async ({ page }) => {
    // 输入搜索关键词
    const searchInput = page.locator('input[placeholder="搜索资源..."]');
    await searchInput.fill('测试资源');

    // 点击搜索按钮
    const searchBtn = page.locator('button:has-text("搜索")');
    await searchBtn.click();

    // 等待搜索结果
    await page.waitForSelector('.result-item', { timeout: 10000 });
  });

  test('应正确展示搜索结果', async ({ page }) => {
    // 输入搜索关键词并搜索
    const searchInput = page.locator('input[placeholder="搜索资源..."]');
    await searchInput.fill('测试');
    await page.locator('button:has-text("搜索")').click();

    // 验证结果列表
    const results = page.locator('.result-item');
    await expect(results).toHaveCount(2);
  });

  test('应支持从结果创建任务', async ({ page }) => {
    // 输入搜索关键词并搜索
    const searchInput = page.locator('input[placeholder="搜索资源..."]');
    await searchInput.fill('测试');
    await page.locator('button:has-text("搜索")').click();

    // 点击创建任务按钮
    const createBtn = page.locator('.result-item').first().locator('button:has-text("创建任务")');
    await createBtn.click();

    // 验证跳转到任务创建页面
    await expect(page).toHaveURL(/.*\/tasks/);
  });
});
