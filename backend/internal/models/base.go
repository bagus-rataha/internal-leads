package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel adalah shared base struct pengganti gorm.Model dengan UUID v7 sebagai primary key.
// Embed BaseModel di model yang ingin pakai UUID + timestamp standard.
//
// Untuk tipe ID lain (auto-increment, NanoID, dll), jangan embed BaseModel.
// Define ID field manual + BeforeCreate hook sendiri.
type BaseModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// BeforeCreate generates UUID v7 (time-sortable) before INSERT bila ID belum diset.
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.Must(uuid.NewV7())
	}
	return nil
}
