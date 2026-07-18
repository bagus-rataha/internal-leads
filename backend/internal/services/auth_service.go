package services

import (
	"errors"
	"fiber-api-boilerplate/internal/config"
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/models"
	"fiber-api-boilerplate/internal/utils"
	"sync"
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

// refreshGraceWindow is how long a just-rotated refresh token still returns
// the pair its rotation produced instead of "invalid or revoked refresh
// token". This absorbs a real race: two refresh calls landing close
// together for the same old token (most plausibly two browser tabs
// bootstrapping around the same time — the frontend's single-flight guard
// only dedupes within one tab) would otherwise have the loser bounced to
// login even though nothing was actually expired. The window is
// deliberately short: a delayed revoke (deactivate/role-change) racing a
// grace-served refresh could still hand out a valid pair for up to this
// long after the revoke - an accepted, bounded trade-off for this app's
// scale, not something to close further without discussion.
const refreshGraceWindow = 5 * time.Second

// refreshGraceEntry is one cached rotation result, keyed by the OLD token
// string that produced it.
type refreshGraceEntry struct {
	response  *dto.TokenResponse
	expiresAt time.Time
}

type AuthService struct {
	userRepo         userRepository
	refreshTokenRepo refreshTokenRepository
	config           *config.Config

	graceMu    sync.Mutex
	graceCache map[string]refreshGraceEntry
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
		graceCache:       make(map[string]refreshGraceEntry),
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

// checkRefreshGrace returns the token pair a previous rotation of oldToken
// already produced, if that rotation happened within refreshGraceWindow.
// Expired entries are evicted on access - no background sweep needed at
// this app's scale.
func (s *AuthService) checkRefreshGrace(oldToken string) (*dto.TokenResponse, bool) {
	s.graceMu.Lock()
	defer s.graceMu.Unlock()

	entry, ok := s.graceCache[oldToken]
	if !ok {
		return nil, false
	}
	if time.Now().After(entry.expiresAt) {
		delete(s.graceCache, oldToken)
		return nil, false
	}
	return entry.response, true
}

// storeRefreshGrace records the result of rotating oldToken so a racing
// duplicate refresh call gets the same pair instead of an error. Also
// opportunistically evicts any other expired entries, keeping the map from
// growing unbounded without a dedicated cleanup goroutine.
func (s *AuthService) storeRefreshGrace(oldToken string, response *dto.TokenResponse) {
	s.graceMu.Lock()
	defer s.graceMu.Unlock()

	now := time.Now()
	for token, entry := range s.graceCache {
		if now.After(entry.expiresAt) {
			delete(s.graceCache, token)
		}
	}
	s.graceCache[oldToken] = refreshGraceEntry{response: response, expiresAt: now.Add(refreshGraceWindow)}
}

func (s *AuthService) RefreshToken(refreshToken string) (*dto.TokenResponse, error) {
	if cached, ok := s.checkRefreshGrace(refreshToken); ok {
		return cached, nil
	}

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

	result, err := s.generateAndStoreTokens(user)
	if err != nil {
		return nil, err
	}

	s.storeRefreshGrace(refreshToken, result)
	return result, nil
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
