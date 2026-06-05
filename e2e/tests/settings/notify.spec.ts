import { test, expect } from '@playwright/test';

test.describe('消息推送通道：通知渠道列表展示', () => {
  test('通知 Tab 包含四个通知子 Tab', async ({ page }) => {
    await page.goto('/settings');
    await page.click('#tab-notify');

    // 验证四个通知渠道 Tab 存在
    await expect(page.getByRole('tab', { name: '企业微信' })).toBeVisible();
    await expect(page.getByRole('tab', { name: 'Telegram' })).toBeVisible();
    await expect(page.getByRole('tab', { name: 'WxPusher' })).toBeVisible();
    await expect(page.getByRole('tab', { name: 'Bark' })).toBeVisible();
  });
});

test.describe('消息推送通道：企业微信配置', () => {
  test('配置企业微信通知并保存', async ({ page }) => {
    // Mock SSE + notify API
    await page.route('**/api/dashboard/logs', async route => {
      await route.fulfill({ status: 200, contentType: 'text/event-stream', body: '' });
    });
    await page.route('**/api/notify/wechat', async route => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ name: 'wechat', enabled: false, notify_on_success: true, notify_on_failure: true, config: {} }),
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
    await page.click('#tab-notify');
    await page.getByRole('tab', { name: '企业微信' }).click();

    const form = page.locator('.wechat-form');

    // 启用开关
    const switchEl = form.locator('.el-switch').first();
    if (!(await switchEl.evaluate(el => el.classList.contains('is-checked')))) {
      await switchEl.click();
    }

    // 填写 Webhook URL
    await form.getByPlaceholder(/qyapi\.weixin\.qq\.com/).fill('https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=e2e-test-key');

    // 保存
    await form.locator('.save-wechat-btn').click();
    await expect(page.getByText(/配置已保存|wechat/)).toBeVisible({ timeout: 5000 });
  });

  test('测试企业微信通知推送', async ({ page }) => {
    await page.route('**/api/dashboard/logs', async route => {
      await route.fulfill({ status: 200, contentType: 'text/event-stream', body: '' });
    });
    await page.route('**/api/notify/wechat', async route => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ name: 'wechat', enabled: true, notify_on_success: true, notify_on_failure: true, config: { webhook_url: 'https://qyapi.weixin.qq.com/test' } }),
        });
      } else {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ message: '配置已更新' }),
        });
      }
    });
    await page.route('**/api/notify/wechat/test', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ message: '测试成功' }),
      });
    });

    await page.goto('/settings');
    await page.click('#tab-notify');
    await page.getByRole('tab', { name: '企业微信' }).click();

    const form = page.locator('.wechat-form');
    await form.locator('.test-wechat-btn').click();

    await expect(page.getByText(/测试消息已发送|测试成功/)).toBeVisible({ timeout: 5000 });
  });
});

test.describe('消息推送通道：WxPusher 配置', () => {
  test('配置 WxPusher 通知并保存', async ({ page }) => {
    await page.route('**/api/dashboard/logs', async route => {
      await route.fulfill({ status: 200, contentType: 'text/event-stream', body: '' });
    });
    await page.route('**/api/notify/wxpusher', async route => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ name: 'wxpusher', enabled: false, notify_on_success: true, notify_on_failure: true, config: {} }),
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
    await page.click('#tab-notify');
    await page.getByRole('tab', { name: 'WxPusher' }).click();

    const form = page.locator('.wxpusher-form');

    // 启用开关
    const switchEl = form.locator('.el-switch').first();
    if (!(await switchEl.evaluate(el => el.classList.contains('is-checked')))) {
      await switchEl.click();
    }

    // 填写 App Token 和 UID
    await form.getByPlaceholder('AT_xxx').fill('AT_e2e_test_token');
    await form.getByPlaceholder('UID_xxx').fill('UID_e2e_test_uid');

    // 保存
    await form.locator('.save-wxpusher-btn').click();
    await expect(page.getByText(/配置已保存|wxpusher/)).toBeVisible({ timeout: 5000 });
  });
});

test.describe('消息推送通道：Telegram 通知配置', () => {
  test('配置 Telegram 通知渠道并保存', async ({ page }) => {
    await page.route('**/api/dashboard/logs', async route => {
      await route.fulfill({ status: 200, contentType: 'text/event-stream', body: '' });
    });
    await page.route('**/api/notify/telegram', async route => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ name: 'telegram', enabled: false, notify_on_success: true, notify_on_failure: true, config: {} }),
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
    await page.click('#tab-notify');
    await page.getByRole('tab', { name: 'Telegram' }).click();

    const form = page.locator('.telegram-form');

    // 启用开关
    const switchEl = form.locator('.el-switch').first();
    if (!(await switchEl.evaluate(el => el.classList.contains('is-checked')))) {
      await switchEl.click();
    }

    // 填写 Bot Token 和 Chat ID
    await form.getByPlaceholder(/123456789:ABCdef/).fill('123456789:E2E_TEST_TOKEN');
    await form.getByPlaceholder('123456789').fill('987654321');

    // 保存
    await form.locator('.save-telegram-btn').click();
    await expect(page.getByText(/配置已保存|telegram/)).toBeVisible({ timeout: 5000 });
  });
});
