package services

import (
	"fiber-api-boilerplate/internal/config"
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/models"
	"fiber-api-boilerplate/internal/utils"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func newTestConfig() *config.Config {
	return &config.Config{
		JWTAccessSecret:  "test-access-secret",
		JWTRefreshSecret: "test-refresh-secret",
		JWTAccessExpire:  time.Hour,
		JWTRefreshExpire: 24 * time.Hour,
	}
}

func newAuthService(userRepo *MockUserRepository, rtRepo *MockRefreshTokenRepository) *AuthService {
	return NewAuthService(userRepo, rtRepo, newTestConfig())
}

func TestRegister_Success(t *testing.T) {
	userRepo := new(MockUserRepository)
	rtRepo := new(MockRefreshTokenRepository)

	userRepo.On("FindByEmail", "test@test.com").Return(nil, gorm.ErrRecordNotFound)
	userRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)
	rtRepo.On("Create", mock.AnythingOfType("*models.RefreshToken")).Return(nil)

	svc := newAuthService(userRepo, rtRepo)
	result, err := svc.Register(dto.RegisterInput{
		Email:    "test@test.com",
		Password: "123456",
		Name:     "Test",
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.Equal(t, "test@test.com", result.User.Email)
}

func TestRegister_EmailExists(t *testing.T) {
	userRepo := new(MockUserRepository)
	rtRepo := new(MockRefreshTokenRepository)

	existingUser := &models.User{Email: "test@test.com"}
	userRepo.On("FindByEmail", "test@test.com").Return(existingUser, nil)

	svc := newAuthService(userRepo, rtRepo)
	_, err := svc.Register(dto.RegisterInput{
		Email:    "test@test.com",
		Password: "123456",
		Name:     "Test",
	})

	assert.Error(t, err)
	assert.Equal(t, "email already registered", err.Error())
}

func TestLogin_Success(t *testing.T) {
	userRepo := new(MockUserRepository)
	rtRepo := new(MockRefreshTokenRepository)

	hashedPassword, _ := utils.HashPassword("123456")
	user := &models.User{
		BaseModel: models.BaseModel{ID: uuid.Must(uuid.NewV7())},
		Email:     "test@test.com",
		Password:  hashedPassword,
		Name:      "Test",
		Role:      "user",
	}

	userRepo.On("FindByEmail", "test@test.com").Return(user, nil)
	rtRepo.On("Create", mock.AnythingOfType("*models.RefreshToken")).Return(nil)

	svc := newAuthService(userRepo, rtRepo)
	result, err := svc.Login(dto.LoginInput{
		Email:    "test@test.com",
		Password: "123456",
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
}

func TestLogin_UserNotFound(t *testing.T) {
	userRepo := new(MockUserRepository)
	rtRepo := new(MockRefreshTokenRepository)

	userRepo.On("FindByEmail", "test@test.com").Return(nil, gorm.ErrRecordNotFound)

	svc := newAuthService(userRepo, rtRepo)
	_, err := svc.Login(dto.LoginInput{
		Email:    "test@test.com",
		Password: "123456",
	})

	assert.Error(t, err)
	assert.Equal(t, "invalid credentials", err.Error())
}

func TestLogin_WrongPassword(t *testing.T) {
	userRepo := new(MockUserRepository)
	rtRepo := new(MockRefreshTokenRepository)

	hashedPassword, _ := utils.HashPassword("123456")
	user := &models.User{
		Email:    "test@test.com",
		Password: hashedPassword,
	}

	userRepo.On("FindByEmail", "test@test.com").Return(user, nil)

	svc := newAuthService(userRepo, rtRepo)
	_, err := svc.Login(dto.LoginInput{
		Email:    "test@test.com",
		Password: "wrongpassword",
	})

	assert.Error(t, err)
	assert.Equal(t, "invalid credentials", err.Error())
}

func TestLogout_Success(t *testing.T) {
	userRepo := new(MockUserRepository)
	rtRepo := new(MockRefreshTokenRepository)

	rtRepo.On("DeleteByToken", "some-token").Return(nil)

	svc := newAuthService(userRepo, rtRepo)
	err := svc.Logout("some-token")

	assert.NoError(t, err)
}

func TestLogoutAll_Success(t *testing.T) {
	userRepo := new(MockUserRepository)
	rtRepo := new(MockRefreshTokenRepository)
	userID := uuid.Must(uuid.NewV7())

	rtRepo.On("DeleteAllByUserID", userID).Return(nil)

	svc := newAuthService(userRepo, rtRepo)
	err := svc.LogoutAll(userID)

	assert.NoError(t, err)
}
