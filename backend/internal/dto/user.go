package dto

import (
	"fiber-api-boilerplate/internal/models"
	"time"

	"github.com/google/uuid"
)

// UserResponse for user API responses
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// UpdateProfileInput for profile updates
type UpdateProfileInput struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

// ToUserResponse converts model to DTO
func ToUserResponse(user *models.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Role:      user.Role,
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
