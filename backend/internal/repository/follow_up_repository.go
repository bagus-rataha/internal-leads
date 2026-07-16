package repository

import (
	"fiber-api-boilerplate/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FollowUpRepository handles database operations for follow-ups
type FollowUpRepository struct {
	db *gorm.DB
}

// NewFollowUpRepository creates new follow-up repository
func NewFollowUpRepository(db *gorm.DB) *FollowUpRepository {
	return &FollowUpRepository{db: db}
}

// Create inserts a new follow-up. Callers that need this atomic with the
// lead's last_follow_up_at/follow_up_count/status updates construct this
// repository with a transaction-scoped *gorm.DB.
func (r *FollowUpRepository) Create(f *models.FollowUp) error {
	return r.db.Create(f).Error
}

// ListByLeadID returns every follow-up for a lead, oldest first.
func (r *FollowUpRepository) ListByLeadID(leadID uuid.UUID) ([]models.FollowUp, error) {
	var followUps []models.FollowUp
	err := r.db.Where("lead_id = ?", leadID).Order("created_at ASC").Find(&followUps).Error
	return followUps, err
}
