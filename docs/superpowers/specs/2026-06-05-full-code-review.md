# 全量代码 Review 与 Bug 检查报告

> 审查范围：后端 Go（全部核心模块）+ 前端 Vue 3（API 层 + 核心组件）
> 审查日期：2026-06-05
> 审查维度：安全性、并发/数据竞争、错误处理/健壮性、架构/代码质量

---

## 🔴 P0 — 必须立即修复（安全/数据丢失风险）

### BUG-001: `isSafeURL` SSRF 防护不完整

**文件**: `internal/api/search.go:91-111`

172.16.0.0/12 网段仅检查了 `172.` 前缀，未精确校验 172.16-172.31 范围，导致 172.32.x.x - 172.255.x.x 被误拦截，而攻击者理论上可利用 IPv6 地址或 DNS rebinding 绕过。

```go
// 当前代码
if strings.HasPrefix(host, "172.") {
    return false
}
```

**修复方案**: 使用 `net.ParseIP` 替代字符串前缀匹配，标准判定 RFC 1918 私有地址段，同时拦截 IPv6 回环地址和 IPv4-mapped IPv6。

---

### BUG-002: `updateAccount` 可覆盖任意字段（DTO 越权）

**文件**: `internal/api/router.go:278-308`

`updateAccount` 直接将 `ShouldBindJSON` 反序列化到 `db.Account` 结构体，攻击者可通过 JSON 注入覆盖 `ID`、`CreatedAt`、`UpdatedAt`、`Status`、`CapacityUsed`、`CapacityTotal` 等敏感字段。`createAccount` 存在同样问题。

```go
// 当前代码：直接绑定到 ORM 模型
if err := c.ShouldBindJSON(&account); err != nil { ... }
db.DB.Save(&account)
```

**修复方案**: 创建 `createAccountDTO` / `updateAccountDTO` 白名单结构体，仅允许 `Platform`、`AccountName`、`Cookie`、`AuthToken` 字段，在 handler 层映射到 ORM 模型后再持久化。

---

### BUG-003: `updateGlobalSettings` 前缀保护不充分

**文件**: `internal/api/router.go:691-743`

虽然已保护 `notify_config_`、`telegram_config_` 等前缀，但攻击者仍可通过此接口写入任意 key-value 到 Setting 表（如注入 `global_schedule_cron` 为恶意值、覆盖 `openlist_api_token` 等）。缺少 key 白名单校验。

**修复方案**: 定义允许通过此接口修改的 key 白名单集合，拒绝白名单外的 key。

---

### BUG-004: `math/rand` 全局随机数在 `computeMcloudSign` 中未加锁

**文件**: `internal/core/cloud139/client.go:117-119`

`rand.Intn(len(chars))` 使用的是全局 `math/rand` 源，Go 1.20+ 虽然已自动 seed，但在高并发下仍可能产生可预测的随机序列。签名中的 randomStr 应使用 `crypto/rand` 确保不可预测性，否则签名可能被伪造。`computeAnySign` 方法也存在相同问题（`client.go:231`）。

**修复方案**: 将 `math/rand.Intn` 替换为 `crypto/rand.Read` + 取模，或使用 `math/rand/v2` 的 per-goroutine Rand 实例。

---

## 🟡 P1 — 近期修复（并发/数据竞争/健壮性）

### BUG-005: `sysinfo.go` 的 `cachedCPU` 存在数据竞争

**文件**: `internal/utils/sysinfo.go:23-24, 37, 46, 62`

`cachedCPU` 是一个 `float64` 包级变量，被后台 goroutine（`StartCPUCollector`）写入、被 `GetSysInfo`（HTTP handler goroutine）读取，无任何同步保护。Go 的 `float64` 不是原子类型，这是 **真正的 data race**。

```go
var (
    cachedCPU float64   // ← 无锁/atomic 保护
    cpuOnce   sync.Once
    cpuStop   chan struct{}
)
```

**修复方案**: 使用 `atomic.Value` 或 `sync.RWMutex` 保护 `cachedCPU` 的读写。推荐 `atomic.Pointer[float64]`（Go 1.19+）或简单使用 `sync.RWMutex`。

---

### BUG-006: `Broadcaster.run()` 持锁广播存在性能风险

**文件**: `internal/utils/broadcaster.go:49-67`

