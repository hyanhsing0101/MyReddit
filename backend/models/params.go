package models

// =============================================================================
// 用户 / 鉴权（注册、登录、刷新 Token）
// =============================================================================

type ParamSignUp struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	RePassword string `json:"re_password" binding:"required,eqfield=Password"`
}

type ParamLogin struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ParamRefresh struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// =============================================================================
// 帖子
// =============================================================================

type PostSort string

const (
	PostSortNew PostSort = "new"
	PostSortHot PostSort = "hot"
	PostSortTop PostSort = "top"
)

type ParamCreatePost struct {
	BoardID int64   `json:"board_id" binding:"required"`
	TagIDs  []int64 `json:"tag_ids" binding:"required"`
	Title   string  `json:"title" binding:"required"`
	Content string  `json:"content" binding:"required"`
}

type ParamUpdatePost struct {
	TagIDs  []int64 `json:"tag_ids" binding:"required"`
	Title   string  `json:"title" binding:"required"`
	Content string  `json:"content" binding:"required"`
}

type ParamPostList struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	BoardID  *int64 `form:"board_id"`
	Sort     PostSort `form:"sort"`
}

func (p *ParamPostList) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 10
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
	switch p.Sort {
	case PostSortNew:
		p.Sort = PostSortNew
	case PostSortHot:
		p.Sort = PostSortHot
	case PostSortTop:
		p.Sort = PostSortTop
	default:
		p.Sort = PostSortNew
	}
}

type ParamFavoritePostList struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

func (p *ParamFavoritePostList) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
}

// ParamVotePost value：1 上票，-1 下票，0 取消投票（须显式传，不可省略）。
type ParamVotePost struct {
	Value *int8 `json:"value" binding:"required"`
}

// =============================================================================
// 板块
// =============================================================================

type ParamBoardList struct {
	Page              int  `form:"page"`
	PageSize          int  `form:"page_size"`
	IncludeSystemSink bool `form:"include_system_sink"`
}

func (p *ParamBoardList) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
}

type ParamCreateBoard struct {
	Slug        string `json:"slug" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type ParamFavoriteBoardList struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

func (p *ParamFavoriteBoardList) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
}

// =============================================================================
// 评论
// =============================================================================

type ParamCreateComment struct {
	Content  string `json:"content" binding:"required"`
	ParentID *int64 `json:"parent_id"`
}

type ParamCommentList struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

func (p *ParamCommentList) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 50
	}
	if p.PageSize > 200 {
		p.PageSize = 200
	}
}

// =============================================================================
// 搜索
// =============================================================================

type ParamSearch struct {
	Q          string `form:"q"`
	Scope      string `form:"scope"`
	PostLimit  int    `form:"post_limit"`
	BoardLimit int    `form:"board_limit"`
}

func (p *ParamSearch) Normalize() {
	if len(p.Q) > 200 {
		p.Q = p.Q[:200]
	}
	if p.Scope == "" {
		p.Scope = "all"
	}
	if p.Scope != "all" && p.Scope != "posts" && p.Scope != "boards" {
		p.Scope = "all"
	}

	if p.PostLimit < 1 {
		p.PostLimit = 20
	}
	if p.PostLimit > 50 {
		p.PostLimit = 50
	}
	if p.BoardLimit < 1 {
		p.BoardLimit = 10
	}
	if p.BoardLimit > 30 {
		p.BoardLimit = 30
	}
}

// =============================================================================
// 标签
// =============================================================================

type ParamTagList struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

func (p *ParamTagList) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 50
	}
	if p.PageSize > 200 {
		p.PageSize = 200
	}
}

// =============================================================================
// 用户主页
// =============================================================================

type ParamUserHome struct {
	PostPage       int `form:"post_page"`
	PostPageSize   int `form:"post_page_size"`
	CommentPage    int `form:"comment_page"`
	CommentPageSize int `form:"comment_page_size"`
}

func (p *ParamUserHome) Normalize() {
	if p.PostPage < 1 {
		p.PostPage = 1
	}
	if p.PostPageSize < 1 {
		p.PostPageSize = 10
	}
	if p.PostPageSize > 50 {
		p.PostPageSize = 50
	}
	if p.CommentPage < 1 {
		p.CommentPage = 1
	}
	if p.CommentPageSize < 1 {
		p.CommentPageSize = 10
	}
	if p.CommentPageSize > 50 {
		p.CommentPageSize = 50
	}
}
