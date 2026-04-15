package controller

import "net/http"

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
	CodePostCommentsLocked

	// =============================================================================
	// 精细业务语义
	// =============================================================================
	CodeCommentNotExist
	CodeInvalidVoteValue
	CodeTagNotExist
	CodeTagCountExceeded
	CodeCannotPostToSystemBoard
	CodeBoardModeratorNotExist
	CodeCannotRemoveLastOwner
	CodeInvalidCommentParent
	CodeParentCommentMismatch
	CodeInvalidBoardID
	CodeCannotReportOwnPost
	CodeDuplicatePostReport
	CodePostReportNotExist
	CodeCannotAppealUnsealedPost
	CodePostAppealNotExist
	CodePostNotSoftDeleted
	CodeCannotReportOwnComment
	CodeDuplicateCommentReport
	CodeCommentReportNotExist
	CodeUploadDisabled
	CodeUploadTooLarge
	CodeUploadInvalidImage
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
	CodeForbidden:                 "forbidden",
	CodeNotBoardMember:            "not board member",
	CodeCannotFavoritePublicBoard: "cannot favorite public board",
	CodePostSealed:                "post sealed",
	CodePostCommentsLocked:        "post comments locked",

	// ----- 精细业务语义 -----
	CodeCommentNotExist:          "comment not exist",
	CodeInvalidVoteValue:         "invalid vote value",
	CodeTagNotExist:              "tag not exist",
	CodeTagCountExceeded:         "tag count exceeded",
	CodeCannotPostToSystemBoard:  "cannot post to system board",
	CodeBoardModeratorNotExist:   "board moderator not exist",
	CodeCannotRemoveLastOwner:    "cannot remove last owner",
	CodeInvalidCommentParent:     "invalid parent comment",
	CodeParentCommentMismatch:    "parent comment mismatch",
	CodeInvalidBoardID:           "invalid board id",
	CodeCannotReportOwnPost:      "cannot report own post",
	CodeDuplicatePostReport:      "duplicate post report",
	CodePostReportNotExist:       "post report not exist",
	CodeCannotAppealUnsealedPost: "cannot appeal unsealed post",
	CodePostAppealNotExist:       "post appeal not exist",
	CodePostNotSoftDeleted:       "post not soft deleted",
	CodeCannotReportOwnComment:   "cannot report own comment",
	CodeDuplicateCommentReport:   "duplicate comment report",
	CodeCommentReportNotExist:    "comment report not exist",
	CodeUploadDisabled:           "upload disabled",
	CodeUploadTooLarge:           "upload too large",
	CodeUploadInvalidImage:       "upload invalid image",
}

func (c ResCode) Msg() string {
	msg, ok := codeMsgMap[c]
	if !ok {
		msg = codeMsgMap[CodeServerBusy]
	}
	return msg
}

// HTTPStatus 返回业务码对应的标准 HTTP 状态码。
func (c ResCode) HTTPStatus() int {
	switch c {
	case CodeSuccess:
		return http.StatusOK
	case CodeInvalidParam:
		return http.StatusBadRequest
	case CodeUserExist, CodeBoardSlugTaken:
		return http.StatusConflict
	case CodeUserNotExist, CodeInvalidPassword, CodeInvalidToken, CodeNeedLogin:
		return http.StatusUnauthorized
	case CodePostNotExist, CodeBoardNotExist, CodeCommentNotExist, CodeTagNotExist, CodeBoardModeratorNotExist:
		return http.StatusNotFound
	case CodePostReportNotExist, CodePostAppealNotExist, CodeCommentReportNotExist:
		return http.StatusNotFound
	case CodeForbidden, CodeNotBoardMember, CodeCannotFavoritePublicBoard, CodeCannotPostToSystemBoard, CodeUploadDisabled:
		return http.StatusForbidden
	case CodePostSealed, CodePostCommentsLocked, CodeCannotRemoveLastOwner, CodeDuplicatePostReport, CodeDuplicateCommentReport:
		return http.StatusConflict
	case CodeCannotReportOwnPost, CodeCannotAppealUnsealedPost, CodePostNotSoftDeleted, CodeCannotReportOwnComment:
		return http.StatusBadRequest
	case CodeInvalidVoteValue, CodeTagCountExceeded, CodeInvalidCommentParent, CodeParentCommentMismatch, CodeInvalidBoardID, CodeUploadInvalidImage:
		return http.StatusBadRequest
	case CodeUploadTooLarge:
		return http.StatusRequestEntityTooLarge
	case CodeServerBusy:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
