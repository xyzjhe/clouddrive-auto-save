// internal/core/search/sources.go
package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Source 搜索源接口
type Source interface {
	Name() string
	Search(query string, page int) (*SearchResult, error)
}

// SearchResult 搜索结果
type SearchResult struct {
	Total int          `json:"total"`
	Page  int          `json:"page"`
	Items []SearchItem `json:"items"`
}

// SearchItem 搜索结果项
type SearchItem struct {
	Title     string `json:"title" binding:"required"`
	Source    string `json:"source" binding:"required"`
	Platform  string `json:"platform" binding:"required"`
	URL       string `json:"url" binding:"required"`
	Size      string `json:"size"`
	UpdatedAt string `json:"updated_at"`
	Summary   string `json:"summary"`
}

// CloudSaverSource CloudSaver 搜索源
type CloudSaverSource struct {
	baseURL string
}

func NewCloudSaverSource(baseURL string) *CloudSaverSource {
	return &CloudSaverSource{baseURL: baseURL}
}

func (s *CloudSaverSource) Name() string {
	return "CloudSaver"
}

func (s *CloudSaverSource) Search(query string, page int) (*SearchResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/search?q=%s&page=%d", s.baseURL, query, page)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) UCAS/1.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 CloudSaver 搜索引擎失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("搜索引擎响应异常，状态码: %d", resp.StatusCode)
	}

	var result SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析搜索结果失败: %w", err)
	}

	return &result, nil
}

// PanSouSource PanSou 搜索源
type PanSouSource struct {
	baseURL string
}

func NewPanSouSource(baseURL string) *PanSouSource {
	return &PanSouSource{baseURL: baseURL}
}

func (s *PanSouSource) Name() string {
	return "PanSou"
}

func (s *PanSouSource) Search(query string, page int) (*SearchResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/search?q=%s&page=%d", s.baseURL, query, page)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) UCAS/1.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 PanSou 搜索引擎失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("搜索引擎响应异常，状态码: %d", resp.StatusCode)
	}

	var result SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析搜索结果失败: %w", err)
	}

	return &result, nil
}
