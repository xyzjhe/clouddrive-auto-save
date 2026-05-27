// internal/api/telegram.go
package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/zcq/clouddrive-auto-save/internal/core/telegram"
	"github.com/zcq/clouddrive-auto-save/internal/db"
)

// TelegramHandler Telegram API 处理器
type TelegramHandler struct {
	bot *telegram.Bot
}

// NewTelegramHandler 创建 Telegram API 处理器
func NewTelegramHandler(bot *telegram.Bot) *TelegramHandler {
	return &TelegramHandler{bot: bot}
}

// GetConfig 获取 Telegram 配置
func (h *TelegramHandler) GetConfig(c *gin.Context) {
	var setting db.Setting
	err := db.DB.Where("key = ?", "telegram_config").First(&setting).Error
	if err != nil {
		c.PureJSON(http.StatusOK, telegram.DefaultConfig())
		return
	}

	var config telegram.Config
	if err := json.Unmarshal([]byte(setting.Value), &config); err != nil {
		c.PureJSON(http.StatusInternalServerError, gin.H{"error": "反序列化配置失败: " + err.Error()})
		return
	}

	c.PureJSON(http.StatusOK, config)
}

// UpdateConfig 更新 Telegram 配置
func (h *TelegramHandler) UpdateConfig(c *gin.Context) {
	var config telegram.Config
	if err := c.ShouldBindJSON(&config); err != nil {
		c.PureJSON(http.StatusBadRequest, gin.H{"error": "无效的配置格式"})
		return
	}

	val, err := json.Marshal(config)
	if err != nil {
		c.PureJSON(http.StatusInternalServerError, gin.H{"error": "序列化配置失败"})
		return
	}

	setting := db.Setting{
		Key:   "telegram_config",
		Value: string(val),
	}

	if err := db.DB.Save(&setting).Error; err != nil {
		c.PureJSON(http.StatusInternalServerError, gin.H{"error": "保存配置失败: " + err.Error()})
		return
	}

	// 触发热重载以实现配置即时生效
	h.bot.Stop()
	h.bot.UpdateConfig(&config)
	if config.Enabled {
		go func() {
			if err := h.bot.Start(); err != nil {
				slog.Error("热重载 Telegram 机器人失败", "error", err)
			}
		}()
	}

	c.PureJSON(http.StatusOK, gin.H{"message": "配置已更新"})
}

// TestConnection 测试连接
func (h *TelegramHandler) TestConnection(c *gin.Context) {
	var config telegram.Config
	if err := c.ShouldBindJSON(&config); err != nil {
		c.PureJSON(http.StatusBadRequest, gin.H{"error": "无效的配置格式"})
		return
	}

	if config.BotToken == "" {
		c.PureJSON(http.StatusBadRequest, gin.H{"error": "Bot Token 不能为空"})
		return
	}

	// 向 Telegram API 发送请求来测试连通性和 Token 的有效性
	api, err := tgbotapi.NewBotAPI(config.BotToken)
	if err != nil {
		c.PureJSON(http.StatusInternalServerError, gin.H{"error": "连接失败: " + err.Error()})
		return
	}

	c.PureJSON(http.StatusOK, gin.H{"username": api.Self.UserName, "message": "连接成功"})
}
