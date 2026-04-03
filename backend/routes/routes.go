package routes

import (
	"myreddit/controller"
	"myreddit/logger"
	"myreddit/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRouter(mode string) *gin.Engine {
	// ...不变...
	r := gin.New()
	r.Use(logger.GinLogger(), logger.GinRecovery(true))
	r.Use(middleware.CORS())

	r.POST("/signup", controller.SignUpHandler)
	r.POST("/login", controller.LoginHandler)

	r.POST("/refresh", controller.RefreshTokenHandler)

	r.GET("/ping", middleware.JWTAuthMiddleware(), func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	return r
}