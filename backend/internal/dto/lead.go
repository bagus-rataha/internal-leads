package dto

import (
	"fiber-api-boilerplate/internal/models"
	"time"

	"github.com/google/uuid"
)

// CreateLeadInput for POST /leads. owner_id is role-dependent: ignored and
// forced to self for SALES, validated as a team member for LEADER, free for
// ADMIN_SALES/SU. created_by_id is never part of the body - always the token.
type CreateLeadInput struct {
	CompanyName   string     `json:"company_name" validate:"required,min=2,max=200"`
	BusinessField *string    `json:"business_field" validate:"omitempty,max=200"`
	Website       *string    `json:"website" validate:"omitempty,max=200"`
	ProvinceID    *int       `json:"province_id"`
	CityID        *int       `json:"city_id"`
	DistrictID    *int       `json:"district_id"`
	VillageID     *int       `json:"village_id"`
	ZipID         *int       `json:"zip_id"`
	Rt            *string    `json:"rt" validate:"omitempty,max=10"`
	Rw            *string    `json:"rw" validate:"omitempty,max=10"`
	Street        *string    `json:"street" validate:"omitempty,max=300"`
	PicName       *string    `json:"pic_name" validate:"omitempty,max=200"`
	PicPosition   *string    `json:"pic_position" validate:"omitempty,max=100"`
	OfficePhone   *string    `json:"office_phone" validate:"omitempty,max=30"`
	MobilePhone   *string    `json:"mobile_phone" validate:"omitempty,max=30"`
	Email         *string    `json:"email" validate:"omitempty,email"`
	ServiceTypeID *uuid.UUID `json:"service_type_id"`
	CapacityMbps  *int       `json:"capacity_mbps" validate:"omitempty,gt=0"`
	ExistingIsp   *string    `json:"existing_isp" validate:"omitempty,max=100"`
	Price         *float64   `json:"price" validate:"omitempty,gte=0"`
	OtherServices *string    `json:"other_services" validate:"omitempty,max=500"`
	LeadSourceID  *uuid.UUID `json:"lead_source_id"`
	OwnerID       *uuid.UUID `json:"owner_id"`
}

// UpdateLeadInput for PATCH /leads/:code. All pointers so only submitted
// fields are applied. status/lost_reason are excluded on purpose - those
// only change through PATCH /leads/:code/status. owner_id reassignment is
// honored only for ADMIN_SALES/SU; other roles submitting it are silently
// ignored by the service, not rejected.
type UpdateLeadInput struct {
	CompanyName   *string    `json:"company_name" validate:"omitempty,min=2,max=200"`
	BusinessField *string    `json:"business_field" validate:"omitempty,max=200"`
	Website       *string    `json:"website" validate:"omitempty,max=200"`
	ProvinceID    *int       `json:"province_id"`
	CityID        *int       `json:"city_id"`
	DistrictID    *int       `json:"district_id"`
	VillageID     *int       `json:"village_id"`
	ZipID         *int       `json:"zip_id"`
	Rt            *string    `json:"rt" validate:"omitempty,max=10"`
	Rw            *string    `json:"rw" validate:"omitempty,max=10"`
	Street        *string    `json:"street" validate:"omitempty,max=300"`
	PicName       *string    `json:"pic_name" validate:"omitempty,max=200"`
	PicPosition   *string    `json:"pic_position" validate:"omitempty,max=100"`
	OfficePhone   *string    `json:"office_phone" validate:"omitempty,max=30"`
	MobilePhone   *string    `json:"mobile_phone" validate:"omitempty,max=30"`
	Email         *string    `json:"email" validate:"omitempty,email"`
	ServiceTypeID *uuid.UUID `json:"service_type_id"`
	CapacityMbps  *int       `json:"capacity_mbps" validate:"omitempty,gt=0"`
	ExistingIsp   *string    `json:"existing_isp" validate:"omitempty,max=100"`
	Price         *float64   `json:"price" validate:"omitempty,gte=0"`
	OtherServices *string    `json:"other_services" validate:"omitempty,max=500"`
	LeadSourceID  *uuid.UUID `json:"lead_source_id"`
	OwnerID       *uuid.UUID `json:"owner_id"`
}

// UpdateLeadStatusInput for PATCH /leads/:code/status. FOLLOW_UP is a valid
// enum value here but the service still rejects BARU->FOLLOW_UP - that
// transition only happens through the follow-up endpoint.
type UpdateLeadStatusInput struct {
	Status     string  `json:"status" validate:"required,oneof=FOLLOW_UP HANDOFF_ODOO LOST"`
	LostReason *string `json:"lost_reason" validate:"omitempty,max=500"`
}

// LeadListQuery is the fully-parsed, typed form of GET /leads's query
// params. The handler builds this from raw c.Query() strings (matching the
// manual-parsing convention already used in reference_handler.go) rather
// than Fiber's QueryParser, since that pattern isn't established anywhere
// else in this codebase yet.
type LeadListQuery struct {
	Q            string
	Status       string
	SourceID     *uuid.UUID
	TeamID       *uuid.UUID
	OwnerID      *uuid.UUID
	ProvinceID   *int
	CityID       *int
	DateFrom     *time.Time
	DateTo       *time.Time
	FollowUpFrom *time.Time
	FollowUpTo   *time.Time
	Stale        bool
	Sort         string
	Page         int
	Limit        int
}

