package models

// User database model
type User struct {
	BaseModel
	Email    string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null"`
	Name     string `gorm:"not null"`
	Role     string `gorm:"default:user"`
}
