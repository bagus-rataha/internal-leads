package models

// Zip adalah data referensi wilayah read-only hasil seed dari Odoo.
type Zip struct {
	ID         int    `gorm:"primaryKey"`
	ExternalID string `gorm:"uniqueIndex;not null"`
	Code       string `gorm:"not null"`
	CityID     int    `gorm:"not null"`
	DistrictID int    `gorm:"not null;index"`
}
