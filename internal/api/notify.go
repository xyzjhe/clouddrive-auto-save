package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zcq/clouddrive-auto-save/internal/core/notify"
	"github.com/zcq/clouddrive-auto-save/internal/db"
)

// NotifyHandler 通知配置 API 处理器
type NotifyHandler struct {
	manager *notify.Manager
}

// NewNotifyHandler 创建通知配置 API 处理器
func NewNotifyHandler(manager *notify.Manager) *NotifyHandler {
	return &NotifyHandler{manager: manager}
}

// ListNotifiers 列出所有通知渠道
func (h *NotifyHandler) ListNotifiers(c *gin.Context) {
	notifiers := h.manager.ListNotifiers()
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": notifiers,
	})
}

// GetNotifier 获取通知渠道配置
func (h *NotifyHandler) GetNotifier(c *gin.Context) {
	name := c.Param("name")

	var setting db.Setting
	err := db.DB.Where("key = ?", "notify_config_"+name).First(&setting).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"name":              name,
				"enabled":           false,
				"notify_on_success": true,
				"notify_on_failure": true,
				"config":            gin.H{},
			},
		})
		return
	}

	var config notify.NotifierConfig
	if err := json.Unmarshal([]byte(setting.Value), &config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "反序列化配置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": config,
	})
}

// UpdateNotifier 更新通知渠道配置
func (h *NotifyHandler) UpdateNotifier(c *gin.Context) {
	name := c.Param("name")

	var config notify.NotifierConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的配置格式",
		})
		return
	}

	config.Name = name

	val, err := json.Marshal(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "序列化配置失败",
		})
		return
	}

	setting := db.Setting{
		Key:   "notify_config_" + name,
		Value: string(val),
	}

	if err := db.DB.Save(&setting).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "保存配置失败: " + err.Error(),
		})
		return
	}

	// 重新初始化通知渠道以实现配置热重载
	if err := notify.InitGlobal(db.DB); err != nil {
		slog.Error("重新初始化全局通知管理器失败", "error", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "配置已更新",
	})
}

// TestNotifier 测试通知渠道
func (h *NotifyHandler) TestNotifier(c *gin.Context) {
	name := c.Param("name")

	if err := h.manager.Test(c.Request.Context(), name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "测试失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "测试成功",
	})
}
