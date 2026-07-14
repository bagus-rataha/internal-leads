package utils

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGenerateToken_Success(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	token, err := GenerateToken(userID, "test@test.com", "user", "secret", time.Hour)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestValidateToken_Success(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	token, _ := GenerateToken(userID, "test@test.com", "user", "secret", time.Hour)

	claims, err := ValidateToken(token, "secret")

	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, "test@test.com", claims.Email)
	assert.Equal(t, "user", claims.Role)
}

func TestValidateToken_InvalidSecret(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	token, _ := GenerateToken(userID, "test@test.com", "user", "secret", time.Hour)

	_, err := ValidateToken(token, "wrong-secret")

	assert.Error(t, err)
}

func TestValidateToken_Expired(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	token, _ := GenerateToken(userID, "test@test.com", "user", "secret", -time.Hour)

	_, err := ValidateToken(token, "secret")

	assert.Error(t, err)
}
