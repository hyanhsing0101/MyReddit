package controller

import (
	"myreddit/models"
	"myreddit/logic"

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
		zap.L().Error("Create Post With Invalid Param", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, nil)
}