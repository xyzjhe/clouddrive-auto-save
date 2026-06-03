// internal/core/search/client_test.go
package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockSource Mock 搜索源实现
type MockSource struct {
	name      string
	results   *SearchResult
	searchErr error
}

func NewMockSource(name string, results *SearchResult) *MockSource {
	return &MockSource{
		name:    name,
		results: results,
	}
}

func (s *MockSource) Name() string {
	return s.name
}

func (s *MockSource) Search(query string, platforms []string, page int) (*SearchResult, error) {
	_ = platforms
	if s.searchErr != nil {
		return nil, s.searchErr
	}
	return s.results, nil
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"http转https", "http://pan.quark.cn/s/abc", "https://pan.quark.cn/s/abc"},
		{"尾部分隔符去除", "https://pan.quark.cn/s/abc/", "https://pan.quark.cn/s/abc"},
		{"查询参数排序", "https://pan.quark.cn/s/abc?b=2&a=1", "https://pan.quark.cn/s/abc?a=1&b=2"},
		{"片段去除", "https://pan.quark.cn/s/abc#section", "https://pan.quark.cn/s/abc"},
		{"主机名小写", "https://PAN.Quark.CN/s/abc", "https://pan.quark.cn/s/abc"},
		{"空URL", "", ""},
		{"纯http+尾部斜杠+参数乱序", "http://PAN.Quark.CN/s/abc/?b=2&a=1#frag", "https://pan.quark.cn/s/abc?a=1&b=2"},
		{"已是标准格式", "https://pan.quark.cn/s/abc", "https://pan.quark.cn/s/abc"},
		{"尾部emoji去除", "https://caiyun.139.com/m/i?2j3ahPrfeGrp0🏷", "https://caiyun.139.com/m/i?2j3ahPrfeGrp0"},
		{"尾部特殊符号去除", "https://pan.quark.cn/s/abc!!!", "https://pan.quark.cn/s/abc"},
		{"仅尾部垃圾", "!!!", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, normalizeURL(tt.input))
		})
	}
}

func TestClient_Search_MergeAndDedup(t *testing.T) {
	config := &SearchConfig{
		PanSou: PanSouConfig{Server: "http://test"},
	}
	client := NewClient(config, nil)

	// 替换为 mock 源
	client.sources = []Source{
		NewMockSource("CloudSaver", &SearchResult{
			Items: []SearchItem{
				{Title: "A", URL: "https://pan.quark.cn/s/aaa", UpdatedAt: "2025-01-01 10:00:00", Source: "CloudSaver", Platform: "quark"},
				{Title: "B", URL: "https://pan.quark.cn/s/bbb", UpdatedAt: "2025-01-02 10:00:00", Source: "CloudSaver", Platform: "quark"},
			},
		}),
		NewMockSource("PanSou", &SearchResult{
			Items: []SearchItem{
				{Title: "B-dup", URL: "https://pan.quark.cn/s/bbb", UpdatedAt: "2025-01-03 10:00:00", Source: "PanSou", Platform: "quark"},
				{Title: "C", URL: "https://pan.quark.cn/s/ccc", UpdatedAt: "2025-01-01 08:00:00", Source: "PanSou", Platform: "quark"},
			},
		}),
	}

	result, err := client.Search("test", []string{}, nil, 1)
	require.NoError(t, err)
	assert.Len(t, result.Items, 3)
	// 验证按时间降序
	assert.Equal(t, "https://pan.quark.cn/s/bbb", result.Items[0].URL) // 01-02
	assert.Equal(t, "https://pan.quark.cn/s/aaa", result.Items[1].URL) // 01-01 10:00
	assert.Equal(t, "https://pan.quark.cn/s/ccc", result.Items[2].URL) // 01-01 08:00
}

func TestClient_Search_FilterSources(t *testing.T) {
	config := &SearchConfig{}
	client := NewClient(config, nil)

	client.sources = []Source{
		NewMockSource("CloudSaver", &SearchResult{
			Items: []SearchItem{{Title: "CS", URL: "https://cs.test", Source: "CloudSaver", Platform: "quark"}},
		}),
		NewMockSource("PanSou", &SearchResult{
			Items: []SearchItem{{Title: "PS", URL: "https://ps.test", Source: "PanSou", Platform: "quark"}},
		}),
	}

	result, err := client.Search("test", []string{"PanSou"}, nil, 1)
	require.NoError(t, err)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, "PanSou", result.Items[0].Source)
}

func TestClient_ListSources(t *testing.T) {
	config := &SearchConfig{}
	client := NewClient(config, nil)

	client.sources = []Source{
		NewMockSource("CloudSaver", nil),
		NewMockSource("PanSou", nil),
	}

	sources := client.ListSources()
	assert.Len(t, sources, 2)
	assert.Contains(t, sources, "CloudSaver")
	assert.Contains(t, sources, "PanSou")
}

func TestClient_UpdateConfig(t *testing.T) {
	config := &SearchConfig{}
	client := NewClient(config, nil)
	assert.Len(t, client.sources, 0)

	client.UpdateConfig(&SearchConfig{
		PanSou: PanSouConfig{Server: "https://so.252035.xyz"},
	})
	assert.Len(t, client.sources, 1)
	assert.Equal(t, "PanSou", client.sources[0].Name())
}
