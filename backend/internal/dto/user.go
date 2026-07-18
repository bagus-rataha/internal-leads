package dto

import (
	"fiber-api-boilerplate/internal/models"
	"time"

	"github.com/google/uuid"
)

// UserResponse for user API responses
type UserResponse struct {
	ID        uuid.UUID  `json:"id"`
	Email     string     `json:"email"`
	Name      string     `json:"name"`
	Role      string     `json:"role"`
	TeamID    *uuid.UUID `json:"team_id"`
	TeamName  *string    `json:"team_name"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
}

// UpdateProfileInput for profile updates
type UpdateProfileInput struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

// CreateUserInput for admin-created accounts. There is no self-registration
// endpoint — this is the only way a new user is created.
type CreateUserInput struct {
	Name     string     `json:"name" validate:"required,min=2,max=100"`
	Email    string     `json:"email" validate:"required,email"`
	Password string     `json:"password" validate:"required,min=6,max=100"`
	Role     string     `json:"role" validate:"required,oneof=SALES LEADER ADMIN_SALES SU"`
	TeamID   *uuid.UUID `json:"team_id"`
}

// UpdateUserInput for admin edits to a user. Fields are pointers so only the
// ones actually provided by the caller are applied.
type UpdateUserInput struct {
	Name   *string    `json:"name" validate:"omitempty,min=2,max=100"`
	Role   *string    `json:"role" validate:"omitempty,oneof=SALES LEADER ADMIN_SALES SU"`
	TeamID *uuid.UUID `json:"team_id"`

	// IsActive, when true, reactivates a previously deactivated user. false
	// is rejected by the service — deactivation only happens through
	// POST /users/:id/deactivate, which carries the active-lead-reassignment
	// and last-active-administrator safety checks this endpoint doesn't.
	IsActive *bool `json:"is_active"`
}

// ResetPasswordInput for admin-triggered password resets.
type ResetPasswordInput struct {
	NewPassword string `json:"new_password" validate:"required,min=6,max=100"`
}

// DeactivateUserInput for admin deactivation. If the user still owns active
// leads, ReassignToUserID must be set or the request is rejected.
type DeactivateUserInput struct {
	ReassignToUserID *uuid.UUID `json:"reassign_to_user_id"`
}

// teamName reads user.Team.Name, nil when the relation wasn't preloaded or
// the user legitimately has no team (ADMIN_SALES/SU can be teamless).
func teamName(user *models.User) *string {
	if user.Team == nil {
		return nil
	}
	return &user.Team.Name
}

// ToUserResponse converts model to DTO
func ToUserResponse(user *models.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Role:      user.Role,
		TeamID:    user.TeamID,
		TeamName:  teamName(user),
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
	}
}

// ToUserResponseList converts models to DTOs
func ToUserResponseList(users []models.User) []UserResponse {
	responses := make([]UserResponse, len(users))
	for i, user := range users {
		responses[i] = ToUserResponse(&user)
	}
	return responses
}
