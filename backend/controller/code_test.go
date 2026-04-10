package controller

import (
	"net/http"
	"testing"
)

func TestResCodeHTTPStatus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		code ResCode
		want int
	}{
		{name: "success", code: CodeSuccess, want: http.StatusOK},
		{name: "invalid param", code: CodeInvalidParam, want: http.StatusBadRequest},
		{name: "user exists", code: CodeUserExist, want: http.StatusConflict},
		{name: "board slug taken", code: CodeBoardSlugTaken, want: http.StatusConflict},
		{name: "need login", code: CodeNeedLogin, want: http.StatusUnauthorized},
		{name: "invalid token", code: CodeInvalidToken, want: http.StatusUnauthorized},
		{name: "invalid password", code: CodeInvalidPassword, want: http.StatusUnauthorized},
		{name: "post not exist", code: CodePostNotExist, want: http.StatusNotFound},
		{name: "board not exist", code: CodeBoardNotExist, want: http.StatusNotFound},
		{name: "comment not exist", code: CodeCommentNotExist, want: http.StatusNotFound},
		{name: "tag not exist", code: CodeTagNotExist, want: http.StatusNotFound},
		{name: "board moderator not exist", code: CodeBoardModeratorNotExist, want: http.StatusNotFound},
		{name: "forbidden", code: CodeForbidden, want: http.StatusForbidden},
		{name: "not board member", code: CodeNotBoardMember, want: http.StatusForbidden},
		{name: "cannot favorite public board", code: CodeCannotFavoritePublicBoard, want: http.StatusForbidden},
		{name: "cannot post to system board", code: CodeCannotPostToSystemBoard, want: http.StatusForbidden},
		{name: "post sealed", code: CodePostSealed, want: http.StatusConflict},
		{name: "cannot remove last owner", code: CodeCannotRemoveLastOwner, want: http.StatusConflict},
		{name: "invalid vote value", code: CodeInvalidVoteValue, want: http.StatusBadRequest},
		{name: "tag count exceeded", code: CodeTagCountExceeded, want: http.StatusBadRequest},
		{name: "invalid comment parent", code: CodeInvalidCommentParent, want: http.StatusBadRequest},
		{name: "parent comment mismatch", code: CodeParentCommentMismatch, want: http.StatusBadRequest},
		{name: "invalid board id", code: CodeInvalidBoardID, want: http.StatusBadRequest},
		{name: "server busy", code: CodeServerBusy, want: http.StatusInternalServerError},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.code.HTTPStatus(); got != tc.want {
				t.Fatalf("HTTPStatus()=%d, want=%d", got, tc.want)
			}
		})
	}
}
