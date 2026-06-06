# CLAUDE.md

本文件为 Claude Code (claude.ai/code) 在本仓库中工作提供指导。

## 项目概述

**统一云盘自动转存系统 (UCAS)** 是一个云盘自动化工具，整合了移动云盘 (139) 和夸克网盘 (Quark) 的转存能力。支持多任务并发转存、正则表达式文件过滤/重命名、双层 Cron 调度。后端 (Go/Gin) 将嵌入式 Vue 3 SPA 作为单一二进制文件提供服务。

## 常用命令

所有任务通过 `make` 管理（详见 `Makefile`）：

```bash
make dev            # 同时启动前端 (5173) 和后端 (8080)，自动清理端口占用
make dev-server     # 在 :8080 端口启动 Go 后端（DEBUG 模式，自动清理端口占用）
make dev-web        # 在 :5173 端口启动 Vue 3 开发服务器（代理 /api 到 :8080，自动清理端口占用）
make build          # 完整生产构建：前端编译 -> 嵌入 Go 二进制 -> bin/ucas
make build-web      # 仅构建前端（web/dist -> internal/api/dist）
make test           # 运行 Go 单元测试（带 -race 和覆盖率）
make lint           # gofmt 代码格式检查
make vet            # go vet ./...
make check          # 完整 CI 流水线：lint + vet + test
make e2e-setup      # 安装 Playwright + Chromium（仅首次需要）
make e2e-test       # 构建二进制，以 E2E 模式启动服务，运行 Playwright 测试
make docker-build   # 构建 Docker 镜像
make docker-up      # 通过 docker-compose 启动
make clean          # 清理 bin/、web/dist/、coverage.out
```

运行单个 Go 测试：`go test -run TestName ./internal/path/...`

## 架构

### 后端 (Go 1.25)

- **入口点**：`cmd/server/main.go` — 初始化日志 (slog)、SQLite (GORM)、工作池 (3 个 goroutine)、Cron 调度器、Gin HTTP 服务器
- **`internal/core/drive.go`** — `CloudDrive` 接口，所有驱动实现该接口。工厂模式：`RegisterDriver(platform, factory)` / `GetDriver(account)`
- **驱动注册**：通过 `init()` 副作用导入在 `internal/api/router.go` 中注册：
  - `internal/core/cloud139/client.go` — 移动云盘
  - `internal/core/quark/client.go` — 夸克网盘
- **工作池** (`internal/core/worker/`) — 带缓冲的任务队列 (容量 100)，`Submit` 为非阻塞模式（队列满时返回 error）。每个 worker 执行完整流水线：解析分享链接 → 去重检查 → 正则过滤 → 保存链接 → 重命名文件 → Bark 通知。重试使用 `select + ctx.Done()` 支持优雅关闭。致命错误分两层：驱动层 `quarkErrorCodeMap` 命中直接 `[Fatal]`；`isFatalError` 作为兜底安全网，通过子串匹配（如 `token无效`、`cookie过期`）拦截漏网的不可恢复错误
- **夸克网盘错误码**（`internal/core/quark/client.go`）— `quarkErrorCodeMap` 包级变量，命中即返回 `[Fatal]`（与 139 对齐），不可重试：
  - `41010`：文件涉及违规内容
  - `41012`：好友已取消了分享
  - `41008`：需要提取码
  - `41007` / `41009`：提取码错误
  - `24000`：提取码不正确
  - `24001`：该分享已失效
  - `20002`：账号登录已失效
- **调度器** (`internal/core/scheduler/`) — 封装 robfig/cron，支持秒级精度。"global" 模式共享一个 cron 触发所有全局任务；"custom" 模式为每个任务独立 cron。带有 `[Fatal]` 消息的任务会被自动跳过
- **重命名器** (`internal/core/renamer/`) — 支持魔法变量 `{TASKNAME}`、`{OLDNAME}`、`{CHINESE}`、`{DATE}`、`{YEAR}`、`{EXT}`，正则捕获组 `${1}`，以及 Go `text/template` 表达式
- **SSE/事件系统** (`internal/utils/`) — `Broadcaster` 发布/订阅系统，向所有 SSE 客户端广播实时日志和 `[EVENT:task_update|task_delete|stats_update|search_validate]` 结构化 JSON 事件。`DashboardLogger` 双写 slog 输出到控制台 + SSE
- **插件系统** (`internal/core/plugin/`) — 支持模块化扩展，插件有三个生命周期钩子：`task_before`、`task_after`、`run`
- **Telegram 集成** (`internal/core/telegram/`) — 支持通过 Telegram 远程管理任务，包括命令处理和消息推送
- **资源搜索** (`internal/core/search/`) — 集成 CloudSaver/PanSou 等资源搜索引擎，支持搜索后一键创建任务
- **多渠道通知** (`internal/core/notify/`) — 统一的 NotifyManager 调度器，支持企业微信、Telegram、WxPusher、Bark 四种推送渠道

### 前端 (Vue 3 + Vite)

5 个页面位于 `web/src/views/`：Dashboard（左右两栏面板 + SSE 实时日志）、Accounts（139/Quark 卡片网格账号管理与空间进度环）、Tasks（抽屉式 CRUD 任务管理 + 智能提取解析）、Settings（Tab 式集中管理：系统调度、四通道推送及扩展插件）、Search（云盘资源搜索引擎对接与跨页面联动创建）

共享工具模块 `web/src/utils/`：
- `format.js` — 统一的 `formatSize`（`parseFloat` 去尾零）、`formatTime`（相对时间）、`getStatusTagType`/`getStatusLabel` 状态映射
- `sse.js` — 统一的 SSE 连接管理（自动重连、事件解析）

