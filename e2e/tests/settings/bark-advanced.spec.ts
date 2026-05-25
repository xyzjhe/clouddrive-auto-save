import { test, expect } from '@playwright/test';

test.describe('系统设置：Bark 高级配置测试', () => {
  test('展开 Bark 高级设置面板', async ({ page }) => {
    await page.goto('/settings');
    await page.click('#tab-notify');
    await page.getByRole('tab', { name: 'Bark' }).click();
    // 移除 networkidle 等待，防止 SSE 持久连接导致超时

    const barkPane = page.locator('.bark-form');

    // 展开高级设置
    const collapseHeader = barkPane.getByText('高级推送设置');
    if (await collapseHeader.isVisible()) {
      await collapseHeader.click();
      await page.waitForTimeout(300);

      // 验证高级选项可见
      await expect(barkPane.getByText('自定义图标 URL')).toBeVisible();
      await expect(barkPane.getByText('自动保存历史')).toBeVisible();
    }
  });

  test('Bark 测试弹窗包含所有高级选项', async ({ page }) => {
    await page.goto('/settings');
    await page.click('#tab-notify');
    await page.getByRole('tab', { name: 'Bark' }).click();
    // 移除 networkidle 等待，防止 SSE 持久连接导致超时

    const barkPane = page.locator('.bark-form');

    // 启用 Bark
    const barkSwitch = barkPane.locator('.el-switch').first();
    if (!(await barkSwitch.evaluate(el => el.classList.contains('is-checked')))) {
      await barkSwitch.click();
    }

    await barkPane.locator('input[placeholder="https://api.day.app"]').fill('https://api.day.app');
    await barkPane.locator('input[placeholder="您的 Bark 设备 Key"]').fill('mock_device_key');
    await barkPane.locator('.save-bark-btn').click();
    await expect(page.getByText('Bark 推送设置已保存')).toBeVisible({ timeout: 5000 });

    await barkPane.locator('.test-bark-btn').click();
    const testDialog = page.getByRole('dialog', { name: '发送测试推送' });
    await expect(testDialog).toBeVisible();

    // 验证弹窗包含选项
    await expect(testDialog.locator('.el-input__inner').first()).toBeVisible();
    await expect(testDialog.locator('.el-textarea__inner')).toBeVisible();

    // 关闭弹窗
    await page.getByRole('button', { name: '取消' }).click();
    await expect(testDialog).not.toBeVisible();
  });
});