`run()` 在处理 `messages` 时持有 `mu.Lock()`，并在锁内遍历所有客户端逐个发送。如果某个客户端 channel 满了（走 `default` 分支跳过），不会阻塞；但如果客户端数量极大，遍历本身在锁内完成会阻塞 register/unregister 操作。

**修复方案**: 在锁内仅复制 clients map 的快照，锁外再遍历发送。当前实现在客户端数量有限时（SSE 连接通常不多）影响不大，但属于潜在瓶颈。

---

### BUG-007: `scheduler.UpdateTask` 锁序不一致可能导致死锁

**文件**: `internal/core/scheduler/scheduler.go:108-149`

`UpdateTask` 先调用 `RemoveTask`（获取 `s.mu.Lock()`），然后在 `mode == "custom"` 分支中再次获取 `s.mu.Lock()`。虽然 Go 的 `sync.Mutex` 是不可重入的，第二次 Lock 会死锁。但由于 `RemoveTask` 使用 `defer s.mu.Unlock()` 先释放了锁，所以第二次加锁不会死锁——不过这种锁释放后立即再加锁的模式在高并发下可能导致调度任务在极短窗口被创建两次。

**修复方案**: 将 `UpdateTask` 整体重构为单次加锁操作，在锁内完成移除+添加。

---

### BUG-008: `worker.execute` 重试 goroutine 泄漏

**文件**: `internal/core/worker/worker.go:360-371`

重试逻辑通过 `go func()` 启动匿名 goroutine，但如果 Manager 已 Stop（`m.ctx.Done()` 已关闭），这个 goroutine 在 `timer.C` 到达后仍会尝试 `m.Submit`。虽然 Submit 本身不会 panic，但此 goroutine 未被 `m.wg` 跟踪，无法被优雅等待。

```go
go func() {
    timer := time.NewTimer(time.Duration(delay) * time.Second)
    defer timer.Stop()
    select {
    case <-m.ctx.Done():
        slog.Info("重试已取消（服务关闭）", "task_id", task.ID)
    case <-timer.C:
        if err := m.Submit(Job{Task: task, BatchID: job.BatchID}); err != nil {
            // ...
        }
    }
}()
```

**修复方案**: 在 Manager 中维护一个 `retryWg sync.WaitGroup`，在 `Stop()` 时等待所有重试 goroutine 完成。

---

### BUG-009: `getNewSignHash` 中 `json.Marshal` 错误被静默忽略

**文件**: `internal/core/cloud139/client.go:101`

```go
jsonBytes, _ := json.Marshal(body)
s = jsEncodeURIComponent(string(jsonBytes))
```

如果 `body` 包含无法序列化的类型（如 chan、func），`jsonBytes` 为 nil，`string(nil)` 为 `""`，签名将基于空字符串计算，导致请求被服务端拒绝。`doRequest` 中也存在相同问题（`client.go:143`）。

**修复方案**: 显式检查 `json.Marshal` 的错误并返回。

---

### BUG-010: `worker.execute` 中 `ListFiles` 错误被忽略

**文件**: `internal/core/worker/worker.go:280`

```go
newFiles, _ := driver.ListFiles(m.ctx, targetID)
```

重命名阶段重新列出文件时，错误被静默忽略。如果 `ListFiles` 失败，`newFiles` 为 nil，重命名流程会被静默跳过，用户不会收到任何提示。

**修复方案**: 检查错误并记录日志，如果失败则通过 SSE 通知用户重命名阶段出现问题。

---

### BUG-011: Quark `ParseShare` 分页未完整支持

**文件**: `internal/core/quark/client.go:581`

`_size` 参数硬编码为 `100`，如果分享链接中文件数量超过 100 个，只会获取前 100 个。139 驱动同样存在硬编码分页限制。

**修复方案**: 添加分页循环，直到获取全部文件（类似 `ListFiles` 的实现）。

---

## 🔵 P2 — 计划重构（架构/代码质量）

### ARCH-001: `router.go` 承担了过多职责

**文件**: `internal/api/router.go`（126 个符号）

`router.go` 同时包含了路由注册、所有 API handler（账号/任务/仪表盘/设置/通知/插件/Telegram/搜索/OpenList）的实现，文件过大。按 CLAUDE.md 架构描述，`notify.go`、`plugin.go`、`search.go`、`telegram.go` 已经抽取了部分 handler，但核心的账号/任务/设置 handler 仍然在 router.go 中。

