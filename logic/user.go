package logic

import (
	"myreddit/dao/mysql"
	"myreddit/models"
	"myreddit/pkg/snowflake"
)

func SignUp(p *models.ParamSignUp) {
	// 判断用户是否存在
	mysql.QueryUserByUsername()
	// 生成 UID
	snowflake.GenID()
	// 保存进数据库
	mysql.InsertUser()
}
