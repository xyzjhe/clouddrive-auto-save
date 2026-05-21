import { test, expect } from '@playwright/test';

test.describe('系统设置：Bark 高级配置测试', () => {
  test('展开 Bark 高级设置面板', async ({ page }) => {
    await page.goto('/settings');
    await page.waitForLoadState('networkidle');

    const barkCard = page.locator('.el-card').filter({ hasText: 'Bark 消息推送' });

    // 展开高级设置
    const collapseHeader = barkCard.getByText('高级设置');
    if (await collapseHeader.isVisible()) {
      await collapseHeader.click();
      await page.waitForTimeout(300);

      // 验证高级选项可见
      await expect(barkCard.getByText('自定义图标')).toBeVisible();
      await expect(barkCard.getByText('通知级别')).toBeVisible();
      await expect(barkCard.getByText('提醒铃声')).toBeVisible();
    }
  });

  test('Bark 测试弹窗包含所有高级选项', async ({ page }) => {
    await page.goto('/settings');
    await page.waitForLoadState('networkidle');

    const barkCard = page.locator('.el-card').filter({ hasText: 'Bark 消息推送' });

    // 启用 Bark
    const barkSwitch = barkCard.locator('.el-switch').first();
    if (!(await barkSwitch.evaluate(el => el.classList.contains('is-checked')))) {
      await barkSwitch.click();
    }

    await page.getByLabel('Bark 服务器地址').fill('https://api.day.app');
    await page.getByLabel('Device Key').fill('mock_device_key');
    await barkCard.getByRole('button', { name: '保存配置' }).click();
    await expect(page.getByText('Bark 推送设置已保存')).toBeVisible({ timeout: 5000 });

    // 打开测试弹窗
    await page.getByRole('button', { name: '发送测试消息' }).click();
    const testDialog = page.getByRole('dialog', { name: '发送测试推送' });
    await expect(testDialog).toBeVisible();

    // 验证弹窗包含高级选项
    await expect(testDialog.getByLabel('推送标题')).toBeVisible();
    await expect(testDialog.getByLabel('推送内容')).toBeVisible();
    await expect(testDialog.getByText('通知级别')).toBeVisible();
    await expect(testDialog.getByText('提醒铃声')).toBeVisible();

    // 关闭弹窗
    await page.getByRole('button', { name: '取消' }).click();
    await expect(testDialog).not.toBeVisible();
  });
});
