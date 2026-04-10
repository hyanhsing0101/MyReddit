package controller

import (
	"errors"
	"fmt"
	"myreddit/dao/postgres"
	"myreddit/logic"
	"myreddit/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

func SignUpHandler(c *gin.Context) {
	p := new(models.ParamSignUp)
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("SignUp With Invalid Param", zap.Error(err))
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			ResponseError(c, CodeInvalidParam)
			return
		}
		ResponseErrorWithMsg(c, CodeInvalidParam, errs.Translate(trans))
		return
	}

	if err := logic.SignUp(p); err != nil {
		zap.L().Error("SignUp With Invalid Param", zap.Error(err))
		if errors.Is(err, postgres.ErrorUserExist) {
			ResponseError(c, CodeUserExist)
			return
		}
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, nil)
}

func LoginHandler(c *gin.Context) {
	p := new(models.ParamLogin)
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("Login With Invalid Param", zap.Error(err))
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			ResponseError(c, CodeInvalidParam)
			return
		}
		ResponseErrorWithMsg(c, CodeInvalidParam, removeTopStruct(errs.Translate(trans)))
		return
	}

	fmt.Println(p)

	tokenPair, err := logic.Login(p)
	if err != nil {
		zap.L().Error("Login With Invalid Param", zap.Error(err), zap.String("username", p.Username))
		if errors.Is(err, postgres.ErrorUserNotExist) {
			ResponseError(c, CodeUserNotExist)
			return
		}
		ResponseError(c, CodeInvalidPassword)
		return
	}

	ResponseSuccess(c, tokenPair)
}

// GetUserHomeHandler 获取用户主页数据（帖子与评论双列表）。
func GetUserHomeHandler(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || userID < 1 {
		ResponseError(c, CodeInvalidParam)
		return
	}
	p := new(models.ParamUserHome)
	if err := c.ShouldBindQuery(p); err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}
	data, err := logic.GetUserHome(userID, p, GetOptionalUserID(c))
	if err != nil {
		if errors.Is(err, postgres.ErrorUserNotExist) {
			ResponseError(c, CodeUserNotExist)
			return
		}
		zap.L().Error("GetUserHome Failed", zap.Error(err), zap.Int64("user_id", userID))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
}