### 构建标签分离

- `internal/api/fs.go`（标签 `!embed`）：从磁盘 `web/dist/` 目录提供静态文件（开发环境）
- `internal/api/fs_embed.go`（标签 `embed`）：从 `embed.FS` 提供静态文件（生产二进制）

### E2E 测试

- `E2E_TEST_MODE=true` 激活 `internal/core/mock_http.go`，替换 `http.DefaultTransport` 拦截所有云盘 API 调用并返回预设响应
- Playwright 测试用例位于 `e2e/tests/`（74 个测试），覆盖账号、仪表盘、任务、设置和布局导航模块
- Dashboard 页面有 SSE 长连接 (`/api/dashboard/logs`)，使用 `page.route` mock 数据时必须同时 mock SSE 端点，否则真实后端事件会触发 `fetchStats()` 与 mock 竞态导致 flaky
- Playwright 选择器：当多个按钮的可访问名称有包含关系时（如"选择目录" vs "浏览分享内容并选择目录"），必须使用 `getByRole('button', { name: '...', exact: true })` 避免严格模式冲突
- 账号通过 `e2e/tests/global.setup.ts` 预置并保存 storageState，各测试文件无需重复创建
- 表格中"原始文件名"和"预览文件名"列显示相同文本时，`getByText` 会匹配多个元素，需使用 `.first()` 限定
- 139 平台 mock 中文件夹 ID 是完整 path（如 `root/139_sub_dir`），不是短 ID（`139_sub_dir`）
- Element Plus `ElMessageBox.confirm` 的按钮不在 `dialog` role 内，需用 `.el-message-box` CSS 选择器定位容器，再用 `.el-button--primary` 定位确认按钮
- Element Plus `el-switch` 不是 checkbox/radio，不能用 `isChecked()`，需用 `evaluate(el => el.classList.contains('is-checked'))` 判断状态
- Element Plus `el-radio` 的 input 被 label span 遮挡，点击时用 `getByText('标签文字')` 而非 `getByRole('radio')`
- 嵌套 el-tabs 定位冲突防范：多层嵌套的选项卡面板 DOM 结构中，极易因为外层或隐藏面板中存在同名元素（如多个 `el-switch` 或 “保存” 按钮）引发 Playwright 模糊定位到不可见元素导致超时挂起。必须在前端为各自的表单和关键按钮分配专属 Class（如 `.bark-form`, `.save-bark-btn`），并在 E2E 中改用精准的类定位选择器
- `formatSize` 使用 `parseFloat` 去尾零（`”2.00”` → `”2”`），E2E 断言容量的文本时应使用 `”1 GB”` 而非 `”1.00 GB”`

### 前端注意事项

- Element Plus `el-input` 的 `#append` 插槽中放置多个按钮时，需要显式 CSS：`:deep(.el-input-group__append) { display: flex; align-items: center; }`，按钮需设置 `margin-left: 0`
- `@phosphor-icons/vue` 图标使用 `Ph` 前缀导入（如 `PhPlay`、`PhTrash`），按钮中图标需设置 `display: inline-flex; align-items: center; justify-content: center`
- 139 平台分享链接 URL 不包含目录信息（不像夸克可通过 URL 中的 pdirFID 区分），子目录需通过 `share_parent_id` 字段单独存储

## 核心约定

- **语言**：所有注释、文档、提交信息和 UI 文本必须使用**中文**
- **提交规范**：严格遵循 Angular Conventional Commits — `feat(scope): ...`、`fix(scope): ...`、`docs(scope): ...`，重点阐述 "原因 (Why)" 和 "改动 (What)"
- **纯 Go 架构**：使用 `glebarez/sqlite`（无 CGO 依赖）以确保跨平台交叉编译，无需 C 编译器
- **错误处理**：不可吞噬错误。严重异常使用 `[Fatal]` 级别日志，通过 SSE 同步至前端 UI 展示
- **API 响应格式**：统一使用扁平格式 — 成功直接返回业务数据（`c.PureJSON(200, data)`），错误返回 `gin.H{"error": "..."}`。禁止使用信封格式 `{code, data}`。前端统一通过 `request.js`（axios）调用，禁止绕过用 raw `fetch()`
- **环境变量**：`LOG_LEVEL`（DEBUG/INFO/WARN/ERROR）、`DB_PATH`（默认 `data.db`）、`LISTEN_ADDR`（默认 `0.0.0.0:8080`）

## 数据库模型 (`internal/db/db.go`)

- **Account**：平台 (139/quark)、昵称、凭证、状态、容量
- **Task**：关联账号、分享链接、提取码、保存路径、正则表达式、Cron、状态/进度、重试计数、运行星期、忽略后缀去重
- **CommonFolder**：每个账号的收藏文件夹路径
- **Setting**：全局配置的键值存储（调度、Bark 通知）

## API 路由

所有路由以 `/api` 为前缀（定义在 `internal/api/router.go`）：
- 账号管理：增删改查 + 校验 + 文件夹
- 任务管理：增删改查 + 执行 + 全部执行 + 预览 + 解析分享 + 忽略
- 仪表盘：统计信息 + SSE 日志流 + 历史日志 + 清空日志
- 设置：调度配置 + 全局设置 + 测试 Bark
- 插件管理：列表 + 详情 + 配置更新
- Telegram 配置：获取配置 + 更新配置 + 测试连接
- 资源搜索：搜索资源 + 搜索源列表 + 搜索配置（GET/PUT /api/search/config）+ 链接验证 + 批量验证（POST /api/search/validate_batch）
- 魔法匹配：预定义正则规则列表（GET /api/magic_patterns）
- 通知配置：列表 + 详情 + 更新 + 测试
