package models

// SalesTeam database model
type SalesTeam struct {
	BaseModel
	Name     string `gorm:"not null"`
	IsActive bool   `gorm:"not null;default:true"`
}
