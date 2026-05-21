// internal/core/search/client.go
package search

import (
	"fmt"
	"sync"
)

// Client 搜索客户端
type Client struct {
	sources []Source
	mu      sync.RWMutex
}

// NewClient 创建搜索客户端
func NewClient() *Client {
	return &Client{
		sources: []Source{
			NewCloudSaverSource("https://api.cloudsaver.com"),
			NewPanSouSource("https://api.pansou.com"),
		},
	}
}

// Search 搜索资源
func (c *Client) Search(query string, sources []string, page int) (*SearchResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var allItems []SearchItem

	for _, source := range c.sources {
		// 如果指定了搜索源，只搜索指定的源
		if len(sources) > 0 && !contains(sources, source.Name()) {
			continue
		}

		result, err := source.Search(query, page)
		if err != nil {
			// 记录错误，继续搜索其他源
			fmt.Printf("搜索源 %s 失败: %v\n", source.Name(), err)
			continue
		}

		allItems = append(allItems, result.Items...)
	}

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
