package models

// City adalah data referensi wilayah read-only hasil seed dari Odoo.
type City struct {
	ID         int    `gorm:"primaryKey"`
	ExternalID string `gorm:"uniqueIndex;not null"`
	ProvinceID int    `gorm:"not null;index"`
	Name       string `gorm:"not null"`
}
