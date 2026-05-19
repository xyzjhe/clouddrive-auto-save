import { test, expect } from '@playwright/test';

test.describe('任务管理：编辑测试', () => {
  test('编辑任务名称和保存路径', async ({ page }) => {
    const originalName = `E2E_编辑_原始_${Date.now()}`;
    const updatedName = `E2E_编辑_更新_${Date.now()}`;

    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();
    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).first().click();
    await page.getByLabel('任务名称').fill(originalName);
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_success');
    await page.getByLabel('保存路径').fill('/edit_test');
    await page.getByRole('button', { name: '确认并保存' }).click();

    const taskRow = page.locator('tr').filter({ hasText: originalName });
    await expect(taskRow).toBeVisible();

    await taskRow.getByRole('button', { name: '编辑' }).click();
    const dialog = page.getByRole('dialog', { name: '编辑任务' });
    await expect(dialog).toBeVisible();

    await page.getByLabel('任务名称').fill(updatedName);
    await page.getByLabel('保存路径').fill('/edit_test_updated');
    await page.getByRole('button', { name: '确认并保存' }).click();

    await expect(page.locator('tr').filter({ hasText: updatedName })).toBeVisible();
    await expect(page.locator('tr').filter({ hasText: originalName })).not.toBeVisible();
  });

  test('编辑任务切换调度模式', async ({ page }) => {
    const taskName = `E2E_调度_${Date.now()}`;

    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();
    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).first().click();
    await page.getByLabel('任务名称').fill(taskName);
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_success');
    await page.getByRole('button', { name: '确认并保存' }).click();

    const taskRow = page.locator('tr').filter({ hasText: taskName });
    await expect(taskRow).toBeVisible();

    await taskRow.getByRole('button', { name: '编辑' }).click();
    await page.getByRole('radio', { name: '自定义频率' }).click();
    await expect(page.getByLabel('自定义频率 (Cron)')).toBeVisible();
    await page.getByRole('button', { name: '确认并保存' }).click();

    await expect(taskRow.getByText('自定义')).toBeVisible();
  });

  test('子目录重置：点击提示条清除按钮重置为根目录', async ({ page }) => {
    const taskName = `E2E_重置_${Date.now()}`;

    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();
    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).first().click();
    await page.getByLabel('任务名称').fill(taskName);
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_success');

    await page.getByRole('button', { name: '浏览分享内容并选择目录' }).click();
    const browseDialog = page.getByRole('dialog', { name: '浏览分享内容' });
    await expect(browseDialog).toBeVisible();
    await browseDialog.getByText('139分享子目录').first().click();
    await browseDialog.getByRole('button', { name: '进入' }).click();
    await browseDialog.getByRole('button', { name: /选择当前目录/ }).click();

    await expect(page.getByText('当前目录：')).toBeVisible();

    await page.getByRole('button', { name: 'Close this tag' }).click();

    await expect(page.getByText('当前目录：')).not.toBeVisible();
  });
});
