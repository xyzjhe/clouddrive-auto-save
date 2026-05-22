// internal/api/notify.go
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zcq/clouddrive-auto-save/internal/core/notify"
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

	// TODO: 从数据库获取配置
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"name": name,
		},
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

	// TODO: 保存配置到数据库

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
