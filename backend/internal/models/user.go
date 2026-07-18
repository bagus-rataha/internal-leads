package models

import "github.com/google/uuid"

// User database model
type User struct {
	BaseModel
	Email    string     `gorm:"uniqueIndex;not null"`
	Password string     `gorm:"not null"`
	Name     string     `gorm:"not null"`
	Role     string     `gorm:"default:SALES"`
	TeamID   *uuid.UUID `gorm:"type:uuid;index"`
	IsActive bool       `gorm:"not null;default:true"`

	// Team is read-only, populated via Preload for display purposes (e.g.
	// owner_team_name in lead API responses). Never written to by this model.
	Team *SalesTeam `gorm:"foreignKey:TeamID"`
}
