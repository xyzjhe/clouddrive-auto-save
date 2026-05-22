// internal/core/notify/wechat_test.go
package notify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWeChatNotifier_Init(t *testing.T) {
	notifier := NewWeChatNotifier()

	// 初始化应成功
	config := map[string]interface{}{
		"webhook_url": "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test",
	}
	err := notifier.Init(config)
	require.NoError(t, err)

	// 空 webhook_url 应返回错误
	config = map[string]interface{}{}
	err = notifier.Init(config)
	assert.Error(t, err)
}

func TestWeChatNotifier_Send(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求方法
		assert.Equal(t, "POST", r.Method)

		// 验证请求体
		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)

		// 验证消息类型
		assert.Equal(t, "markdown", body["msgtype"])

		// 返回成功响应
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"errcode": 0})
	}))
	defer server.Close()

	notifier := NewWeChatNotifier()
	notifier.Init(map[string]interface{}{
		"webhook_url": server.URL,
	})

	// 发送消息应成功
	message := &Message{
		Title:   "测试标题",
		Content: "测试内容",
		Level:   LevelInfo,
	}
	err := notifier.Send(context.Background(), message)
	require.NoError(t, err)
}

func TestWeChatNotifier_Test(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"errcode": 0})
	}))
	defer server.Close()

	notifier := NewWeChatNotifier()
	notifier.Init(map[string]interface{}{
		"webhook_url": server.URL,
	})

	// 测试应成功
	err := notifier.Test(context.Background())
	require.NoError(t, err)
}

func TestWeChatNotifier_Close(t *testing.T) {
	notifier := NewWeChatNotifier()
	err := notifier.Close()
	require.NoError(t, err)
}
