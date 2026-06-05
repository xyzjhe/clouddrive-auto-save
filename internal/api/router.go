package api

import (
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zcq/clouddrive-auto-save/internal/core/notify"
	"github.com/zcq/clouddrive-auto-save/internal/core/plugin"
	"github.com/zcq/clouddrive-auto-save/internal/core/search"
	"github.com/zcq/clouddrive-auto-save/internal/core/telegram"
	"github.com/zcq/clouddrive-auto-save/internal/core/worker"
)

var WorkerManager *worker.Manager

// 版本信息（由 main 包通过 InitRouter 传入）
var (
	appVersion = "dev"
	appCommit  = "unknown"
	appDate    = "unknown"
)

func InitRouter(wm *worker.Manager, version, commit, date string) *gin.Engine {
	WorkerManager = wm
	appVersion = version
	appCommit = commit
	appDate = date
	r := gin.Default()

	// 基础 API 路由组
	api := r.Group("/api")
	{
		api.GET("/version", getVersion)
		api.GET("/magic_patterns", listMagicPatterns)

		api.GET("/accounts", listAccounts)
		api.POST("/accounts", createAccount)
		api.PUT("/accounts/:id", updateAccount)
		api.DELETE("/accounts/:id", deleteAccount)
		api.POST("/accounts/:id/check", checkAccount)
		api.GET("/accounts/:id/folders", getAccountFolders)
		api.POST("/accounts/:id/folders", createAccountFolder)

		api.GET("/tasks", listTasks)
		api.POST("/tasks", createTask)
		api.PUT("/tasks/:id", updateTask)
		api.DELETE("/tasks/:id", deleteTask)
		api.POST("/tasks/:id/run", runTask)
		api.POST("/tasks/run_all", runAllTasks)
		api.POST("/tasks/:id/dismiss", dismissTaskAPI)
		api.POST("/tasks/preview", previewTask)
		api.POST("/tasks/parse_share", parseShareLinkInfo)

		api.GET("/dashboard/stats", getDashboardStats)
		api.GET("/dashboard/logs", streamLogs)
		api.GET("/dashboard/logs/recent", getRecentLogs)
		api.DELETE("/dashboard/logs/recent", clearRecentLogs)

		api.GET("/settings/schedule", getScheduleSettings)
		api.POST("/settings/schedule", updateScheduleSettings)
		api.GET("/settings/global", getGlobalSettings)
		api.POST("/settings/global", updateGlobalSettings)
		api.POST("/settings/test_bark", testBarkNotification)

		api.POST("/openlist/scan", triggerOpenListScan)

		// 插件管理
		api.GET("/plugins", listPlugins)
		api.GET("/plugins/:name", getPlugin)
		api.PUT("/plugins/:name/config", updatePluginConfig)

		// Telegram 配置
		api.GET("/telegram/config", getTelegramConfig)
		api.PUT("/telegram/config", updateTelegramConfig)
		api.POST("/telegram/test", testTelegramConnection)

		// 资源搜索
		api.GET("/search", searchResources)
		api.GET("/search/sources", listSearchSources)
		api.GET("/search/config", getSearchConfig)
		api.PUT("/search/config", updateSearchConfig)
		api.GET("/search/validate", validateSearchLink)
		api.POST("/search/validate_batch", validateSearchBatch)

		// 通知配置
		api.GET("/notify", listNotifiers)
		api.GET("/notify/:name", getNotifier)
		api.PUT("/notify/:name", updateNotifier)
		api.POST("/notify/:name/test", testNotifier)
	}

	// 静态资源处理
	staticFS := GetStaticFS()
	fileServer := http.FileServer(staticFS)

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// 1. 如果是 API 请求，直接返回 404
		if strings.HasPrefix(path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API route not found"})
			return
		}

		// 2. 如果是请求具体的文件（带点 .），尝试通过 FileServer 处理
		if strings.Contains(path, ".") {
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		// 3. 对于页面路由（如 /tasks 或 /），手动返回 index.html 的内容
		// 使用 Open + ReadAll 绕过 http.ServeFile 的自动重定向逻辑
		f, err := staticFS.Open("index.html")
		if err != nil {
			slog.Error("无法打开 index.html", "error", err)
			c.String(http.StatusNotFound, "Frontend assets not found")
			return
		}
		defer f.Close()

		content, err := io.ReadAll(f)
		if err != nil {
			slog.Error("读取 index.html 失败", "error", err)
			c.String(http.StatusInternalServerError, "Internal Server Error")
			return
		}

		c.Data(http.StatusOK, "text/html; charset=utf-8", content)
	})

	return r
}

// 插件管理处理函数
var pluginHandler *PluginHandler

