package dto

import (
	"fiber-api-boilerplate/internal/models"
	"time"

	"github.com/google/uuid"
)

// ServiceTypeAdminResponse for service type management API responses
// (distinct from ServiceTypeResponse in reference.go, id+name only, used by
// the read-only /refs/service-types dropdown).
type ServiceTypeAdminResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateServiceTypeInput for creating a service type
type CreateServiceTypeInput struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

// UpdateServiceTypeInput for updating a service type. Fields are pointers so
// only the ones actually provided by the caller are applied.
type UpdateServiceTypeInput struct {
	Name     *string `json:"name" validate:"omitempty,min=2,max=100"`
	IsActive *bool   `json:"is_active"`
}

// ToServiceTypeAdminResponse converts model to DTO
func ToServiceTypeAdminResponse(serviceType *models.ServiceType) ServiceTypeAdminResponse {
	return ServiceTypeAdminResponse{
		ID:        serviceType.ID,
		Name:      serviceType.Name,
		IsActive:  serviceType.IsActive,
		CreatedAt: serviceType.CreatedAt,
	}
}

// ToServiceTypeAdminResponseList converts models to DTOs
func ToServiceTypeAdminResponseList(types []models.ServiceType) []ServiceTypeAdminResponse {
	responses := make([]ServiceTypeAdminResponse, len(types))
	for i, st := range types {
		responses[i] = ToServiceTypeAdminResponse(&st)
	}
	return responses
}
