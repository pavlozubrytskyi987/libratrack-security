package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"libratrack-security/internal/model"
	"libratrack-security/internal/store"
)

type LoanHandler struct {
	store *store.Store
}

func NewLoanHandler(s *store.Store) *LoanHandler {
	return &LoanHandler{store: s}
}

func (h *LoanHandler) Create(c *gin.Context) {
	var req model.CreateLoanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	loan, err := h.store.CreateLoan(req.UserID, req.BookID)
	if errors.Is(err, store.ErrAlreadyLoaned) {
		c.JSON(http.StatusConflict, gin.H{"error": "book already on loan"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, loan)
}

func (h *LoanHandler) Return(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	loan, err := h.store.ReturnLoan(id)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "loan not found"})
		return
	}
	if errors.Is(err, store.ErrNotOnLoan) {
		c.JSON(http.StatusConflict, gin.H{"error": "loan already returned"})
		return
	}

	c.JSON(http.StatusOK, loan)
}

func (h *LoanHandler) MyLoans(c *gin.Context) {
	userID, _ := c.Get("user_id")
	loans := h.store.UserLoans(userID.(int))
	if loans == nil {
		loans = []*model.Loan{}
	}
	c.JSON(http.StatusOK, loans)
}
