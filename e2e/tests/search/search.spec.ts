import { test, expect } from '@playwright/test';

// 搜索页面公共 mock 数据
const MOCK_SOURCES = ['CloudSaver', 'PanSou'];
const MOCK_SEARCH_RESULTS = {
  total: 3,
  page: 1,
  search_id: 'srch_mock1234abcd',
  items: [
    {
      title: 'E2E测试电影合集',
      url: 'https://pan.quark.cn/s/mock_link_id',
      source: 'CloudSaver',
      channel: '影视资源',
      updated_at: '2024-04-20',
      size: '10 GB',
      platform: 'quark',
      tags: ['电影', '合集'],
      summary: '包含多部经典电影的合集资源'
    },
    {
      title: 'E2E测试文档资料',
      url: 'https://yun.139.com/w/#/share/link/mock_doc_id',
      source: 'PanSou',
      updated_at: '2024-04-19',
      size: '500 MB',
      platform: '139',
      tags: ['文档'],
      summary: '学习资料文档合集'
    },
    {
      title: 'E2E测试音乐包',
      url: 'https://pan.quark.cn/s/mock_music_id',
      source: 'CloudSaver',
      updated_at: '2024-04-18',
      platform: 'quark',
      tags: []
    }
  ]
};

/**
 * 拦截搜索页面依赖的所有 API 端点
 * - SSE 日志流必须 mock，否则真实后端事件会触发竞态
 * - /api/search/sources 提供搜索源列表
 * - /api/search 提供搜索结果
 */
async function mockSearchAPIs(page, options: { sources?: any; searchResults?: any } = {}) {
  // Mock SSE 日志流（防止真实后端事件触发竞态）
  await page.route('**/api/dashboard/logs', async route => {
    await route.fulfill({
      status: 200,
      contentType: 'text/event-stream',
      body: 'data: \n\n',
    });
  });

  // Mock 搜索接口（宽泛模式，注册在前 = 优先级最低）
  // Playwright 按注册逆序匹配路由，宽泛路由必须先注册，具体路由后注册
  if (options.searchResults !== null) {
    await page.route('**/api/search**', async route => {
      const url = route.request().url();
      if (!url.match(/\/api\/search\?/) && !url.endsWith('/api/search')) {
        await route.continue();
        return;
      }
      if (route.request().method() !== 'GET') {
        await route.continue();
        return;
      }
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(options.searchResults || MOCK_SEARCH_RESULTS),
      });
    });
  }

  // Mock 搜索源列表（注册在后 = 优先级最高，避免被宽泛搜索路由拦截）
  await page.route('**/api/search/sources', async route => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(options.sources || MOCK_SOURCES),
    });
  });

  // Mock 批量验证接口（注册在后 = 优先级最高）
  await page.route('**/api/search/validate_batch', async route => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ message: '验证已启动', count: 0 }),
    });
  });
}

test.describe('资源搜索：页面加载与布局', () => {
  test('搜索页面正确加载标题、搜索框和筛选区域', async ({ page }) => {
    await mockSearchAPIs(page);
    await page.goto('/search');

    // 验证页面标题与副标题
    await expect(page.getByRole('heading', { name: '资源搜索' })).toBeVisible();
    await expect(page.getByText('搜索云盘资源，一键创建转存任务')).toBeVisible();

    // 验证搜索输入框存在
    await expect(page.getByPlaceholder('搜索资源...')).toBeVisible();

    // 验证搜索按钮
    await expect(page.getByRole('button', { name: '搜索' })).toBeVisible();

    // 验证搜索源筛选区域
    await expect(page.getByText('搜索源：')).toBeVisible();
    await expect(page.getByText('CloudSaver', { exact: true })).toBeVisible();
    await expect(page.getByText('PanSou', { exact: true })).toBeVisible();

    // 验证网盘类型筛选区域
    await expect(page.getByText('网盘类型：')).toBeVisible();
    const allCheckbox = page.locator('.platform-filter').getByText('全部');
    await expect(allCheckbox).toBeVisible();
  });

  test('初始状态下无搜索结果，不显示空状态提示', async ({ page }) => {
    await mockSearchAPIs(page);
    await page.goto('/search');

    // 未搜索时不显示"未找到相关资源"
    await expect(page.getByText('未找到相关资源')).not.toBeVisible();

    // 不显示分页器
    await expect(page.locator('.pagination-wrapper')).not.toBeVisible();
  });
});

