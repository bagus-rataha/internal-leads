package models

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken database model
type RefreshToken struct {
	BaseModel
	Token     string    `gorm:"uniqueIndex;type:text;not null"`
	UserID    uuid.UUID `gorm:"type:uuid;index;not null"`
	ExpiredAt time.Time `gorm:"index;not null"`
}
