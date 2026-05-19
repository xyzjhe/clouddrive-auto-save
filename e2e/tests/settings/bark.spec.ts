import { test, expect } from '@playwright/test';

test.describe('系统设置：Bark 通知测试', () => {
  test('配置 Bark 并保存', async ({ page }) => {
    await page.goto('/settings');

    const barkCard = page.locator('.el-card').filter({ hasText: 'Bark 消息推送' });

    const barkSwitch = barkCard.locator('.el-switch').first();
    if (!(await barkSwitch.evaluate(el => el.classList.contains('is-checked')))) {
      await barkSwitch.click();
    }

    await page.getByLabel('Bark 服务器地址').fill('https://api.day.app');
    await page.getByLabel('Device Key').fill('mock_device_key');

    await barkCard.getByRole('button', { name: '保存配置' }).click();
    await expect(page.getByText('Bark 推送设置已保存')).toBeVisible({ timeout: 5000 });
  });

  test('发送测试消息', async ({ page }) => {
    await page.goto('/settings');

    const barkCard = page.locator('.el-card').filter({ hasText: 'Bark 消息推送' });

    const barkSwitch = barkCard.locator('.el-switch').first();
    if (!(await barkSwitch.evaluate(el => el.classList.contains('is-checked')))) {
      await barkSwitch.click();
    }
    await page.getByLabel('Bark 服务器地址').fill('https://api.day.app');
    await page.getByLabel('Device Key').fill('mock_device_key');
    await barkCard.getByRole('button', { name: '保存配置' }).click();
    await expect(page.getByText('Bark 推送设置已保存')).toBeVisible({ timeout: 5000 });

    await page.getByRole('button', { name: '发送测试消息' }).click();

    const testDialog = page.getByRole('dialog', { name: '发送测试推送' });
    await expect(testDialog).toBeVisible();

    await page.getByLabel('推送标题').fill('E2E 测试推送');
    await page.getByLabel('推送内容').fill('这是一条 E2E 测试消息');

    await page.getByRole('button', { name: '立即发送' }).click();

    // 验证对话框关闭（表示发送完成）或出现成功消息
    await expect(testDialog).not.toBeVisible({ timeout: 15000 });
  });
});
