import { test, expect } from '@playwright/test';

test.describe('资源搜索：页面加载与基础渲染', () => {
  test('搜索页面加载显示搜索输入框和筛选区', async ({ page }) => {
    // Mock SSE 日志流，防止持久连接干扰
    await page.route('**/api/dashboard/logs', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'text/event-stream',
        body: '',
      });
    });

    // Mock 搜索源列表
    await page.route('**/api/search/sources', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(['CloudSaver', 'PanSou']),
      });
    });

    await page.goto('/search');
    await page.waitForLoadState('networkidle');

    // 验证页面标题
    await expect(page.getByRole('heading', { name: '资源搜索' })).toBeVisible();
    await expect(page.getByText('搜索云盘资源，一键创建转存任务')).toBeVisible();

    // 验证搜索输入框
    await expect(page.getByPlaceholder('搜索资源...')).toBeVisible();

    // 验证筛选区
    await expect(page.getByText('搜索源：')).toBeVisible();
    await expect(page.getByText('网盘类型：')).toBeVisible();
  });
});

test.describe('资源搜索：核心搜索流程', () => {
  test('输入关键词搜索并展示结果列表', async ({ page }) => {
    await page.route('**/api/dashboard/logs', async route => {
      await route.fulfill({ status: 200, contentType: 'text/event-stream', body: '' });
    });
    await page.route('**/api/search/sources', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(['CloudSaver']),
      });
    });

    // Mock 搜索结果
    await page.route('**/api/search?q=**', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          total: 2,
          page: 1,
          search_id: 'srch_mock1234',
          items: [
            { title: '测试电影合集', url: 'https://pan.quark.cn/s/mock123', source: 'CloudSaver', platform: 'quark', updated_at: '2024-04-20', size: '10 GB' },
            { title: '测试文档', url: 'https://caiyun.139.com/m/i/mock456', source: 'CloudSaver', platform: '139', updated_at: '2024-04-19', tags: ['文档', 'PDF'] },
          ],
        }),
      });
    });

    // Mock 批量验证
    await page.route('**/api/search/validate_batch', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ message: '验证已启动', count: 2 }),
      });
    });

    await page.goto('/search');
    await page.waitForLoadState('networkidle');

    // 输入搜索词
    await page.getByPlaceholder('搜索资源...').fill('测试');
    await page.getByRole('button', { name: /搜索/ }).click();

    // 验证搜索结果展示
    await expect(page.getByText('测试电影合集')).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('测试文档')).toBeVisible();

    // 验证结果元信息
    await expect(page.getByText('CloudSaver').first()).toBeVisible();
    await expect(page.getByText('10 GB')).toBeVisible();

    // 验证标签
    await expect(page.getByText('文档', { exact: false }).first()).toBeVisible();

    // 验证创建任务按钮
    await expect(page.getByRole('button', { name: '创建任务' }).first()).toBeVisible();
  });

  test('空搜索结果提示未找到相关资源', async ({ page }) => {
    await page.route('**/api/dashboard/logs', async route => {
      await route.fulfill({ status: 200, contentType: 'text/event-stream', body: '' });
    });
    await page.route('**/api/search/sources', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(['CloudSaver']),
      });
    });

    // Mock 空搜索结果
    await page.route('**/api/search?q=**', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ total: 0, page: 1, search_id: '', items: [] }),
      });
    });

    await page.goto('/search');
    await page.waitForLoadState('networkidle');

    await page.getByPlaceholder('搜索资源...').fill('不存在的资源xyz');
    await page.getByRole('button', { name: /搜索/ }).click();

    // 验证空状态提示
    await expect(page.getByText('未找到相关资源')).toBeVisible({ timeout: 5000 });
  });

  test('不输入关键词搜索显示警告提示', async ({ page }) => {
    await page.route('**/api/dashboard/logs', async route => {
      await route.fulfill({ status: 200, contentType: 'text/event-stream', body: '' });
    });
    await page.route('**/api/search/sources', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(['CloudSaver']),
      });
    });

    await page.goto('/search');
    await page.waitForLoadState('networkidle');

    // 不输入关键词直接点击搜索
    await page.getByRole('button', { name: /搜索/ }).click();

    // 验证警告提示
    await expect(page.getByText('请输入搜索关键词')).toBeVisible({ timeout: 5000 });
  });
});

