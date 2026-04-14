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

// PostFeed 帖子列表范围：全站或仅「已收藏板块」（订阅流，数据源为 board_favorite）。
type PostFeed string

const (
	PostFeedAll        PostFeed = "all"
	PostFeedSubscribed PostFeed = "subscribed"
)

type ParamPostList struct {
	Page     int      `form:"page"`
	PageSize int      `form:"page_size"`
	BoardID  *int64   `form:"board_id"`
	Sort     PostSort `form:"sort"`
	// Feed 可选：空或 all 为全站；subscribed 为仅已收藏板块下的帖子（须登录，且不可与 board_id 同用）。
	Feed string `form:"feed"`
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
	switch PostFeed(p.Feed) {
	case PostFeedSubscribed:
		p.Feed = string(PostFeedSubscribed)
	default:
		p.Feed = string(PostFeedAll)
	}
}

// SubscribedFeed 是否请求订阅流（已收藏板块下的帖子）。
func (p *ParamPostList) SubscribedFeed() bool {
	return PostFeed(p.Feed) == PostFeedSubscribed
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
	// Visibility 可选，默认 public；仅支持 public | private。
	Visibility string `json:"visibility" binding:"omitempty,oneof=public private"`
}

type ParamFavoriteBoardList struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

type ParamAddBoardModerator struct {
	UserID int64  `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=owner moderator"`
}

type ParamUpdateBoardModeratorRole struct {
	Role string `json:"role" binding:"required,oneof=owner moderator"`
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
// 举报
// =============================================================================

type ParamCreatePostReport struct {
	Reason string `json:"reason" binding:"required,max=120"`
	Detail string `json:"detail" binding:"omitempty,max=1000"`
}

type ParamReportList struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Status   string `form:"status"`
}

func (p *ParamReportList) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
	switch PostReportStatus(p.Status) {
	case PostReportStatusOpen, PostReportStatusReview, PostReportStatusResolved, PostReportStatusRejected:
	default:
		p.Status = ""
	}
}

type ParamUpdatePostReport struct {
	Status      string `json:"status" binding:"required,oneof=open in_review resolved rejected"`
	HandlerNote string `json:"handler_note" binding:"omitempty,max=1000"`
}

type ParamBatchUpdatePostReports struct {
	ReportIDs   []int64 `json:"report_ids" binding:"required,min=1,max=100,dive,gt=0"`
	Status      string  `json:"status" binding:"required,oneof=open in_review resolved rejected"`
	HandlerNote string  `json:"handler_note" binding:"omitempty,max=1000"`
}

// ParamBatchUpdateCommentReports 与帖子举报批量结构一致（JSON 字段名 report_ids）。
type ParamBatchUpdateCommentReports struct {
	ReportIDs   []int64 `json:"report_ids" binding:"required,min=1,max=100,dive,gt=0"`
	Status      string  `json:"status" binding:"required,oneof=open in_review resolved rejected"`
	HandlerNote string  `json:"handler_note" binding:"omitempty,max=1000"`
}

type ParamCreatePostAppeal struct {
	Reason         string `json:"reason" binding:"required,max=500"`
	RequestedTitle string `json:"requested_title" binding:"required,max=300"`
	RequestedBody  string `json:"requested_content" binding:"required,max=20000"`
	UserReply      string `json:"user_reply" binding:"omitempty,max=2000"`
}

type ParamPostAppealList struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Status   string `form:"status"`
}

func (p *ParamPostAppealList) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
	switch PostAppealStatus(p.Status) {
	case PostAppealStatusOpen, PostAppealStatusReview, PostAppealStatusApproved, PostAppealStatusRejected:
	default:
		p.Status = ""
	}
}

type ParamHandlePostAppeal struct {
	Status         string `json:"status" binding:"required,oneof=in_review approved rejected"`
	ModeratorReply string `json:"moderator_reply" binding:"omitempty,max=2000"`
	ApplyUpdate    bool   `json:"apply_update"`
}

type ParamModerationLogList struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
	Action     string `form:"action"`
	TargetType string `form:"target_type"`
	TargetID   *int64 `form:"target_id"`
}

func (p *ParamModerationLogList) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
	switch ModerationAction(p.Action) {
	case ModerationActionSealPost,
		ModerationActionUnsealPost,
		ModerationActionDeletePost,
		ModerationActionRestorePost,
		ModerationActionHandlePostAppeal,
		ModerationActionLockPostComments,
		ModerationActionUnlockPostComments,
		ModerationActionPinPost,
		ModerationActionUnpinPost,
		ModerationActionResolvePostReport,
		ModerationActionResolveCommentReport,
		ModerationActionUpsertBoardModerator,
		ModerationActionUpdateBoardModerator,
		ModerationActionRemoveBoardModerator:
	default:
		p.Action = ""
	}
	switch ModerationTargetType(p.TargetType) {
	case ModerationTargetPost, ModerationTargetPostReport, ModerationTargetCommentReport, ModerationTargetModerator:
	default:
		p.TargetType = ""
	}
	if p.TargetID != nil && *p.TargetID < 1 {
		p.TargetID = nil
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
	PostPage        int `form:"post_page"`
	PostPageSize    int `form:"post_page_size"`
	CommentPage     int `form:"comment_page"`
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
