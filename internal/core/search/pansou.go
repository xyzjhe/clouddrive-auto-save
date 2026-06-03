// internal/core/search/pansou.go
package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// PanSouSource PanSou 搜索源
type PanSouSource struct {
	baseURL string
}

// NewPanSouSource 创建 PanSou 搜索源
func NewPanSouSource(baseURL string) *PanSouSource {
	return &PanSouSource{baseURL: strings.TrimRight(baseURL, "/")}
}

func (s *PanSouSource) Name() string {
	return "PanSou"
}

// panToPlatform 将 PanSou 的 pan 字段映射到平台常量
func panToPlatform(pan string) string {
	switch strings.ToLower(pan) {
	case "quark":
		return PlatformQuark
	case "139", "mobile", "china_mobile":
		return Platform139
	default:
		return pan // 返回原始值，如 "xunlei", "aliyundrive" 等
	}
}

// platformToCloudTypes 将平台常量映射为 PanSou cloud_types 参数值
func platformToCloudTypes(platforms []string) []string {
	var result []string
	for _, p := range platforms {
		switch strings.ToLower(p) {
		case "139":
			result = append(result, "mobile") // 本地 PanSou 使用 "mobile" 而非 "139"
		case "quark":
			result = append(result, "quark")
		default:
			result = append(result, p)
		}
	}
	return result
}

// extractURLFromContent 从 content HTML 中提取链接
func extractURLFromContent(content string) string {
	re := regexp.MustCompile(`href="([^"]+)"`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}
	re2 := regexp.MustCompile(`https?://[^\s<"]+`)
	matches2 := re2.FindStringSubmatch(content)
	if len(matches2) > 0 {
		return matches2[0]
	}
	return ""
}

// extractTitleFromContent 从 content 中提取标题
func extractTitleFromContent(content string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	cleaned := re.ReplaceAllString(content, "")

	lines := strings.Split(cleaned, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "名称：") {
			title := strings.TrimPrefix(line, "名称：")
			title = strings.ReplaceAll(title, "'", "")
			return strings.TrimSpace(title)
		}
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "描述：") && !strings.HasPrefix(line, "链接：") {
			return line
		}
	}
	return ""
}

// extractSummaryFromContent 从 content 中提取描述
func extractSummaryFromContent(content string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	cleaned := re.ReplaceAllString(content, "")

	lines := strings.Split(cleaned, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "描述：") {
			return strings.TrimPrefix(line, "描述：")
		}
	}
	return ""
}

// Search 搜索资源（兼容新旧两种 PanSou API 格式）
func (s *PanSouSource) Search(query string, platforms []string, page int) (*SearchResult, error) {
	// 优先尝试旧版 API（kw + merged_by_type），因为本地部署的 PanSou 服务器使用此格式
	items, err := s.searchLegacy(query, platforms, page)
	if err == nil && len(items) > 0 {
		return &SearchResult{
			Total: len(items),
			Page:  page,
			Items: items,
		}, nil
	}

	// 如果旧版 API 失败或无结果，尝试新版 API（keyword + flat array）
	items, err = s.searchNew(query, platforms, page)
	if err == nil {
		return &SearchResult{
			Total: len(items),
			Page:  page,
			Items: items,
		}, nil
	}

	// 两种都失败，返回空结果
	return &SearchResult{Page: page}, nil
}

// searchLegacy 旧版 API：kw 参数 + merged_by_type 响应
func (s *PanSouSource) searchLegacy(query string, platforms []string, page int) ([]SearchItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	params := url.Values{}
	params.Set("kw", query)
	params.Set("res", "merge")

	if len(platforms) > 0 {
		cloudTypes := platformToCloudTypes(platforms)
		params.Set("cloud_types", strings.Join(cloudTypes, ","))
	} else {
		params.Set("cloud_types", "quark,mobile")
	}

	reqURL := fmt.Sprintf("%s/api/search?%s", s.baseURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 UCAS/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// 旧版响应格式
	var legacyResult struct {
		Code int `json:"code"`
		Data struct {
			MergedByType map[string][]struct {
				URL      string `json:"url"`
				Note     string `json:"note"`
				DateTime string `json:"datetime"`
				Source   string `json:"source"`
			} `json:"merged_by_type"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&legacyResult); err != nil {
		return nil, err
	}

	if legacyResult.Code != 0 && legacyResult.Code != 200 {
		return nil, fmt.Errorf("API error code: %d", legacyResult.Code)
	}

	// 构建平台过滤集合
	platformFilter := make(map[string]bool)
	for _, p := range platforms {
		platformFilter[strings.ToLower(p)] = true
	}

	var items []SearchItem
	seen := make(map[string]bool)
	pattern := regexp.MustCompile(`^(.*?)(?:【(?:简介|介绍|描述)】|\[(?:简介|介绍|描述)\]|(?:简介|介绍|描述)[:：])(.*)$`)

	for platform, dataItems := range legacyResult.Data.MergedByType {
		mappedPlatform := panToPlatform(platform)

		// 如果指定了平台过滤，检查是否匹配
		if len(platformFilter) > 0 && !platformFilter[strings.ToLower(mappedPlatform)] {
			continue
		}

		for _, item := range dataItems {
			if item.URL == "" || seen[item.URL] {
				continue
			}
			seen[item.URL] = true

			title := item.Note
			content := ""

			if m := pattern.FindStringSubmatch(item.Note); len(m) > 2 {
				title = strings.TrimSpace(m[1])
				content = strings.TrimSpace(m[2])
			}

			items = append(items, SearchItem{
				Title:     title,
				URL:       item.URL,
				Source:    "PanSou",
				Platform:  mappedPlatform,
				Summary:   content,
				UpdatedAt: toCST(item.DateTime),
				Channel:   item.Source,
			})
		}
	}

	return items, nil
}

// searchNew 新版 API：keyword 参数 + 扁平数组响应
func (s *PanSouSource) searchNew(query string, platforms []string, page int) ([]SearchItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	params := url.Values{}
	params.Set("keyword", query)

	reqURL := fmt.Sprintf("%s/api/search?%s", s.baseURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 UCAS/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// 新版响应格式
	var newResult struct {
		Total int `json:"total"`
		Data  []struct {
			ID      int    `json:"id"`
			Content string `json:"content"`
			Pan     string `json:"pan"`
			Image   string `json:"image"`
			Time    string `json:"time"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&newResult); err != nil {
		return nil, err
	}

	// 构建平台过滤集合
	platformFilter := make(map[string]bool)
	for _, p := range platforms {
		platformFilter[strings.ToLower(p)] = true
	}

	var items []SearchItem
	seen := make(map[string]bool)

	for _, item := range newResult.Data {
		itemURL := extractURLFromContent(item.Content)
		if itemURL == "" || seen[itemURL] {
			continue
		}

		platform := panToPlatform(item.Pan)

		if len(platformFilter) > 0 && !platformFilter[strings.ToLower(platform)] {
			continue
		}

		seen[itemURL] = true

		title := extractTitleFromContent(item.Content)
		if title == "" {
			title = "未知资源"
		}

		items = append(items, SearchItem{
			Title:     title,
			URL:       itemURL,
			Source:    "PanSou",
			Platform:  platform,
			Summary:   extractSummaryFromContent(item.Content),
			UpdatedAt: toCST(item.Time),
			Channel:   item.Pan,
		})
	}

	return items, nil
}
