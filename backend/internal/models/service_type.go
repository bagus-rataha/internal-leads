package models

// ServiceType database model. Jenis layanan (dropdown), seed statis dari migration.
type ServiceType struct {
	BaseModel
	Name     string `gorm:"uniqueIndex;not null"`
	IsActive bool   `gorm:"not null;default:true"`
}
