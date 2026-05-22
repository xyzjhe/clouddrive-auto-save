// internal/api/plugin_test.go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zcq/clouddrive-auto-save/internal/core/plugin"
)

func setupPluginRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 初始化插件管理器
	manager := plugin.NewManager()
	InitPluginHandler(manager)

	// 注册路由
	r.GET("/api/plugins", listPlugins)
	r.GET("/api/plugins/:name", getPlugin)
	r.PUT("/api/plugins/:name/config", updatePluginConfig)

	return r
}

func TestPluginAPI_ListPlugins(t *testing.T) {
	router := setupPluginRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/plugins", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(0), response["code"])
}

func TestPluginAPI_GetPlugin(t *testing.T) {
	router := setupPluginRouter()

	// 获取不存在的插件应返回 404
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/plugins/not_exist", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
}

func TestPluginAPI_UpdatePluginConfig(t *testing.T) {
	router := setupPluginRouter()

	// 更新配置应成功
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/plugins/test/config", nil)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// 空请求体应返回 400
	assert.Equal(t, 400, w.Code)
}
