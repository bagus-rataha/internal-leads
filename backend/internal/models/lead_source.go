package models

// LeadSource database model. Sumber lead (dropdown), seed statis dari migration.
type LeadSource struct {
	BaseModel
	Name     string `gorm:"uniqueIndex;not null"`
	IsActive bool   `gorm:"not null;default:true"`
}
