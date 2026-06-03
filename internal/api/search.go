// internal/api/search.go
package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zcq/clouddrive-auto-save/internal/core"
	"github.com/zcq/clouddrive-auto-save/internal/core/search"
	"github.com/zcq/clouddrive-auto-save/internal/utils"
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
	platforms := c.QueryArray("platform")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	result, err := h.client.Search(query, sources, platforms, page)
	if err != nil {
		c.PureJSON(http.StatusInternalServerError, gin.H{"error": "搜索失败"})
		return
	}

	// 生成 search_id 用于关联 SSE 验证事件
	searchID := "srch_" + generateHexID(8)

	// 异步启动链接验证（仅验证分页范围内的结果，避免海量结果冲刷 SSE 通道）
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	validateItems := result.Items
	if len(validateItems) > pageSize {
		validateItems = validateItems[:pageSize]
	}
	if len(validateItems) > 0 {
		go validateSearchResults(searchID, validateItems)
	}

	c.PureJSON(http.StatusOK, gin.H{
		"total":          result.Total,
		"page":           result.Page,
		"items":          result.Items,
		"search_id":      searchID,
		"validate_count": len(validateItems),
	})
}

// ListSources 列出搜索源
func (h *SearchHandler) ListSources(c *gin.Context) {
	sources := h.client.ListSources()
	c.PureJSON(http.StatusOK, sources)
}

// GetConfig 获取搜索源配置
// 密码/token 返回真实值，前端通过 type="password" + show-password 做视觉隐藏
func (h *SearchHandler) GetConfig(c *gin.Context) {
	config := h.client.GetConfig()
	c.PureJSON(http.StatusOK, config)
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

// isSafeURL 基础 URL 安全校验：仅阻止内网地址和非 HTTP(S) 协议（防止 SSRF）
// 平台域名合法性由 GetDriverByURL 判定，此处不重复
func isSafeURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	// 阻止内网地址
	internal := []string{"localhost", "127.0.0.1", "0.0.0.0"}
	for _, h := range internal {
		if host == h {
			return false
		}
	}
	if strings.HasPrefix(host, "192.168.") || strings.HasPrefix(host, "10.") || strings.HasPrefix(host, "172.") {
		return false
	}
	return true
}

// ValidateLink 验证分享链接有效性
func (h *SearchHandler) ValidateLink(c *gin.Context) {
	rawURL := c.Query("url")
	if rawURL == "" {
		c.PureJSON(http.StatusBadRequest, gin.H{"error": "请提供链接"})
		return
	}
	if !isSafeURL(rawURL) {
		c.PureJSON(http.StatusBadRequest, gin.H{"error": "链接地址不合法"})
		return
	}

	driver := core.GetDriverByURL(rawURL)
	if driver == nil {
		c.PureJSON(http.StatusOK, gin.H{"valid": false, "message": "不支持的链接格式"})
		return
	}

	_, err := driver.ParseShare(c.Request.Context(), rawURL, "", "")
	if err != nil {
		c.PureJSON(http.StatusOK, gin.H{"valid": false, "message": err.Error()})
		return
	}

	c.PureJSON(http.StatusOK, gin.H{"valid": true, "message": "链接有效"})
}

// generateHexID 生成指定长度的随机 hex 字符串
func generateHexID(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)[:n]
}

// validateSearchResults 并发验证搜索结果中的链接有效性
// 通过 SSE 推送每条结果的验证状态
func validateSearchResults(searchID string, items []search.SearchItem) {
	sem := make(chan struct{}, 15) // 15 并发
	var wg sync.WaitGroup

	for i, item := range items {
		wg.Add(1)
		sem <- struct{}{} // 获取信号量
		go func(idx int, url string) {
			defer wg.Done()
			defer func() { <-sem }() // 释放信号量

			valid, message := validateSingleLink(url)
			utils.BroadcastSearchValidate(utils.SearchValidateEvent{
				SearchID: searchID,
				Index:    idx,
				Valid:    valid,
				Message:  message,
			})
		}(i, item.URL)
	}
	wg.Wait()
}

// validateSingleLink 验证单个分享链接有效性
func validateSingleLink(rawURL string) (bool, string) {
	if !isSafeURL(rawURL) {
		return false, "链接地址不合法"
	}

	driver := core.GetDriverByURL(rawURL)
	if driver == nil {
		return false, "不支持的链接格式"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := driver.ParseShare(ctx, rawURL, "", "")
	if err != nil {
		return false, err.Error()
	}
	return true, ""
}
