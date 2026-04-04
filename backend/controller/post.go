package controller

import (
	"errors"
	"myreddit/dao/postgres"
	"myreddit/logic"
	"myreddit/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

func CreatePostHandler(c *gin.Context) {
	userID, err := GetCurrentUser(c)
	if err != nil {
		zap.L().Error("Get Current User With Invalid Param", zap.Error(err))
		ResponseError(c, CodeNeedLogin)
		return
	}
	p := new(models.ParamCreatePost)
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("Create Post With Invalid Param", zap.Error(err))
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			ResponseError(c, CodeInvalidParam)
			return
		}
		ResponseErrorWithMsg(c, CodeInvalidParam, errs.Translate(trans))
		return
	}

	if err := logic.CreatePost(p, userID); err != nil {
		if errors.Is(err, postgres.ErrorBoardNotExist) {
			ResponseError(c, CodeBoardNotExist)
			return
		}
		if errors.Is(err, logic.ErrCannotPostToSystemBoard) {
			ResponseErrorWithMsg(c, CodeInvalidParam, err.Error())
			return
		}
		zap.L().Error("Create Post Failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, nil)
}

func ListPostHandler(c *gin.Context) {
	p := new(models.ParamPostList)
	if err := c.ShouldBindQuery(p); err != nil {
		zap.L().Error("List Post With Invalid Param", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	data, err := logic.ListPost(p)
	if err != nil {
		if errors.Is(err, postgres.ErrorBoardNotExist) {
			ResponseError(c, CodeBoardNotExist)
			return
		}
		if errors.Is(err, logic.ErrInvalidBoardID) {
			ResponseError(c, CodeInvalidParam)
			return
		}
		zap.L().Error("List Post Failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
}

func GetPostHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		zap.L().Error("Get Post With Invalid Param", zap.String("id", idStr))
		ResponseError(c, CodeInvalidParam)
		return
	}
	data, err := logic.GetPost(id)
	if err != nil {
		if errors.Is(err, postgres.ErrorPostNotExist) {
			ResponseError(c, CodePostNotExist)
			return
		}
		zap.L().Error("Get Post Failed", zap.Error(err), zap.Int64("id", id))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
}