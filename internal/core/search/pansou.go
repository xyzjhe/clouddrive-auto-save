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

// psSearchResponse PanSou 搜索响应
type psSearchResponse struct {
	Code int `json:"code"`
	Data struct {
		MergedByType struct {
			Quark    []psItem `json:"quark"`
			Cloud139 []psItem `json:"139"`
		} `json:"merged_by_type"`
	} `json:"data"`
}

type psItem struct {
	URL      string `json:"url"`
	Note     string `json:"note"`
	DateTime string `json:"datetime"`
	Source   string `json:"source"`
}

// Search 搜索资源
func (s *PanSouSource) Search(query string, platforms []string, page int) (*SearchResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	params := url.Values{}
	params.Set("kw", query)
	params.Set("res", "merge")

	// 根据 platforms 动态构建 cloud_types
	if len(platforms) > 0 {
		params.Set("cloud_types", strings.Join(platforms, ","))
	} else {
		params.Set("cloud_types", PlatformQuark+","+Platform139)
	}

	reqURL := fmt.Sprintf("%s/api/search?%s", s.baseURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return &SearchResult{Page: page}, nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 UCAS/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &SearchResult{Page: page}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &SearchResult{Page: page}, nil
	}

	var result psSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &SearchResult{Page: page}, nil
	}

	var allItems []SearchItem
	seen := make(map[string]bool)
	allItems = append(allItems, s.formatResults(result.Data.MergedByType.Quark, PlatformQuark, seen)...)
	allItems = append(allItems, s.formatResults(result.Data.MergedByType.Cloud139, Platform139, seen)...)

	return &SearchResult{
		Total: len(allItems),
		Page:  page,
		Items: allItems,
	}, nil
}

// formatResults 格式化搜索结果，seen 用于跨平台去重
func (s *PanSouSource) formatResults(data []psItem, platform string, seen map[string]bool) []SearchItem {
	pattern := regexp.MustCompile(`^(.*?)(?:【(?:简介|介绍|描述)】|\[(?:简介|介绍|描述)\]|(?:简介|介绍|描述)[:：])(.*)$`)

	var items []SearchItem

	for _, item := range data {
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
			Platform:  platform,
			Summary:   content,
			UpdatedAt: toCST(item.DateTime),
			Channel:   item.Source,
		})
	}
	return items
}
