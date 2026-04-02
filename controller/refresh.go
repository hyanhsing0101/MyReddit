package controller

import (
	"myreddit/logic"
	"myreddit/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

func RefreshTokenHandler(c *gin.Context) {
	p := new(models.ParamRefresh)
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("Refresh With Invalid Param", zap.Error(err))
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			ResponseError(c, CodeInvalidParam)
			return
		}
		ResponseErrorWithMsg(c, CodeInvalidParam, removeTopStruct(errs.Translate(trans)))
		return
	}

	tokenPair, err := logic.RefreshToken(p.RefreshToken)
	if err != nil {
		zap.L().Error("Refresh With Invalid Token", zap.Error(err))
		ResponseError(c, CodeInvalidToken)
		return
	}

	ResponseSuccess(c, tokenPair)
}