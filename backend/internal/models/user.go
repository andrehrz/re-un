package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID				uuid.UUID	`json:"id" db:"id"`
	Email			string		`json:"email" db:"email"`
	PasswordHash	string		`json:"-" db:"password_hash"` 
	CreatedAt 		time.Time	`json:"create_at" db:"created_at"`
}

type UserRegisterRequest struct {
	Email		string	`json:"email" binding:"required,email"`
	Password	string	`json:"password" binding:"required,min=8"`

}

type UserLoginRequest struct {
	Email		string	`json:"email" binding:"required,email"`
	Password	string	`json:"password" binding:"required"`
}

type UserAuthResponse struct {
	User	User	`json:"user"`
	Token	string	`json:"token"`
}

