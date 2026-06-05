import { test, expect } from '@playwright/test';

test.describe('搜索源配置：Settings 中的搜索源 Tab', () => {
  test('搜索源 Tab 加载显示 CloudSaver 和 PanSou 配置', async ({ page }) => {
    await page.route('**/api/dashboard/logs', async route => {
      await route.fulfill({ status: 200, contentType: 'text/event-stream', body: '' });
    });
    await page.route('**/api/search/config', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          cloudsaver: { server: 'http://localhost:8080', username: 'admin', password: '***', token: 'mock_token' },
          pansou: { server: 'https://so.252035.xyz' },
        }),
      });
    });

    await page.goto('/settings');
    await page.waitForLoadState('networkidle');

    // 切换到搜索源 Tab
    await page.getByRole('tab', { name: /搜索源/ }).click();

    // 验证 CloudSaver 配置卡片（等待 tab 内容渲染完成）
    await expect(page.getByText('CloudSaver 配置')).toBeVisible({ timeout: 10000 });
    // CloudSaver 和 PanSou 都有"服务地址"字段，用 .first() 限定
    await expect(page.locator('.el-form-item__label').filter({ hasText: '服务地址' }).first()).toBeVisible();
    await expect(page.locator('.el-form-item__label').filter({ hasText: '用户名' })).toBeVisible();
    await expect(page.locator('.el-form-item__label').filter({ hasText: '密码' })).toBeVisible();

    // 验证 PanSou 配置卡片
    await expect(page.getByText('PanSou 配置')).toBeVisible();
  });

  test('保存搜索源配置', async ({ page }) => {
    await page.route('**/api/dashboard/logs', async route => {
      await route.fulfill({ status: 200, contentType: 'text/event-stream', body: '' });
    });
    await page.route('**/api/search/config', async route => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            cloudsaver: { server: '', username: '', password: '', token: '' },
            pansou: { server: '' },
          }),
        });
      } else {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ message: '配置已更新' }),
        });
      }
    });

    await page.goto('/settings');
    await page.waitForLoadState('networkidle');
    await page.getByRole('tab', { name: /搜索源/ }).click();

    // 填写 CloudSaver 配置
    const csInputs = page.locator('.inner-settings-card').filter({ hasText: 'CloudSaver' });
    await csInputs.getByPlaceholder('http://localhost:8080').fill('http://192.168.1.100:8080');
    await csInputs.getByPlaceholder('用户名').fill('e2e_test_user');
    await csInputs.getByPlaceholder('密码').fill('e2e_test_pass');

    // 保存
    await page.getByRole('button', { name: '保存配置' }).click();
    await expect(page.getByText('搜索配置已保存')).toBeVisible({ timeout: 5000 });
  });
});

test.describe('系统设置：全局设置编辑保存增强', () => {
  test('编辑全局设置字段并成功保存', async ({ page }) => {
    await page.goto('/settings');

    const scheduleCard = page.locator('.el-card').filter({ hasText: '全局定时任务' });

    // 确保调度已启用
    const scheduleSwitch = scheduleCard.locator('.el-switch').first();
    if (!(await scheduleSwitch.evaluate(el => el.classList.contains('is-checked')))) {
      await scheduleSwitch.click();
    }

    // 切换到高级 Cron 模式
    await page.getByText('高级 Cron').click();

    // 输入自定义 Cron
    await page.getByLabel('全局 Cron 表达式').clear();
    await page.getByLabel('全局 Cron 表达式').fill('0 30 8 * * 1-5');

    // 保存
    await scheduleCard.getByRole('button', { name: '保存配置' }).click();
    await expect(page.getByText('全局调度设置已保存')).toBeVisible({ timeout: 5000 });
  });
});