// LeadResponse mirrors models.Lead's field set exactly (see
// ToLeadResponse), plus FollowUpCount, OwnerName, CityName, and IsStale.
type LeadResponse struct {
	ID             uuid.UUID  `json:"id"`
	Code           string     `json:"code"`
	Status         string     `json:"status"`
	LostReason     *string    `json:"lost_reason"`
	OwnerID        uuid.UUID  `json:"owner_id"`
	CreatedByID    uuid.UUID  `json:"created_by_id"`
	CompanyName    string     `json:"company_name"`
	BusinessField  *string    `json:"business_field"`
	Website        *string    `json:"website"`
	ProvinceID     *int       `json:"province_id"`
	CityID         *int       `json:"city_id"`
	DistrictID     *int       `json:"district_id"`
	VillageID      *int       `json:"village_id"`
	ZipID          *int       `json:"zip_id"`
	Rt             *string    `json:"rt"`
	Rw             *string    `json:"rw"`
	Street         *string    `json:"street"`
	PicName        *string    `json:"pic_name"`
	PicPosition    *string    `json:"pic_position"`
	OfficePhone    *string    `json:"office_phone"`
	MobilePhone    *string    `json:"mobile_phone"`
	Email          *string    `json:"email"`
	ServiceTypeID  *uuid.UUID `json:"service_type_id"`
	CapacityMbps   *int       `json:"capacity_mbps"`
	ExistingIsp    *string    `json:"existing_isp"`
	Price          *float64   `json:"price"`
	OtherServices  *string    `json:"other_services"`
	LeadSourceID   *uuid.UUID `json:"lead_source_id"`
	LastFollowUpAt *time.Time `json:"last_follow_up_at"`
	FollowUpCount  int        `json:"follow_up_count"`
	CreatedAt      time.Time  `json:"created_at"`
	OwnerName      string     `json:"owner_name"`
	CityName       *string    `json:"city_name"`
	IsStale        bool       `json:"is_stale"`
}

// PaginatedLeadResponse is GET /leads's Data payload - an object, not a bare
// array, so the frontend gets page counts without a separate count request.
type PaginatedLeadResponse struct {
	Items []LeadResponse `json:"items"`
	Total int64          `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

// ownerName reads lead.Owner.Name, defensively falling back to "" if the
// Owner association wasn't preloaded rather than panicking on a nil pointer.
// This depends only on models, unlike is_stale - which needs
// repository.StaleLeadThresholdDays, so that computation lives in the
// service instead and is set on the response after ToLeadResponse returns.
func ownerName(lead *models.Lead) string {
	if lead.Owner == nil {
		return ""
	}
	return lead.Owner.Name
}

// cityName reads lead.City.Name, returning nil when City wasn't preloaded or
// the lead simply has no city_id (unlike Owner, City is legitimately absent
// for a lead whose address hasn't been filled in yet). Pointer, not "", so
// the frontend can distinguish "no city" from an empty name.
func cityName(lead *models.Lead) *string {
	if lead.City == nil {
		return nil
	}
	return &lead.City.Name
}

// ToLeadResponse converts model to DTO
func ToLeadResponse(lead *models.Lead) LeadResponse {
	return LeadResponse{
		ID:             lead.ID,
		Code:           lead.Code,
		Status:         lead.Status,
		LostReason:     lead.LostReason,
		OwnerID:        lead.OwnerID,
		CreatedByID:    lead.CreatedByID,
		CompanyName:    lead.CompanyName,
		BusinessField:  lead.BusinessField,
		Website:        lead.Website,
		ProvinceID:     lead.ProvinceID,
		CityID:         lead.CityID,
		DistrictID:     lead.DistrictID,
		VillageID:      lead.VillageID,
		ZipID:          lead.ZipID,
		Rt:             lead.Rt,
		Rw:             lead.Rw,
		Street:         lead.Street,
		PicName:        lead.PicName,
		PicPosition:    lead.PicPosition,
		OfficePhone:    lead.OfficePhone,
		MobilePhone:    lead.MobilePhone,
		Email:          lead.Email,
		ServiceTypeID:  lead.ServiceTypeID,
		CapacityMbps:   lead.CapacityMbps,
		ExistingIsp:    lead.ExistingIsp,
		Price:          lead.Price,
		OtherServices:  lead.OtherServices,
		LeadSourceID:   lead.LeadSourceID,
		LastFollowUpAt: lead.LastFollowUpAt,
		FollowUpCount:  lead.FollowUpCount,
		CreatedAt:      lead.CreatedAt,
		OwnerName:      ownerName(lead),
		CityName:       cityName(lead),
		// IsStale is intentionally left at its zero value (false) here -
		// computing it needs repository.StaleLeadThresholdDays, and dto must
		// not import repository. The service sets it after calling this.
	}
}

// ToLeadResponseList converts models to DTOs
func ToLeadResponseList(leads []models.Lead) []LeadResponse {
	responses := make([]LeadResponse, len(leads))
	for i, lead := range leads {
		responses[i] = ToLeadResponse(&lead)
	}
	return responses
}
