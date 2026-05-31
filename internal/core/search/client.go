// internal/core/search/client.go
package search

import (
	"log/slog"
	"sort"
	"sync"

	"gorm.io/gorm"
)

// Client 搜索客户端
type Client struct {
	sources []Source
	config  *SearchConfig
	db      *gorm.DB
	mu      sync.RWMutex
}

// NewClient 创建搜索客户端
func NewClient(config *SearchConfig, db *gorm.DB) *Client {
	c := &Client{
		config: config,
		db:     db,
	}
	c.buildSources()
	return c
}

// buildSources 根据配置构建搜索源
func (c *Client) buildSources() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.sources = nil

	if c.config.CloudSaver.Server != "" {
		cs := NewCloudSaverSource(
			c.config.CloudSaver.Server,
			c.config.CloudSaver.Username,
			c.config.CloudSaver.Password,
			c.config.CloudSaver.Token,
		)
		cs.OnTokenUpdate = func(token string) {
			c.config.CloudSaver.Token = token
			if c.db != nil {
				if err := SaveConfig(c.db, c.config); err != nil {
					slog.Error("持久化 CloudSaver Token 失败", "error", err)
				}
			}
		}
		c.sources = append(c.sources, cs)
	}

	if c.config.PanSou.Server != "" {
		c.sources = append(c.sources, NewPanSouSource(c.config.PanSou.Server))
	}
}

// UpdateConfig 热更新配置
func (c *Client) UpdateConfig(config *SearchConfig) {
	c.config = config
	c.buildSources()
}

// GetConfig 获取当前配置
func (c *Client) GetConfig() *SearchConfig {
	return c.config
}

// SaveAndUpdateConfig 保存配置并热更新
func (c *Client) SaveAndUpdateConfig(config *SearchConfig) error {
	// 保留脱敏字段的原值：前端返回 *** 表示未修改，不应覆盖真实密码
	if config.CloudSaver.Password == "***" {
		config.CloudSaver.Password = c.config.CloudSaver.Password
	}
	if config.CloudSaver.Token == "***" {
		config.CloudSaver.Token = c.config.CloudSaver.Token
	}
	if c.db != nil {
		if err := SaveConfig(c.db, config); err != nil {
			return err
		}
	}
	c.UpdateConfig(config)
	return nil
}

// Search 搜索资源（并发 + 去重 + 排序）
func (c *Client) Search(query string, sources []string, page int) (*SearchResult, error) {
	c.mu.RLock()
	activeSources := make([]Source, len(c.sources))
	copy(activeSources, c.sources)
	c.mu.RUnlock()

	var allItems []SearchItem
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, source := range activeSources {
		if len(sources) > 0 && !contains(sources, source.Name()) {
			continue
		}

		wg.Add(1)
		go func(src Source) {
			defer wg.Done()
			result, err := src.Search(query, nil, page)
			if err != nil {
				slog.Error("搜索源失败", "name", src.Name(), "error", err)
				return
			}
			if result != nil && len(result.Items) > 0 {
				mu.Lock()
				allItems = append(allItems, result.Items...)
				mu.Unlock()
			}
		}(source)
	}

	wg.Wait()

	// 去重
	seen := make(map[string]bool)
	var deduped []SearchItem
	for _, item := range allItems {
		if !seen[item.URL] {
			seen[item.URL] = true
			deduped = append(deduped, item)
		}
	}

	// 按时间降序排序
	sort.Slice(deduped, func(i, j int) bool {
		return deduped[i].UpdatedAt > deduped[j].UpdatedAt
	})

	return &SearchResult{
		Total: len(deduped),
		Page:  page,
		Items: deduped,
	}, nil
}

// ListSources 列出所有搜索源
func (c *Client) ListSources() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var names []string
	for _, source := range c.sources {
		names = append(names, source.Name())
	}
	return names
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
