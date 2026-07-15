package models

// District adalah data referensi wilayah read-only hasil seed dari Odoo.
type District struct {
	ID         int    `gorm:"primaryKey"`
	ExternalID string `gorm:"uniqueIndex;not null"`
	CityID     int    `gorm:"not null;index"`
	Name       string `gorm:"not null"`
}
