package handlers

import (
	"database/sql"
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
	
	// generating tokens
	accessTokenString, err := auth.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	refreshTokenString, err := auth.GenerateRefreshToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	// 
	_, err = handler.db.Exec(
		"INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at) VALUES ($1, $2, $3, $4, $5)",
		uuid.New(),
		user.ID, 
		refreshTokenString, 
		time.Now().Add(auth.RefreshTokenDuration),
		time.Now(),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store refresh token in database"})
		return
	}

	tokens := models.TokenReponse{
		AccessToken: accessTokenString,
		RefreshToken: refreshTokenString,
	}

	c.JSON(http.StatusCreated, gin.H{
		"user": user,
		"access_token": tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})

}

func (handler *AuthHandler) LoginUser(c *gin.Context) {
	var req models.UserLoginRequest

	// bind and validate json using gin
	// does the checks for valid email and password length
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// get the user from database
	var user models.User
	err := handler.db.QueryRow(
		"SELECT id, email, password_hash, created_at FROM users where email = $1",
		req.Email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)

	// email not registered
	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Database Error"})
		return
	}

	// incorrect password for email
	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// generating tokens
	accessTokenString, err := auth.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	refreshTokenString, err := auth.GenerateRefreshToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	// 
	_, err = handler.db.Exec(
		"INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at) VALUES ($1, $2, $3, $4, $5)",
		uuid.New(),
		user.ID, 
		refreshTokenString, 
		time.Now().Add(auth.RefreshTokenDuration),
		time.Now(),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store refresh token in database"})
		return
	}

	tokens := models.TokenReponse{
		AccessToken: accessTokenString,
		RefreshToken: refreshTokenString,
	}

	c.JSON(http.StatusCreated, gin.H{
		"user": user,
		"access_token": tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
}

// generates a new access token using refresh token
func (handler *AuthHandler) RefreshToken(c *gin.Context){
	var req models.RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create refresh token request model form request"})
		return
	}

	var storedToken models.RefreshToken
	err := handler.db.QueryRow(
		"SELECT id, user_id, token, expires_at, created_at FROM refresh_tokens where token = $1",
		req.RefreshToken,
	).Scan(&storedToken.ID, &storedToken.UserID, &storedToken.Token, &storedToken.ExpiresAt, &storedToken.CreatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query for refresh token error"})
		return
	}

	// found refresh token in database. need to validate expiration now
	if time.Now().After(storedToken.ExpiresAt) {
		// delete expired token
		handler.db.Exec("DELETE FROM refresh_tokens WHERE id = $1", storedToken.ID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token expired"})
		return
	}

	// get user from refresh token details 
	var email string
	err = handler.db.QueryRow(
		"SELECT email FROM users WHERE id = $1",
		storedToken.UserID,
	).Scan(&email)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	// new access token
	newAccessToken, err := auth.GenerateAccessToken(storedToken.UserID, email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new access token"})
		return
	}
	
	// rotate refresh token
	newRefreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new refresh token"})
		return
	}

	// delete old refresh token and store new refresh token
	_, err = handler.db.Exec("DELETE FROM refresh_tokens WHERE id = $1", storedToken.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete old token"})
		return
	}

	_, err = handler.db.Exec(
		"INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at) VALUES ($1, $2, $3, $4, $5)",
		uuid.New(),
		storedToken.UserID,
		newRefreshToken,
		time.Now().Add(auth.RefreshTokenDuration),
		time.Now(),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store new refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": newAccessToken,
		"refresh_token": newRefreshToken,
	})

}

func (handler *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, ok := auth.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
		return 
	}

	// have a user id 
	var user models.User
	err := handler.db.QueryRow(
		"SELECT id, email, created_at FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Email, &user.CreatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
	}

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Database error"})
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}