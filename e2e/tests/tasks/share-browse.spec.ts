import { test, expect } from '@playwright/test';
import { add139Account, addQuarkAccount } from '../../fixtures/account.fixture';

test.describe('分享链接子目录浏览测试', () => {
  test.beforeEach(async ({ page }) => {
    await add139Account(page);
    await addQuarkAccount(page);
  });

  test('139 平台：浏览分享内容弹窗展示根目录内容并可选择目录', async ({ page }) => {
    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();

    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).first().click();

    await page.getByLabel('任务名称').fill('E2E_139_浏览测试');
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_link_id');

    // 点击"浏览分享内容并选择目录"按钮
    await page.getByRole('button', { name: '浏览分享内容并选择目录' }).click();
    const dialog = page.getByRole('dialog', { name: '浏览分享内容' });
    await expect(dialog).toBeVisible();

    // 验证根目录内容：应显示文件夹和文件
    await expect(dialog.getByText('139分享子目录').first()).toBeVisible();
    await expect(dialog.getByText('readme.txt').first()).toBeVisible();

    // 验证底部按钮显示"选择当前目录（根目录）"
    await expect(dialog.getByRole('button', { name: /选择当前目录/ })).toBeVisible();

    // 确认选择根目录
    await dialog.getByRole('button', { name: /选择当前目录/ }).click();

    // 保存任务
    await page.getByRole('button', { name: '确认并保存' }).click();
    await expect(page.getByText('E2E_139_浏览测试')).toBeVisible({ timeout: 10000 });
  });

  test('夸克平台：浏览分享内容弹窗展示根目录内容', async ({ page }) => {
    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();

    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E夸克用户' }).first().click();

    await page.getByLabel('任务名称').fill('E2E_夸克_浏览测试');
    await page.getByLabel('分享链接').fill('https://pan.quark.cn/s/mock_link_id');

    // 点击"浏览分享内容并选择目录"按钮
    await page.getByRole('button', { name: '浏览分享内容并选择目录' }).click();
    const dialog = page.getByRole('dialog', { name: '浏览分享内容' });
    await expect(dialog).toBeVisible();

    // 验证根目录内容
    await expect(dialog.getByText('夸克分享子目录').first()).toBeVisible();
    await expect(dialog.getByText('[2024.04.20] E2E测试电影.mp4').first()).toBeVisible();
    await expect(dialog.getByText('readme.txt').first()).toBeVisible();

    // 确认选择根目录
    await dialog.getByRole('button', { name: /选择当前目录/ }).click();

    // 验证 URL 被更新为包含根目录 pdirFID（0）
    await expect(page.getByLabel('分享链接')).toHaveValue(/#\/list\/share\/0/);
  });

  test('按钮布局：两个按钮独立可点击且互不干扰', async ({ page }) => {
    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();

    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E夸克用户' }).first().click();

    await page.getByLabel('分享链接').fill('https://pan.quark.cn/s/mock_link_id');

    const browseBtn = page.getByRole('button', { name: '浏览分享内容并选择目录' });
    const openLinkBtn = page.getByRole('button', { name: '在新标签页中打开链接' });

    // 验证两个按钮都存在且可点击
    await expect(browseBtn).toBeVisible();
    await expect(browseBtn).toBeEnabled();
    await expect(openLinkBtn).toBeVisible();
    await expect(openLinkBtn).toBeEnabled();

    // 点击浏览按钮，验证弹窗打开
    await browseBtn.click();
    const dialog = page.getByRole('dialog', { name: '浏览分享内容' });
    await expect(dialog).toBeVisible();
    await dialog.getByRole('button', { name: '取消' }).click();
    await expect(dialog).not.toBeVisible();

    // 验证取消后另一个按钮仍然可点击
    await expect(openLinkBtn).toBeEnabled();
  });
});
