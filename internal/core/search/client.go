// internal/core/search/client.go
package search

import (
	"log/slog"
	"sync"
)

// Client 搜索客户端
type Client struct {
	sources []Source
	mu      sync.RWMutex
}

// NewClient 创建搜索客户端
func NewClient(config *SearchConfig) *Client {
	var sources []Source

	if config.CloudSaver.Server != "" {
		sources = append(sources, NewCloudSaverSource(
			config.CloudSaver.Server,
			config.CloudSaver.Username,
			config.CloudSaver.Password,
			config.CloudSaver.Token,
		))
	}

	// TODO: Task 4 实现 PanSou 源后启用

	return &Client{sources: sources}
}

// Search 搜索资源
func (c *Client) Search(query string, sources []string, page int) (*SearchResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var allItems []SearchItem
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, source := range c.sources {
		// 如果指定了搜索源，只搜索指定的源
		if len(sources) > 0 && !contains(sources, source.Name()) {
			continue
		}

		wg.Add(1)
		go func(src Source) {
			defer wg.Done()
			result, err := src.Search(query, page)
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

	return &SearchResult{
		Total: len(allItems),
		Page:  page,
		Items: allItems,
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
