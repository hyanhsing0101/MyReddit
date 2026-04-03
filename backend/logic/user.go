package logic

import (
	"errors"
	"myreddit/dao/postgres"
	redisDao "myreddit/dao/redis"
	"myreddit/models"
	"myreddit/pkg/jwt"
	"myreddit/pkg/snowflake"
)

func SignUp(p *models.ParamSignUp) (err error) {
	if err = postgres.CheckUserExist(p.Username); err != nil {
		return err
	}

	userID := snowflake.GenID()

	u := models.User{
		UserID:   userID,
		Username: p.Username,
		Password: p.Password,
	}

	if err := postgres.InsertUser(&u); err != nil {
		return err
	}
	return nil
}

// Login 改为返回 TokenPair（access + refresh）
// refresh token 会在 Redis 中落库（并可轮换）
func Login(p *models.ParamLogin) (models.TokenPair, error) {
	user := &models.User{
		Username: p.Username,
		Password: p.Password,
	}
	if err := postgres.Login(user); err != nil {
		return models.TokenPair{}, err
	}

	accessToken, err := jwt.GenAccessToken(user.UserID, user.Username)
	if err != nil {
		return models.TokenPair{}, err
	}

	refreshToken, jti, err := jwt.GenRefreshToken(user.UserID, user.Username)
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