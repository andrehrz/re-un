package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "mySecurePassword123"
	
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	
	// Hash should be different from password
	if hash == password {
		t.Error("Hash should not equal plain text password")
	}
	
	// Should be able to verify correct password
	if !CheckPassword(password, hash) {
		t.Error("Should verify correct password")
	}
	
	// Should reject wrong password
	if CheckPassword("wrongPassword", hash) {
		t.Error("Should reject wrong password")
	}
}