package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"myreddit/controller"

	"github.com/gin-gonic/gin"
)

type testResp struct {
	Code int64 `json:"code"`
}

type endpointCase struct {
	name         string
	method       string
	path         string
	body         string
}

func TestSetupRouter_AllRoutesRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := SetupRouter("test")

	got := make(map[string]struct{}, len(r.Routes()))
	for _, ri := range r.Routes() {
		got[ri.Method+" "+ri.Path] = struct{}{}
	}

	all := []string{
		"POST /signup", "POST /login", "POST /refresh",
		"GET /posts", "POST /posts", "GET /posts/:id", "PATCH /posts/:id", "DELETE /posts/:id",
		"POST /posts/:id/vote", "POST /posts/:id/seal", "DELETE /posts/:id/seal",
		"POST /posts/:id/favorites", "DELETE /posts/:id/favorites",
		"GET /posts/:id/comments", "POST /posts/:id/comments", "POST /posts/:id/comments/:cid/vote",
		"POST /post", "POST /posts/:id/edit", "POST /posts/:id/unseal", "POST /posts/:id/favorite", "POST /posts/:id/unfavorite",
		"GET /boards", "POST /boards", "GET /boards/:id", "GET /boards/by-slug/:slug",
		"POST /boards/:id/favorites", "DELETE /boards/:id/favorites",
		"GET /boards/:id/moderators", "POST /boards/:id/moderators", "PATCH /boards/:id/moderators/:uid", "DELETE /boards/:id/moderators/:uid",
		"GET /boards/slug/:slug", "POST /boards/:id/favorite", "POST /boards/:id/unfavorite", "POST /boards/:id/moderators/:uid/role",
		"GET /tags", "GET /search", "GET /users/:id/home",
		"GET /me/permissions", "GET /me/favorite-boards", "GET /me/favorite-posts",
		"GET /debug/auth/any", "GET /debug/auth/admin", "GET /ping",
	}

	for _, ep := range all {
		if _, ok := got[ep]; !ok {
			t.Fatalf("missing route: %s", ep)
		}
	}
}

func TestSetupRouter_ProtectedEndpointsRequireLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := SetupRouter("test")

	cases := []endpointCase{
		{name: "create posts", method: http.MethodPost, path: "/posts", body: `{"board_id":1,"title":"t","content":"c","tag_ids":[]}`},
		{name: "patch post", method: http.MethodPatch, path: "/posts/1", body: `{"title":"t","content":"c","tag_ids":[]}`},
		{name: "delete post", method: http.MethodDelete, path: "/posts/1"},
		{name: "vote post", method: http.MethodPost, path: "/posts/1/vote", body: `{"value":1}`},
		{name: "seal post", method: http.MethodPost, path: "/posts/1/seal"},
		{name: "unseal post", method: http.MethodDelete, path: "/posts/1/seal"},
		{name: "favorite post", method: http.MethodPost, path: "/posts/1/favorites"},
		{name: "unfavorite post", method: http.MethodDelete, path: "/posts/1/favorites"},
		{name: "create comment", method: http.MethodPost, path: "/posts/1/comments", body: `{"content":"x"}`},
		{name: "vote comment", method: http.MethodPost, path: "/posts/1/comments/2/vote", body: `{"value":1}`},

		{name: "legacy create post", method: http.MethodPost, path: "/post", body: `{"board_id":1,"title":"t","content":"c","tag_ids":[]}`},
		{name: "legacy edit post", method: http.MethodPost, path: "/posts/1/edit", body: `{"title":"t","content":"c","tag_ids":[]}`},
		{name: "legacy unseal post", method: http.MethodPost, path: "/posts/1/unseal"},
		{name: "legacy favorite post", method: http.MethodPost, path: "/posts/1/favorite"},
		{name: "legacy unfavorite post", method: http.MethodPost, path: "/posts/1/unfavorite"},

		{name: "create board", method: http.MethodPost, path: "/boards", body: `{"slug":"b","name":"n"}`},
		{name: "favorite board", method: http.MethodPost, path: "/boards/1/favorites"},
		{name: "unfavorite board", method: http.MethodDelete, path: "/boards/1/favorites"},
		{name: "list moderators", method: http.MethodGet, path: "/boards/1/moderators"},
		{name: "add moderator", method: http.MethodPost, path: "/boards/1/moderators", body: `{"user_id":1,"role":"moderator"}`},
		{name: "patch moderator", method: http.MethodPatch, path: "/boards/1/moderators/2", body: `{"role":"owner"}`},
		{name: "delete moderator", method: http.MethodDelete, path: "/boards/1/moderators/2"},
		{name: "legacy favorite board", method: http.MethodPost, path: "/boards/1/favorite"},
		{name: "legacy unfavorite board", method: http.MethodPost, path: "/boards/1/unfavorite"},
		{name: "legacy moderator role", method: http.MethodPost, path: "/boards/1/moderators/2/role", body: `{"role":"owner"}`},

		{name: "me permissions", method: http.MethodGet, path: "/me/permissions"},
		{name: "me favorite boards", method: http.MethodGet, path: "/me/favorite-boards"},
		{name: "me favorite posts", method: http.MethodGet, path: "/me/favorite-posts"},
		{name: "debug any", method: http.MethodGet, path: "/debug/auth/any"},
		{name: "debug admin", method: http.MethodGet, path: "/debug/auth/admin"},
		{name: "ping", method: http.MethodGet, path: "/ping"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString(tc.body))
			if tc.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			var got testResp
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("protected api must return json, err=%v, body=%s", err, w.Body.String())
			}
			if got.Code != int64(controller.CodeNeedLogin) {
				t.Fatalf("protected api without token should be need-login, got=%d body=%s", got.Code, w.Body.String())
			}
		})
	}
}

func TestSetupRouter_MethodMismatchRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := SetupRouter("test")

	cases := []struct {
		name   string
		method string
		path   string
	}{
		{name: "posts reject put", method: http.MethodPut, path: "/posts/1"},
		{name: "boards reject patch", method: http.MethodPatch, path: "/boards"},
		{name: "search reject post", method: http.MethodPost, path: "/search"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusNotFound && w.Code != http.StatusMethodNotAllowed {
				t.Fatalf("expect 404/405, got=%d for %s %s", w.Code, tc.method, tc.path)
			}
		})
	}
}
