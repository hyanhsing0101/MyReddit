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

type ParamCreatePost struct {
	BoardID int64   `json:"board_id" binding:"required"`
	TagIDs []int64 `json:"tag_ids" binding:"required"`
	Title   string  `json:"title" binding:"required"`
	Content string  `json:"content" binding:"required"`
}

type ParamPostList struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	BoardID  *int64 `form:"board_id"`
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