test.describe('资源搜索：搜索功能', () => {
  test('输入关键词并点击搜索按钮，展示搜索结果', async ({ page }) => {
    await mockSearchAPIs(page);
    await page.goto('/search');

    await page.getByPlaceholder('搜索资源...').fill('测试电影');
    await page.getByRole('button', { name: '搜索' }).click();

    // 验证搜索结果展示
    await expect(page.getByText('E2E测试电影合集')).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('E2E测试文档资料')).toBeVisible();
    await expect(page.getByText('E2E测试音乐包')).toBeVisible();

    // 验证分页器出现
    await expect(page.locator('.pagination-wrapper')).toBeVisible();
  });

  test('按 Enter 键触发搜索', async ({ page }) => {
    await mockSearchAPIs(page);
    await page.goto('/search');

    await page.getByPlaceholder('搜索资源...').fill('测试');
    await page.getByPlaceholder('搜索资源...').press('Enter');

    await expect(page.getByText('E2E测试电影合集')).toBeVisible({ timeout: 10000 });
  });

  test('空关键词搜索时弹出警告提示', async ({ page }) => {
    await mockSearchAPIs(page);
    await page.goto('/search');

    await page.getByRole('button', { name: '搜索' }).click();

    // Element Plus ElMessage 警告
    await expect(page.locator('.el-message').filter({ hasText: '请输入搜索关键词' })).toBeVisible({ timeout: 5000 });
  });

  test('搜索无结果时显示空状态', async ({ page }) => {
    await mockSearchAPIs(page, {
      searchResults: { total: 0, page: 1, search_id: 'srch_empty', items: [] }
    });
    await page.goto('/search');

    await page.getByPlaceholder('搜索资源...').fill('不存在的资源');
    await page.getByRole('button', { name: '搜索' }).click();

    await expect(page.getByText('未找到相关资源')).toBeVisible({ timeout: 10000 });
  });

  test('搜索失败时不崩溃，页面保持可用', async ({ page }) => {
    await mockSearchAPIs(page, { searchResults: null });
    // 搜索接口返回 500
    await page.route('**/api/search**', async route => {
      const url = route.request().url();
      if (route.request().method() === 'GET' && (url.match(/\/api\/search\?/) || url.endsWith('/api/search'))) {
        await route.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({ error: '搜索服务暂时不可用' }),
        });
      } else {
        await route.continue();
      }
    });

    await page.goto('/search');
    await page.getByPlaceholder('搜索资源...').fill('测试');
    await page.getByRole('button', { name: '搜索' }).click();

    // 页面不崩溃，搜索框仍可用
    await expect(page.getByPlaceholder('搜索资源...')).toBeVisible();
    await expect(page.getByRole('button', { name: '搜索' })).toBeVisible();
  });
});

test.describe('资源搜索：搜索结果展示', () => {
  test('结果项正确展示标题、来源、日期、大小等元信息', async ({ page }) => {
    await mockSearchAPIs(page);
    await page.goto('/search');

    await page.getByPlaceholder('搜索资源...').fill('测试');
    await page.getByRole('button', { name: '搜索' }).click();

    const firstResult = page.locator('.result-item').first();
    await expect(firstResult).toBeVisible({ timeout: 10000 });

    // 标题
    await expect(firstResult.getByText('E2E测试电影合集')).toBeVisible();
    // 来源
    await expect(firstResult.getByText('CloudSaver')).toBeVisible();
    // 频道
    await expect(firstResult.getByText('影视资源')).toBeVisible();
    // 日期
    await expect(firstResult.getByText('2024-04-20')).toBeVisible();
    // 大小
    await expect(firstResult.getByText('10 GB')).toBeVisible();
    // 标签
    await expect(firstResult.locator('.tag-item').filter({ hasText: '电影' })).toBeVisible();
    await expect(firstResult.locator('.tag-item').filter({ hasText: '合集' })).toBeVisible();
    // 摘要
    await expect(firstResult.getByText('包含多部经典电影的合集资源')).toBeVisible();
  });

  test('每条结果都有"创建任务"按钮', async ({ page }) => {
    await mockSearchAPIs(page);
    await page.goto('/search');

    await page.getByPlaceholder('搜索资源...').fill('测试');
    await page.getByRole('button', { name: '搜索' }).click();

    await expect(page.locator('.result-item').first()).toBeVisible({ timeout: 10000 });

    const createButtons = page.locator('.result-item .el-button--primary').filter({ hasText: '创建任务' });
    await expect(createButtons).toHaveCount(3);
  });

  test('结果项无标签时不渲染标签区域', async ({ page }) => {
    await mockSearchAPIs(page);
    await page.goto('/search');

    await page.getByPlaceholder('搜索资源...').fill('测试');
    await page.getByRole('button', { name: '搜索' }).click();

    await expect(page.locator('.result-item').first()).toBeVisible({ timeout: 10000 });

    // 第三条结果 tags 为空数组
    const thirdResult = page.locator('.result-item').nth(2);
    await expect(thirdResult.getByText('E2E测试音乐包')).toBeVisible();
    await expect(thirdResult.locator('.result-tags')).not.toBeVisible();
  });
});