func InitPluginHandler(manager *plugin.Manager) {
	pluginHandler = NewPluginHandler(manager)
}

func listPlugins(c *gin.Context) {
	if pluginHandler == nil {
		c.PureJSON(http.StatusServiceUnavailable, gin.H{"error": "插件系统未初始化"})
		return
	}
	pluginHandler.ListPlugins(c)
}

func getPlugin(c *gin.Context) {
	if pluginHandler == nil {
		c.PureJSON(http.StatusServiceUnavailable, gin.H{"error": "插件系统未初始化"})
		return
	}
	pluginHandler.GetPlugin(c)
}

func updatePluginConfig(c *gin.Context) {
	if pluginHandler == nil {
		c.PureJSON(http.StatusServiceUnavailable, gin.H{"error": "插件系统未初始化"})
		return
	}
	pluginHandler.UpdatePluginConfig(c)
}

// Telegram 处理函数
var telegramHandler *TelegramHandler

func InitTelegramHandler(bot *telegram.Bot) {
	telegramHandler = NewTelegramHandler(bot)
}

func getTelegramConfig(c *gin.Context) {
	if telegramHandler == nil {
		c.PureJSON(http.StatusServiceUnavailable, gin.H{"error": "Telegram 未初始化"})
		return
	}
	telegramHandler.GetConfig(c)
}

func updateTelegramConfig(c *gin.Context) {
	if telegramHandler == nil {
		c.PureJSON(http.StatusServiceUnavailable, gin.H{"error": "Telegram 未初始化"})
		return
	}
	telegramHandler.UpdateConfig(c)
}

func testTelegramConnection(c *gin.Context) {
	if telegramHandler == nil {
		c.PureJSON(http.StatusServiceUnavailable, gin.H{"error": "Telegram 未初始化"})
		return
	}
	telegramHandler.TestConnection(c)
}

// 搜索处理函数
var searchHandler *SearchHandler

func InitSearchHandler(client *search.Client) {
	searchHandler = NewSearchHandler(client)
}

func searchResources(c *gin.Context) {
	if searchHandler == nil {
		c.PureJSON(http.StatusServiceUnavailable, gin.H{"error": "搜索服务未初始化"})
		return
	}
	searchHandler.Search(c)
}

func listSearchSources(c *gin.Context) {
	if searchHandler == nil {
		c.PureJSON(http.StatusServiceUnavailable, gin.H{"error": "搜索服务未初始化"})
		return
	}
	searchHandler.ListSources(c)
}

func getSearchConfig(c *gin.Context) {
	if searchHandler == nil {
		c.PureJSON(http.StatusServiceUnavailable, gin.H{"error": "搜索服务未初始化"})
		return
	}
	searchHandler.GetConfig(c)
}

func updateSearchConfig(c *gin.Context) {
	if searchHandler == nil {
		c.PureJSON(http.StatusServiceUnavailable, gin.H{"error": "搜索服务未初始化"})
		return
	}
	searchHandler.UpdateConfig(c)
}

func validateSearchLink(c *gin.Context) {
	if searchHandler == nil {
		c.PureJSON(http.StatusServiceUnavailable, gin.H{"error": "搜索服务未初始化"})
		return
	}
	searchHandler.ValidateLink(c)
}

func validateSearchBatch(c *gin.Context) {
	if searchHandler == nil {
		c.PureJSON(http.StatusServiceUnavailable, gin.H{"error": "搜索服务未初始化"})
		return
	}
	searchHandler.ValidateBatch(c)
}

// 通知处理函数
var notifyHandler *NotifyHandler

func InitNotifyHandler(manager *notify.Manager) {
	notifyHandler = NewNotifyHandler(manager)
}

func listNotifiers(c *gin.Context) {
	if notifyHandler == nil {
		c.PureJSON(http.StatusServiceUnavailable, gin.H{"error": "通知服务未初始化"})
		return
	}
	notifyHandler.ListNotifiers(c)
}

func getNotifier(c *gin.Context) {
	if notifyHandler == nil {
		c.PureJSON(http.StatusServiceUnavailable, gin.H{"error": "通知服务未初始化"})
		return
	}
	notifyHandler.GetNotifier(c)
}

func updateNotifier(c *gin.Context) {
	if notifyHandler == nil {
		c.PureJSON(http.StatusServiceUnavailable, gin.H{"error": "通知服务未初始化"})
		return
	}
	notifyHandler.UpdateNotifier(c)
}

func testNotifier(c *gin.Context) {
	if notifyHandler == nil {
		c.PureJSON(http.StatusServiceUnavailable, gin.H{"error": "通知服务未初始化"})
		return
	}
	notifyHandler.TestNotifier(c)
}
