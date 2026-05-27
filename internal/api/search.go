// internal/api/search.go
package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zcq/clouddrive-auto-save/internal/core/search"
)

// SearchHandler 搜索 API 处理器
type SearchHandler struct {
	client *search.Client
}

// NewSearchHandler 创建搜索 API 处理器
func NewSearchHandler(client *search.Client) *SearchHandler {
	return &SearchHandler{client: client}
}

// Search 搜索资源
func (h *SearchHandler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请提供搜索关键词",
		})
		return
	}

	sources := c.QueryArray("source")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	result, err := h.client.Search(query, sources, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "搜索失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": result,
	})
}

// ListSources 列出搜索源
func (h *SearchHandler) ListSources(c *gin.Context) {
	sources := h.client.ListSources()
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": sources,
	})
}

// GetConfig 获取搜索源配置（密码脱敏）
func (h *SearchHandler) GetConfig(c *gin.Context) {
	config := h.client.GetConfig()
	masked := *config
	if masked.CloudSaver.Password != "" {
		masked.CloudSaver.Password = "***"
	}
	if masked.CloudSaver.Token != "" {
		masked.CloudSaver.Token = "***"
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": masked})
}

// UpdateConfig 更新搜索源配置
func (h *SearchHandler) UpdateConfig(c *gin.Context) {
	var config search.SearchConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误"})
		return
	}
	if err := h.client.SaveAndUpdateConfig(&config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存配置失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "配置已更新"})
}