test.describe('资源搜索：筛选与分页', () => {
  test('搜索源筛选仅选择指定搜索源', async ({ page }) => {
    await page.route('**/api/dashboard/logs', async route => {
      await route.fulfill({ status: 200, contentType: 'text/event-stream', body: '' });
    });
    await page.route('**/api/search/sources', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(['CloudSaver', 'PanSou']),
      });
    });

    let capturedUrl = '';
    await page.route('**/api/search?q=**', async route => {
      capturedUrl = route.request().url();
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ total: 0, page: 1, search_id: '', items: [] }),
      });
    });

    await page.goto('/search');
    await page.waitForLoadState('networkidle');

    // 取消"CloudSaver"，只选"PanSou"
    const cloudSaverCheckbox = page.getByRole('checkbox', { name: 'CloudSaver' });
    if (await cloudSaverCheckbox.isChecked()) {
      await cloudSaverCheckbox.click();
    }

    // 验证 PanSou 仍然可见
    await expect(page.getByRole('checkbox', { name: 'PanSou' })).toBeVisible();

    await page.getByPlaceholder('搜索资源...').fill('测试');
    await page.getByRole('button', { name: /搜索/ }).click();
    await page.waitForTimeout(1000);

    // 验证请求包含 source 参数
    expect(capturedUrl).toContain('source=');
  });

  test('网盘类型筛选可切换平台', async ({ page }) => {
    await page.route('**/api/dashboard/logs', async route => {
      await route.fulfill({ status: 200, contentType: 'text/event-stream', body: '' });
    });
    await page.route('**/api/search/sources', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(['CloudSaver']),
      });
    });

    await page.goto('/search');
    await page.waitForLoadState('networkidle');

    // 验证默认"全部"被选中
    const allCheckbox = page.getByRole('checkbox', { name: '全部' });
    await expect(allCheckbox).toBeChecked();

    // 取消全部，选择具体平台
    await allCheckbox.click();
    await expect(page.getByRole('checkbox', { name: '夸克网盘' })).toBeEnabled();
    await expect(page.getByRole('checkbox', { name: '移动云盘' })).toBeEnabled();

    // 选择夸克网盘
    await page.getByRole('checkbox', { name: '夸克网盘' }).click();

    // 验证"全部"被取消
    await expect(allCheckbox).not.toBeChecked();
  });

  test('搜索结果分页切换', async ({ page }) => {
    await page.route('**/api/dashboard/logs', async route => {
      await route.fulfill({ status: 200, contentType: 'text/event-stream', body: '' });
    });
    await page.route('**/api/search/sources', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(['CloudSaver']),
      });
    });

    // 生成 25 条搜索结果以触发分页
    const items = Array.from({ length: 25 }, (_, i) => ({
      title: `搜索结果 ${i + 1}`,
      url: `https://pan.quark.cn/s/mock${i}`,
      source: 'CloudSaver',
      platform: 'quark',
      updated_at: '2024-04-20',
    }));

    await page.route('**/api/search?q=**', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ total: 25, page: 1, search_id: 'srch_mock', items }),
      });
    });
    await page.route('**/api/search/validate_batch', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ message: '验证已启动', count: 20 }),
      });
    });

    await page.goto('/search');
    await page.waitForLoadState('networkidle');

    await page.getByPlaceholder('搜索资源...').fill('测试');
    await page.getByRole('button', { name: /搜索/ }).click();

    // 验证分页组件出现
    await expect(page.locator('.el-pagination')).toBeVisible({ timeout: 5000 });

    // 验证总数显示
    await expect(page.getByText('共 25 条')).toBeVisible();
  });
});

test.describe('资源搜索：链接验证与任务创建', () => {
  test('搜索结果点击可打开分享内容弹窗', async ({ page }) => {
    await page.route('**/api/dashboard/logs', async route => {
      await route.fulfill({ status: 200, contentType: 'text/event-stream', body: '' });
    });
    await page.route('**/api/search/sources', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(['CloudSaver']),
      });
    });

    await page.route('**/api/search?q=**', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          total: 1,
          page: 1,
          search_id: 'srch_mock',
          items: [
            { title: '测试资源', url: 'https://pan.quark.cn/s/mockabc', source: 'CloudSaver', platform: 'quark', updated_at: '2024-04-20' },
          ],
        }),
      });
    });
    await page.route('**/api/search/validate_batch', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ message: '验证已启动', count: 1 }),
      });
    });

    await page.goto('/search');
    await page.waitForLoadState('networkidle');

    await page.getByPlaceholder('搜索资源...').fill('测试');
    await page.getByRole('button', { name: /搜索/ }).click();

    await expect(page.getByText('测试资源')).toBeVisible({ timeout: 5000 });

    // 点击搜索结果项（非按钮区域）
    const resultItem = page.locator('.result-item.clickable').first();
    await resultItem.click();

    // 验证分享内容弹窗打开
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });
  });

  test('从搜索结果点击创建任务跳转到任务页面', async ({ page }) => {
    await page.route('**/api/dashboard/logs', async route => {
      await route.fulfill({ status: 200, contentType: 'text/event-stream', body: '' });
    });
    await page.route('**/api/search/sources', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(['CloudSaver']),
      });
    });

    await page.route('**/api/search?q=**', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          total: 1,
          page: 1,
          search_id: 'srch_mock',
          items: [
            { title: '测试资源', url: 'https://pan.quark.cn/s/mockabc', source: 'CloudSaver', platform: 'quark', updated_at: '2024-04-20' },
          ],
        }),
      });
    });
    await page.route('**/api/search/validate_batch', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ message: '验证已启动', count: 1 }),
      });
    });

    await page.goto('/search');
    await page.waitForLoadState('networkidle');

    await page.getByPlaceholder('搜索资源...').fill('测试');
    await page.getByRole('button', { name: /搜索/ }).click();

    await expect(page.getByText('测试资源')).toBeVisible({ timeout: 5000 });

    // 点击创建任务按钮
    await page.getByRole('button', { name: '创建任务' }).click();

    // 验证跳转到任务页面并携带参数
    await page.waitForURL(/\/tasks/);
    const url = page.url();
    expect(url).toContain('/tasks');
    expect(url).toContain('share_url=');
    expect(url).toContain('platform=quark');
  });

  test('搜索关键词回车触发搜索', async ({ page }) => {
    await page.route('**/api/dashboard/logs', async route => {
      await route.fulfill({ status: 200, contentType: 'text/event-stream', body: '' });
    });
    await page.route('**/api/search/sources', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(['CloudSaver']),
      });
    });

    await page.route('**/api/search?q=**', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ total: 0, page: 1, search_id: '', items: [] }),
      });
    });

    await page.goto('/search');
    await page.waitForLoadState('networkidle');

    await page.getByPlaceholder('搜索资源...').fill('回车测试');
    await page.getByPlaceholder('搜索资源...').press('Enter');

    // 验证空结果提示出现（说明搜索已触发）
    await expect(page.getByText('未找到相关资源')).toBeVisible({ timeout: 5000 });
  });
});
