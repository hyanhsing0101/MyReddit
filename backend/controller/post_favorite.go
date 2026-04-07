package controller

import (
	"errors"
	"myreddit/dao/postgres"
	"myreddit/logic"
	"myreddit/models"
	"strconv"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

func AddPostFavoriteHandler(c *gin.Context) {
	userID, err := GetCurrentUser(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	idStr := c.Param("id")
	postID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || postID < 1 {
		ResponseError(c, CodeInvalidParam)
		return
	}
	if err := logic.AddPostFavorite(userID, postID); err != nil {
		if errors.Is(err, postgres.ErrorPostNotExist) {
			ResponseError(c, CodePostNotExist)
			return
		}
		zap.L().Error("AddPostFavorite Failed", zap.Error(err), zap.Int64("post_id", postID))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, nil)
}

func RemovePostFavoriteHandler(c *gin.Context) {
	userID, err := GetCurrentUser(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	idStr := c.Param("id")
	postID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || postID < 1 {
		ResponseError(c, CodeInvalidParam)
		return
	}
	if err := logic.RemovePostFavorite(userID, postID); err != nil {
		if errors.Is(err, postgres.ErrorPostNotExist) {
			ResponseError(c, CodePostNotExist)
			return
		}
		zap.L().Error("RemovePostFavorite Failed", zap.Error(err), zap.Int64("post_id", postID))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, nil)
}

func ListMyFavoritePostsHandler(c *gin.Context) {
	userID, err := GetCurrentUser(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	p := new(models.ParamFavoritePostList)
	if err := c.ShouldBindQuery(p); err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}
	data, err := logic.ListMyFavoritePosts(userID, p)
	if err != nil {
		zap.L().Error("ListMyFavoritePosts Failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
}
