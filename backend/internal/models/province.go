package models

// Province adalah data referensi wilayah read-only hasil seed dari Odoo.
type Province struct {
	ID         int    `gorm:"primaryKey"`
	ExternalID string `gorm:"uniqueIndex;not null"`
	Code       *string
	Name       string `gorm:"not null"`
}
