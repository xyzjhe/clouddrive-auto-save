# UCAS 回归测试设计方案

**日期**: 2026-05-22
**版本**: 1.0
**状态**: 已确认

## 1. 背景与目标

### 1.1 背景

UCAS 项目刚刚完成了全面升级（阶段一：UI/UX 改进、阶段二：功能扩展、阶段三：架构增强），引入了插件系统、Telegram 集成、资源搜索、多渠道通知等新功能。为了确保新功能的引入不会导致现有功能的回归，需要制定完整的回归测试策略。

### 1.2 目标

1. **全面覆盖**：覆盖新增模块和现有核心模块
2. **分层测试**：建立单元测试、集成测试、API 测试、E2E 测试的完整体系
3. **核心优先**：优先确保核心功能（任务执行、转存、调度）的测试覆盖
4. **持续集成**：将测试集成到 CI/CD 流程中

## 2. 测试策略

### 2.1 分层测试结构

```
┌─────────────────────────────────────────────────────────────┐
│                    E2E 测试 (Playwright)                      │
│  用户场景、页面交互、跨页面流程                                  │
├─────────────────────────────────────────────────────────────┤
│                    API 测试 (httptest)                        │
│  请求/响应格式、状态码、错误处理                                 │
├─────────────────────────────────────────────────────────────┤
│                   集成测试 (内存DB + Mock)                     │
│  模块间交互、数据流、业务逻辑                                    │
├─────────────────────────────────────────────────────────────┤
│                   单元测试 (testing + testify)                │
│  函数逻辑、边界条件、错误处理                                    │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 各层次测试范围

#### 单元测试（目标：核心函数 100% 覆盖）

**插件系统 (`internal/core/plugin/`)**
- `manager_test.go`：插件注册、初始化、执行钩子、关闭
- 测试用例：
  - 注册重复插件应返回错误
  - 初始化不存在的插件应跳过
  - 执行钩子应按顺序调用所有支持的插件
  - 关闭应释放所有插件资源

**Telegram 集成 (`internal/core/telegram/`)**
- `bot_test.go`：机器人启动、停止、消息发送
- `handler_test.go`：命令处理、权限检查
- 测试用例：
  - 未启用时启动应返回错误
  - 未授权用户应被拒绝
  - 各命令应返回正确响应

**资源搜索 (`internal/core/search/`)**
- `client_test.go`：搜索查询、源选择、结果聚合
- `sources_test.go`：各搜索源的实现
- 测试用例：
  - 空查询应返回错误
  - 指定源应只搜索该源
  - 多源结果应正确聚合

**通知系统 (`internal/core/notify/`)**
- `manager_test.go`：渠道注册、消息调度
- `wechat_test.go`：企业微信 Webhook 推送
- `telegram_test.go`：Telegram Bot 推送
- `wxpusher_test.go`：WxPusher 推送
- 测试用例：
  - 禁用的渠道不应发送消息
  - 成功/失败通知应按配置发送
  - 测试消息应正确格式化

#### 集成测试（目标：核心流程 100% 覆盖）

**任务执行流程**
- 测试场景：解析分享 → 去重 → 过滤 → 转存 → 通知
- 测试文件：`internal/core/worker/worker_test.go`（已有，需扩展）
- 新增测试：
  - 插件钩子应在任务前后正确调用
  - 多渠道通知应在任务完成后发送

**插件生命周期**
- 测试场景：注册 → 初始化 → 执行 → 关闭
- 测试文件：`internal/core/plugin/integration_test.go`（新增）
- 测试用例：
  - 完整生命周期应无错误
  - 中途失败应正确清理资源

**资源搜索流程**
- 测试场景：查询 → 多源搜索 → 结果聚合 → 创建任务
- 测试文件：`internal/core/search/integration_test.go`（新增）
- 测试用例：
  - 搜索结果应正确转换为任务参数
  - 无效结果应被过滤

#### API 测试（目标：所有端点 100% 覆盖）

**插件管理 API**
- 测试文件：`internal/api/plugin_test.go`（新增）
- 测试用例：
  - `GET /api/plugins`：应返回插件列表
  - `GET /api/plugins/:name`：应返回插件详情
  - `PUT /api/plugins/:name/config`：应更新配置

**通知配置 API**
- 测试文件：`internal/api/notify_test.go`（新增）
- 测试用例：
  - `GET /api/notify`：应返回通知渠道列表
  - `PUT /api/notify/:name`：应更新配置
  - `POST /api/notify/:name/test`：应发送测试消息

**资源搜索 API**
- 测试文件：`internal/api/search_test.go`（新增）
- 测试用例：
  - `GET /api/search`：应返回搜索结果
  - `GET /api/search/sources`：应返回搜索源列表

**Telegram 配置 API**
- 测试文件：`internal/api/telegram_test.go`（新增）
- 测试用例：
  - `GET /api/telegram/config`：应返回配置
  - `PUT /api/telegram/config`：应更新配置
  - `POST /api/telegram/test`：应测试连接

#### E2E 测试（目标：核心用户场景 100% 覆盖）

**插件管理页面**
- 测试文件：`e2e/tests/plugins/list.spec.ts`（新增）
- 测试用例：
  - 应正确展示插件列表
  - 应支持启用/禁用插件
  - 应支持配置插件

**资源搜索页面**
- 测试文件：`e2e/tests/search/search.spec.ts`（新增）
- 测试用例：
  - 应支持输入关键词搜索
  - 应正确展示搜索结果
  - 应支持从结果创建任务

**通知配置页面**
- 测试文件：`e2e/tests/notify/config.spec.ts`（新增）
- 测试用例：
  - 应正确展示通知渠道列表
  - 应支持配置各渠道参数
  - 应支持发送测试消息

**侧边栏导航**
- 测试文件：`e2e/tests/layout/sidebar.spec.ts`（扩展）
- 新增测试用例：
  - 应支持分类分组折叠
  - 应支持搜索功能
  - 应正确导航到新页面

## 3. 测试工具和框架

### 3.1 单元测试

```go
// 使用标准库 testing + testify
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestPluginManager_Register(t *testing.T) {
    manager := NewManager()
    plugin := &MockPlugin{name: "test"}

    err := manager.Register(plugin)
    require.NoError(t, err)

    // 注册重复插件应返回错误
    err = manager.Register(plugin)
    assert.Error(t, err)
}
```

### 3.2 集成测试

```go
// 使用内存数据库 + Mock HTTP
func TestTaskExecution_WithPlugin(t *testing.T) {
    // 初始化内存数据库
    db.InitDB("file::memory:?cache=shared")

    // 注册 Mock 插件
    plugin := &MockPlugin{hooks: []HookType{HookTaskBefore, HookTaskAfter}}
    manager.Register(plugin)

    // 执行任务
    worker.ExecuteTask(task)

    // 验证插件钩子被调用
    assert.True(t, plugin.taskBeforeCalled)
    assert.True(t, plugin.taskAfterCalled)
}
```

### 3.3 API 测试

```go
// 使用 httptest
func TestPluginAPI_ListPlugins(t *testing.T) {
    router := setupRouter()
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/api/plugins", nil)
    router.ServeHTTP(w, req)

    assert.Equal(t, 200, w.Code)

    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, 0, response["code"])
}
```

### 3.4 E2E 测试

```typescript
// 使用 Playwright
import { test, expect } from '@playwright/test';

