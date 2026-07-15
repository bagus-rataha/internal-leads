package services

import (
	"errors"
	"fiber-api-boilerplate/internal/config"
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/models"
	"fiber-api-boilerplate/internal/utils"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// userRepository is the data-access contract AuthService depends on.
// Defined consumer-side so it can be satisfied by the real repository
// (structural typing) or by a mock in tests.
type userRepository interface {
	FindByID(id uuid.UUID) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	List() ([]models.User, error)
}

// refreshTokenRepository is the refresh-token data-access contract used by AuthService.
type refreshTokenRepository interface {
	Create(rt *models.RefreshToken) error
	FindByToken(token string) (*models.RefreshToken, error)
	DeleteByToken(token string) error
	DeleteAllByUserID(userID uuid.UUID) error
}

type AuthService struct {
	userRepo         userRepository
	refreshTokenRepo refreshTokenRepository
	config           *config.Config
}

func NewAuthService(
	userRepo userRepository,
	refreshTokenRepo refreshTokenRepository,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
 		config:           cfg,
	}
}

// generateAndStoreTokens issues access + refresh tokens and persists the
// refresh token row so it can be rotated and revoked.
func (s *AuthService) generateAndStoreTokens(user *models.User) (*dto.TokenResponse, error) {
	accessToken, err := utils.GenerateToken(
		user.ID, user.Email, user.Role,
		s.config.JWTAccessSecret, s.config.JWTAccessExpire,
	)
	if err != nil {
		return nil, err
	}

	refreshToken, err := utils.GenerateToken(
		user.ID, user.Email, user.Role,
		s.config.JWTRefreshSecret, s.config.JWTRefreshExpire,
	)
	if err != nil {
		return nil, err
	}

	rt := &models.RefreshToken{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiredAt: time.Now().Add(s.config.JWTRefreshExpire),
	}
	if err := s.refreshTokenRepo.Create(rt); err != nil {
		return nil, err
	}

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         dto.ToUserResponse(user),
	}, nil
}

func (s *AuthService) Login(input dto.LoginInput) (*dto.TokenResponse, error) {
	user, err := s.userRepo.FindByEmail(input.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	if !utils.VerifyPassword(input.Password, user.Password) {
		return nil, errors.New("invalid credentials")
	}

	// A deactivated user must not be able to obtain fresh tokens. Revoking
	// their existing refresh tokens on deactivation only closes the old-token
	// window; without this gate they could simply log in again. Same generic
	// message as a bad password so a deactivated account isn't enumerable.
	if !user.IsActive {
		return nil, errors.New("invalid credentials")
	}

	return s.generateAndStoreTokens(user)
}

func (s *AuthService) RefreshToken(refreshToken string) (*dto.TokenResponse, error) {
	claims, err := utils.ValidateToken(refreshToken, s.config.JWTRefreshSecret)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	if _, err := s.refreshTokenRepo.FindByToken(refreshToken); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid or revoked refresh token")
		}
		return nil, err
	}

	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Defense in depth: deactivation revokes stored refresh tokens, so a
	// revoked token is normally already rejected above. This also stops any
	// still-valid token from being rotated into a fresh pair.
	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	if err := s.refreshTokenRepo.DeleteByToken(refreshToken); err != nil {
		return nil, err
	}

	return s.generateAndStoreTokens(user)
}

// Logout revokes a single refresh token. Idempotent — it does not error if
// the token was already absent from the store.
func (s *AuthService) Logout(refreshToken string) error {
	return s.refreshTokenRepo.DeleteByToken(refreshToken)
}

// LogoutAll revokes every refresh token belonging to the user.
func (s *AuthService) LogoutAll(userID uuid.UUID) error {
	return s.refreshTokenRepo.DeleteAllByUserID(userID)
}
