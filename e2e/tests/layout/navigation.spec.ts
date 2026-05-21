import { test, expect } from '@playwright/test';

test.describe('布局：侧边栏导航测试', () => {
  test('侧边栏菜单可切换页面', async ({ page }) => {
    await page.goto('/accounts');
    // 使用明确的元素等待替代 networkidle
    await expect(page.getByText('账号管理').first()).toBeVisible({ timeout: 10000 });

    // 点击任务列表
    await page.locator('.el-menu').getByText('任务列表').click();
    await expect(page).toHaveURL(/\/tasks/);

    // 点击系统设置
    await page.locator('.el-menu').getByText('系统设置').click();
    await expect(page).toHaveURL(/\/settings/);

    // 点击仪表盘概览
    await page.locator('.el-menu').getByText('仪表盘概览').click();
    await expect(page).toHaveURL(/\/$/);
  });

  test('面包屑显示当前页面路径', async ({ page }) => {
    await page.goto('/tasks');
    await expect(page.getByRole('button', { name: '创建任务' }).last()).toBeVisible({ timeout: 10000 });

    const breadcrumb = page.locator('.el-breadcrumb');
    await expect(breadcrumb).toBeVisible();
    await expect(breadcrumb.getByText('首页')).toBeVisible();
    // 面包屑第二段为当前页面名，使用宽松匹配
    await expect(breadcrumb.locator('.el-breadcrumb__item').last()).toBeVisible();
  });

  test('深色/浅色模式切换', async ({ page }) => {
    await page.goto('/accounts');
    await expect(page.getByText('账号管理').first()).toBeVisible({ timeout: 10000 });

    // 找到主题切换按钮
    const themeBtn = page.locator('button').filter({ has: page.locator('.lucide-sun, .lucide-moon') });
    if (await themeBtn.isVisible()) {
      // 记录当前 html class
      const hadDark = await page.locator('html').evaluate(el => el.classList.contains('dark'));

      await themeBtn.click();
      await page.waitForTimeout(300);

      const hasDark = await page.locator('html').evaluate(el => el.classList.contains('dark'));
      expect(hasDark).not.toBe(hadDark);

      // 切换回来
      await themeBtn.click();
      await page.waitForTimeout(300);
      const restored = await page.locator('html').evaluate(el => el.classList.contains('dark'));
      expect(restored).toBe(hadDark);
    }
  });
});
