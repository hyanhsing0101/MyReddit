package logic

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"myreddit/dao/postgres"
	redisDao "myreddit/dao/redis"
	"myreddit/models"
	"time"
)

var (
	ErrCannotPostToSystemBoard = errors.New("cannot post to system board")
	ErrInvalidBoardID          = errors.New("invalid board id")
	ErrDeletePostForbidden     = errors.New("delete post forbidden")
	ErrEditPostForbidden       = errors.New("edit post forbidden")
	ErrTagCountExceedsMaxLimit = errors.New("tag count exceeds the maximum limit")
	ErrNotBoardMember          = errors.New("not board member")
	ErrPostSealed              = errors.New("post sealed")
	ErrPostCommentsLocked      = errors.New("post comments locked")
	ErrSealPostForbidden       = errors.New("seal post forbidden")
	ErrUnsealPostForbidden     = errors.New("unseal post forbidden")
	ErrLockPostForbidden       = errors.New("lock post forbidden")
	ErrUnlockPostForbidden     = errors.New("unlock post forbidden")
	ErrPinPostForbidden        = errors.New("pin post forbidden")
	ErrUnpinPostForbidden      = errors.New("unpin post forbidden")
	// ErrSubscribedFeedNeedLogin 使用 feed=subscribed 但未登录。
	ErrSubscribedFeedNeedLogin = errors.New("subscribed feed requires login")
	// ErrSubscribedFeedWithBoardID feed=subscribed 与 board_id 不能同时使用。
	ErrSubscribedFeedWithBoardID = errors.New("subscribed feed cannot combine with board_id")
	ErrPostNotSoftDeleted        = errors.New("post not soft deleted")
)

const MaxPostTagCount = 5

func calcHotScore(score int64, createdAt time.Time) float64 {
	s := score
	if s < -50 {
		s = -50
	}
	hours := time.Since(createdAt).Hours()
	if hours < 0 {
		hours = 0
	}
	return float64(s) / math.Pow(hours+2.0, 1.8)
}

// CreatePost 创建帖子并同步维护标签与热榜缓存（缓存失败不影响主流程）。
func CreatePost(p *models.ParamCreatePost, userID int64) error {
	board, err := postgres.GetBoardByID(p.BoardID)
	if err != nil {
		return err
	}
	if board.IsSystemSink {
		return ErrCannotPostToSystemBoard
	}
	okPost, err := canPostToBoard(userID, board)
	if err != nil {
		return err
	}
	if !okPost {
		return ErrNotBoardMember
	}
	tagIDs, err := postgres.ValidateTagIDs(p.TagIDs)
	if err != nil {
		return err
	}
	if len(tagIDs) > MaxPostTagCount {
		return ErrTagCountExceedsMaxLimit
	}
	post := models.Post{
		BoardID:    p.BoardID,
		Title:      p.Title,
		Content:    p.Content,
		AuthorID:   sql.NullInt64{Int64: userID, Valid: true},
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	postid, err := postgres.CreatePost(&post)
	if err != nil {
		return err
	}
	if err := postgres.ReplacePostTags(postid, tagIDs); err != nil {
		return err
	}
	// 热榜缓存为加速层：写失败不影响主流程。
	_ = redisDao.UpsertHotPost(postid, calcHotScore(0, post.CreateTime))
	return nil
}

// ListPost 分页获取帖子列表，并按登录态附加 my_vote / is_favorited。
func ListPost(p *models.ParamPostList, viewerID *int64) (*models.PostListData, error) {
	p.Normalize()
	if p.SubscribedFeed() {
		if p.BoardID != nil {
			return nil, ErrSubscribedFeedWithBoardID
		}
		if viewerID == nil {
			return nil, ErrSubscribedFeedNeedLogin
		}
	}
	r, err := postReader(viewerID)
	if err != nil {
		return nil, err
	}
	var boardFilter *int64
	if p.BoardID != nil {
		if *p.BoardID < 1 {
			return nil, ErrInvalidBoardID
		}
		board, err := postgres.GetBoardByID(*p.BoardID)
		if err != nil {
			return nil, err
		}
		ok, err := canReadBoard(viewerID, board)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, postgres.ErrorBoardNotExist
		}
		boardFilter = p.BoardID
	}
	subscribedOnly := p.SubscribedFeed()
	total, err := postgres.CountPosts(boardFilter, subscribedOnly, r)
	if err != nil {
		return nil, err
	}
	offset := (p.Page - 1) * p.PageSize
	var posts []models.Post
	// 全站 hot 优先走 Redis 热榜；分板块、订阅流或其他排序直接走 PG。
	if p.Sort == models.PostSortHot && boardFilter == nil && !subscribedOnly {
		cachedIDs, err := redisDao.GetHotPostIDs(p.Page, p.PageSize)
		if err == nil && len(cachedIDs) > 0 {
			posts, err = postgres.ListPostsByIDsOrdered(cachedIDs, r)
			if err != nil {
				return nil, err
			}
		}
		// 缓存未命中（或命中后被过滤空）则回源 PG，再回填一页缓存。
		if len(posts) == 0 {
			posts, err = postgres.ListPosts(boardFilter, subscribedOnly, p.Sort, p.PageSize, offset, r)
			if err != nil {
				return nil, err
			}
			scores := make(map[int64]float64, len(posts))
			for _, row := range posts {
				scores[row.ID] = calcHotScore(row.Score, row.CreateTime)
			}
			_ = redisDao.SetHotPostScores(scores)
		}
	} else {
		posts, err = postgres.ListPosts(boardFilter, subscribedOnly, p.Sort, p.PageSize, offset, r)
		if err != nil {
			return nil, err
		}
	}
	// 批量取标签，避免逐帖查询导致 N+1。
	postIDs := make([]int64, len(posts))
	for i := range posts {
		postIDs[i] = posts[i].ID
	}
	tagsByPost, err := postgres.GetTagsByPostIDs(postIDs)
	if err != nil {
		return nil, err
	}

	list := make([]models.PostView, 0, len(posts))
	for _, row := range posts {
		v := models.PostToView(row)
		if t := tagsByPost[row.ID]; t != nil {
			v.Tags = t
		} else {
			v.Tags = []models.Tag{}
		}
		list = append(list, v)
	}
	if viewerID != nil {
		if err := attachPostMyVotes(list, *viewerID); err != nil {
			return nil, err
		}
		if err := attachPostFavoriteFlags(list, *viewerID); err != nil {
			return nil, err
		}
	}
	return &models.PostListData{
		List:     list,
		Total:    total,
		Page:     p.Page,
		PageSize: p.PageSize,
	}, nil
}

