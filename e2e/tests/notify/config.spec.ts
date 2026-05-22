// e2e/tests/notify/config.spec.ts
import { test, expect } from '@playwright/test';

test.describe('通知配置页面', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/notify');
  });

  test('应正确展示通知渠道列表', async ({ page }) => {
    // 验证标签页存在
    const tabs = page.locator('.el-tabs__item');
    await expect(tabs).toHaveCount(3); // 企业微信, Telegram, WxPusher
  });

  test('应支持配置企业微信', async ({ page }) => {
    // 切换到企业微信标签
    await page.locator('.el-tabs__item:has-text("企业微信")').click();

    // 输入 Webhook URL
    const webhookInput = page.locator('input[placeholder*="qyapi.weixin.qq.com"]');
    await webhookInput.fill('https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test');

    // 点击保存按钮
    const saveBtn = page.locator('button:has-text("保存")');
    await saveBtn.click();

    // 验证保存成功
    await expect(page.locator('.el-message--success')).toBeVisible();
  });

  test('应支持发送测试消息', async ({ page }) => {
    // 切换到企业微信标签
    await page.locator('.el-tabs__item:has-text("企业微信")').click();

    // 输入 Webhook URL
    const webhookInput = page.locator('input[placeholder*="qyapi.weixin.qq.com"]');
    await webhookInput.fill('https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test');

    // 点击测试按钮
    const testBtn = page.locator('button:has-text("测试")');
    await testBtn.click();

    // 验证测试消息发送
    await expect(page.locator('.el-message--success')).toBeVisible();
  });
});
