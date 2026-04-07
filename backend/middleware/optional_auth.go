package middleware

import (
	"myreddit/controller"
	"myreddit/dao/postgres"
	"myreddit/pkg/jwt"
	"strings"

	"github.com/gin-gonic/gin"
)

// OptionalAuthMiddleware 若带合法 Bearer access_token 则写入 UserID；否则等同匿名，不中断请求。
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}
		mc, err := jwt.ParseToken(parts[1])
		if err != nil || mc.TokenType != jwt.TokenTypeAccess {
			c.Next()
			return
		}
		ok, err := postgres.CheckUserByIDAndName(mc.UserID, mc.Username)
		if err != nil || !ok {
			c.Next()
			return
		}
		c.Set(controller.ContextUserIDkey, mc.UserID)
		c.Next()
	}
}
