package logic

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"myreddit/dao/postgres"
	redisDao "myreddit/dao/redis"
	"myreddit/models"
	"myreddit/pkg/jwt"
	"myreddit/pkg/snowflake"
)

const passwordSecret = "hyanhsing0101"

var ErrInvalidPassword = errors.New("wrong password")

func encryptPassword(rawPassword, salt string) string {
	h := md5.New()
	h.Write([]byte(passwordSecret))
	h.Write([]byte(salt))
	h.Write([]byte(rawPassword))
	return hex.EncodeToString(h.Sum(nil))
}

// SignUp 注册用户并写入认证信息。
func SignUp(p *models.ParamSignUp) (err error) {
	if err = postgres.CheckUserExist(p.Username); err != nil {
		return err
	}

	userID := snowflake.GenID()

	passwordHash := encryptPassword(p.Password, p.Username)
	if err := postgres.InsertUser(userID, p.Username, passwordHash); err != nil {
		return err
	}
	return nil
}

// Login 改为返回 TokenPair（access + refresh）
// refresh token 会在 Redis 中落库（并可轮换）
func Login(p *models.ParamLogin) (models.TokenPair, error) {
	authRow, err := postgres.GetUserAuthByUsername(p.Username)
	if err != nil {
		return models.TokenPair{}, err
	}
	if encryptPassword(p.Password, authRow.Username) != authRow.Password {
		return models.TokenPair{}, ErrInvalidPassword
	}

	accessToken, err := jwt.GenAccessToken(authRow.UserID, authRow.Username)
	if err != nil {
		return models.TokenPair{}, err
	}

	refreshToken, jti, err := jwt.GenRefreshToken(authRow.UserID, authRow.Username)
	if err != nil {
		return models.TokenPair{}, err
	}

	if err := redisDao.SaveRefreshToken(jti, refreshToken, jwt.RefreshTokenExpireDuration); err != nil {
		return models.TokenPair{}, err
	}

	return models.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken：校验 refresh token -> 轮换 refresh token -> 签发新的 access token
func RefreshToken(refreshToken string) (models.TokenPair, error) {
	mc, err := jwt.ParseToken(refreshToken)
	if err != nil {
		return models.TokenPair{}, err
	}
	if mc.TokenType != jwt.TokenTypeRefresh {
		return models.TokenPair{}, errors.New("not refresh token")
	}
	if mc.Id == "" {
		return models.TokenPair{}, errors.New("missing jti")
	}

	// 可选但建议：refresh 绑定的用户是否仍存在
	ok, err := postgres.CheckUserByIDAndName(mc.UserID, mc.Username)
	if err != nil || !ok {
		return models.TokenPair{}, errors.New("user not exist")
	}

	// 校验 Redis 中 jti 对应的 token hash 是否匹配
	valid, err := redisDao.VerifyRefreshToken(mc.Id, refreshToken)
	if err != nil {
		return models.TokenPair{}, err
	}
	if !valid {
		return models.TokenPair{}, errors.New("refresh token invalid")
	}

	// 轮换：删除旧 refresh token（实现一次性使用/防重放）
	if err := redisDao.DeleteRefreshToken(mc.Id); err != nil {
		return models.TokenPair{}, err
	}

	// 签发新的 access + refresh
	accessToken, err := jwt.GenAccessToken(mc.UserID, mc.Username)
	if err != nil {
		return models.TokenPair{}, err
	}

	newRefreshToken, newJti, err := jwt.GenRefreshToken(mc.UserID, mc.Username)
	if err != nil {
		return models.TokenPair{}, err
	}

	if err := redisDao.SaveRefreshToken(newJti, newRefreshToken, jwt.RefreshTokenExpireDuration); err != nil {
		return models.TokenPair{}, err
	}

	return models.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// MePermissions 返回当前用户权限视图（含站点管理员标记）。
func MePermissions(userID int64) (*models.MePermissionsView, error) {
	admin, err := postgres.IsSiteAdmin(userID)
	if err != nil {
		return nil, err
	}
	username, err := postgres.GetUsernameByUserID(userID)
	if err != nil {
		return nil, err
	}
	modBoards, err := postgres.ListModeratedBoardIDsByUser(userID)
	if err != nil {
		return nil, err
	}
	roles := []string{"user"}
	if admin {
		roles = append(roles, "admin")
	}
	if len(modBoards) > 0 {
		roles = append(roles, "moderator")
	}
	return &models.MePermissionsView{
		UserID:            userID,
		Username:          username,
		Roles:             roles,
		IsSiteAdmin:       admin,
		ModeratedBoardIDs: modBoards,
	}, nil
}

// GetUserHome 获取用户主页数据（帖子与评论双列表分页）。
func GetUserHome(userID int64, p *models.ParamUserHome, viewerID *int64) (*models.UserHomeData, error) {
	p.Normalize()
	username, err := postgres.GetUsernameByUserID(userID)
	if err != nil {
		return nil, err
	}
	r, err := postReader(viewerID)
	if err != nil {
		return nil, err
	}

	postsTotal, err := postgres.CountPostsByAuthorIDForViewer(userID, r)
	if err != nil {
		return nil, err
	}
	postOffset := (p.PostPage - 1) * p.PostPageSize
	postRows, err := postgres.ListPostsByAuthorIDForViewer(userID, r, p.PostPageSize, postOffset)
	if err != nil {
		return nil, err
	}
	posts := make([]models.UserHomePostItem, 0, len(postRows))
	for _, row := range postRows {
		item := models.UserHomePostItem{
			ID:         row.ID,
			BoardID:    row.BoardID,
			Title:      row.Title,
			Score:      row.Score,
			CreateTime: row.CreateTime,
			UpdateTime: row.UpdateTime,
		}
		if row.BoardSlug.Valid {
			item.BoardSlug = row.BoardSlug.String
		}
		if row.BoardName.Valid {
			item.BoardName = row.BoardName.String
		}
		posts = append(posts, item)
	}

	commentsTotal, err := postgres.CountCommentsByAuthorIDForViewer(userID, r)
	if err != nil {
		return nil, err
	}
	commentOffset := (p.CommentPage - 1) * p.CommentPageSize
	comments, err := postgres.ListCommentsByAuthorIDForViewer(userID, r, p.CommentPageSize, commentOffset)
	if err != nil {
		return nil, err
	}

	return &models.UserHomeData{
		UserID:          userID,
		Username:        username,
		Posts:           posts,
		PostsTotal:      postsTotal,
		PostPage:        p.PostPage,
		PostPageSize:    p.PostPageSize,
		Comments:        comments,
		CommentsTotal:   commentsTotal,
		CommentPage:     p.CommentPage,
		CommentPageSize: p.CommentPageSize,
	}, nil
}
