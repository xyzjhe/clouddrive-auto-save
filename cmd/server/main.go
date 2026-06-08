package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/zcq/clouddrive-auto-save/internal/api"
	"github.com/zcq/clouddrive-auto-save/internal/core/notify"
	"github.com/zcq/clouddrive-auto-save/internal/core/plugin"
	"github.com/zcq/clouddrive-auto-save/internal/core/scheduler"
	"github.com/zcq/clouddrive-auto-save/internal/core/search"
	"github.com/zcq/clouddrive-auto-save/internal/core/telegram"
	"github.com/zcq/clouddrive-auto-save/internal/core/worker"
	"github.com/zcq/clouddrive-auto-save/internal/crypto"
	"github.com/zcq/clouddrive-auto-save/internal/db"
	"github.com/zcq/clouddrive-auto-save/internal/utils"
)

// 版本信息，构建时通过 -ldflags 注入
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	// 0. 初始化日志系统
	logLevelStr := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	minLevel := slog.LevelInfo
	switch logLevelStr {
	case "DEBUG":
		minLevel = slog.LevelDebug
	case "WARN":
		minLevel = slog.LevelWarn
	case "ERROR":
		minLevel = slog.LevelError
	}
	utils.InitLogger(minLevel, os.Stdout)
	slog.Info("UCAS starting", "version", version, "commit", commit, "date", date, "logLevel", minLevel.String())

	// 0.5 初始化凭据加密模块
	if err := crypto.Init(os.Getenv("UCAS_SECRET_KEY")); err != nil {
		slog.Error("初始化凭据加密失败", "error", err)
		os.Exit(1)
	}

	// 1. 初始化数据库
	dbPath := os.Getenv("DB_PATH")
	isE2E := os.Getenv("E2E_TEST_MODE") == "true"
	if isE2E {
		dbPath = "file::memory:?cache=shared"
		slog.Info("Running in E2E TEST MODE (using memory database)")
		// 开启 HTTP 层 Mock 拦截，让系统走真实驱动逻辑进行 JSON 解析测试
		setupE2EMock()
	} else if dbPath == "" {
		dbPath = "data.db"
	}

	slog.Info("Initializing database...", "path", dbPath)
	if err := db.InitDB(dbPath); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}

	// 1.5 清理异常中断的任务（重置卡在 running 状态的任务）
	db.DB.Model(&db.Task{}).Where("status = ?", "running").Updates(map[string]interface{}{
		"status":  "pending",
		"message": "服务重启，已重置执行状态",
	})

	// 1.6 凭据加密迁移：将明文凭据自动加密
	if crypto.Enabled() {
		var accounts []db.Account
		db.DB.Find(&accounts)
		migrated := 0
		for i := range accounts {
			if accounts[i].Cookie != "" && !crypto.IsEncrypted(accounts[i].Cookie) {
				accounts[i].Cookie = crypto.Encrypt(accounts[i].Cookie)
				accounts[i].AuthToken = crypto.Encrypt(accounts[i].AuthToken)
				db.DB.Save(&accounts[i])
				migrated++
			}
		}
		if migrated > 0 {
			slog.Info("凭据加密迁移完成", "count", migrated)
		}
	}

	// 2. 启动任务管理器（可通过环境变量配置并发数和队列容量）
	numWorkers := 3
	if v := os.Getenv("WORKER_COUNT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 32 {
			numWorkers = n
		}
	}
	queueSize := 100
	if v := os.Getenv("WORKER_QUEUE_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 10000 {
			queueSize = n
		}
	}
	slog.Info("Starting worker manager...", "workers", numWorkers, "queue_size", queueSize)
	wm := worker.NewManager(numWorkers, queueSize, db.DB)
	wm.Start()
	defer wm.Stop()
	defer utils.GlobalBroadcaster.Shutdown()

	// 2.5 启动定时调度器
	scheduler.Init(wm)
	scheduler.Global.Start()
	defer scheduler.Global.Stop()

	// 加载全局调度设置
	var enabledSetting db.Setting
	var cronSetting db.Setting
	db.DB.Where("key = ?", "global_schedule_enabled").First(&enabledSetting)
	db.DB.Where("key = ?", "global_schedule_cron").First(&cronSetting)

	// 设置默认值：每天凌晨
	if cronSetting.Value == "" {
		cronSetting.Key = "global_schedule_cron"
		cronSetting.Value = "0 0 0 * * *"
		db.DB.Save(&cronSetting)
		slog.Info("Initialized default global schedule cron: 0 0 0 * * *")
	}

	scheduler.Global.UpdateGlobalSchedule(cronSetting.Value, enabledSetting.Value == "true")

	// 加载所有任务的调度
	var tasks []db.Task
	db.DB.Find(&tasks)
	for _, t := range tasks {
		scheduler.Global.UpdateTask(t.ID, t.ScheduleMode, t.Cron)
	}

	// 3. 初始化插件管理器
	slog.Info("Initializing plugin manager...")
	pluginManager := plugin.NewManager()
	api.InitPluginHandler(pluginManager)

	// 4. 初始化 Telegram 机器人
	slog.Info("Initializing Telegram bot...")
	telegramConfig := telegram.DefaultConfig()
	var tgSetting db.Setting
	if err := db.DB.Where("key = ?", "telegram_config").First(&tgSetting).Error; err == nil {
		if err := json.Unmarshal([]byte(tgSetting.Value), telegramConfig); err != nil {
			slog.Error("反序列化 Telegram 配置失败", "error", err)
		}
	}
	telegramBot := telegram.NewBot(telegramConfig)
	telegramHandler := telegram.NewHandler(telegramBot, db.DB, wm)
	telegramBot.SetHandler(telegramHandler)
	if telegramConfig.Enabled {
		if err := telegramBot.Start(); err != nil {
			slog.Error("启动 Telegram 机器人失败", "error", err)
		}
	}
	api.InitTelegramHandler(telegramBot)

	// 4.5 Bark 配置迁移：将旧 bark_* 单独键迁移为 notify_config_bark JSON
	migrateBarkConfig()

	// 4.6 初始化全局通知管理器
	slog.Info("Initializing notify manager...")
	if err := notify.InitGlobal(db.DB); err != nil {
		slog.Error("Failed to initialize global notify manager", "error", err)
	}
	api.InitNotifyHandler(notify.Global)

	// 5. 初始化搜索客户端
	slog.Info("Initializing search client...")
	searchConfig, err := search.LoadConfig(db.DB)
	if err != nil {
		slog.Warn("加载搜索配置失败，使用空配置", "error", err)
		searchConfig = &search.SearchConfig{}
	}
	searchClient := search.NewClient(searchConfig, db.DB)
	searchClient.WarmupToken() // 后台预热 token，避免首次搜索延迟
	api.InitSearchHandler(searchClient)

	// 5.5 启动系统遥测采集（后台周期采样 CPU）
	utils.StartCPUCollector()

	// 6. 启动 API 服务
	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = "0.0.0.0:8080"
	}
	slog.Info("Starting API server", "addr", listenAddr)
	r := api.InitRouter(wm, version, commit, date)
	if err := r.Run(listenAddr); err != nil {
		slog.Error("Failed to start API server", "error", err)
		os.Exit(1)
	}
}

