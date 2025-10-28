package handlers

import (
	"github.com/gin-gonic/gin"

	"backend/internal/database"
	"backend/internal/models"
)

type AuthHandler struct {
	db *database.DB
}

func NewAuthHandler(db *database.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

func (h *AuthHandler) RegisterUser(c *gin.Context) {
	var req models.UserRegisterRequest

	// bind and validate json using gin 
	if err := c.ShouldBindJSON()
}
