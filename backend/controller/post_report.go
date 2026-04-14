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

func CreatePostReportHandler(c *gin.Context) {
	userID, err := GetCurrentUser(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || postID < 1 {
		ResponseError(c, CodeInvalidParam)
		return
	}
	p := new(models.ParamCreatePostReport)
	if err := c.ShouldBindJSON(p); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			ResponseErrorWithMsg(c, CodeInvalidParam, errs.Translate(trans))
			return
		}
		ResponseError(c, CodeInvalidParam)
		return
	}
	if err := logic.CreatePostReport(postID, userID, p); err != nil {
		switch {
		case errors.Is(err, postgres.ErrorPostNotExist):
			ResponseError(c, CodePostNotExist)
		case errors.Is(err, logic.ErrCannotReportOwnPost):
			ResponseError(c, CodeCannotReportOwnPost)
		case errors.Is(err, logic.ErrDuplicateActivePostReport):
			ResponseError(c, CodeDuplicatePostReport)
		default:
			zap.L().Error("CreatePostReport Failed", zap.Error(err), zap.Int64("post_id", postID), zap.Int64("user_id", userID))
			ResponseError(c, CodeServerBusy)
		}
		return
	}
	ResponseCreated(c, nil)
}

func ListBoardPostReportsHandler(c *gin.Context) {
	userID, err := GetCurrentUser(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	boardID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || boardID < 1 {
		ResponseError(c, CodeInvalidParam)
		return
	}
	p := new(models.ParamReportList)
	if err := c.ShouldBindQuery(p); err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}
	data, err := logic.ListBoardPostReports(boardID, userID, p)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrorBoardNotExist):
			ResponseError(c, CodeBoardNotExist)
		case errors.Is(err, logic.ErrManageReportForbidden):
			ResponseError(c, CodeForbidden)
		default:
			zap.L().Error("ListBoardPostReports Failed", zap.Error(err), zap.Int64("board_id", boardID), zap.Int64("user_id", userID))
			ResponseError(c, CodeServerBusy)
		}
		return
	}
	ResponseSuccess(c, data)
}

func UpdatePostReportStatusHandler(c *gin.Context) {
	userID, err := GetCurrentUser(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	reportID, err := strconv.ParseInt(c.Param("rid"), 10, 64)
	if err != nil || reportID < 1 {
		ResponseError(c, CodeInvalidParam)
		return
	}
	p := new(models.ParamUpdatePostReport)
	if err := c.ShouldBindJSON(p); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			ResponseErrorWithMsg(c, CodeInvalidParam, errs.Translate(trans))
			return
		}
		ResponseError(c, CodeInvalidParam)
		return
	}
	if err := logic.UpdatePostReportStatus(reportID, userID, p); err != nil {
		switch {
		case errors.Is(err, postgres.ErrorPostReportNotExist):
			ResponseError(c, CodePostReportNotExist)
		case errors.Is(err, logic.ErrManageReportForbidden):
			ResponseError(c, CodeForbidden)
		default:
			zap.L().Error("UpdatePostReportStatus Failed", zap.Error(err), zap.Int64("report_id", reportID), zap.Int64("user_id", userID))
			ResponseError(c, CodeServerBusy)
		}
		return
	}
	ResponseSuccess(c, nil)
}

func BatchUpdatePostReportStatusHandler(c *gin.Context) {
	userID, err := GetCurrentUser(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	boardID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || boardID < 1 {
		ResponseError(c, CodeInvalidParam)
		return
	}
	p := new(models.ParamBatchUpdatePostReports)
	if err := c.ShouldBindJSON(p); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			ResponseErrorWithMsg(c, CodeInvalidParam, errs.Translate(trans))
			return
		}
		ResponseError(c, CodeInvalidParam)
		return
	}
	if err := logic.BatchUpdatePostReportStatus(boardID, userID, p); err != nil {
		switch {
		case errors.Is(err, postgres.ErrorBoardNotExist):
			ResponseError(c, CodeBoardNotExist)
		case errors.Is(err, postgres.ErrorPostReportNotExist):
			ResponseError(c, CodePostReportNotExist)
		case errors.Is(err, logic.ErrManageReportForbidden):
			ResponseError(c, CodeForbidden)
		default:
			zap.L().Error("BatchUpdatePostReportStatus Failed", zap.Error(err), zap.Int64("board_id", boardID), zap.Int64("user_id", userID))
			ResponseError(c, CodeServerBusy)
		}
		return
	}
	ResponseSuccess(c, nil)
}
