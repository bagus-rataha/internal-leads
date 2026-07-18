package dto

import (
	"fiber-api-boilerplate/internal/models"
	"time"

	"github.com/google/uuid"
)

// LeadSourceAdminResponse for lead source management API responses (distinct
// from LeadSourceResponse in reference.go, which only exposes id+name for
// the read-only /refs/lead-sources dropdown).
type LeadSourceAdminResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateLeadSourceInput for creating a lead source
type CreateLeadSourceInput struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

// UpdateLeadSourceInput for updating a lead source. Fields are pointers so
// only the ones actually provided by the caller are applied.
type UpdateLeadSourceInput struct {
	Name     *string `json:"name" validate:"omitempty,min=2,max=100"`
	IsActive *bool   `json:"is_active"`
}

// ToLeadSourceAdminResponse converts model to DTO
func ToLeadSourceAdminResponse(source *models.LeadSource) LeadSourceAdminResponse {
	return LeadSourceAdminResponse{
		ID:        source.ID,
		Name:      source.Name,
		IsActive:  source.IsActive,
		CreatedAt: source.CreatedAt,
	}
}

// ToLeadSourceAdminResponseList converts models to DTOs
func ToLeadSourceAdminResponseList(sources []models.LeadSource) []LeadSourceAdminResponse {
	responses := make([]LeadSourceAdminResponse, len(sources))
	for i, s := range sources {
		responses[i] = ToLeadSourceAdminResponse(&s)
	}
	return responses
}