// GetPost 获取单个帖子详情，并按登录态附加 my_vote / is_favorited。
func GetPost(id int64, viewerID *int64) (*models.PostView, error) {
	r, err := postReader(viewerID)
	if err != nil {
		return nil, err
	}
	post, err := postgres.GetPostByID(id, r)
	if err != nil {
		return nil, err
	}
	v := models.PostToView(*post)
	Tags, err := postgres.GetTagsByPostID(post.ID)
	if err != nil {
		return nil, err
	}
	v.Tags = Tags
	if viewerID != nil {
		slice := []models.PostView{v}
		if err := attachPostMyVotes(slice, *viewerID); err != nil {
			return nil, err
		}
		if err := attachPostFavoriteFlags(slice, *viewerID); err != nil {
			return nil, err
		}
		v = slice[0]
		act, err := moderationActionsForPost(post, *viewerID)
		if err != nil {
			return nil, err
		}
		if act != nil && (act.CanSeal || act.CanUnseal || act.CanLockComments || act.CanUnlockComments || act.CanPin || act.CanUnpin) {
			v.ModerationActions = act
		}
	}
	return &v, nil
}

// attachPostMyVotes 批量附加当前用户对列表帖子投票状态。
func attachPostMyVotes(list []models.PostView, userID int64) error {
	if len(list) == 0 {
		return nil
	}
	ids := make([]int64, len(list))
	for i := range list {
		ids[i] = list[i].ID
	}
	m, err := postgres.GetPostVotesForUser(userID, ids)
	if err != nil {
		return err
	}
	for i := range list {
		if val, ok := m[list[i].ID]; ok {
			v := val
			list[i].MyVote = &v
		}
	}
	return nil
}

// attachPostFavoriteFlags 批量附加当前用户对列表帖子的收藏状态。
func attachPostFavoriteFlags(list []models.PostView, userID int64) error {
	if len(list) == 0 {
		return nil
	}
	ids := make([]int64, len(list))
	for i := range list {
		ids[i] = list[i].ID
	}
	m, err := postgres.ListPostIDsFavoritedByUser(userID, ids)
	if err != nil {
		return err
	}
	for i := range list {
		if _, ok := m[list[i].ID]; ok {
			t := true
			list[i].IsFavorited = &t
		} else {
			f := false
			list[i].IsFavorited = &f
		}
	}
	return nil
}

