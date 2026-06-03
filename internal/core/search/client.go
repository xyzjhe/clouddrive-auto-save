// internal/core/search/client.go
package search

import (
	"log/slog"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"

	"gorm.io/gorm"
)

// reTrailingGarbage 匹配 URL 尾部非 URL 安全字符（搜索源数据常带 emoji、标签符号等）
var reTrailingGarbage = regexp.MustCompile(`[^\w\-./?:@&=#]+$`)

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

// WarmupToken 启动时主动校验 CloudSaver token，过期则重登，避免首次搜索延迟
func (c *Client) WarmupToken() {
	go func() {
		c.mu.RLock()
		var cs *CloudSaverSource
		for _, src := range c.sources {
			if v, ok := src.(*CloudSaverSource); ok {
				cs = v
				break
			}
		}
		c.mu.RUnlock()
		if cs != nil {
			cs.EnsureToken()
		}
	}()
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
func (c *Client) Search(query string, sources []string, platforms []string, page int) (*SearchResult, error) {
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
			result, err := src.Search(query, platforms, page)
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

	// 按 URL 去重（归一化后再比较）
	seen := make(map[string]bool, len(allItems))
	var deduped []SearchItem
	for _, item := range allItems {
		key := normalizeURL(item.URL)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		deduped = append(deduped, item)
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

// normalizeURL 对 URL 进行归一化，用于去重比较
// 处理：剥离尾部垃圾字符、协议统一 https、去除尾部分隔符、排序查询参数、去除片段
func normalizeURL(rawURL string) string {
	// 剥离尾部非 URL 安全字符（搜索源数据常带 emoji、标签符号等）
	rawURL = reTrailingGarbage.ReplaceAllString(strings.TrimSpace(rawURL), "")
	if rawURL == "" {
		return ""
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return strings.ToLower(strings.TrimSpace(rawURL))
	}

	// 协议统一为小写，http → https 视为相同
	u.Scheme = strings.ToLower(u.Scheme)
	if u.Scheme == "http" {
		u.Scheme = "https"
	}

	// 主机名统一小写
	u.Host = strings.ToLower(u.Host)

	// 去除尾部分隔符
	u.Path = strings.TrimRight(u.Path, "/")

	// 排序查询参数，确保不同顺序的相同参数被视为相同
	if u.RawQuery != "" {
		q := u.Query()
		if len(q) > 0 {
			// 收集并排序键值对
			var pairs []string
			for k, vals := range q {
				for _, v := range vals {
					if v == "" {
						pairs = append(pairs, k)
					} else {
						pairs = append(pairs, k+"="+v)
					}
				}
			}
			sort.Strings(pairs)
			u.RawQuery = strings.Join(pairs, "&")
		}
	}

	// 去除片段
	u.Fragment = ""

	return u.String()
}
