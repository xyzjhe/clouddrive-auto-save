// internal/api/search.go
package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net"
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

	// 不再自动验证，由前端按分页触发 POST /api/search/validate_batch

	c.PureJSON(http.StatusOK, gin.H{
		"total":     result.Total,
		"page":      result.Page,
		"items":     result.Items,
		"search_id": searchID,
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
	return isPublicHost(u.Hostname())
}

// isPublicHost 检查主机名是否为公网可达地址，拦截 RFC 1918 私有网段、回环、链路本地等
func isPublicHost(hostname string) bool {
	host := strings.ToLower(hostname)
	// 域名黑名单（非 IP 的内网主机名）
	switch host {
	case "localhost", "0.0.0.0":
		return false
	}

	// 尝试解析为 IP（含 IPv6）
	ip := net.ParseIP(host)
	if ip == nil {
		// 不是 IP 格式（域名），放行，由 DNS 解析后的实际连接层面防护
		return true
	}

	// IPv4 / IPv6 回环
	if ip.IsLoopback() {
		return false
	}
	// 链路本地（169.254.x.x / fe80::）
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return false
	}
	// IPv4 私有网段：10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16
	if ip.IsPrivate() {
		return false
	}
	// 未指定地址 0.0.0.0 / ::
	if ip.IsUnspecified() {
		return false
	}

	return true
}

// validateItemRequest 批量验证请求中的单条记录
type validateItemRequest struct {
	Index int    `json:"index"`
	URL   string `json:"url"`
}

// ValidateBatch 批量验证搜索结果中的链接有效性
// 由前端按分页触发，仅验证当前页可见的链接
func (h *SearchHandler) ValidateBatch(c *gin.Context) {
	var req struct {
		SearchID string                `json:"search_id"`
		Items    []validateItemRequest `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.SearchID == "" || len(req.Items) == 0 {
		c.PureJSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 限制单次最多验证 100 条，防止滥用
	if len(req.Items) > 100 {
		req.Items = req.Items[:100]
	}

	go validateBatchItems(req.SearchID, req.Items)
	c.PureJSON(http.StatusOK, gin.H{"message": "验证已启动", "count": len(req.Items)})
}

// validateBatchItems 并发验证一批链接
func validateBatchItems(searchID string, items []validateItemRequest) {
	sem := make(chan struct{}, 15)
	var wg sync.WaitGroup

	for _, item := range items {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, rawURL string) {
			defer wg.Done()
			defer func() { <-sem }()

			valid, message := validateSingleLink(rawURL)
			utils.BroadcastSearchValidate(utils.SearchValidateEvent{
				SearchID: searchID,
				Index:    idx,
				Valid:    valid,
				Message:  message,
			})
		}(item.Index, item.URL)
	}
	wg.Wait()
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