test.describe('资源搜索：分页功能', () => {
  test('分页器正确显示总条数和分页控件', async ({ page }) => {
    await mockSearchAPIs(page);
    await page.goto('/search');

    await page.getByPlaceholder('搜索资源...').fill('测试');
    await page.getByRole('button', { name: '搜索' }).click();

    await expect(page.locator('.pagination-wrapper')).toBeVisible({ timeout: 10000 });
    // Element Plus 分页器显示总数（格式可能含空格差异）
    await expect(page.locator('.el-pagination__total')).toContainText('3');
    // 默认每页 20 条，3 条结果在第一页
    await expect(page.locator('.result-item')).toHaveCount(3);
  });

  test('修改每页条数后正确更新展示', async ({ page }) => {
    // 构造 25 条搜索结果以测试分页
    const manyItems = Array.from({ length: 25 }, (_, i) => ({
      title: `分页测试文件_${String(i + 1).padStart(2, '0')}`,
      url: `https://pan.quark.cn/s/mock_page_${i}`,
      source: 'CloudSaver',
      updated_at: '2024-04-20',
      platform: 'quark',
      tags: []
    }));

    await mockSearchAPIs(page, {
      searchResults: { total: 25, page: 1, search_id: 'srch_paginate', items: manyItems }
    });
    await page.goto('/search');

    await page.getByPlaceholder('搜索资源...').fill('分页');
    await page.getByRole('button', { name: '搜索' }).click();

    await expect(page.locator('.result-item').first()).toBeVisible({ timeout: 10000 });
    // 默认每页 20 条
    await expect(page.locator('.result-item')).toHaveCount(20);

    // 修改每页条数为 10（el-select 下拉 teleport 到 body，选项格式为 "10/page"）
    const pageSizeSelect = page.locator('.el-pagination .el-pagination__sizes');
    await pageSizeSelect.click();
    await page.getByRole('option', { name: '10/page' }).click();

    // 切换后应只显示 10 条
    await expect(page.locator('.result-item')).toHaveCount(10);
  });
});

test.describe('资源搜索：筛选功能', () => {
  test('勾选搜索源后，搜索请求包含 source 参数', async ({ page }) => {
    let capturedParams: string | null = null;

    await mockSearchAPIs(page);
    // 监听搜索请求参数（不注册额外路由，避免覆盖 sources mock）
    page.on('request', request => {
      const url = request.url();
      if (request.method() === 'GET' && url.match(/\/api\/search\?/)) {
        capturedParams = url;
      }
    });

    await page.goto('/search');

    // 勾选 CloudSaver 搜索源
    await page.locator('.source-filter').getByText('CloudSaver').click();

    await page.getByPlaceholder('搜索资源...').fill('测试');
    await page.getByRole('button', { name: '搜索' }).click();

    await expect(page.getByText('E2E测试电影合集')).toBeVisible({ timeout: 10000 });

    // 验证请求 URL 包含 source 参数
    expect(capturedParams).toContain('source=');
  });

  test('勾选具体网盘平台后，"全部"取消选中', async ({ page }) => {
    await mockSearchAPIs(page);
    await page.goto('/search');

    // "全部"默认已勾选，具体平台被禁用，需先取消"全部"
    const allCheckbox = page.locator('.platform-filter').getByText('全部');
    await allCheckbox.click();

    // 勾选夸克网盘
    await page.locator('.platform-filter').getByText('夸克网盘').click();

    // "全部"应保持未选中
    await expect(allCheckbox).not.toBeChecked();
  });

  test('勾选全部具体平台后，自动切回"全部"模式', async ({ page }) => {
    await mockSearchAPIs(page);
    await page.goto('/search');

    // 先取消"全部"以启用具体平台复选框
    const allCheckbox = page.locator('.platform-filter').getByText('全部');
    await allCheckbox.click();

    // 逐个勾选夸克和移动云盘
    await page.locator('.platform-filter').getByText('夸克网盘').click();
    await page.locator('.platform-filter').getByText('移动云盘').click();

    // 两个平台都勾选后，等效于"全部"，allPlatforms 为 true
    await expect(allCheckbox).toBeChecked();
  });

  test('选中"全部"后，具体平台复选框被禁用', async ({ page }) => {
    await mockSearchAPIs(page);
    await page.goto('/search');

    // 默认"全部"已勾选
    const allCheckbox = page.locator('.platform-filter').getByText('全部');
    await expect(allCheckbox).toBeChecked();

    // 夸克和移动云盘的复选框应被禁用
    const quarkCheckbox = page.locator('.platform-filter').getByText('夸克网盘');
    const cloud139Checkbox = page.locator('.platform-filter').getByText('移动云盘');
    await expect(quarkCheckbox).toBeDisabled();
    await expect(cloud139Checkbox).toBeDisabled();
  });
});