**建议**: 将 `createAccount`、`updateAccount`、`deleteAccount`、`checkAccount` 等提取到 `account.go`；将 `createTask`、`updateTask`、`runTask` 等提取到 `task.go`；将 `getDashboardStats`、`dashboardLogs` 等提取到 `dashboard.go`。

---

### ARCH-002: `CloudDrive` 接口缺少 context 传播的统一超时

各驱动方法接受 `context.Context` 参数，但调用方（`worker.execute`）使用的是 `m.ctx`（Manager 级别的 context），单个 API 调用没有独立的超时控制。如果某个云盘 API 挂起，会阻塞整个 worker goroutine 直到 Manager 关闭。

**建议**: 在 `worker.execute` 中为每个阶段（ParseShare、ListFiles、SaveLink、RenameFile）创建带超时的子 context（如 60s）。

---

### ARCH-003: 错误处理模式不统一

- 部分使用 `fmt.Errorf` 直接返回原始错误
- 部分包装为 `fmt.Errorf("xxx: %w", err)`
- 部分使用 `fmt.Errorf("xxx: %v", err)`
- 日志和返回值中的错误信息有时中英混杂

**建议**: 统一使用 `%w` 错误包装以支持 `errors.Is/As` 链式判断，日志信息全中文。

---

### ARCH-004: 缺少核心模块的单元测试覆盖

`codegraph` 分析显示以下核心模块无测试覆盖：
- `CloudDrive` 接口及其工厂
- `Scheduler`（调度器）
- `Worker.Manager`（任务执行器）
- `Telegram.Bot/Handler`
- `Search.Client`
- `notify.Manager`

当前只有 `router_test.go`、`notify_test.go`、`plugin_test.go`、`task_preview_test.go`、`broadcaster_test.go` 几个测试文件。核心业务逻辑（worker 流水线、调度器、驱动）的测试完全依赖 E2E。

**建议**: 至少为 worker 的 `execute`、`finishTask`、`isFatalError` 编写单元测试，使用 `MockDriver` 替代真实驱动。

---

### ARCH-005: 前端 `AccountCard.vue` 字段映射不一致

**文件**: `web/src/components/cards/AccountCard.vue:16-17`

```javascript
const storagePercentage = computed(() => {
  if (!props.account.capacity) return 0
  return Math.round((props.account.usedSpace / props.account.capacity) * 100)
})
```

使用 `capacity` 和 `usedSpace`，但后端 `Account` 模型字段为 `capacity_total` 和 `capacity_used`。如果数据源直接来自后端，这个组件可能无法正确计算存储百分比。需要确认数据传输过程中是否有字段映射层。

---

### ARCH-006: 硬编码的 Worker 数量和队列容量

**文件**: `cmd/server/main.go:69`、`internal/core/worker/worker.go:43`

```go
wm := worker.NewManager(3, db.DB)   // 硬编码 3 个 worker
jobQueue: make(chan Job, 100),       // 硬编码 100 容量
```

无法通过配置调整，对于大量任务并发场景或低性能设备都不够灵活。

**建议**: 通过环境变量或配置文件可配置化。

---

### ARCH-007: `Broadcaster` 历史日志无上限保护

**文件**: `internal/utils/broadcaster.go:52-56`

历史日志限制为 50 条，使用切片截断实现 `b.history = b.history[1:]`。在极端高频场景下（如大批量任务同时执行），频繁的切片重组会产生 GC 压力。

**建议**: 使用环形缓冲区（ring buffer）替代切片截断。

---

## 总结

| 级别 | 数量 | 关键问题 |
|------|------|----------|
| 🔴 P0 | 4 | SSRF 防护不完整、DTO 越权、全局设置注入、签名可预测 |
| 🟡 P1 | 7 | 数据竞争、锁序、goroutine 泄漏、错误吞没、分页缺失 |
| 🔵 P2 | 7 | 文件过大、超时控制、错误包装、测试覆盖、字段映射、硬编码 |
| **合计** | **18** | |

**建议修复优先级**: P0 → P1 中 BUG-005（数据竞争）→ P1 其余 → P2
