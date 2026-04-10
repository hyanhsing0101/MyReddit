package controller

type ResCode int64

// 业务码自 1000 起由 iota 连续递增；请勿在中间插入常量，以免与客户端/文档约定冲突。
// 分段：通用与用户 → 鉴权 → 帖子 → 板块 → 权限。
const (
	CodeSuccess ResCode = 1000 + iota

	// =============================================================================
	// 通用 & 用户
	// =============================================================================
	CodeInvalidParam
	CodeUserExist
	CodeUserNotExist
	CodeInvalidPassword
	CodeServerBusy

	// =============================================================================
	// 鉴权 / Token
	// =============================================================================
	CodeInvalidToken
	CodeNeedLogin

	// =============================================================================
	// 帖子
	// =============================================================================
	CodePostNotExist

	// =============================================================================
	// 板块
	// =============================================================================
	CodeBoardNotExist
	CodeBoardSlugTaken

	// =============================================================================
	// 权限
	// =============================================================================
	CodeForbidden
	CodeNotBoardMember
	CodeCannotFavoritePublicBoard
	CodePostSealed
)

var codeMsgMap = map[ResCode]string{
	// ----- 通用 & 用户 -----
	CodeSuccess:         "success",
	CodeInvalidParam:    "invalid param",
	CodeUserExist:       "user exist",
	CodeUserNotExist:    "user not exist",
	CodeInvalidPassword: "invalid password",
	CodeServerBusy:      "server busy",

	// ----- 鉴权 -----
	CodeInvalidToken: "invalid token",
	CodeNeedLogin:    "need login",

	// ----- 帖子 -----
	CodePostNotExist: "post not exist",

	// ----- 板块 -----
	CodeBoardNotExist:  "board not exist",
	CodeBoardSlugTaken: "board slug taken",

	// ----- 权限 -----
	CodeForbidden:                "forbidden",
	CodeNotBoardMember:           "not board member",
	CodeCannotFavoritePublicBoard: "cannot favorite public board",
	CodePostSealed:                "post sealed",
}

func (c ResCode) Msg() string {
	msg, ok := codeMsgMap[c]
	if !ok {
		msg = codeMsgMap[CodeServerBusy]
	}
	return msg
}