test.describe('资源搜索：分享内容弹窗', () => {
  test('点击搜索结果项打开分享内容弹窗', async ({ page }) => {
    await mockSearchAPIs(page);
    // Mock 分享链接解析接口（ShareContentDialog 组件调用 parseShareLink）
    await page.route('**/api/tasks/parse_share**', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([
          { id: 'file1', name: 'E2E测试电影.mp4', size: 1024, is_folder: false },
          { id: 'dir1', name: '子目录文件夹', size: 0, is_folder: true }
        ]),
      });
    });

    await page.goto('/search');
    await page.getByPlaceholder('搜索资源...').fill('测试');
    await page.getByRole('button', { name: '搜索' }).click();

    await expect(page.locator('.result-item').first()).toBeVisible({ timeout: 10000 });

    // 点击第一条结果（整个卡片可点击）
    await page.locator('.result-item').first().click();

    // 验证分享内容弹窗打开
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });
    await expect(dialog.getByText('分享内容：E2E测试电影合集')).toBeVisible();

    // 验证文件列表渲染
    await expect(dialog.getByText('E2E测试电影.mp4')).toBeVisible({ timeout: 5000 });
    await expect(dialog.getByText('子目录文件夹')).toBeVisible();
  });

  test('弹窗中点击"创建任务"跳转到任务页面', async ({ page }) => {
    await mockSearchAPIs(page);
    await page.route('**/api/tasks/parse_share**', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([
          { id: 'file1', name: 'test.mp4', size: 1024, is_folder: false }
        ]),
      });
    });

    await page.goto('/search');
    await page.getByPlaceholder('搜索资源...').fill('测试');
    await page.getByRole('button', { name: '搜索' }).click();

    await expect(page.locator('.result-item').first()).toBeVisible({ timeout: 10000 });
    await page.locator('.result-item').first().click();

    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });

    // 点击弹窗内的创建任务按钮
    await dialog.getByRole('button', { name: '创建任务' }).click();

    // 验证跳转到任务页面
    await expect(page).toHaveURL(/\/tasks/, { timeout: 5000 });
    expect(page.url()).toContain('share_url=');
  });

  test('弹窗中点击"关闭"按钮关闭弹窗', async ({ page }) => {
    await mockSearchAPIs(page);
    await page.route('**/api/tasks/parse_share**', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([
          { id: 'file1', name: 'test.mp4', size: 1024, is_folder: false }
        ]),
      });
    });

    await page.goto('/search');
    await page.getByPlaceholder('搜索资源...').fill('测试');
    await page.getByRole('button', { name: '搜索' }).click();

    await expect(page.locator('.result-item').first()).toBeVisible({ timeout: 10000 });
    await page.locator('.result-item').first().click();

    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });

    // 点击关闭按钮
    await dialog.getByRole('button', { name: '关闭' }).click();

    // 弹窗关闭
    await expect(page.getByRole('dialog')).not.toBeVisible({ timeout: 3000 });
  });
});

test.describe('资源搜索：创建任务联动', () => {
  test('点击结果项的"创建任务"按钮跳转到任务页面并携带参数', async ({ page }) => {
    await mockSearchAPIs(page);
    await page.goto('/search');

    await page.getByPlaceholder('搜索资源...').fill('测试');
    await page.getByRole('button', { name: '搜索' }).click();

    await expect(page.locator('.result-item').first()).toBeVisible({ timeout: 10000 });

    // 点击第一条结果的"创建任务"按钮（stopPropagation，不会触发卡片点击）
    const firstCreateBtn = page.locator('.result-item').first().getByRole('button', { name: '创建任务' });
    await firstCreateBtn.click();

    // 验证跳转到任务页面并携带正确的 query 参数
    await expect(page).toHaveURL(/\/tasks/, { timeout: 5000 });
    expect(page.url()).toContain('share_url=');
    expect(page.url()).toContain('platform=quark');
  });
});
