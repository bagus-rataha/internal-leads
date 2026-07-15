package models

// Village adalah data referensi wilayah read-only hasil seed dari Odoo.
type Village struct {
	ID         int    `gorm:"primaryKey"`
	ExternalID string `gorm:"uniqueIndex;not null"`
	DistrictID int    `gorm:"not null;index"`
	ZipID      *int
	Name       string `gorm:"not null;index"`
}