// VotePost 上票/下票/取消；返回最新 score 与 my_vote（取消后为 null）。
func VotePost(postID, userID int64, value int8) (*models.PostVoteResult, error) {
	r, err := postReader(&userID)
	if err != nil {
		return nil, err
	}
	if _, err := postgres.GetPostByID(postID, r); err != nil {
		return nil, err
	}
	postFull, err := postgres.GetPostByIDIncludingDeleted(postID)
	if err != nil {
		return nil, err
	}
	if postFull.SealedAt.Valid {
		return nil, ErrPostSealed
	}
	score, myVote, err := postgres.ApplyPostVote(postID, userID, value)
	if err != nil {
		return nil, err
	}
	// 更新热榜缓存分数（失败可容忍，后续读路径会回源修复）。
	if p, err := postgres.GetPostByID(postID, r); err == nil {
		_ = redisDao.UpsertHotPost(postID, calcHotScore(score, p.CreateTime))
	}
	return &models.PostVoteResult{
		Score:  score,
		MyVote: myVote,
	}, nil
}

// DeletePost 软删：作者可删自己帖子；站主可删任意帖；版主可删本板帖子。
func DeletePost(postID, operatorUserID int64) error {
	post, err := postgres.GetPostByIDIncludingDeleted(postID)
	if err != nil {
		return err
	}
	if post.DeletedAt.Valid {
		return postgres.ErrorPostNotExist
	}
	admin, err := postgres.IsSiteAdmin(operatorUserID)
	if err != nil {
		return err
	}
	isAuthor := post.AuthorID.Valid && post.AuthorID.Int64 == operatorUserID
	isModerator := false
	if !admin {
		isModerator, err = canModerateBoard(operatorUserID, post.BoardID)
		if err != nil {
			return err
		}
	}
	if !isAuthor && !admin && !isModerator {
		return ErrDeletePostForbidden
	}
	if err := postgres.SoftDeletePost(postID, time.Now()); err != nil {
		return err
	}
	if admin || isModerator {
		desc := "by=moderator"
		if admin {
			desc = "by=site_admin"
		}
		appendModerationLog(
			post.BoardID,
			operatorUserID,
			models.ModerationActionDeletePost,
			models.ModerationTargetPost,
			postID,
			desc,
		)
	}
	_ = redisDao.RemoveHotPost(postID)
	return nil
}

// UpdatePost 编辑帖子：作者本人或站点管理员；无主帖仅管理员可编辑；已软删返回 post not exist
func UpdatePost(postID, operatorUserID int64, p *models.ParamUpdatePost) error {
	post, err := postgres.GetPostByIDIncludingDeleted(postID)
	if err != nil {
		return err
	}
	if post.DeletedAt.Valid {
		return postgres.ErrorPostNotExist
	}
	admin, err := postgres.IsSiteAdmin(operatorUserID)
	if err != nil {
		return err
	}
	if !admin {
		if !post.AuthorID.Valid || post.AuthorID.Int64 != operatorUserID {
			return ErrEditPostForbidden
		}
	}
	tagIDs, err := postgres.ValidateTagIDs(p.TagIDs)
	if err != nil {
		return err
	}
	if len(tagIDs) > MaxPostTagCount {
		return ErrTagCountExceedsMaxLimit
	}
	now := time.Now()
	if err := postgres.UpdatePostContent(postID, p.Title, p.Content, now); err != nil {
		return err
	}
	if err := postgres.ReplacePostTags(postID, tagIDs); err != nil {
		return err
	}
	return nil
}

// AddPostFavorite 收藏帖子（帖子不存在时返回错误）。
func AddPostFavorite(userID, postID int64) error {
	r, err := postReader(&userID)
	if err != nil {
		return err
	}
	if _, err := postgres.GetPostByID(postID, r); err != nil {
		return err
	}
	return postgres.AddPostFavorite(userID, postID)
}

// RemovePostFavorite 取消用户对帖子的收藏。
func RemovePostFavorite(userID, postID int64) error {
	r, err := postReader(&userID)
	if err != nil {
		return err
	}
	if _, err := postgres.GetPostByID(postID, r); err != nil {
		return err
	}
	return postgres.RemovePostFavorite(userID, postID)
}

