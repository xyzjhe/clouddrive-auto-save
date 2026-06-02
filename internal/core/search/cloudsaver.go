// internal/core/search/cloudsaver.go
package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

// CloudSaverSource CloudSaver 搜索源
type CloudSaverSource struct {
	baseURL       string
	username      string
	password      string
	token         string
	mu            sync.RWMutex
	OnTokenUpdate func(token string)
}

// NewCloudSaverSource 创建 CloudSaver 搜索源
func NewCloudSaverSource(baseURL, username, password, token string) *CloudSaverSource {
	return &CloudSaverSource{
		baseURL:  strings.TrimRight(baseURL, "/"),
		username: username,
		password: password,
		token:    token,
	}
}

func (s *CloudSaverSource) Name() string {
	return "CloudSaver"
}

// EnsureToken 确保 token 有效，过期则重新登录（启动时预热用，不执行搜索）
func (s *CloudSaverSource) EnsureToken() {
	s.mu.RLock()
	token := s.token
	s.mu.RUnlock()
	if token == "" {
		if err := s.login(); err != nil {
			slog.Warn("CloudSaver token 预热失败", "error", err)
		}
	}
}

// login 登录获取 Token
func (s *CloudSaverSource) login() error {
	reqURL := fmt.Sprintf("%s/api/user/login", s.baseURL)
	// 使用 json.Marshal 防止用户名/密码中的特殊字符导致 JSON 注入
	loginPayload, err := json.Marshal(map[string]string{
		"username": s.username,
		"password": s.password,
	})
	if err != nil {
		return fmt.Errorf("构造登录请求失败: %w", err)
	}
	req, err := http.NewRequestWithContext(context.Background(), "POST", reqURL, bytes.NewReader(loginPayload))
	if err != nil {
		return fmt.Errorf("创建登录请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("登录请求失败: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("解析登录响应失败: %w", err)
	}
	if !result.Success {
		return fmt.Errorf("登录失败: %s", result.Message)
	}

	s.mu.Lock()
	s.token = result.Data.Token
	s.mu.Unlock()

	if s.OnTokenUpdate != nil {
		s.OnTokenUpdate(result.Data.Token)
	}
	return nil
}

// Search 搜索资源
func (s *CloudSaverSource) Search(query string, platforms []string, page int) (*SearchResult, error) {
	result, err := s.doSearch(query, "")
	if err != nil {
		slog.Error("CloudSaver 首次搜索失败", "error", err)
		return nil, err
	}

	slog.Info("CloudSaver 搜索响应", "success", result.Success, "message", result.Message, "data_len", len(result.Data))

	if result.Message == "无效的 token" || result.Message == "未提供 token" {
		slog.Info("CloudSaver token 无效，尝试自动登录")
		if loginErr := s.login(); loginErr != nil {
			return nil, fmt.Errorf("自动登录失败: %w", loginErr)
		}
		result, err = s.doSearch(query, "")
		if err != nil {
			return nil, err
		}
	}

	if !result.Success {
		return nil, fmt.Errorf("搜索失败: %s", result.Message)
	}

	items := s.cleanResults(result.Data, platforms)
	slog.Info("CloudSaver 清洗结果", "input_channels", len(result.Data), "output_items", len(items))
	return &SearchResult{
		Total: len(items),
		Page:  page,
		Items: items,
	}, nil
}

// csSearchResponse CloudSaver 搜索响应
type csSearchResponse struct {
	Success bool                     `json:"success"`
	Message string                   `json:"message"`
	Data    []map[string]interface{} `json:"data"`
}

// doSearch 执行搜索请求
func (s *CloudSaverSource) doSearch(query, lastMessageID string) (*csSearchResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	params := url.Values{}
	params.Set("keyword", query)
	if lastMessageID != "" {
		params.Set("lastMessageId", lastMessageID)
	}
	reqURL := fmt.Sprintf("%s/api/search?%s", s.baseURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建搜索请求失败: %w", err)
	}

	s.mu.RLock()
	token := s.token
	s.mu.RUnlock()

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 UCAS/1.0")

	slog.Debug("CloudSaver 请求", "url", reqURL, "token_len", len(token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("搜索请求失败: %w", err)
	}
	defer resp.Body.Close()

	slog.Debug("CloudSaver 响应头", "status", resp.StatusCode, "content_length", resp.ContentLength, "transfer_encoding", resp.TransferEncoding, "uncompressed", resp.Uncompressed, "content_encoding", resp.Header.Get("Content-Encoding"))

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	slog.Info("CloudSaver 响应大小", "bytes", len(body), "content_length", resp.ContentLength)

	var result csSearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析搜索响应失败: %w", err)
	}
	return &result, nil
}

// 预编译 CloudSaver 清洗正则，避免每次调用重新编译
var (
	csPatternTitle   = regexp.MustCompile(`(?:名称|标题)[：:]?\s*(.*)`)
	csPatternContent = regexp.MustCompile(`(?:描述|简介)[：:]?\s*(.*?)(?:链接|标签|$)`)
	csPatternHTML    = regexp.MustCompile(`<[^>]+>`)
)

// cleanResults 清洗搜索结果
func (s *CloudSaverSource) cleanResults(data []map[string]interface{}, platforms []string) []SearchItem {
	var items []SearchItem
	seen := make(map[string]bool)

	for _, channel := range data {
		list, ok := channel["list"].([]interface{})
		if !ok {
			continue
		}
		for _, item := range list {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			title, _ := itemMap["title"].(string)
			content, _ := itemMap["content"].(string)
			pubDate, _ := itemMap["pubDate"].(string)
			channelID, _ := itemMap["channelId"].(string)

			var tags []string
			if tagsRaw, ok := itemMap["tags"].([]interface{}); ok {
				for _, t := range tagsRaw {
					if s, ok := t.(string); ok {
						tags = append(tags, s)
					}
				}
			}

			if m := csPatternTitle.FindStringSubmatch(title); len(m) > 1 {
				title = strings.TrimSpace(m[1])
			}

			if m := csPatternContent.FindStringSubmatch(content); len(m) > 1 {
				content = strings.TrimSpace(m[1])
			}
			content = csPatternHTML.ReplaceAllString(content, "")
			content = strings.TrimSpace(content)

			cloudLinks, _ := itemMap["cloudLinks"].([]interface{})
			for _, link := range cloudLinks {
				linkMap, ok := link.(map[string]interface{})
				if !ok {
					continue
				}
				cloudType, _ := linkMap["cloudType"].(string)

				// 根据 platforms 过滤（如果指定了）
				if len(platforms) > 0 && !contains(platforms, cloudType) {
					continue
				}

				linkURL, _ := linkMap["link"].(string)
				if linkURL == "" || seen[linkURL] {
					continue
				}
				seen[linkURL] = true

				items = append(items, SearchItem{
					Title:     title,
					URL:       linkURL,
					Source:    "CloudSaver",
					Platform:  cloudType,
					Summary:   content,
					UpdatedAt: toCST(pubDate),
					Tags:      tags,
					Channel:   channelID,
				})
			}
		}
	}
	return items
}
