import { test, expect } from '@playwright/test';

test.describe('系统设置：调度配置测试', () => {
  test('启用全局调度并设置简易定时', async ({ page }) => {
    await page.goto('/settings');
    await expect(page.getByText('系统设置')).toBeVisible();

    const scheduleCard = page.locator('.el-card').filter({ hasText: '全局定时任务' });

    const scheduleSwitch = scheduleCard.locator('.el-switch');
    if (!(await scheduleSwitch.isChecked())) {
      await scheduleSwitch.click();
    }

    await page.getByRole('radio', { name: '简易定时' }).click();

    await page.getByRole('button', { name: '凌晨' }).click();

    await scheduleCard.getByRole('button', { name: '保存配置' }).click();

    await expect(page.getByText('保存成功')).toBeVisible({ timeout: 5000 });
  });

  test('设置自定义 Cron 表达式', async ({ page }) => {
    await page.goto('/settings');

    const scheduleCard = page.locator('.el-card').filter({ hasText: '全局定时任务' });

    const scheduleSwitch = scheduleCard.locator('.el-switch');
    if (!(await scheduleSwitch.isChecked())) {
      await scheduleSwitch.click();
    }

    await page.getByRole('radio', { name: '高级 Cron' }).click();

    await page.getByLabel('全局 Cron 表达式').fill('0 0 */6 * * *');

    await scheduleCard.getByRole('button', { name: '保存配置' }).click();
    await expect(page.getByText('保存成功')).toBeVisible({ timeout: 5000 });
  });

  test('禁用全局调度', async ({ page }) => {
    await page.goto('/settings');

    const scheduleCard = page.locator('.el-card').filter({ hasText: '全局定时任务' });
    const scheduleSwitch = scheduleCard.locator('.el-switch');

    if (await scheduleSwitch.isChecked()) {
      await scheduleSwitch.click();
    }

    await scheduleCard.getByRole('button', { name: '保存配置' }).click();
    await expect(page.getByText('保存成功')).toBeVisible({ timeout: 5000 });
  });
});
