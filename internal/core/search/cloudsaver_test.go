// internal/core/search/cloudsaver_test.go
package search

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloudSaver_Search_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/user/login":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"data":    map[string]string{"token": "test-token"},
			})
		case "/api/search":
			assert.Equal(t, "黑镜", r.URL.Query().Get("keyword"))
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"data": []map[string]interface{}{
					{
						"channelId": "ch1",
						"list": []map[string]interface{}{
							{
								"title":   "名称：黑镜 第七季",
								"content": "描述：科幻美剧 链接：https://pan.quark.cn/s/xxx 标签：#科幻",
								"pubDate": "2025-01-15T10:30:00+08:00",
								"tags":    []string{"科幻", "美剧"},
								"cloudLinks": []map[string]interface{}{
									{"cloudType": "quark", "link": "https://pan.quark.cn/s/abc123"},
									{"cloudType": "alipan", "link": "https://www.alipan.com/s/xyz"},
								},
							},
						},
					},
				},
			})
		}
	}))
	defer server.Close()

	src := NewCloudSaverSource(server.URL, "admin", "pass", "test-token")

	result, err := src.Search("黑镜", nil, 1)
	require.NoError(t, err)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, "黑镜 第七季", result.Items[0].Title)
	assert.Equal(t, "https://pan.quark.cn/s/abc123", result.Items[0].URL)
	assert.Equal(t, "科幻美剧", result.Items[0].Summary)
	assert.Equal(t, "CloudSaver", result.Items[0].Source)
	assert.Equal(t, "quark", result.Items[0].Platform)
	assert.Equal(t, "alipan", result.Items[1].Platform)
}

func TestCloudSaver_Search_TokenExpired(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/user/login":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"data":    map[string]string{"token": "new-token"},
			})
		case "/api/search":
			callCount++
			if callCount == 1 {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"message": "无效的 token",
				})
			} else {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": true,
					"data":    []map[string]interface{}{},
				})
			}
		}
	}))
	defer server.Close()

	src := NewCloudSaverSource(server.URL, "admin", "pass", "expired-token")
	result, err := src.Search("test", nil, 1)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, callCount)
}

func TestCloudSaver_CleanResults_FilterNonQuark(t *testing.T) {
	src := &CloudSaverSource{}
	raw := []map[string]interface{}{
		{
			"channelId": "ch1",
			"list": []interface{}{
				map[string]interface{}{
					"title":   "测试资源",
					"content": "描述内容",
					"pubDate": "2025-01-15T10:30:00+08:00",
					"cloudLinks": []interface{}{
						map[string]interface{}{"cloudType": "quark", "link": "https://pan.quark.cn/s/aaa"},
						map[string]interface{}{"cloudType": "alipan", "link": "https://www.alipan.com/s/bbb"},
						map[string]interface{}{"cloudType": "quark", "link": "https://pan.quark.cn/s/aaa"}, // 重复
					},
				},
			},
		},
	}
	items := src.cleanResults(raw, []string{"quark"})
	assert.Len(t, items, 1)
	assert.Equal(t, "quark", items[0].Platform)
}
