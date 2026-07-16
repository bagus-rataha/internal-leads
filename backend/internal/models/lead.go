package models

import (
	"time"

	"github.com/google/uuid"
)

// Lead database model
type Lead struct {
	BaseModel
	Code           string `gorm:"uniqueIndex;not null"`
	Status         string `gorm:"not null;default:BARU;index"`
	LostReason     *string
	OwnerID        uuid.UUID `gorm:"type:uuid;not null;index"`
	CreatedByID    uuid.UUID `gorm:"type:uuid;not null"`
	CompanyName    string    `gorm:"not null"`
	BusinessField  *string
	Website        *string
	ProvinceID     *int
	CityID         *int `gorm:"index"`
	DistrictID     *int
	VillageID      *int
	ZipID          *int
	Rt             *string
	Rw             *string
	Street         *string
	PicName        *string
	PicPosition    *string
	OfficePhone    *string
	MobilePhone    *string
	Email          *string
	ServiceTypeID  *uuid.UUID `gorm:"type:uuid"`
	CapacityMbps   *int
	ExistingIsp    *string
	Price          *float64
	OtherServices  *string
	LeadSourceID   *uuid.UUID `gorm:"type:uuid"`
	LastFollowUpAt *time.Time `gorm:"index"`
	FollowUpCount  int        `gorm:"not null;default:0"`
}
