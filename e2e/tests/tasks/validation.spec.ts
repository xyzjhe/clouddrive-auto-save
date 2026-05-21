import { test, expect } from '@playwright/test';

test.describe('任务管理：表单验证测试', () => {
  test('不填写必填字段时无法提交', async ({ page }) => {
    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();

    // 不填写任何字段，直接点击确认
    await page.getByRole('button', { name: '确认并保存' }).click();

    // 验证弹窗仍在（未成功提交）
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();
    await page.keyboard.press('Escape');
  });

  test('手动输入保存路径', async ({ page }) => {
    const taskName = `E2E_手动路径_${Date.now()}`;

    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();

    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).first().click();

    await page.getByLabel('任务名称').fill(taskName);
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_success');

    // 手动输入保存路径
    await page.getByLabel('保存路径').fill('/manual/path/test');

    await page.getByRole('button', { name: '确认并保存' }).click();

    const taskRow = page.locator('tr').filter({ hasText: taskName });
    await expect(taskRow).toBeVisible({ timeout: 10000 });
  });

  test('创建任务后取消关闭弹窗', async ({ page }) => {
    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();

    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();

    // 填写部分字段
    await page.getByLabel('任务名称').fill('应该被取消的任务');

    // 点击取消
    await page.getByRole('button', { name: '取消' }).click();
    await expect(dialog).not.toBeVisible();

    // 验证任务未被创建
    await expect(page.getByText('应该被取消的任务')).not.toBeVisible();
  });
});
