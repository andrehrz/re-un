package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"backend/internal/auth"
	"backend/internal/database"
	"backend/internal/models"
)

type AuthHandler struct {
	db *database.DB
}

func NewAuthHandler(db *database.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

func (handler *AuthHandler) RegisterUser(c *gin.Context) {
	var req models.UserRegisterRequest

	// bind and validate json using gin
	// does the checks for valid email and password length
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// will check if the user already exists
	var exists bool
	err := handler.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)",
		req.Email,
	).Scan(&exists)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database read error. Failed to check if user exists"})
		return
	}

	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	// all checks complete, will now start adding user to database
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to has password"})
		return
	}

	user := models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
	}

	// inserting into database
	_, err = handler.db.Exec(
		"INSERT INTO USERS (id, email, password_hash, created_at) VALUES ($1, $2, $3, $4)",
		user.ID, user.Email, user.PasswordHash, user.CreatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	response := models.UserAuthResponse{
		User:  user,
		Token: "todo",
	}

	c.JSON(http.StatusCreated, response)

}
