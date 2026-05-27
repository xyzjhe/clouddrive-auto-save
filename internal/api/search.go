// internal/api/search.go
package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zcq/clouddrive-auto-save/internal/core"
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
		c.PureJSON(http.StatusBadRequest, gin.H{"error": "请提供搜索关键词"})
		return
	}

	sources := c.QueryArray("source")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	result, err := h.client.Search(query, sources, page)
	if err != nil {
		c.PureJSON(http.StatusInternalServerError, gin.H{"error": "搜索失败"})
		return
	}

	c.PureJSON(http.StatusOK, result)
}

// ListSources 列出搜索源
func (h *SearchHandler) ListSources(c *gin.Context) {
	sources := h.client.ListSources()
	c.PureJSON(http.StatusOK, sources)
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
	c.PureJSON(http.StatusOK, masked)
}

// UpdateConfig 更新搜索源配置
func (h *SearchHandler) UpdateConfig(c *gin.Context) {
	var config search.SearchConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.PureJSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}
	if err := h.client.SaveAndUpdateConfig(&config); err != nil {
		c.PureJSON(http.StatusInternalServerError, gin.H{"error": "保存配置失败"})
		return
	}
	c.PureJSON(http.StatusOK, gin.H{"message": "配置已更新"})
}

// ValidateLink 验证分享链接有效性
func (h *SearchHandler) ValidateLink(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.PureJSON(http.StatusBadRequest, gin.H{"error": "请提供链接"})
		return
	}

	driver := core.GetDriverByURL(url)
	if driver == nil {
		c.PureJSON(http.StatusOK, gin.H{"valid": false, "message": "不支持的链接格式"})
		return
	}

	_, err := driver.ParseShare(c.Request.Context(), url, "", "")
	if err != nil {
		c.PureJSON(http.StatusOK, gin.H{"valid": false, "message": err.Error()})
		return
	}

	c.PureJSON(http.StatusOK, gin.H{"valid": true, "message": "链接有效"})
}
