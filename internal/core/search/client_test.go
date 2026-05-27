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

func (s *MockSource) Search(query string, page int) (*SearchResult, error) {
	if s.searchErr != nil {
		return nil, s.searchErr
	}
	return s.results, nil
}

func TestClient_Search(t *testing.T) {
	config := &SearchConfig{
		CloudSaver: CloudSaverConfig{
			Server:   "https://api.cloudsaver.com",
			Username: "admin",
			Password: "pass",
		},
	}
	client := NewClient(config)

	// 测试搜索
	result, err := client.Search("test", []string{}, 1)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestClient_Search_WithSources(t *testing.T) {
	config := &SearchConfig{
		CloudSaver: CloudSaverConfig{
			Server:   "https://api.cloudsaver.com",
			Username: "admin",
			Password: "pass",
		},
	}
	client := NewClient(config)

	// 测试指定搜索源
	result, err := client.Search("test", []string{"CloudSaver"}, 1)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestClient_ListSources(t *testing.T) {
	config := &SearchConfig{
		CloudSaver: CloudSaverConfig{
			Server:   "https://api.cloudsaver.com",
			Username: "admin",
			Password: "pass",
		},
	}
	client := NewClient(config)

	// 测试列出搜索源
	sources := client.ListSources()
	assert.Len(t, sources, 1)
	assert.Contains(t, sources, "CloudSaver")
}
