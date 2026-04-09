package logic

import (
	"database/sql"
	"errors"
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

func CreatePost(p *models.ParamCreatePost, userID int64) error {
	board, err := postgres.GetBoardByID(p.BoardID)
	if err != nil {
		return err
	}
	if board.IsSystemSink {
		return ErrCannotPostToSystemBoard
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

func ListPost(p *models.ParamPostList, viewerID *int64) (*models.PostListData, error) {
	p.Normalize()
	var boardFilter *int64
	if p.BoardID != nil {
		if *p.BoardID < 1 {
			return nil, ErrInvalidBoardID
		}
		if _, err := postgres.GetBoardByID(*p.BoardID); err != nil {
			return nil, err
		}
		boardFilter = p.BoardID
	}
	total, err := postgres.CountPosts(boardFilter)
	if err != nil {
		return nil, err
	}
	offset := (p.Page - 1) * p.PageSize
	var posts []models.Post
	if p.Sort == models.PostSortHot && boardFilter == nil {
		cachedIDs, err := redisDao.GetHotPostIDs(p.Page, p.PageSize)
		if err == nil && len(cachedIDs) > 0 {
			posts, err = postgres.ListPostsByIDsOrdered(cachedIDs)
			if err != nil {
				return nil, err
			}
		}
		// 缓存未命中（或命中后被过滤空）则回源 PG，再回填一页缓存。
		if len(posts) == 0 {
			posts, err = postgres.ListPosts(boardFilter, p.Sort, p.PageSize, offset)
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
		var err error
		posts, err = postgres.ListPosts(boardFilter, p.Sort, p.PageSize, offset)
		if err != nil {
			return nil, err
		}
	}
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

func GetPost(id int64, viewerID *int64) (*models.PostView, error) {
	post, err := postgres.GetPostByID(id)
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
	}
	return &v, nil
}

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
	score, myVote, err := postgres.ApplyPostVote(postID, userID, value)
	if err != nil {
		return nil, err
	}
	// 更新热榜缓存分数（失败可容忍，后续读路径会回源修复）。
	if p, err := postgres.GetPostByID(postID); err == nil {
		_ = redisDao.UpsertHotPost(postID, calcHotScore(score, p.CreateTime))
	}
	return &models.PostVoteResult{
		Score:  score,
		MyVote: myVote,
	}, nil
}

// DeletePost 软删：作者本人或站点管理员；无主帖仅管理员可删；已软删返回 post not exist
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
	if admin {
		if err := postgres.SoftDeletePost(postID, time.Now()); err != nil {
			return err
		}
		_ = redisDao.RemoveHotPost(postID)
		return nil
	}
	if !post.AuthorID.Valid {
		return ErrDeletePostForbidden
	}
	if post.AuthorID.Int64 != operatorUserID {
		return ErrDeletePostForbidden
	}
	if err := postgres.SoftDeletePost(postID, time.Now()); err != nil {
		return err
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

func AddPostFavorite(userID, postID int64) error {
	if _, err := postgres.GetPostByID(postID); err != nil {
		return err
	}
	return postgres.AddPostFavorite(userID, postID)
}

func RemovePostFavorite(userID, postID int64) error {
	if _, err := postgres.GetPostByID(postID); err != nil {
		return err
	}
	return postgres.RemovePostFavorite(userID, postID)
}

func ListMyFavoritePosts(userID int64, p *models.ParamFavoritePostList) (*models.PostFavoriteListData, error) {
	p.Normalize()
	total, err := postgres.CountPostFavoritesByUser(userID)
	if err != nil {
		return nil, err
	}
	offset := (p.Page - 1) * p.PageSize
	posts, favTimes, err := postgres.ListPostFavoritesByUser(userID, p.PageSize, offset)
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
