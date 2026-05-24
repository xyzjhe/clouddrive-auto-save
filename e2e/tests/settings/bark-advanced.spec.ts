import { test, expect } from '@playwright/test';

test.describe('系统设置：Bark 高级配置测试', () => {
  test('展开 Bark 高级设置面板', async ({ page }) => {
    await page.goto('/notify');
    await page.getByRole('tab', { name: 'Bark' }).click();
    await page.waitForLoadState('networkidle');

    const barkPane = page.locator('.el-tab-pane').filter({ hasText: 'Device Key' });

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
    await page.goto('/notify');
    await page.getByRole('tab', { name: 'Bark' }).click();
    await page.waitForLoadState('networkidle');

    const barkPane = page.locator('.el-tab-pane').filter({ hasText: 'Device Key' });

    // 启用 Bark
    const barkSwitch = barkPane.locator('.el-switch').first();
    if (!(await barkSwitch.evaluate(el => el.classList.contains('is-checked')))) {
      await barkSwitch.click();
    }

    await page.getByLabel('服务器地址').fill('https://api.day.app');
    await page.getByLabel('Device Key').fill('mock_device_key');
    await barkPane.getByRole('button', { name: '保存' }).click();
    await expect(page.getByText('Bark 推送设置已保存')).toBeVisible({ timeout: 5000 });

    // 打开测试弹窗
    await barkPane.getByRole('button', { name: '测试' }).click();
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
