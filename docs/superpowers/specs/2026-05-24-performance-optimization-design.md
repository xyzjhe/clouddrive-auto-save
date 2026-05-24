# 统一云盘自动转存系统性能优化设计规格书 (Performance Spec)

本规格书旨在记录并规范在不改变项目既有业务逻辑和前端交互界面的前提下，针对统一云盘自动转存系统（UCAS）后端服务进行的一系列底层性能调优设计。

## 1. 优化目标
* **并发吞吐提升**：解决 SQLite 并发定时任务冲突时的写锁表隐患，降低频繁进度更新带来的磁盘寻道开销。
* **CPU 负载降低**：消除重命名处理器在遍历大目录文件时，因高频、重复编译正则表达式而造成的垃圾回收（GC）延迟和 CPU 周期浪费。
* **外部 API 降载**：通过优化重命名时的对比匹配机制，避免每次重命名都遍历云盘目标目录的全部旧文件，实现常数级 O(1) 的重命名匹配性能。

---

## 2. 系统设计与技术细节

### 2.1 数据库与 I/O 并发调优
在 [db.go](file:///home/zcq/Github/clouddrive-auto-save/internal/db/db.go) 中，对底层 SQLite 与 GORM 的连接参数进行配置优化，解决传统 Delete 日志模式的独占锁瓶颈：

1. **禁用 GORM 默认事务**：
   在 `db.InitDB` 中开启数据库连接时，为 `gorm.Open` 传入 `SkipDefaultTransaction: true` 配置，避免非必要的单条状态更新和进度修改在隐式事务中反复触发 fsync 写盘。
2. **激活 WAL (Write-Ahead Logging) 预写式日志**：
   在数据库成功挂载后，获取底层 `sql.DB` 驱动并执行以下配置语句：
   * `PRAGMA journal_mode=WAL;`：开启预写日志模式，实现“读不阻塞写，写不阻塞读”的并发事务。
   * `PRAGMA synchronous=NORMAL;`：在 WAL 模式下将同步盘写级别降为 `NORMAL`。在确保系统级崩溃数据不丢失的同时，显著减少因强制磁盘同步引起的寻道时间。
   * `PRAGMA busy_timeout=5000;`：配置冲突退避超时时间为 5 秒，遇到写并发锁表时协程会自动轮询排队等待，保证后台调度的平稳运行。

---

### 2.2 重命名正则预编译与 CPU 减负
在重命名引擎 [renamer.go](file:///home/zcq/Github/clouddrive-auto-save/internal/core/renamer/renamer.go) 中，通过将固定匹配表达式剥离为包级只读已编译变量，彻底消除运行期间的重复解释开销：

1. **静态全局预编译正则表**：
   定义只读全局 Map 缓存编译好的 `*regexp.Regexp` 指针，移除在 `Process` 方法循环内部的 `regexp.MustCompile` 声明：
   ```go
   var magicRegexps = map[string]*regexp.Regexp{
       "{YEAR}":    regexp.MustCompile(`\b(?:18|19|20)\d{2}\b`),
       "{DATE}":    regexp.MustCompile(`\b(?:18|19|20)?\d{2}[\.\-/年]\d{1,2}[\.\-/月]\d{1,2}\b`),
       "{CHINESE}": regexp.MustCompile(`\p{Han}{2,}`),
       "{EXT}":     regexp.MustCompile(`\.(\w+)$`),
   }
   
   var nonDigitRegexp = regexp.MustCompile(`\D`)
   ```
2. **逻辑内编译剔除**：
   * `cleanDate` 直接复用 `nonDigitRegexp` 进行替换；
   * `Process` 内查找魔法变量直接使用已编译指针进行子组提取匹配；
   * `RenameOptions` 增加对外部预编译指针（如已编译的用户 Pattern 正则）的入参承载，避免循环内编译用户的过滤表达式。

---

### 2.3 重命名扫描增量降载与防 Rate Limit
优化 [worker.go](file:///home/zcq/Github/clouddrive-auto-save/internal/core/worker/worker.go) 中的重命名处理逻辑，彻底消除大文件夹下对全部老文件做无用正则计算与 O(N) 遍历的弊端：

1. **缓存映射结构**：
   在任务准备过滤阶段，内存中直接构造 `renameMap := make(map[string]string)`（键为**原始分享文件名**，值为**计算后的预期新文件名**）。
2. **按修改时间降序的早期退出 (Early Break)**：
   由于 139 和夸克 Driver 返回的文件列表均是按修改时间 `updated_at DESC`（最新在最前）进行排序的，Worker 只需对最新拉取的一批文件根据 `renameMap` 定向比对重命名。
3. **完成退出机制**：
   ```go
   for _, tf := range newFiles {
       if expectedNewName, ok := renameMap[tf.Name]; ok {
           if expectedNewName != tf.Name {
               err := driver.RenameFile(m.ctx, tf.ID, expectedNewName)
               if err == nil {
                   delete(renameMap, tf.Name)
               }
           }
       }
       // 只要待重命名的文件列表全部处理完毕，立即 break 退出
       if len(renameMap) == 0 {
           break
       }
   }
   ```
   该机制使多次转存大目录下少数几个新文件时的重命名耗时和网盘 API 调用次数由 O(N) 直接降为 O(1)。

---

## 3. 验证与性能衡量指标

### 3.1 性能衡量 (Benchmark)
* **并发数据库写操作**：在高并发模拟（同时发起 10 个任务并发）下，不应再发生 `database is locked` 写阻塞报错。
* **重命名 CPU 吞吐量**：针对 1000 次重命名循环测试，CPU 消耗及内存分配应下降 80% 以上，GC 暂停频率显著降低。
* **重命名网络 IO 次数**：对包含 1000 个老文件的文件夹执行转存 2 个新文件并重命名，调用的重命名接口次数应且仅为 2 次，不遍历老文件。

### 3.2 回归测试
* 运行全量 `make check`，确保优化改动后，现存的所有账号转存及去重过滤单元测试维持 `PASS` 状态。
