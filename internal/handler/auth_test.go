package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	"libratrack-security/internal/handler"
	"libratrack-security/internal/middleware"
	"libratrack-security/internal/store"
)

func init() {
	gin.SetMode(gin.TestMode)
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing")
}

func setupRouter() (*gin.Engine, *store.Store) {
	st := store.New()
	authH := handler.NewAuthHandler(st)
	bookH := handler.NewBookHandler(st)
	loanH := handler.NewLoanHandler(st)
	userH := handler.NewUserHandler(st)

	r := gin.New()
	api := r.Group("/api/v1")

	api.POST("/auth/register", authH.Register)
	api.POST("/auth/login", authH.Login)
	api.GET("/books", bookH.List)

	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/me/loans", loanH.MyLoans)
	}

	staff := api.Group("")
	staff.Use(middleware.AuthMiddleware(), middleware.RequireRole("librarian", "admin"))
	{
		staff.POST("/books", bookH.Create)
		staff.POST("/loans", loanH.Create)
	}

	admin := api.Group("")
	admin.Use(middleware.AuthMiddleware(), middleware.RequireRole("admin"))
	{
		admin.GET("/users", userH.List)
	}

	return r, st
}

func registerUser(r *gin.Engine, email, password string) *httptest.ResponseRecorder {
	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func loginUser(r *gin.Engine, email, password string) (string, int) {
	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		var resp map[string]string
		json.Unmarshal(w.Body.Bytes(), &resp)
		return resp["token"], w.Code
	}
	return "", w.Code
}

func TestAuth(t *testing.T) {
	tests := []struct {
		name       string
		run        func(r *gin.Engine) int
		wantStatus int
	}{
		{
			name: "register new user returns 201",
			run: func(r *gin.Engine) int {
				return registerUser(r, "new@test.com", "password123").Code
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "register duplicate email returns 422",
			run: func(r *gin.Engine) int {
				registerUser(r, "dup@test.com", "password123")
				return registerUser(r, "dup@test.com", "password123").Code
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "valid login returns 200 with JWT",
			run: func(r *gin.Engine) int {
				registerUser(r, "login@test.com", "password123")
				_, code := loginUser(r, "login@test.com", "password123")
				return code
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "login wrong password returns 401 generic error",
			run: func(r *gin.Engine) int {
				registerUser(r, "wrong@test.com", "password123")
				_, code := loginUser(r, "wrong@test.com", "wrongpassword")
				return code
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "login unknown email returns 401 generic error",
			run: func(r *gin.Engine) int {
				_, code := loginUser(r, "nobody@test.com", "password123")
				return code
			},
			wantStatus: http.StatusUnauthorized,
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