// migrateBarkConfig 将旧 bark_* 单独键迁移为统一的 notify_config_bark JSON
func migrateBarkConfig() {
	// 检查是否已有新配置
	var existing db.Setting
	if err := db.DB.Where("key = ?", "notify_config_bark").First(&existing).Error; err == nil {
		// 新配置已存在，跳过迁移
		return
	}

	// 读取旧配置键
	var settings []db.Setting
	db.DB.Where("key LIKE ?", "bark_%").Find(&settings)
	if len(settings) == 0 {
		return
	}

	// 构建 map 便于查找
	old := make(map[string]string)
	for _, s := range settings {
		old[s.Key] = s.Value
	}

	// 仅在 enabled=true 且 device_key 不为空时迁移有效配置
	if old["bark_enabled"] != "true" || old["bark_device_key"] == "" {
		slog.Info("Bark 旧配置未启用或缺少 device_key，跳过迁移")
		return
	}

	config := map[string]interface{}{
		"name":              "bark",
		"type":              "bark",
		"enabled":           true,
		"notify_on_success": old["bark_notify_on_success"] != "false",
		"notify_on_failure": old["bark_notify_on_failure"] != "false",
		"config": map[string]interface{}{
			"server":        old["bark_server"],
			"device_key":    old["bark_device_key"],
			"icon":          old["bark_icon"],
			"archive":       old["bark_archive"],
			"success_level": old["bark_success_level"],
			"success_sound": old["bark_success_sound"],
			"failure_level": old["bark_failure_level"],
			"failure_sound": old["bark_failure_sound"],
		},
	}

	data, err := json.Marshal(config)
	if err != nil {
		slog.Error("序列化迁移后 Bark 配置失败", "error", err)
		return
	}

	if err := db.DB.Save(&db.Setting{Key: "notify_config_bark", Value: string(data)}).Error; err != nil {
		slog.Error("保存迁移后 Bark 配置失败", "error", err)
		return
	}

	slog.Info("Bark 配置已从旧格式迁移为 notify_config_bark", "keys", len(settings))
}
