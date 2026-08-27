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
	ForecastMrr    *float64
	LastFollowUpAt *time.Time `gorm:"index"`
	FollowUpCount  int        `gorm:"not null;default:0"`

	// Owner is read-only, populated via Preload for display purposes (e.g.
	// owner_name in the API response). Never written to by this model.
	Owner *User `gorm:"foreignKey:OwnerID"`

	// City is read-only, populated via Preload for display purposes (e.g.
	// city_name in the API response). Never written to by this model.
	City *City `gorm:"foreignKey:CityID"`

	// The following are read-only, populated via Preload ONLY on the Detail
	// read path (LeadRepository.FindDetailByCode) - List and the plain
	// FindByCode (used by Update/UpdateStatus/FollowUp's write paths) never
	// load these, so those queries don't pay for Preloads they don't render.
	Province    *Province    `gorm:"foreignKey:ProvinceID"`
	District    *District    `gorm:"foreignKey:DistrictID"`
	Village     *Village     `gorm:"foreignKey:VillageID"`
	Zip         *Zip         `gorm:"foreignKey:ZipID"`
	ServiceType *ServiceType `gorm:"foreignKey:ServiceTypeID"`
	LeadSource  *LeadSource  `gorm:"foreignKey:LeadSourceID"`
	CreatedBy   *User        `gorm:"foreignKey:CreatedByID"`
}
