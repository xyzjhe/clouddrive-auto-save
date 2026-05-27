// internal/api/notify_test.go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zcq/clouddrive-auto-save/internal/core/notify"
	"github.com/zcq/clouddrive-auto-save/internal/db"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupNotifyRouter(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	testDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	err = testDB.AutoMigrate(&db.Setting{})
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}
	db.DB = testDB

	// 初始化通知管理器
	manager := notify.NewManager()
	InitNotifyHandler(manager)

	// 注册路由
	r.GET("/api/notify", listNotifiers)
	r.GET("/api/notify/:name", getNotifier)
	r.PUT("/api/notify/:name", updateNotifier)
	r.POST("/api/notify/:name/test", testNotifier)

	return r
}

func TestNotifyAPI_ListNotifiers(t *testing.T) {
	router := setupNotifyRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/notify", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	// 响应为扁平数组格式
	var response []interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
}

func TestNotifyAPI_GetNotifier(t *testing.T) {
	router := setupNotifyRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/notify/wechat", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestNotifyAPI_UpdateNotifier(t *testing.T) {
	router := setupNotifyRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/notify/wechat", nil)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// 空请求体应返回 400
	assert.Equal(t, 400, w.Code)
}

func TestNotifyAPI_TestNotifier(t *testing.T) {
	router := setupNotifyRouter(t)

	// 测试不存在的渠道应返回错误
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/notify/not_exist/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}