test('插件管理页面应正确展示插件列表', async ({ page }) => {
    await page.goto('/plugins');
    await expect(page.locator('.plugin-card')).toHaveCount(3);
});
```

## 4. 测试文件组织

### 4.1 新增测试文件

```
internal/
├── core/
│   ├── plugin/
│   │   ├── manager_test.go          # 单元测试
│   │   └── integration_test.go      # 集成测试
│   ├── telegram/
│   │   ├── bot_test.go              # 单元测试
│   │   └── handler_test.go          # 单元测试
│   ├── search/
│   │   ├── client_test.go           # 单元测试
│   │   ├── sources_test.go          # 单元测试
│   │   └── integration_test.go      # 集成测试
│   └── notify/
│       ├── manager_test.go          # 单元测试
│       ├── wechat_test.go           # 单元测试
│       ├── telegram_test.go         # 单元测试
│       ├── wxpusher_test.go         # 单元测试
│       └── integration_test.go      # 集成测试
├── api/
│   ├── plugin_test.go               # API 测试
│   ├── telegram_test.go             # API 测试
│   ├── search_test.go               # API 测试
│   └── notify_test.go               # API 测试
└── db/
    └── db_test.go                   # 单元测试

e2e/tests/
├── plugins/
│   ├── list.spec.ts                 # E2E 测试
│   └── config.spec.ts               # E2E 测试
├── search/
│   ├── search.spec.ts               # E2E 测试
│   └── create-task.spec.ts          # E2E 测试
├── notify/
│   ├── config.spec.ts               # E2E 测试
│   └── test.spec.ts                 # E2E 测试
└── layout/
    └── sidebar.spec.ts              # 扩展现有测试
