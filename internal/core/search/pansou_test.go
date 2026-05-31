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

func TestPanSou_Search_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "哪吒", r.URL.Query().Get("kw"))
		assert.Equal(t, "quark", r.URL.Query().Get("cloud_types"))
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

func TestPanSou_Search_EmptyNote(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"merged_by_type": map[string]interface{}{
					"quark": []map[string]interface{}{
						{
							"url":      "https://pan.quark.cn/s/xyz",
							"note":     "简单标题没有简介",
							"datetime": "2025-02-01T12:00:00+08:00",
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	src := NewPanSouSource(server.URL)
	result, err := src.Search("test", nil, 1)
	require.NoError(t, err)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, "简单标题没有简介", result.Items[0].Title)
	assert.Equal(t, "", result.Items[0].Summary)
}

func TestPanSou_Search_Dedup(t *testing.T) {
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

func TestPanSou_Search_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	src := NewPanSouSource(server.URL)
	result, err := src.Search("test", nil, 1)
	require.NoError(t, err)
	assert.Len(t, result.Items, 0)
}
