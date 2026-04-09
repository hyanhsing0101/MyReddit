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

// CreatePostHandler 创建帖子并处理参数/业务错误映射。
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
		if errors.Is(err, postgres.ErrorTagNotExist) {
			ResponseErrorWithMsg(c, CodeInvalidParam, err.Error())
			return
		}
		if errors.Is(err, logic.ErrTagCountExceedsMaxLimit) {
			ResponseErrorWithMsg(c, CodeInvalidParam, err.Error())
			return
		}
		zap.L().Error("Create Post Failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, nil)
}

// ListPostHandler 分页获取帖子列表（支持 board_id 与 sort）。
func ListPostHandler(c *gin.Context) {
	p := new(models.ParamPostList)
	if err := c.ShouldBindQuery(p); err != nil {
		zap.L().Error("List Post With Invalid Param", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	data, err := logic.ListPost(p, GetOptionalUserID(c))
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

// GetPostHandler 获取单个帖子详情（可选附带当前用户状态）。
func GetPostHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		zap.L().Error("Get Post With Invalid Param", zap.String("id", idStr))
		ResponseError(c, CodeInvalidParam)
		return
	}
	data, err := logic.GetPost(id, GetOptionalUserID(c))
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

// DeletePostHandler 软删帖子（作者本人或站点管理员）。
func DeletePostHandler(c *gin.Context) {
	userID, err := GetCurrentUser(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		zap.L().Error("Delete Post With Invalid Param", zap.String("id", idStr))
		ResponseError(c, CodeInvalidParam)
		return
	}
	if err := logic.DeletePost(id, userID); err != nil {
		if errors.Is(err, postgres.ErrorPostNotExist) {
			ResponseError(c, CodePostNotExist)
			return
		}
		if errors.Is(err, logic.ErrDeletePostForbidden) {
			ResponseError(c, CodeForbidden)
			return
		}
		zap.L().Error("Delete Post Failed", zap.Error(err), zap.Int64("id", id))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, nil)
}

// UpdatePostHandler 编辑帖子内容与标签（作者本人或站点管理员）。
func UpdatePostHandler(c *gin.Context) {
	userID, err := GetCurrentUser(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		zap.L().Error("Update Post With Invalid Param", zap.String("id", idStr))
		ResponseError(c, CodeInvalidParam)
		return
	}
	p := new(models.ParamUpdatePost)
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("Update Post With Invalid Body", zap.Error(err))
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			ResponseError(c, CodeInvalidParam)
			return
		}
		ResponseErrorWithMsg(c, CodeInvalidParam, errs.Translate(trans))
		return
	}
	if err := logic.UpdatePost(id, userID, p); err != nil {
		if errors.Is(err, postgres.ErrorPostNotExist) {
			ResponseError(c, CodePostNotExist)
			return
		}
		if errors.Is(err, postgres.ErrorTagNotExist) ||
			errors.Is(err, logic.ErrTagCountExceedsMaxLimit) {
			ResponseErrorWithMsg(c, CodeInvalidParam, err.Error())
			return
		}
		if errors.Is(err, logic.ErrEditPostForbidden) {
			ResponseError(c, CodeForbidden)
			return
		}
		zap.L().Error("Update Post Failed", zap.Error(err), zap.Int64("id", id))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, nil)
}
