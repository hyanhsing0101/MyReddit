package middleware

import (
	"myreddit/controller"
	"myreddit/dao/postgres"

	"github.com/gin-gonic/gin"
)

// RequireSiteAdmin 必须在 JWTAuthMiddleware 之后使用
func RequireSiteAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, err := controller.GetCurrentUser(c)
		if err != nil {
			controller.ResponseError(c, controller.CodeNeedLogin)
			c.Abort()
			return
		}
		ok, err := postgres.IsSiteAdmin(uid)
		if err != nil {
			controller.ResponseError(c, controller.CodeServerBusy)
			c.Abort()
			return
		}
		if !ok {
			controller.ResponseError(c, controller.CodeForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}