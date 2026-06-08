// internal/api/plugin.go
package api

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zcq/clouddrive-auto-save/internal/core/plugin"
	"github.com/zcq/clouddrive-auto-save/internal/db"
)

// PluginHandler 插件 API 处理器
type PluginHandler struct {
	manager *plugin.Manager
}

// NewPluginHandler 创建插件 API 处理器
func NewPluginHandler(manager *plugin.Manager) *PluginHandler {
	return &PluginHandler{manager: manager}
}

// ListPlugins 列出所有插件
func (h *PluginHandler) ListPlugins(c *gin.Context) {
	plugins := h.manager.ListPlugins()
	c.PureJSON(http.StatusOK, plugins)
}

// GetPlugin 获取插件详情
func (h *PluginHandler) GetPlugin(c *gin.Context) {
	name := c.Param("name")

	plugin, exists := h.manager.GetPlugin(name)
	if !exists {
		c.PureJSON(http.StatusNotFound, gin.H{"error": "插件不存在"})
		return
	}

	c.PureJSON(http.StatusOK, gin.H{
		"name":        plugin.Name(),
		"version":     plugin.Version(),
		"description": plugin.Description(),
		"hooks":       plugin.Hooks(),
	})
}

// UpdatePluginConfig 更新插件配置
func (h *PluginHandler) UpdatePluginConfig(c *gin.Context) {
	name := c.Param("name")

	var config map[string]interface{}
	if err := c.ShouldBindJSON(&config); err != nil {
		c.PureJSON(http.StatusBadRequest, gin.H{"error": "无效的配置格式"})
		return
	}

	// 校验插件是否存在
	_, exists := h.manager.GetPlugin(name)
	if !exists {
		c.PureJSON(http.StatusNotFound, gin.H{"error": "插件不存在"})
		return
	}

	// 持久化配置到 Setting 表
	configJSON, err := json.Marshal(config)
	if err != nil {
		c.PureJSON(http.StatusInternalServerError, gin.H{"error": "序列化配置失败"})
		return
	}
	db.DB.Save(&db.Setting{Key: "plugin_config_" + name, Value: string(configJSON)})

	c.PureJSON(http.StatusOK, gin.H{"message": "配置已更新"})
}
