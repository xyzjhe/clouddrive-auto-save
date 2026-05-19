import { test, expect } from '@playwright/test';

test.describe('仪表盘：日志管理测试', () => {
  test('查看历史日志', async ({ page }) => {
    const taskName = `E2E_日志_${Date.now()}`;
    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();
    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).first().click();
    await page.getByLabel('任务名称').fill(taskName);
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_success');
    await page.getByRole('button', { name: '确认并保存' }).click();

    const taskRow = page.locator('tr').filter({ hasText: taskName });
    await expect(taskRow).toBeVisible();
    await taskRow.getByRole('button', { name: '运行' }).click();

    await expect(taskRow.locator('.el-tag').filter({ hasText: 'SUCCESS' })).toBeVisible({ timeout: 60000 });

    await page.goto('/');
    await expect(page.getByText('实时日志流')).toBeVisible();

    const logArea = page.locator('.log-terminal, .log-content, pre').first();
    await expect(logArea).toBeVisible();
  });

  test('清空日志', async ({ page }) => {
    await page.goto('/');

    const clearBtn = page.getByRole('button', { name: '清空日志' });
    if (await clearBtn.isVisible()) {
      await clearBtn.click();
      await page.getByRole('button', { name: '确定' }).click();
      await expect(page.getByText(/已清空|清空成功/)).toBeVisible({ timeout: 5000 });
    }
  });
});
