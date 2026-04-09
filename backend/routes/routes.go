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

	// 功能：新用户注册。
	// Postman：POST {{baseUrl}}/signup · Body raw JSON：{"username":"alice","password":"yourpass","re_password":"yourpass"}。
	r.POST("/signup", controller.SignUpHandler)
	// 功能：登录，返回 access_token / refresh_token。
	// Postman：POST {{baseUrl}}/login · Body raw JSON：{"username":"alice","password":"yourpass"}。
	r.POST("/login", controller.LoginHandler)

	// 功能：用 refresh_token 换新 access_token（与 refresh_token 对）。
	// Postman：POST {{baseUrl}}/refresh · Body raw JSON：{"refresh_token":"<login 返回的 refresh_token>"}。
	r.POST("/refresh", controller.RefreshTokenHandler)

	// 功能：在指定板块发帖（不可发往系统归档板）。
	// Postman：POST {{baseUrl}}/post · Bearer · Body raw JSON：{"board_id":1,"title":"标题","content":"正文"}。
	r.POST("/post", middleware.JWTAuthMiddleware(), controller.CreatePostHandler)
	// 功能：分页拉取某帖评论（未软删、按时间正序）；帖子不存在或已删返回与详情一致。
	// Postman：GET {{baseUrl}}/posts/1/comments?page=1&page_size=50。可选 Bearer：每条含 my_vote。
	r.GET("/posts/:id/comments", middleware.OptionalAuthMiddleware(), controller.ListCommentsHandler)
	// 功能：登录用户对评论投票：Body {"value":1|-1|0}，评论须属于该帖。
	// Postman：POST {{baseUrl}}/posts/1/comments/2/vote · Bearer。
	r.POST("/posts/:id/comments/:cid/vote", middleware.JWTAuthMiddleware(), controller.VoteCommentHandler)
	// 功能：登录用户发表评论；JSON 可选 parent_id 表示回复该条评论（须同属本帖）。
	// Postman：POST {{baseUrl}}/posts/1/comments · Bearer · Body：{"content":"正文","parent_id":2}（parent_id 可省略）。
	r.POST("/posts/:id/comments", middleware.JWTAuthMiddleware(), controller.CreateCommentHandler)
	// 功能：按 id 查帖子详情（含 board_id、board_slug、board_name）；已软删的帖子对所有人不可见。
	// Postman：GET {{baseUrl}}/posts/1（把 1 换成帖子 id）。可选 Bearer：带合法 access_token 时返回 my_vote / is_favorited。
	r.GET("/posts/:id", middleware.OptionalAuthMiddleware(), controller.GetPostHandler)
	// 功能：登录用户对帖子投票：Body {"value":1} 上票，{"value":-1} 下票，{"value":0} 取消。
	// Postman：POST {{baseUrl}}/posts/1/vote · Bearer · JSON 如上。
	r.POST("/posts/:id/vote", middleware.JWTAuthMiddleware(), controller.VotePostHandler)
	// 功能：编辑帖子；作者可编辑自己的帖，站点管理员可编辑任意帖；已删帖不可编辑。
	// Postman：POST {{baseUrl}}/posts/1/edit · Bearer · Body {"title":"新标题","content":"新正文","tag_ids":[1,2]}。
	r.POST("/posts/:id/edit", middleware.JWTAuthMiddleware(), controller.UpdatePostHandler)
	// 功能：登录用户收藏帖子（重复请求幂等）。
	// Postman：POST {{baseUrl}}/posts/1/favorite · Bearer。
	r.POST("/posts/:id/favorite", middleware.JWTAuthMiddleware(), controller.AddPostFavoriteHandler)
	// 功能：取消收藏帖子（未收藏也返回成功）。统一用 POST，避免部分环境对 DELETE 跨域/网关限制。
	// Postman：POST {{baseUrl}}/posts/1/unfavorite · Bearer。
	r.POST("/posts/:id/unfavorite", middleware.JWTAuthMiddleware(), controller.RemovePostFavoriteHandler)
	// 功能：软删帖子；作者可删自己的帖，站点管理员可删任意帖；无主帖仅管理员可删。
	// Postman：DELETE {{baseUrl}}/posts/1 · Bearer。
	r.DELETE("/posts/:id", middleware.JWTAuthMiddleware(), controller.DeletePostHandler)
	// 功能：分页帖子列表；可选 board_id 只拉该板帖子；sort 支持 new|top|hot（默认 new）。
	// Postman：GET {{baseUrl}}/posts?page=1&page_size=10&sort=hot · sort=new|hot|top；可选 &board_id=1。前端首页/板块「最新·热门·高分」与此一致。可选 Bearer：带合法 access_token 时每条含 my_vote / is_favorited。
	r.GET("/posts", middleware.OptionalAuthMiddleware(), controller.ListPostHandler)

	// 功能：分页获取全站标签。
	// Postman：GET {{baseUrl}}/tags?page=1&page_size=50。
	r.GET("/tags", controller.ListTagsHandler)

	// 功能：按 slug 查板块详情。可选 Bearer：返回 is_favorited。
	// Postman：GET {{baseUrl}}/boards/slug/general（把 general 换成板块 slug）。
	r.GET("/boards/slug/:slug", middleware.OptionalAuthMiddleware(), controller.GetBoardBySlugHandler)
	// 功能：登录用户收藏板块（重复请求幂等）；系统归档板不可收藏。
	// Postman：POST {{baseUrl}}/boards/1/favorite · Bearer。
	r.POST("/boards/:id/favorite", middleware.JWTAuthMiddleware(), controller.AddBoardFavoriteHandler)
	// 功能：取消收藏板块（未收藏也返回成功）。统一用 POST，避免部分环境对 DELETE 跨域/网关限制。
	// Postman：POST {{baseUrl}}/boards/1/unfavorite · Bearer。
	r.POST("/boards/:id/unfavorite", middleware.JWTAuthMiddleware(), controller.RemoveBoardFavoriteHandler)
	// 功能：按数字 id 查板块详情。可选 Bearer：返回 is_favorited。
	// Postman：GET {{baseUrl}}/boards/1（把 1 换成板块 id）。
	r.GET("/boards/:id", middleware.OptionalAuthMiddleware(), controller.GetBoardByIDHandler)
	// 功能：分页板块列表；include_system_sink=true 时包含系统归档板 _archived。可选 Bearer：每条含 is_favorited。
	// Postman：GET {{baseUrl}}/boards?page=1&page_size=20 · 可选 &include_system_sink=true。
	r.GET("/boards", middleware.OptionalAuthMiddleware(), controller.ListBoardsHandler)
	// 功能：登录用户创建板块（slug 小写+数字+下划线，且不能占用保留名 _archived）。
	// Postman：POST {{baseUrl}}/boards · Bearer · Body raw JSON：{"slug":"my_board","name":"展示名","description":"可选"}。
	r.POST("/boards", middleware.JWTAuthMiddleware(), controller.CreateBoardHandler)

	// 功能：全站搜索（FTS，scope=all|posts|boards 控制范围）。
	// Postman：GET {{baseUrl}}/search?q=t1&scope=posts&post_limit=20&board_limit=10。
	r.GET("/search", controller.SearchHandler)

	// 任意登录用户：查看自己的权限
	r.GET("/me/permissions", middleware.JWTAuthMiddleware(), controller.MePermissionsHandler)
	// 功能：当前用户收藏的板块列表（按收藏时间倒序分页）。
	// Postman：GET {{baseUrl}}/me/favorite-boards?page=1&page_size=20 · Bearer。
	r.GET("/me/favorite-boards", middleware.JWTAuthMiddleware(), controller.ListMyFavoriteBoardsHandler)
	// 功能：当前用户收藏的帖子列表（按收藏时间倒序分页）。
	// Postman：GET {{baseUrl}}/me/favorite-posts?page=1&page_size=20 · Bearer。
	r.GET("/me/favorite-posts", middleware.JWTAuthMiddleware(), controller.ListMyFavoritePostsHandler)
	// 测试：仅登录
	r.GET("/debug/auth/any", middleware.JWTAuthMiddleware(), controller.DebugAuthAnyHandler)
	// 测试：仅站点管理员（注意中间件顺序）
	r.GET("/debug/auth/admin", middleware.JWTAuthMiddleware(), middleware.RequireSiteAdmin(), controller.DebugAuthAdminHandler)

	// 功能：鉴权探活，成功返回纯文本 pong。
	// Postman：GET {{baseUrl}}/ping · Bearer（任意有效 access_token）。
	r.GET("/ping", middleware.JWTAuthMiddleware(), func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	return r
}