```

### 4.2 测试辅助文件

```
internal/
├── core/
│   ├── mock_plugin.go               # Mock 插件实现
│   └── mock_notifier.go             # Mock 通知渠道实现
└── testutil/
    └── testutil.go                  # 测试工具函数
```

## 5. 测试执行流程

### 5.1 Makefile 新增目标

```makefile
## unit-test: 运行单元测试
unit-test:
	@echo "=> Running unit tests..."
	go test -v -race -short ./...

## integration-test: 运行集成测试
integration-test:
	@echo "=> Running integration tests..."
	go test -v -race -run Integration ./...

## api-test: 运行 API 测试
api-test:
	@echo "=> Running API tests..."
	go test -v -race -run API ./internal/api/...

## regression-test: 运行完整回归测试
regression-test: unit-test api-test e2e-test
	@echo "=> All regression tests passed!"
```

### 5.2 CI/CD 集成

```yaml
# .github/workflows/regression.yml
name: Regression Tests

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  unit-test:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - run: make unit-test

  api-test:
    runs-on: ubuntu-24.04
    needs: unit-test
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - run: make api-test

  e2e-test:
    runs-on: ubuntu-24.04
    needs: api-test
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - uses: actions/setup-node@v4
        with:
          node-version: '24'
      - run: make e2e-test
```

## 6. 测试覆盖率目标

| 模块 | 单元测试 | 集成测试 | API 测试 | E2E 测试 |
|------|---------|---------|---------|---------|
| 插件系统 | 100% | 100% | 100% | 100% |
| Telegram 集成 | 100% | 80% | 100% | 100% |
| 资源搜索 | 100% | 100% | 100% | 100% |
| 通知系统 | 100% | 100% | 100% | 100% |
| 核心功能 | 90% | 100% | 100% | 100% |
| **整体目标** | **80%+** | **90%+** | **100%** | **100%** |

## 7. 验证方案

### 7.1 测试验证

```bash
# 运行所有测试
make regression-test

# 检查覆盖率
make test-html

# 验证 E2E 测试
make e2e-test
```

### 7.2 CI 验证

- 推送代码到 GitHub
- 检查 CI 流水线是否通过
- 查看测试报告和覆盖率报告

## 8. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 测试编写工作量大 | 延期 | 分阶段实施，优先核心功能 |
| 测试维护成本高 | 长期负担 | 建立测试规范，定期重构 |
| Mock 与真实行为不一致 | 测试失效 | 定期验证 Mock 准确性 |
| E2E 测试不稳定 | CI 反复失败 | 增加重试机制，优化等待策略 |

## 9. 总结

本方案采用分层测试策略，从单元测试到 E2E 测试全面覆盖 UCAS 项目的所有功能模块。通过优先测试核心功能、混合使用测试工具、集成到 CI/CD 流程，确保新功能的引入不会导致现有功能的回归。

方案分阶段实施：
1. **第一阶段**：为新增模块编写单元测试和 API 测试
2. **第二阶段**：扩展 E2E 测试覆盖新功能页面
3. **第三阶段**：补充现有模块的测试覆盖
4. **第四阶段**：优化测试基础设施和 CI/CD 集成
