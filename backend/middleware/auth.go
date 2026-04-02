package middleware

import (
	"myreddit/controller"
	"myreddit/dao/mysql"
	"myreddit/pkg/jwt"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWTAuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			controller.ResponseError(c, controller.CodeNeedLogin)
			c.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			controller.ResponseError(c, controller.CodeInvalidToken)
			c.Abort()
			return
		}

		mc, err := jwt.ParseToken(parts[1])
		if err != nil {
			controller.ResponseError(c, controller.CodeInvalidToken)
			c.Abort()
			return
		}

		// 只允许 access token 访问受保护接口
		if mc.TokenType != jwt.TokenTypeAccess {
			controller.ResponseError(c, controller.CodeInvalidToken)
			c.Abort()
			return
		}

		ok, err := mysql.CheckUserByIDAndName(mc.UserID, mc.Username)
		if err != nil || !ok {
			controller.ResponseError(c, controller.CodeInvalidToken)
			c.Abort()
			return
		}

		c.Set(controller.ContextUserIDkey, mc.UserID)
		c.Next()
	}
}
