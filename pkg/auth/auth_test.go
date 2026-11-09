package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPasswordHashing(t *testing.T) {
	password := "SecretP@ssw0rd!"

	hash, err := HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)

	assert.True(t, CheckPassword(hash, password))
	assert.False(t, CheckPassword(hash, "WrongPassword"))
}

func TestJWT_GenerateAndValidate(t *testing.T) {
	secret := "test-secret-key-12345"
	userID := uuid.New()
	email := "user@example.com"
	role := "admin"

	tokenStr, err := GenerateToken(userID, email, role, "access", secret, 5*time.Minute)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	claims, err := ValidateToken(tokenStr, secret)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, role, claims.Role)
	assert.Equal(t, "access", claims.TokenType)
}

func TestJWT_ExpiredToken(t *testing.T) {
	secret := "test-secret-key-12345"
	userID := uuid.New()

	tokenStr, err := GenerateToken(userID, "user@example.com", "user", "access", secret, -1*time.Minute)
	assert.NoError(t, err)

	claims, err := ValidateToken(tokenStr, secret)
	assert.ErrorIs(t, err, ErrExpiredToken)
	assert.Nil(t, claims)
}

func TestJWT_InvalidSecret(t *testing.T) {
	secret := "test-secret-key-12345"
	userID := uuid.New()

	tokenStr, err := GenerateToken(userID, "user@example.com", "user", "access", secret, 5*time.Minute)
	assert.NoError(t, err)

	claims, err := ValidateToken(tokenStr, "wrong-secret")
	assert.ErrorIs(t, err, ErrInvalidToken)
	assert.Nil(t, claims)
}
