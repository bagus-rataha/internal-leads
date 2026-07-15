package models

import "github.com/google/uuid"

// User database model
type User struct {
	BaseModel
	Email    string     `gorm:"uniqueIndex;not null"`
	Password string     `gorm:"not null"`
	Name     string     `gorm:"not null"`
	Role     string     `gorm:"default:user"`
	TeamID   *uuid.UUID `gorm:"type:uuid;index"`
	IsActive bool       `gorm:"not null;default:true"`
}
