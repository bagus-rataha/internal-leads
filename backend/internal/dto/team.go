package dto

import (
	"fiber-api-boilerplate/internal/models"
	"time"

	"github.com/google/uuid"
)

// TeamResponse for sales team API responses
type TeamResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateTeamInput for creating a sales team
type CreateTeamInput struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

// UpdateTeamInput for updating a sales team. Fields are pointers so only the
// ones actually provided by the caller are applied.
type UpdateTeamInput struct {
	Name     *string `json:"name" validate:"omitempty,min=2,max=100"`
	IsActive *bool   `json:"is_active"`
}

// ToTeamResponse converts model to DTO
func ToTeamResponse(team *models.SalesTeam) TeamResponse {
	return TeamResponse{
		ID:        team.ID,
		Name:      team.Name,
		IsActive:  team.IsActive,
		CreatedAt: team.CreatedAt,
	}
}

// ToTeamResponseList converts models to DTOs
func ToTeamResponseList(teams []models.SalesTeam) []TeamResponse {
	responses := make([]TeamResponse, len(teams))
	for i, team := range teams {
		responses[i] = ToTeamResponse(&team)
	}
	return responses
}
