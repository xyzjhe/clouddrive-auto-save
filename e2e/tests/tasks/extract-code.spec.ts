import { test, expect } from '@playwright/test';

test.describe('任务管理：提取码功能测试', () => {
  test('创建带提取码的任务并保存', async ({ page }) => {
    const taskName = `E2E_提取码_${Date.now()}`;

    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();

    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).first().click();

    await page.getByLabel('任务名称').fill(taskName);
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_success');
    await page.getByLabel('提取码').fill('ABCD');
    await page.getByLabel('保存路径').fill('/extract_code_test');
    await page.getByRole('button', { name: '确认并保存' }).click();

    // 验证任务创建成功
    const taskRow = page.locator('tr').filter({ hasText: taskName });
    await expect(taskRow).toBeVisible({ timeout: 10000 });

    // 编辑任务，验证提取码被保留
    await taskRow.getByRole('button', { name: '编辑' }).click();
    const dialog = page.getByRole('dialog', { name: '编辑任务' });
    await expect(dialog).toBeVisible();

    await expect(page.getByLabel('提取码')).toHaveValue('ABCD');

    // 关闭弹窗
    await page.getByRole('button', { name: '取消' }).click();
  });

  test('编辑任务时修改提取码并重置状态', async ({ page }) => {
    const taskName = `E2E_改码_${Date.now()}`;

    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();

    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).first().click();

    await page.getByLabel('任务名称').fill(taskName);
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_success');
    await page.getByRole('button', { name: '确认并保存' }).click();

    const taskRow = page.locator('tr').filter({ hasText: taskName });
    await expect(taskRow).toBeVisible();

    // 编辑：添加提取码
    await taskRow.getByRole('button', { name: '编辑' }).click();
    const dialog = page.getByRole('dialog', { name: '编辑任务' });
    await expect(dialog).toBeVisible();

    await page.getByLabel('提取码').fill('NEWCODE');
    await page.getByRole('button', { name: '确认并保存' }).click();

    await expect(dialog).not.toBeVisible({ timeout: 5000 });
  });
});
