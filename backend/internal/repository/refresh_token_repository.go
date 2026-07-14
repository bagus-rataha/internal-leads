package repository

import (
	"fiber-api-boilerplate/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RefreshTokenRepository handles database operations for refresh tokens
type RefreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository creates new refresh token repository
func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create stores a refresh token
func (r *RefreshTokenRepository) Create(rt *models.RefreshToken) error {
	return r.db.Create(rt).Error
}

// FindByToken finds refresh token by token value
func (r *RefreshTokenRepository) FindByToken(token string) (*models.RefreshToken, error) {
	var rt models.RefreshToken
	err := r.db.Where("token = ?", token).First(&rt).Error
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

// DeleteByToken removes a refresh token by its value
func (r *RefreshTokenRepository) DeleteByToken(token string) error {
	return r.db.Where("token = ?", token).Delete(&models.RefreshToken{}).Error
}

// DeleteAllByUserID removes all refresh tokens for a user
func (r *RefreshTokenRepository) DeleteAllByUserID(userID uuid.UUID) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.RefreshToken{}).Error
}

// DeleteExpired removes all refresh tokens whose ExpiredAt is in the past.
// Intended to be called by a scheduled cleanup job (cron / goroutine ticker).
func (r *RefreshTokenRepository) DeleteExpired() error {
	return r.db.Where("expired_at < NOW()").Delete(&models.RefreshToken{}).Error
}
