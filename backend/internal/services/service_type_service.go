package services

import (
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/models"

	"github.com/google/uuid"
)

// serviceTypeRepository is the subset of repository methods ServiceTypeService
// needs. Defined consumer-side for testability.
type serviceTypeRepository interface {
	Create(serviceType *models.ServiceType) error
	FindByID(id uuid.UUID) (*models.ServiceType, error)
	List() ([]models.ServiceType, error)
	Update(serviceType *models.ServiceType) error
}

type ServiceTypeService struct {
	typeRepo serviceTypeRepository
}

func NewServiceTypeService(typeRepo serviceTypeRepository) *ServiceTypeService {
	return &ServiceTypeService{typeRepo: typeRepo}
}

func (s *ServiceTypeService) Create(input dto.CreateServiceTypeInput) (*dto.ServiceTypeAdminResponse, error) {
	serviceType := &models.ServiceType{
		Name:     input.Name,
		IsActive: true,
	}

	if err := s.typeRepo.Create(serviceType); err != nil {
		return nil, err
	}

	response := dto.ToServiceTypeAdminResponse(serviceType)
	return &response, nil
}

func (s *ServiceTypeService) List() ([]dto.ServiceTypeAdminResponse, error) {
	types, err := s.typeRepo.List()
	if err != nil {
		return nil, err
	}
	return dto.ToServiceTypeAdminResponseList(types), nil
}

// Update applies the given fields. Same no-active-reference-guard behavior
// as LeadSourceService.Update - see that file's comment.
func (s *ServiceTypeService) Update(id uuid.UUID, input dto.UpdateServiceTypeInput) (*dto.ServiceTypeAdminResponse, error) {
	serviceType, err := s.typeRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		serviceType.Name = *input.Name
	}
	if input.IsActive != nil {
		serviceType.IsActive = *input.IsActive
	}

	if err := s.typeRepo.Update(serviceType); err != nil {
		return nil, err
	}

	response := dto.ToServiceTypeAdminResponse(serviceType)
	return &response, nil
}
