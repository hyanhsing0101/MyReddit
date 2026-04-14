package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestResponseError_UsesResCodeHTTPStatus(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name       string
		code       ResCode
		wantStatus int
	}{
		{name: "invalid vote value -> 400", code: CodeInvalidVoteValue, wantStatus: http.StatusBadRequest},
		{name: "cannot remove last owner -> 409", code: CodeCannotRemoveLastOwner, wantStatus: http.StatusConflict},
		{name: "board moderator not exist -> 404", code: CodeBoardModeratorNotExist, wantStatus: http.StatusNotFound},
		{name: "not board member -> 403", code: CodeNotBoardMember, wantStatus: http.StatusForbidden},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			ResponseError(c, tc.code)

			if w.Code != tc.wantStatus {
				t.Fatalf("status=%d, want=%d", w.Code, tc.wantStatus)
			}
			var resp ResponseData
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal failed: %v body=%s", err, w.Body.String())
			}
			if resp.Code != tc.code {
				t.Fatalf("resp code=%d, want=%d", resp.Code, tc.code)
			}
		})
	}
}
