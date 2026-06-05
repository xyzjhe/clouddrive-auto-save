import { test, expect } from '@playwright/test';

test.describe('Telegram 远程管理：配置加载与展示', () => {
  test('Telegram 通知渠道 Tab 加载并展示配置表单', async ({ page }) => {
    await page.route('**/api/dashboard/logs', async route => {
      await route.fulfill({ status: 200, contentType: 'text/event-stream', body: '' });
    });

    await page.goto('/settings');
    await page.waitForLoadState('networkidle');

    // 切换到消息推送通道 Tab
    await page.click('#tab-notify');
    await page.getByRole('tab', { name: 'Telegram' }).click();

    // 验证 Telegram 通知配置表单可见
    const form = page.locator('.telegram-form');
    await expect(form).toBeVisible();
    await expect(form.getByText('启用')).toBeVisible();
    await expect(form.getByText('Bot Token')).toBeVisible();
    await expect(form.getByText('Chat ID')).toBeVisible();
    await expect(form.getByText('通知设置')).toBeVisible();
  });
});

test.describe('Telegram 远程管理：配置保存', () => {
  test('配置 Telegram 并保存', async ({ page }) => {
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

    // 启用
    const switchEl = form.locator('.el-switch').first();
    if (!(await switchEl.evaluate(el => el.classList.contains('is-checked')))) {
      await switchEl.click();
    }

    // 填写凭证
    await form.getByPlaceholder(/123456789:ABCdef/).fill('123456789:AAH_test_token_e2e');
    await form.getByPlaceholder('123456789', { exact: true }).fill('123456789');

    // 设置通知偏好
    const successCheckbox = form.getByRole('checkbox', { name: '成功通知' });
    if (!(await successCheckbox.isChecked())) {
      await successCheckbox.click();
    }
    const failureCheckbox = form.getByRole('checkbox', { name: '失败通知' });
    if (!(await failureCheckbox.isChecked())) {
      await failureCheckbox.click();
    }

    // 保存
    await form.locator('.save-telegram-btn').click();
    await expect(page.getByText(/telegram.*配置已保存|配置已更新/)).toBeVisible({ timeout: 5000 });
  });
});

test.describe('Telegram 远程管理：测试推送', () => {
  test('测试 Telegram 通知推送', async ({ page }) => {
    await page.route('**/api/dashboard/logs', async route => {
      await route.fulfill({ status: 200, contentType: 'text/event-stream', body: '' });
    });
    await page.route('**/api/notify/telegram', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ name: 'telegram', enabled: true, notify_on_success: true, notify_on_failure: true, config: { bot_token: '123456789:test', chat_id: '987654321' } }),
      });
    });
    await page.route('**/api/notify/telegram/test', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ message: '测试成功' }),
      });
    });

    await page.goto('/settings');
    await page.click('#tab-notify');
    await page.getByRole('tab', { name: 'Telegram' }).click();

    const form = page.locator('.telegram-form');
    await form.locator('.test-telegram-btn').click();

    await expect(page.getByText(/测试消息已发送|测试成功/)).toBeVisible({ timeout: 5000 });
  });
});

test.describe('Telegram 远程管理：启用/禁用', () => {
  test('禁用 Telegram 通知渠道', async ({ page }) => {
    await page.route('**/api/dashboard/logs', async route => {
      await route.fulfill({ status: 200, contentType: 'text/event-stream', body: '' });
    });
    await page.route('**/api/notify/telegram', async route => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ name: 'telegram', enabled: true, notify_on_success: true, notify_on_failure: true, config: { bot_token: 'test', chat_id: 'test' } }),
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
    const switchEl = form.locator('.el-switch').first();

    // 确保已启用，然后关闭
    if (!(await switchEl.evaluate(el => el.classList.contains('is-checked')))) {
      await switchEl.click();
    }
    await switchEl.click();

    // 验证开关已关闭
    await expect(switchEl).not.toHaveClass(/is-checked/);
  });
});
