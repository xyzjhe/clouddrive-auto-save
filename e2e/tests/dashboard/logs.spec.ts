import { test, expect } from '@playwright/test';

test.describe('仪表盘：日志管理测试', () => {
  test('查看历史日志', async ({ page }) => {
    const taskName = `E2E_日志_${Date.now()}`;
    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();
    const drawer = page.locator('.el-drawer');
    await expect(drawer).toBeVisible({ timeout: 5000 });
    await page.locator('.el-drawer .el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).first().click();
    await page.getByLabel('任务名称').fill(taskName);
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_success');
    await page.getByRole('button', { name: '确认并保存' }).click();

    const taskRow = page.locator('tr').filter({ hasText: taskName });
    await expect(taskRow).toBeVisible();
    await taskRow.getByRole('button', { name: '运行' }).click();

    await expect(taskRow.locator('.el-tag').filter({ hasText: 'SUCCESS' })).toBeVisible({ timeout: 60000 });

    await page.goto('/');
    await expect(page.getByText('系统日志', { exact: true })).toBeVisible();

    const logArea = page.locator('.log-list').first();
    await expect(logArea).toBeVisible();
  });

  test('清空日志', async ({ page }) => {
    await page.goto('/');

    const clearBtn = page.getByRole('button', { name: '清理结束任务' });
    if (await clearBtn.isVisible()) {
      await clearBtn.click();
      await expect(page.getByText('日志与已完成任务已清空')).toBeVisible({ timeout: 5000 });
    }
  });
});
