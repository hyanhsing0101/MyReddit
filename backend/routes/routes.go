package routes

import (
	"myreddit/controller"
	"myreddit/logger"
	"myreddit/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRouter 注册 HTTP 路由。
// Postman 通用：环境变量 baseUrl=http://127.0.0.1:8081（以实际服务地址为准）；POST JSON 接口在 Headers 加 Content-Type: application/json；标「Bearer」的接口在 Authorization 选 Bearer Token，填登录返回的 access_token。
func SetupRouter(mode string) *gin.Engine {
	r := gin.New()
	r.Use(logger.GinLogger(), logger.GinRecovery(true))
	r.Use(middleware.CORS())

	// === Auth ===
	r.POST("/signup", controller.SignUpHandler)
	r.POST("/login", controller.LoginHandler)
	r.POST("/refresh", controller.RefreshTokenHandler)

	// === Posts ===
	// 主路由（REST 风格）
	r.GET("/posts", middleware.OptionalAuthMiddleware(), controller.ListPostHandler)
	r.POST("/posts", middleware.JWTAuthMiddleware(), controller.CreatePostHandler)
	r.GET("/posts/:id", middleware.OptionalAuthMiddleware(), controller.GetPostHandler)
	r.PATCH("/posts/:id", middleware.JWTAuthMiddleware(), controller.UpdatePostHandler)
	r.DELETE("/posts/:id", middleware.JWTAuthMiddleware(), controller.DeletePostHandler)

	r.POST("/posts/:id/vote", middleware.JWTAuthMiddleware(), controller.VotePostHandler)
	r.POST("/posts/:id/seal", middleware.JWTAuthMiddleware(), controller.SealPostHandler)
	r.DELETE("/posts/:id/seal", middleware.JWTAuthMiddleware(), controller.UnsealPostHandler)

	r.POST("/posts/:id/favorites", middleware.JWTAuthMiddleware(), controller.AddPostFavoriteHandler)
	r.DELETE("/posts/:id/favorites", middleware.JWTAuthMiddleware(), controller.RemovePostFavoriteHandler)

	r.GET("/posts/:id/comments", middleware.OptionalAuthMiddleware(), controller.ListCommentsHandler)
	r.POST("/posts/:id/comments", middleware.JWTAuthMiddleware(), controller.CreateCommentHandler)
	r.POST("/posts/:id/comments/:cid/vote", middleware.JWTAuthMiddleware(), controller.VoteCommentHandler)

	// 兼容旧路由（可逐步下线）
	r.POST("/post", middleware.JWTAuthMiddleware(), controller.CreatePostHandler)
	r.POST("/posts/:id/edit", middleware.JWTAuthMiddleware(), controller.UpdatePostHandler)
	r.POST("/posts/:id/unseal", middleware.JWTAuthMiddleware(), controller.UnsealPostHandler)
	r.POST("/posts/:id/favorite", middleware.JWTAuthMiddleware(), controller.AddPostFavoriteHandler)
	r.POST("/posts/:id/unfavorite", middleware.JWTAuthMiddleware(), controller.RemovePostFavoriteHandler)

	// === Boards ===
	r.GET("/boards", middleware.OptionalAuthMiddleware(), controller.ListBoardsHandler)
	r.POST("/boards", middleware.JWTAuthMiddleware(), controller.CreateBoardHandler)
	r.GET("/boards/:id", middleware.OptionalAuthMiddleware(), controller.GetBoardByIDHandler)
	r.GET("/boards/by-slug/:slug", middleware.OptionalAuthMiddleware(), controller.GetBoardBySlugHandler)

	r.POST("/boards/:id/favorites", middleware.JWTAuthMiddleware(), controller.AddBoardFavoriteHandler)
	r.DELETE("/boards/:id/favorites", middleware.JWTAuthMiddleware(), controller.RemoveBoardFavoriteHandler)

	r.GET("/boards/:id/moderators", middleware.JWTAuthMiddleware(), controller.ListBoardModeratorsHandler)
	r.POST("/boards/:id/moderators", middleware.JWTAuthMiddleware(), controller.AddBoardModeratorHandler)
	r.PATCH("/boards/:id/moderators/:uid", middleware.JWTAuthMiddleware(), controller.UpdateBoardModeratorRoleHandler)
	r.DELETE("/boards/:id/moderators/:uid", middleware.JWTAuthMiddleware(), controller.RemoveBoardModeratorHandler)

	// 兼容旧路由（可逐步下线）
	r.GET("/boards/slug/:slug", middleware.OptionalAuthMiddleware(), controller.GetBoardBySlugHandler)
	r.POST("/boards/:id/favorite", middleware.JWTAuthMiddleware(), controller.AddBoardFavoriteHandler)
	r.POST("/boards/:id/unfavorite", middleware.JWTAuthMiddleware(), controller.RemoveBoardFavoriteHandler)
	r.POST("/boards/:id/moderators/:uid/role", middleware.JWTAuthMiddleware(), controller.UpdateBoardModeratorRoleHandler)

	// === Discovery ===
	r.GET("/tags", controller.ListTagsHandler)
	r.GET("/search", middleware.OptionalAuthMiddleware(), controller.SearchHandler)

	// === Users ===
	r.GET("/users/:id/home", middleware.OptionalAuthMiddleware(), controller.GetUserHomeHandler)

	// === Me ===
	r.GET("/me/permissions", middleware.JWTAuthMiddleware(), controller.MePermissionsHandler)
	r.GET("/me/favorite-boards", middleware.JWTAuthMiddleware(), controller.ListMyFavoriteBoardsHandler)
	r.GET("/me/favorite-posts", middleware.JWTAuthMiddleware(), controller.ListMyFavoritePostsHandler)

	// === Debug ===
	r.GET("/debug/auth/any", middleware.JWTAuthMiddleware(), controller.DebugAuthAnyHandler)
	r.GET("/debug/auth/admin", middleware.JWTAuthMiddleware(), middleware.RequireSiteAdmin(), controller.DebugAuthAdminHandler)

	// === Health ===
	r.GET("/ping", middleware.JWTAuthMiddleware(), func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	return r
}
