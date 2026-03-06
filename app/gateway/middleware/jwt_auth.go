package middleware

import (
	"strings"

	"poptoy-flashsale/pkg/e"
	"poptoy-flashsale/pkg/jwt"
	"poptoy-flashsale/pkg/response"

	"github.com/gin-gonic/gin"
)

// JWTAuth JWT 鉴权中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, e.Unauthorized)
			c.Abort()
			return
		}

		// 按 Bearer <token> 格式解析
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Error(c, e.Unauthorized)
			c.Abort()
			return
		}

		claims, err := jwt.ParseToken(parts[1])
		if err != nil || claims.IsRefresh {
			// 如果是 refresh token，不允许用于常规接口鉴权
			response.Error(c, e.Unauthorized)
			c.Abort()
			return
		}

		// 将 UserID 注入上下文，供后续 Controller 使用
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}