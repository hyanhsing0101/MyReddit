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

func ListBoardsHandler(c *gin.Context) {
	p := new(models.ParamBoardList)
	if err := c.ShouldBindQuery(p); err != nil {
		zap.L().Error("List Boards With Invalid Param", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	data, err := logic.ListBoards(p)
	if err != nil {
		zap.L().Error("List Boards Failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
}

func GetBoardByIDHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		zap.L().Error("Get Board By ID With Invalid Param", zap.String("id", idStr))
		ResponseError(c, CodeInvalidParam)
		return
	}
	data, err := logic.GetBoardByID(id)
	if err != nil {
		if errors.Is(err, postgres.ErrorBoardNotExist) {
			ResponseError(c, CodeBoardNotExist)
			return
		}
		zap.L().Error("Get Board By ID Failed", zap.Error(err), zap.Int64("id", id))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
}

func GetBoardBySlugHandler(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		ResponseError(c, CodeInvalidParam)
		return
	}
	data, err := logic.GetBoardBySlug(slug)
	if err != nil {
		if errors.Is(err, postgres.ErrorBoardNotExist) {
			ResponseError(c, CodeBoardNotExist)
			return
		}
		zap.L().Error("Get Board By Slug Failed", zap.Error(err), zap.String("slug", slug))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
}

func CreateBoardHandler(c *gin.Context) {
	userID, err := GetCurrentUser(c)
	if err != nil {
		zap.L().Error("Create Board With Need Login", zap.Error(err))
		ResponseError(c, CodeNeedLogin)
		return
	}
	p := new(models.ParamCreateBoard)
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("Create Board With Invalid Param", zap.Error(err))
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			ResponseError(c, CodeInvalidParam)
			return
		}
		ResponseErrorWithMsg(c, CodeInvalidParam, errs.Translate(trans))
		return
	}
	if err := logic.CreateBoard(p, userID); err != nil {
		switch {
		case errors.Is(err, logic.ErrBoardSlugReserved),
			errors.Is(err, logic.ErrBoardSlugInvalid),
			errors.Is(err, logic.ErrBoardNameEmpty):
			ResponseErrorWithMsg(c, CodeInvalidParam, err.Error())
			return
		case errors.Is(err, postgres.ErrorBoardSlugTaken):
			ResponseError(c, CodeBoardSlugTaken)
			return
		}
		zap.L().Error("Create Board Failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, nil)
}
