package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FollowUp is append-only: no updated_at, no soft-delete, no update/delete
// endpoint. BaseModel isn't embedded because that pattern brings both of
// those along.
type FollowUp struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	LeadID      uuid.UUID `gorm:"type:uuid;not null;index:idx_follow_ups_lead_id_created_at"`
	Note        string    `gorm:"not null"`
	CreatedByID uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt   time.Time `gorm:"index:idx_follow_ups_lead_id_created_at"`
}

// BeforeCreate generates UUID v7 before INSERT if the ID isn't already set,
// mirroring BaseModel.BeforeCreate without pulling in the timestamp/soft
// -delete fields BaseModel also carries.
func (f *FollowUp) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.Must(uuid.NewV7())
	}
	return nil
}
