package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword_Success(t *testing.T) {
	hash, err := HashPassword("password123")

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, "password123", hash)
}

func TestVerifyPassword_Correct(t *testing.T) {
	hash, _ := HashPassword("password123")
	result := VerifyPassword("password123", hash)

	assert.True(t, result)
}

func TestVerifyPassword_Incorrect(t *testing.T) {
	hash, _ := HashPassword("password123")
	result := VerifyPassword("wrongpassword", hash)

	assert.False(t, result)
}
