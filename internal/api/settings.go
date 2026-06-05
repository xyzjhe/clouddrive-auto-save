package api

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zcq/clouddrive-auto-save/internal/core/notify"
	"github.com/zcq/clouddrive-auto-save/internal/core/openlist"
	"github.com/zcq/clouddrive-auto-save/internal/core/renamer"
	"github.com/zcq/clouddrive-auto-save/internal/core/scheduler"
	"github.com/zcq/clouddrive-auto-save/internal/db"
	"github.com/zcq/clouddrive-auto-save/internal/utils"
)

// allowedSettingKeys 允许通过全局设置接口修改的 key 白名单
var allowedSettingKeys = map[string]bool{
	// 全局调度
	"global_schedule_enabled": true,
	"global_schedule_cron":    true,
	"global_schedule_ui_mode": true,
	// OpenList 扫描
	"openlist_enabled":   true,
	"openlist_api_url":   true,
	"openlist_api_token": true,
	// Bark 通知（兼容新旧字段名）
	"bark_enabled":           true,
	"bark_server":            true,
	"bark_device_key":        true,
	"bark_url":               true,
	"bark_notify_on_success": true,
	"bark_notify_on_failure": true,
	"bark_icon":              true,
	"bark_archive":           true,
	"bark_success_level":     true,
	"bark_success_sound":     true,
	"bark_failure_level":     true,
	"bark_failure_sound":     true,
}

func getScheduleSettings(c *gin.Context) {
	var enabledSetting db.Setting
	var cronSetting db.Setting

	db.DB.Where("key = ?", "global_schedule_enabled").Limit(1).Find(&enabledSetting)
	db.DB.Where("key = ?", "global_schedule_cron").Limit(1).Find(&cronSetting)

	c.PureJSON(http.StatusOK, gin.H{
		"enabled": enabledSetting.Value == "true",
		"cron":    cronSetting.Value,
	})
}

func updateScheduleSettings(c *gin.Context) {
	var input struct {
		Enabled bool   `json:"enabled"`
		Cron    string `json:"cron"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.PureJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 校验全局 Cron 表达式
	if input.Enabled {
		if err := scheduler.ValidateCron(input.Cron); err != nil {
			c.PureJSON(http.StatusBadRequest, gin.H{"error": "全局 Cron 表达式格式错误: " + err.Error()})
			return
		}
	}

	enabledStr := "false"
	if input.Enabled {
		enabledStr = "true"
	}

	// 使用 Updates 或 FirstOrCreate 确保 Key 存在
	db.DB.Save(&db.Setting{Key: "global_schedule_enabled", Value: enabledStr})
	db.DB.Save(&db.Setting{Key: "global_schedule_cron", Value: input.Cron})

	scheduler.Global.UpdateGlobalSchedule(input.Cron, input.Enabled)

	// 推送统计更新
	utils.BroadcastStatsUpdate()

	c.PureJSON(http.StatusOK, gin.H{"message": "settings updated"})
}

func getGlobalSettings(c *gin.Context) {
	var settings []db.Setting
	db.DB.Find(&settings)

	// 返回真实值，前端通过 type="password" + show-password 做视觉隐藏
	res := make(map[string]string)
	for _, s := range settings {
		res[s.Key] = s.Value
	}
	c.PureJSON(http.StatusOK, res)
}

func updateGlobalSettings(c *gin.Context) {
	var input map[string]string
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 白名单校验：仅允许已知的 key
	for k := range input {
		if !allowedSettingKeys[k] {
			c.PureJSON(http.StatusBadRequest, gin.H{"error": "不允许修改未知的配置项: " + k})
			return
		}
	}

	// 提前检查 Cron 校验，避免在循环中重复或次序问题
	if cronExpr, ok := input["global_schedule_cron"]; ok {
		enabled := false
		if e, ok := input["global_schedule_enabled"]; ok {
			enabled = (e == "true")
		} else {
			var s db.Setting
			db.DB.Where("key = ?", "global_schedule_enabled").First(&s)
			enabled = (s.Value == "true")
		}

		if enabled && cronExpr != "" {
			if err := scheduler.ValidateCron(cronExpr); err != nil {
				c.PureJSON(http.StatusBadRequest, gin.H{"error": "全局 Cron 表达式格式错误: " + err.Error()})
				return
			}
		}
	}

	for k, v := range input {
		db.DB.Save(&db.Setting{Key: k, Value: v})
		// 如果更新了定时任务配置，同步更新调度器
		if k == "global_schedule_enabled" || k == "global_schedule_cron" {
			var enabledSetting db.Setting
			var cronSetting db.Setting
			db.DB.Where("key = ?", "global_schedule_enabled").First(&enabledSetting)
			db.DB.Where("key = ?", "global_schedule_cron").First(&cronSetting)
			scheduler.Global.UpdateGlobalSchedule(cronSetting.Value, enabledSetting.Value == "true")
		}
	}

	// 推送统计更新
	utils.BroadcastStatsUpdate()
	c.PureJSON(http.StatusOK, gin.H{"message": "settings updated"})
}

// isSafeBarkURL 校验 Bark 服务器 URL 是否安全（防止 SSRF）
func isSafeBarkURL(serverURL string) bool {
	if serverURL == "" {
		return false
	}
	u, err := url.Parse(strings.TrimSpace(serverURL))
	if err != nil {
		return false
	}
	host := u.Hostname()
	// 禁止内网地址
	if host == "localhost" || host == "127.0.0.1" || host == "0.0.0.0" || strings.HasPrefix(host, "192.168.") || strings.HasPrefix(host, "10.") || strings.HasPrefix(host, "172.") {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

func testBarkNotification(c *gin.Context) {
	var input struct {
		Server  string `json:"bark_server"`
		Key     string `json:"bark_device_key"`
		Title   string `json:"title"`
		Body    string `json:"body"`
		Level   string `json:"level"`
		Sound   string `json:"sound"`
		Icon    string `json:"icon"`
		Archive string `json:"isArchive"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !isSafeBarkURL(input.Server) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bark 服务器地址不合法或为内网地址"})
		return
	}

	title := input.Title
	if title == "" {
		title = "测试通知"
	}
	body := input.Body
	if body == "" {
		body = "这是一条来自 UCAS 的测试推送消息。"
	}
	level := input.Level
	if level == "" {
		level = "active"
	}
	sound := input.Sound
	if sound == "" {
		sound = "birdsong.caf"
	}

	err := notify.SendBarkDirect(input.Server, input.Key, title, body, level, sound, input.Icon, input.Archive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.PureJSON(http.StatusOK, gin.H{"message": "test notification sent"})
}

func getVersion(c *gin.Context) {
	c.PureJSON(http.StatusOK, gin.H{
		"version": appVersion,
		"commit":  appCommit,
		"date":    appDate,
	})
}

func listMagicPatterns(c *gin.Context) {
	c.PureJSON(http.StatusOK, renamer.PredefinedPatterns)
}

func triggerOpenListScan(c *gin.Context) {
	// 重新加载配置（手动扫描忽略全局开关）
	if err := openlist.GlobalScanner.ReloadConfig(true); err != nil {
		c.PureJSON(http.StatusInternalServerError, gin.H{"error": "加载 OpenList 配置失败"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	if err := openlist.GlobalScanner.ScanNow(ctx); err != nil {
		c.PureJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.PureJSON(http.StatusOK, gin.H{"message": "扫描已触发"})
}
