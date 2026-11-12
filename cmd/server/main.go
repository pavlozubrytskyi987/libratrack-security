package main

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"

	"libratrack-security/internal/handler"
	"libratrack-security/internal/middleware"
	"libratrack-security/internal/store"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	st := store.New()

	authH := handler.NewAuthHandler(st)
	bookH := handler.NewBookHandler(st)
	loanH := handler.NewLoanHandler(st)
	userH := handler.NewUserHandler(st)

	r := gin.Default()
	r.Use(middleware.LoggingMiddleware(log))

	api := r.Group("/api/v1")

	// public routes — no auth required
	public := api.Group("")
	{
		public.POST("/auth/register", authH.Register)
		public.POST("/auth/login", authH.Login)
		public.GET("/books", bookH.List)
	}

	// protected routes — any authenticated user
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/me/loans", loanH.MyLoans)
	}

	// staff routes — librarian or admin
	staff := api.Group("")
	staff.Use(middleware.AuthMiddleware(), middleware.RequireRole("librarian", "admin"))
	{
		staff.POST("/books", bookH.Create)
		staff.PUT("/books/:id", bookH.Update)
		staff.POST("/loans", loanH.Create)
		staff.POST("/loans/:id/return", loanH.Return)
	}

	// admin routes — admin only
	admin := api.Group("")
	admin.Use(middleware.AuthMiddleware(), middleware.RequireRole("admin"))
	{
		admin.GET("/users", userH.List)
		admin.PUT("/users/:id/role", userH.UpdateRole)
	}

	addr := ":8080"
	log.Info("server starting", "addr", addr)
	if err := r.Run(addr); err != nil {
		log.Error("server failed", "err", err)
		os.Exit(1)
	}
}
