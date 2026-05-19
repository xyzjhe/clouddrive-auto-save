import { test, expect } from '@playwright/test';

test.describe('系统设置：Bark 通知测试', () => {
  test('配置 Bark 并保存', async ({ page }) => {
    await page.goto('/settings');

    const barkCard = page.locator('.el-card').filter({ hasText: 'Bark 消息推送' });

    const barkSwitch = barkCard.locator('.el-switch').first();
    if (!(await barkSwitch.isChecked())) {
      await barkSwitch.click();
    }

    await page.getByLabel('Bark 服务器地址').fill('https://api.day.app');
    await page.getByLabel('Device Key').fill('mock_device_key');

    await barkCard.getByRole('button', { name: '保存配置' }).click();
    await expect(page.getByText('保存成功')).toBeVisible({ timeout: 5000 });
  });

  test('发送测试消息', async ({ page }) => {
    await page.goto('/settings');

    const barkCard = page.locator('.el-card').filter({ hasText: 'Bark 消息推送' });

    const barkSwitch = barkCard.locator('.el-switch').first();
    if (!(await barkSwitch.isChecked())) {
      await barkSwitch.click();
    }
    await page.getByLabel('Bark 服务器地址').fill('https://api.day.app');
    await page.getByLabel('Device Key').fill('mock_device_key');
    await barkCard.getByRole('button', { name: '保存配置' }).click();
    await expect(page.getByText('保存成功')).toBeVisible({ timeout: 5000 });

    await page.getByRole('button', { name: '发送测试消息' }).click();

    const testDialog = page.getByRole('dialog', { name: '发送测试推送' });
    await expect(testDialog).toBeVisible();

    await page.getByLabel('推送标题').fill('E2E 测试推送');
    await page.getByLabel('推送内容').fill('这是一条 E2E 测试消息');

    await page.getByRole('button', { name: '立即发送' }).click();

    await expect(page.getByText(/发送成功|测试消息已发送/)).toBeVisible({ timeout: 10000 });
  });
});
