// internal/api/telegram.go
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zcq/clouddrive-auto-save/internal/core/telegram"
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
	// TODO: 从数据库获取配置
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": telegram.DefaultConfig(),
	})
}

// UpdateConfig 更新 Telegram 配置
func (h *TelegramHandler) UpdateConfig(c *gin.Context) {
	var config telegram.Config
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的配置格式",
		})
		return
	}

	// TODO: 保存配置到数据库

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "配置已更新",
	})
}

// TestConnection 测试连接
func (h *TelegramHandler) TestConnection(c *gin.Context) {
	// TODO: 实现测试连接逻辑
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "连接成功",
	})
}