// ListMyFavoritePosts 按收藏时间倒序分页返回当前用户收藏帖子。
func ListMyFavoritePosts(userID int64, p *models.ParamFavoritePostList) (*models.PostFavoriteListData, error) {
	p.Normalize()
	r, err := postReader(&userID)
	if err != nil {
		return nil, err
	}
	total, err := postgres.CountPostFavoritesByUser(userID, r)
	if err != nil {
		return nil, err
	}
	offset := (p.Page - 1) * p.PageSize
	posts, favTimes, err := postgres.ListPostFavoritesByUser(userID, p.PageSize, offset, r)
	if err != nil {
		return nil, err
	}
	list := make([]models.PostFavoriteView, 0, len(posts))
	for i := range posts {
		v := models.PostToView(posts[i])
		tags, err := postgres.GetTagsByPostID(posts[i].ID)
		if err != nil {
			return nil, err
		}
		v.Tags = tags
		t := true
		v.IsFavorited = &t
		list = append(list, models.PostFavoriteView{
			PostView:    v,
			FavoritedAt: favTimes[i],
		})
	}
	return &models.PostFavoriteListData{
		List:     list,
		Total:    total,
		Page:     p.Page,
		PageSize: p.PageSize,
	}, nil
}

func moderationActionsForPost(post *models.Post, operatorID int64) (*models.PostModerationActionsView, error) {
	ok, err := canModerateBoard(operatorID, post.BoardID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	admin, err := postgres.IsSiteAdmin(operatorID)
	if err != nil {
		return nil, err
	}
	var act models.PostModerationActionsView
	if !post.SealedAt.Valid {
		act.CanSeal = true
	} else if admin {
		act.CanUnseal = true
	} else if post.SealKind.Valid && post.SealKind.String == "moderator" {
		act.CanUnseal = true
	}
	if !post.LockedAt.Valid {
		act.CanLockComments = true
	} else {
		act.CanUnlockComments = true
	}
	if !post.PinnedAt.Valid {
		act.CanPin = true
	} else {
		act.CanUnpin = true
	}
	return &act, nil
}

// SealPost 版主或站主封帖；站主使用 seal_kind=site，版主使用 moderator。
func SealPost(postID, operatorID int64) error {
	post, err := postgres.GetPostByIDIncludingDeleted(postID)
	if err != nil {
		return err
	}
	if post.DeletedAt.Valid {
		return postgres.ErrorPostNotExist
	}
	if post.SealedAt.Valid {
		return postgres.ErrorPostNotExist
	}
	ok, err := canModerateBoard(operatorID, post.BoardID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrSealPostForbidden
	}
	admin, err := postgres.IsSiteAdmin(operatorID)
	if err != nil {
		return err
	}
	kind := "moderator"
	if admin {
		kind = "site"
	}
	now := time.Now()
	if err := postgres.SealPost(postID, operatorID, kind, now); err != nil {
		return err
	}
	appendModerationLog(
		post.BoardID,
		operatorID,
		models.ModerationActionSealPost,
		models.ModerationTargetPost,
		postID,
		fmt.Sprintf("seal_kind=%s", kind),
	)
	_ = redisDao.RemoveHotPost(postID)
	return nil
}

// UnsealPost 解封：站主任意解封；版主仅可解封 moderator 类封帖。
func UnsealPost(postID, operatorID int64) error {
	post, err := postgres.GetPostByIDIncludingDeleted(postID)
	if err != nil {
		return err
	}
	if post.DeletedAt.Valid || !post.SealedAt.Valid {
		return postgres.ErrorPostNotExist
	}
	admin, err := postgres.IsSiteAdmin(operatorID)
	if err != nil {
		return err
	}
	if admin {
		return finishUnseal(postID, operatorID, post.BoardID)
	}
	mod, err := postgres.IsBoardModerator(operatorID, post.BoardID)
	if err != nil {
		return err
	}
	if !mod {
		return ErrUnsealPostForbidden
	}
	if post.SealKind.Valid && post.SealKind.String == "site" {
		return ErrUnsealPostForbidden
	}
	return finishUnseal(postID, operatorID, post.BoardID)
}

func finishUnseal(postID, operatorID, boardID int64) error {
	now := time.Now()
	if err := postgres.UnsealPost(postID, now); err != nil {
		return err
	}
	appendModerationLog(
		boardID,
		operatorID,
		models.ModerationActionUnsealPost,
		models.ModerationTargetPost,
		postID,
		"",
	)
	if p, err := postgres.GetPostByIDIncludingDeleted(postID); err == nil && !p.DeletedAt.Valid {
		_ = redisDao.UpsertHotPost(postID, calcHotScore(p.Score, p.CreateTime))
	}
	return nil
}

// LockPostComments 锁帖评论：版主或站主可操作。
func LockPostComments(postID, operatorID int64) error {
	post, err := postgres.GetPostByIDIncludingDeleted(postID)
	if err != nil {
		return err
	}
	if post.DeletedAt.Valid || post.LockedAt.Valid {
		return postgres.ErrorPostNotExist
	}
	ok, err := canModerateBoard(operatorID, post.BoardID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrLockPostForbidden
	}
	if err := postgres.LockPostComments(postID, operatorID, time.Now()); err != nil {
		return err
	}
	appendModerationLog(
		post.BoardID,
		operatorID,
		models.ModerationActionLockPostComments,
		models.ModerationTargetPost,
		postID,
		"",
	)
	return nil
}

// UnlockPostComments 解锁评论：版主或站主可操作。
func UnlockPostComments(postID, operatorID int64) error {
	post, err := postgres.GetPostByIDIncludingDeleted(postID)
	if err != nil {
		return err
	}
	if post.DeletedAt.Valid || !post.LockedAt.Valid {
		return postgres.ErrorPostNotExist
	}
	ok, err := canModerateBoard(operatorID, post.BoardID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrUnlockPostForbidden
	}
	if err := postgres.UnlockPostComments(postID, time.Now()); err != nil {
		return err
	}
	appendModerationLog(
		post.BoardID,
		operatorID,
		models.ModerationActionUnlockPostComments,
		models.ModerationTargetPost,
		postID,
		"",
	)
	return nil
}

// PinPost 置顶帖子：版主或站主可操作。
func PinPost(postID, operatorID int64) error {
	post, err := postgres.GetPostByIDIncludingDeleted(postID)
	if err != nil {
		return err
	}
	if post.DeletedAt.Valid || post.PinnedAt.Valid {
		return postgres.ErrorPostNotExist
	}
	ok, err := canModerateBoard(operatorID, post.BoardID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrPinPostForbidden
	}
	if err := postgres.PinPost(postID, operatorID, time.Now()); err != nil {
		return err
	}
	appendModerationLog(
		post.BoardID,
		operatorID,
		models.ModerationActionPinPost,
		models.ModerationTargetPost,
		postID,
		"",
	)
	return nil
}

// UnpinPost 取消置顶：版主或站主可操作。
func UnpinPost(postID, operatorID int64) error {
	post, err := postgres.GetPostByIDIncludingDeleted(postID)
	if err != nil {
		return err
	}
	if post.DeletedAt.Valid || !post.PinnedAt.Valid {
		return postgres.ErrorPostNotExist
	}
	ok, err := canModerateBoard(operatorID, post.BoardID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrUnpinPostForbidden
	}
	if err := postgres.UnpinPost(postID, time.Now()); err != nil {
		return err
	}
	appendModerationLog(
		post.BoardID,
		operatorID,
		models.ModerationActionUnpinPost,
		models.ModerationTargetPost,
		postID,
		"",
	)
	return nil
}

// ListBoardDeletedPosts 版主/站管查看本板已软删帖子列表。
func ListBoardDeletedPosts(boardID, operatorID int64, p *models.ParamReportList) (*models.DeletedPostListData, error) {
	if _, err := postgres.GetBoardByID(boardID); err != nil {
		return nil, err
	}
	ok, err := canModerateBoard(operatorID, boardID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrManageReportForbidden
	}
	p.Normalize()
	total, err := postgres.CountBoardDeletedPosts(boardID)
	if err != nil {
		return nil, err
	}
	offset := (p.Page - 1) * p.PageSize
	list, err := postgres.ListBoardDeletedPosts(boardID, p.PageSize, offset)
	if err != nil {
		return nil, err
	}
	return &models.DeletedPostListData{
		List:     list,
		Total:    total,
		Page:     p.Page,
		PageSize: p.PageSize,
	}, nil
}

// RestorePost 版主/站管恢复已软删帖子；未软删返回 ErrPostNotSoftDeleted。
func RestorePost(postID, operatorID int64) error {
	post, err := postgres.GetPostByIDIncludingDeleted(postID)
	if err != nil {
		return err
	}
	if !post.DeletedAt.Valid {
		return ErrPostNotSoftDeleted
	}
	ok, err := canModerateBoard(operatorID, post.BoardID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrManageReportForbidden
	}
	now := time.Now()
	if err := postgres.RestorePost(postID, now); err != nil {
		return err
	}
	appendModerationLog(
		post.BoardID,
		operatorID,
		models.ModerationActionRestorePost,
		models.ModerationTargetPost,
		postID,
		"restore soft delete",
	)
	if p, err := postgres.GetPostByIDIncludingDeleted(postID); err == nil && !p.DeletedAt.Valid && !p.SealedAt.Valid {
		_ = redisDao.UpsertHotPost(postID, calcHotScore(p.Score, p.CreateTime))
	}
	return nil
}
