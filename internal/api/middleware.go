package api

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// authMiddleware 静态 Token 认证中间件
// 读取环境变量 UCAS_API_KEY，为空则跳过认证（向后兼容）；
// 否则依次检查请求头 X-API-Key 和查询参数 token，匹配则放行，否则返回 401。
func authMiddleware() gin.HandlerFunc {
	apiKey := os.Getenv("UCAS_API_KEY")

	return func(c *gin.Context) {
		// 未配置 API Key，跳过认证
		if apiKey == "" {
			c.Next()
			return
		}

		// 依次检查请求头和查询参数
		token := c.GetHeader("X-API-Key")
		if token == "" {
			token = c.Query("token")
		}

		if token == apiKey {
			c.Next()
			return
		}

		c.PureJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		c.Abort()
	}
}
