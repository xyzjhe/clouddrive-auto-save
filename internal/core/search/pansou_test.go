// internal/core/search/pansou_test.go
package search

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 旧版 API 格式测试（kw + merged_by_type）

func TestPanSou_Legacy_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "哪吒", r.URL.Query().Get("kw"))
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"merged_by_type": map[string]interface{}{
					"quark": []map[string]interface{}{
						{
							"url":      "https://pan.quark.cn/s/abc123",
							"note":     "哪吒之魔童降世【简介】国产动画巅峰",
							"datetime": "2025-01-15T10:30:00+08:00",
							"source":   "channel1",
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	src := NewPanSouSource(server.URL)
	result, err := src.Search("哪吒", nil, 1)
	require.NoError(t, err)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, "哪吒之魔童降世", result.Items[0].Title)
	assert.Equal(t, "国产动画巅峰", result.Items[0].Summary)
	assert.Equal(t, "https://pan.quark.cn/s/abc123", result.Items[0].URL)
	assert.Equal(t, "PanSou", result.Items[0].Source)
	assert.Equal(t, "quark", result.Items[0].Platform)
	assert.Equal(t, "channel1", result.Items[0].Channel)
}

func TestPanSou_Legacy_Platform139(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "mobile", r.URL.Query().Get("cloud_types"))
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"merged_by_type": map[string]interface{}{
					"139": []map[string]interface{}{
						{
							"url":      "https://yun.139.com/s/139abc",
							"note":     "移动云盘资源【简介】139平台测试",
							"datetime": "2025-03-01T08:00:00+08:00",
							"source":   "channel139",
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	src := NewPanSouSource(server.URL)
	result, err := src.Search("test", []string{"139"}, 1)
	require.NoError(t, err)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, "139", result.Items[0].Platform)
	assert.Equal(t, "移动云盘资源", result.Items[0].Title)
	assert.Equal(t, "https://yun.139.com/s/139abc", result.Items[0].URL)
}

func TestPanSou_Legacy_PlatformFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"merged_by_type": map[string]interface{}{
					"quark": []map[string]interface{}{
						{"url": "https://pan.quark.cn/s/q1", "note": "夸克资源", "datetime": "2025-01-01T00:00:00+08:00"},
					},
					"xunlei": []map[string]interface{}{
						{"url": "https://pan.xunlei.com/s/x1", "note": "迅雷资源", "datetime": "2025-01-02T00:00:00+08:00"},
					},
				},
			},
		})
	}))
	defer server.Close()

	src := NewPanSouSource(server.URL)
	result, err := src.Search("test", []string{"quark"}, 1)
	require.NoError(t, err)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, "quark", result.Items[0].Platform)
}

func TestPanSou_Legacy_CrossPlatformDedup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"merged_by_type": map[string]interface{}{
					"quark": []map[string]interface{}{
						{"url": "https://pan.quark.cn/s/dup1", "note": "夸克版本", "datetime": "2025-01-01T00:00:00+08:00"},
					},
					"139": []map[string]interface{}{
						{"url": "https://pan.quark.cn/s/dup1", "note": "139重复", "datetime": "2025-01-02T00:00:00+08:00"},
						{"url": "https://yun.139.com/s/unique", "note": "139独有", "datetime": "2025-01-03T00:00:00+08:00"},
					},
				},
			},
		})
	}))
	defer server.Close()

	src := NewPanSouSource(server.URL)
	result, err := src.Search("test", nil, 1)
	require.NoError(t, err)
	assert.Len(t, result.Items, 2)

	urls := make(map[string]bool)
	for _, item := range result.Items {
		assert.False(t, urls[item.URL], "发现跨平台重复 URL: %s", item.URL)
		urls[item.URL] = true
	}
}

// 新版 API 格式测试（keyword + flat array）

func TestPanSou_New_Success(t *testing.T) {
	// 模拟旧版 API 返回错误，触发新版 API 回退
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		kw := r.URL.Query().Get("kw")
		keyword := r.URL.Query().Get("keyword")

		if kw != "" {
			// 旧版请求返回错误
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// 新版请求
		assert.Equal(t, "哪吒", keyword)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total": 1,
			"data": []map[string]interface{}{
				{
					"id":      1,
					"content": "名称：哪吒之魔童降世\n\n描述：国产动画巅峰\n\n链接：<a class=\"resource-link\" target=\"_blank\" href=\"https://pan.quark.cn/s/abc123\">https://pan.quark.cn/s/abc123</a>",
					"pan":     "quark",
					"image":   "",
					"time":    "2025-01-15T10:30:00+08:00",
				},
			},
			"time": "0.1s",
		})
	}))
	defer server.Close()

	src := NewPanSouSource(server.URL)
	result, err := src.Search("哪吒", nil, 1)
	require.NoError(t, err)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, "哪吒之魔童降世", result.Items[0].Title)
	assert.Equal(t, "国产动画巅峰", result.Items[0].Summary)
	assert.Equal(t, "https://pan.quark.cn/s/abc123", result.Items[0].URL)
	assert.Equal(t, "PanSou", result.Items[0].Source)
	assert.Equal(t, "quark", result.Items[0].Platform)
}

func TestPanSou_New_PlatformFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		kw := r.URL.Query().Get("kw")
		if kw != "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"total": 2,
			"data": []map[string]interface{}{
				{"id": 1, "content": "夸克资源\n\n链接：<a href=\"https://pan.quark.cn/s/q1\">link</a>", "pan": "quark", "time": "2025-01-01T00:00:00+08:00"},
				{"id": 2, "content": "迅雷资源\n\n链接：<a href=\"https://pan.xunlei.com/s/x1\">link</a>", "pan": "xunlei", "time": "2025-01-02T00:00:00+08:00"},
			},
			"time": "0.1s",
		})
	}))
	defer server.Close()

	src := NewPanSouSource(server.URL)
	result, err := src.Search("test", []string{"quark"}, 1)
	require.NoError(t, err)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, "quark", result.Items[0].Platform)
}

// 通用测试

func TestPanSou_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	src := NewPanSouSource(server.URL)
	result, err := src.Search("test", nil, 1)
	require.NoError(t, err)
	assert.Len(t, result.Items, 0)
}

func TestPanSou_Legacy_Dedup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"merged_by_type": map[string]interface{}{
					"quark": []map[string]interface{}{
						{"url": "https://pan.quark.cn/s/aaa", "note": "资源A", "datetime": "2025-01-01T00:00:00+08:00"},
						{"url": "https://pan.quark.cn/s/aaa", "note": "资源A重复", "datetime": "2025-01-02T00:00:00+08:00"},
						{"url": "https://pan.quark.cn/s/bbb", "note": "资源B", "datetime": "2025-01-03T00:00:00+08:00"},
					},
				},
			},
		})
	}))
	defer server.Close()

	src := NewPanSouSource(server.URL)
	result, err := src.Search("test", nil, 1)
	require.NoError(t, err)
	assert.Len(t, result.Items, 2)
}
