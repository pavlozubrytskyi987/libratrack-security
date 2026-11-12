package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func requestWithToken(r *gin.Engine, method, path, token string, body any) *httptest.ResponseRecorder {
	var reqBody *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewReader(b)
	} else {
		reqBody = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRBAC(t *testing.T) {
	tests := []struct {
		name       string
		run        func(r *gin.Engine) int
		wantStatus int
	}{
		{
			name: "protected route without token returns 401",
			run: func(r *gin.Engine) int {
				w := requestWithToken(r, "GET", "/api/v1/me/loans", "", nil)
				return w.Code
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "protected route with invalid token returns 401",
			run: func(r *gin.Engine) int {
				w := requestWithToken(r, "GET", "/api/v1/me/loans", "invalid-token-here", nil)
				return w.Code
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "reader calling POST /books returns 403",
			run: func(r *gin.Engine) int {
				registerUser(r, "reader@test.com", "password123")
				token, _ := loginUser(r, "reader@test.com", "password123")
				w := requestWithToken(r, "POST", "/api/v1/books", token,
					map[string]string{"title": "Test", "author": "Author", "isbn": "123"})
				return w.Code
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "reader can access own loans",
			run: func(r *gin.Engine) int {
				registerUser(r, "loanuser@test.com", "password123")
				token, _ := loginUser(r, "loanuser@test.com", "password123")
				w := requestWithToken(r, "GET", "/api/v1/me/loans", token, nil)
				return w.Code
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "reader calling GET /users returns 403",
			run: func(r *gin.Engine) int {
				registerUser(r, "notadmin@test.com", "password123")
				token, _ := loginUser(r, "notadmin@test.com", "password123")
				w := requestWithToken(r, "GET", "/api/v1/users", token, nil)
				return w.Code
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "public books endpoint needs no auth",
			run: func(r *gin.Engine) int {
				w := requestWithToken(r, "GET", "/api/v1/books", "", nil)
				return w.Code
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, _ := setupRouter()
			got := tt.run(r)
			if got != tt.wantStatus {
				t.Fatalf("status = %d, want %d", got, tt.wantStatus)
			}
		})
	}
}
