package controller

import (
	"errors"
	"myreddit/dao/postgres"
	"myreddit/logic"
	"myreddit/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func AddBoardFavoriteHandler(c *gin.Context) {
	userID, err := GetCurrentUser(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	idStr := c.Param("id")
	boardID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || boardID < 1 {
		ResponseError(c, CodeInvalidParam)
		return
	}
	if err := logic.AddBoardFavorite(userID, boardID); err != nil {
		if errors.Is(err, postgres.ErrorBoardNotExist) {
			ResponseError(c, CodeBoardNotExist)
			return
		}
		if errors.Is(err, logic.ErrCannotFavoriteSystemBoard) {
			ResponseErrorWithMsg(c, CodeInvalidParam, err.Error())
			return
		}
		zap.L().Error("AddBoardFavorite Failed", zap.Error(err), zap.Int64("board_id", boardID))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, nil)
}

func RemoveBoardFavoriteHandler(c *gin.Context) {
	userID, err := GetCurrentUser(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	idStr := c.Param("id")
	boardID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || boardID < 1 {
		ResponseError(c, CodeInvalidParam)
		return
	}
	if err := logic.RemoveBoardFavorite(userID, boardID); err != nil {
		if errors.Is(err, postgres.ErrorBoardNotExist) {
			ResponseError(c, CodeBoardNotExist)
			return
		}
		zap.L().Error("RemoveBoardFavorite Failed", zap.Error(err), zap.Int64("board_id", boardID))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, nil)
}

func ListMyFavoriteBoardsHandler(c *gin.Context) {
	userID, err := GetCurrentUser(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	p := new(models.ParamFavoriteBoardList)
	if err := c.ShouldBindQuery(p); err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}
	data, err := logic.ListMyFavoriteBoards(userID, p)
	if err != nil {
		zap.L().Error("ListMyFavoriteBoards Failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
}
