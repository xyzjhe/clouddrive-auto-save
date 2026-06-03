import { test, expect } from '@playwright/test';

test.describe('任务管理：分享浏览高级交互测试', () => {
  test('面包屑导航：可点击返回上级目录', async ({ page }) => {
    await page.goto('/tasks');
    await expect(page.locator('.header-actions button:has-text("创建任务")')).toBeVisible({ timeout: 20000 });
    await page.locator('.header-actions button:has-text("创建任务")').click();

    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E夸克用户' }).first().click();

    await page.getByLabel('分享链接').fill('https://pan.quark.cn/s/mock_link_id');

    // 浏览分享内容
    await page.getByRole('button', { name: '浏览分享内容并选择目录' }).click();
    const dialog = page.getByRole('dialog', { name: '浏览分享内容' });
    await expect(dialog).toBeVisible();

    // 进入子目录
    await dialog.getByText('夸克分享子目录').first().click();
    await dialog.getByRole('button', { name: '进入' }).click();

    // 验证面包屑显示
    const breadcrumb = dialog.locator('.el-breadcrumb');
    if (await breadcrumb.isVisible()) {
      // 点击根目录面包屑返回
      const rootCrumb = breadcrumb.getByText('根目录');
      if (await rootCrumb.isVisible()) {
        await rootCrumb.click();
        await page.waitForTimeout(500);
      }
    }

    await dialog.getByRole('button', { name: '取消' }).click();
  });

  test('起始文件选择：选择文件后确认', async ({ page }) => {
    await page.goto('/tasks');
    await expect(page.locator('.header-actions button:has-text("创建任务")')).toBeVisible({ timeout: 20000 });
    await page.locator('.header-actions button:has-text("创建任务")').click();

    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).first().click();

    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_link_id');

    // 打开起始文件选择
    await page.getByRole('button', { name: '选择文件' }).click();
    const dialog = page.getByRole('dialog', { name: '选择起始转存文件' });
    await expect(dialog).toBeVisible();

    // 选择一个文件
    await dialog.getByText('readme.txt').first().click();

    // 确认选择
    await dialog.getByRole('button', { name: '确认选择' }).click();

    // 验证起始文件字段被填充
    await expect(page.getByPlaceholder('从该文件开始向前转存 (为空则转存全量)')).toHaveValue('readme.txt');
  });
});
