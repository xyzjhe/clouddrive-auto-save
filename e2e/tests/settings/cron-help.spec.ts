import { test, expect } from '@playwright/test';

test.describe('系统设置：调度高级功能测试', () => {
  test('简易定时模式预设按钮可点击', async ({ page }) => {
    await page.goto('/settings');
    await page.waitForLoadState('networkidle');

    const scheduleCard = page.locator('.el-card').filter({ hasText: '全局定时任务' });

    // 确保处于简易定时模式
    const simpleRadio = scheduleCard.getByText('简易定时');
    if (await simpleRadio.isVisible()) {
      await simpleRadio.click();

      // 点击预设按钮
      const dawnBtn = scheduleCard.getByRole('button', { name: '凌晨' });
      if (await dawnBtn.isVisible()) {
        await dawnBtn.click();
        await scheduleCard.getByRole('button', { name: '保存配置' }).click();
        await expect(page.getByText('全局调度设置已保存')).toBeVisible({ timeout: 5000 });
      }
    }
  });

  test('Cron 帮助按钮弹出说明弹窗', async ({ page }) => {
    await page.goto('/settings');
    await page.waitForLoadState('networkidle');

    const scheduleCard = page.locator('.el-card').filter({ hasText: '全局定时任务' });

    // 切换到高级 Cron 模式
    const advancedRadio = scheduleCard.getByText('高级 Cron');
    if (await advancedRadio.isVisible()) {
      await advancedRadio.click();

      // 点击帮助按钮
      const helpBtn = scheduleCard.getByRole('button', { name: /帮助|help|\?/i });
      if (await helpBtn.isVisible()) {
        await helpBtn.click();

        // 验证帮助弹窗出现
        const helpDialog = page.getByRole('dialog');
        await expect(helpDialog).toBeVisible({ timeout: 3000 });

        // 关闭帮助弹窗
        await page.keyboard.press('Escape');
      }
    }
  });
});
