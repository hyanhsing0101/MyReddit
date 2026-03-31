package logic

import (
	"myreddit/dao/mysql"
	"myreddit/models"
	"myreddit/pkg/snowflake"
)

func SignUp(p *models.ParamSignUp) (err error) {
	// 判断用户是否存在
	if err = mysql.CheckUserExist(p.Username); err != nil {
		return err
	}

	// 生成 UID
	userID := snowflake.GenID()

	// 构造 user 实例
	u := models.User{
		UserID:   userID,
		Username: p.Username,
		Password: p.Password,
	}
	// 保存进数据库
	if err := mysql.InsertUser(&u); err != nil {
		return err
	}
	return nil
}
