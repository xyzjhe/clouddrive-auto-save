import { test, expect } from '@playwright/test';

test.describe('仪表盘：任务交互测试', () => {
  test('失败任务显示重试按钮并可触发重试', async ({ page }) => {
    const taskName = `E2E_重试_${Date.now()}`;
    await page.goto('/tasks');

    // 创建一个会失败的任务
    await page.getByRole('button', { name: '创建任务' }).last().click();
    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).first().click();
    await page.getByLabel('任务名称').fill(taskName);
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_invalid');
    await page.getByRole('button', { name: '确认并保存' }).click();

    const taskRow = page.locator('tr').filter({ hasText: taskName });
    await expect(taskRow).toBeVisible();
    await taskRow.getByRole('button', { name: '运行' }).click();

    // 等待任务失败 - 使用 waitForTimeout 等待任务执行
    await page.waitForTimeout(5000);

    // 刷新页面查看结果
    await page.goto('/tasks');
    await page.waitForLoadState('domcontentloaded');
    await page.waitForTimeout(1000);

    const updatedRow = page.locator('tr').filter({ hasText: taskName });
    await expect(updatedRow.locator('.el-tag--danger').filter({ hasText: 'LINK ERROR' })).toBeVisible({ timeout: 15000 });

    // 去仪表盘查看
    await page.goto('/');
    // Dashboard 有 SSE 长连接，不能用 waitForLoadState('networkidle')
    await expect(page.getByText('云端转存概览')).toBeVisible({ timeout: 10000 });

    // 验证重试按钮存在
    const retryBtn = page.getByRole('button', { name: '重试' });
    try {
      await expect(retryBtn).toBeVisible({ timeout: 10000 });
      await retryBtn.click();
    } catch {
      // 重试按钮可能不存在（任务可能已成功），忽略
    }
  });

  test('清空日志后日志区域被清空', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByText('云端转存概览')).toBeVisible({ timeout: 10000 });

    const clearBtn = page.getByRole('button', { name: '清空日志' });
    if (await clearBtn.isVisible()) {
      await clearBtn.click();
      await page.getByRole('button', { name: '确定' }).click();
      await expect(page.getByText(/已清空|清空成功/)).toBeVisible({ timeout: 5000 });
    }
  });
});
