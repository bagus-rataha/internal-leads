package dto

import (
	"fiber-api-boilerplate/internal/models"
	"time"

	"github.com/google/uuid"
)

// CreateFollowUpInput for POST /leads/:code/followups. created_by_id comes
// from the token, created_at from the server - the body is just the note.
type CreateFollowUpInput struct {
	Note string `json:"note" validate:"required,min=1,max=2000"`
}

// FollowUpResponse for follow-up API responses
type FollowUpResponse struct {
	ID          uuid.UUID `json:"id"`
	LeadID      uuid.UUID `json:"lead_id"`
	Note        string    `json:"note"`
	CreatedByID uuid.UUID `json:"created_by_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// ToFollowUpResponse converts model to DTO
func ToFollowUpResponse(f *models.FollowUp) FollowUpResponse {
	return FollowUpResponse{
		ID:          f.ID,
		LeadID:      f.LeadID,
		Note:        f.Note,
		CreatedByID: f.CreatedByID,
		CreatedAt:   f.CreatedAt,
	}
}

// ToFollowUpResponseList converts models to DTOs
func ToFollowUpResponseList(followUps []models.FollowUp) []FollowUpResponse {
	responses := make([]FollowUpResponse, len(followUps))
	for i, f := range followUps {
		responses[i] = ToFollowUpResponse(&f)
	}
	return responses
}
