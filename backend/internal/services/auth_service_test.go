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

func TestLogin_Success(t *testing.T) {
	userRepo := new(MockUserRepository)
	rtRepo := new(MockRefreshTokenRepository)

	hashedPassword, _ := utils.HashPassword("123456")
	user := &models.User{
		BaseModel: models.BaseModel{ID: uuid.Must(uuid.NewV7())},
		Email:     "test@test.com",
		Password:  hashedPassword,
		Name:      "Test",
		Role:      "SU",
		IsActive:  true,
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

func TestLogin_InactiveUser(t *testing.T) {
	userRepo := new(MockUserRepository)
	rtRepo := new(MockRefreshTokenRepository)

	hashedPassword, _ := utils.HashPassword("123456")
	user := &models.User{
		BaseModel: models.BaseModel{ID: uuid.Must(uuid.NewV7())},
		Email:     "test@test.com",
		Password:  hashedPassword,
		Name:      "Test",
		Role:      "SALES",
		IsActive:  false,
	}

	userRepo.On("FindByEmail", "test@test.com").Return(user, nil)

	svc := newAuthService(userRepo, rtRepo)
	_, err := svc.Login(dto.LoginInput{
		Email:    "test@test.com",
		Password: "123456",
	})

	// Correct password, but the account is deactivated: no tokens are issued,
	// and the message is indistinguishable from a bad password.
	assert.Error(t, err)
	assert.Equal(t, "invalid credentials", err.Error())
	rtRepo.AssertNotCalled(t, "Create", mock.Anything)
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

func TestRefreshToken_GraceWindow_SameOldTokenReturnsSamePair(t *testing.T) {
	userRepo := new(MockUserRepository)
	rtRepo := new(MockRefreshTokenRepository)

	user := &models.User{
		BaseModel: models.BaseModel{ID: uuid.Must(uuid.NewV7())},
		Email:     "test@test.com",
		Role:      "SU",
		IsActive:  true,
	}
	oldToken := "old-refresh-token"

	rtRepo.On("FindByToken", oldToken).Return(&models.RefreshToken{Token: oldToken, UserID: user.ID}, nil).Once()
	rtRepo.On("DeleteByToken", oldToken).Return(nil).Once()
	userRepo.On("FindByID", user.ID).Return(user, nil).Once()
	rtRepo.On("Create", mock.AnythingOfType("*models.RefreshToken")).Return(nil).Once()

	svc := newAuthService(userRepo, rtRepo)
	// The mock config's JWTRefreshSecret must actually sign oldToken for
	// ValidateToken to succeed on the first call — generate it the same way
	// Login/generateAndStoreTokens would, rather than a literal string.
	signedOldToken, err := utils.GenerateToken(user.ID, user.Email, user.Role, newTestConfig().JWTRefreshSecret, time.Hour)
	assert.NoError(t, err)

	rtRepo.ExpectedCalls = nil // reset the string-literal expectations above
	rtRepo.On("FindByToken", signedOldToken).Return(&models.RefreshToken{Token: signedOldToken, UserID: user.ID}, nil).Once()
	rtRepo.On("DeleteByToken", signedOldToken).Return(nil).Once()
	rtRepo.On("Create", mock.AnythingOfType("*models.RefreshToken")).Return(nil).Once()

	first, err := svc.RefreshToken(signedOldToken)
	assert.NoError(t, err)
	assert.NotEmpty(t, first.AccessToken)

	// Second call with the SAME now-rotated old token must NOT hit
	// FindByToken/DeleteByToken/Create again (mock.Once() above would fail
	// the test on a second real call) - it must be served from the grace
	// cache, returning the identical pair.
	second, err := svc.RefreshToken(signedOldToken)
	assert.NoError(t, err)
	assert.Equal(t, first.AccessToken, second.AccessToken)
	assert.Equal(t, first.RefreshToken, second.RefreshToken)

	userRepo.AssertExpectations(t)
	rtRepo.AssertExpectations(t)
}

func TestRefreshToken_GraceWindow_ExpiredEntryFallsThrough(t *testing.T) {
	userRepo := new(MockUserRepository)
	rtRepo := new(MockRefreshTokenRepository)
	user := &models.User{
		BaseModel: models.BaseModel{ID: uuid.Must(uuid.NewV7())},
		Email:     "test@test.com",
		Role:      "SU",
		IsActive:  true,
	}

	svc := newAuthService(userRepo, rtRepo)
	oldToken, err := utils.GenerateToken(user.ID, user.Email, user.Role, newTestConfig().JWTRefreshSecret, time.Hour)
	assert.NoError(t, err)

	// Manually seed an already-expired grace entry (bypassing the real 5s
	// wait) to prove expired entries are not served and fall through to a
	// real (here: not-found) lookup instead of being trusted forever.
	svc.storeRefreshGraceForTest(oldToken, &dto.TokenResponse{AccessToken: "stale"}, time.Now().Add(-time.Second))

	rtRepo.On("FindByToken", oldToken).Return(nil, gorm.ErrRecordNotFound).Once()

	_, err = svc.RefreshToken(oldToken)
	assert.Error(t, err)
	assert.Equal(t, "invalid or revoked refresh token", err.Error())
}

// storeRefreshGraceForTest lets tests seed a grace-cache entry with an
// arbitrary expiry, so expiry behavior can be tested without a real sleep.
func (s *AuthService) storeRefreshGraceForTest(oldToken string, response *dto.TokenResponse, expiresAt time.Time) {
	s.graceMu.Lock()
	defer s.graceMu.Unlock()
	s.graceCache[oldToken] = refreshGraceEntry{response: response, expiresAt: expiresAt}
}
