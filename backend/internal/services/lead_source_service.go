package services

import (
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/models"

	"github.com/google/uuid"
)

// leadSourceRepository is the subset of repository methods LeadSourceService
// needs. Defined consumer-side for testability (satisfied by the real repo
// or a mock).
type leadSourceRepository interface {
	Create(source *models.LeadSource) error
	FindByID(id uuid.UUID) (*models.LeadSource, error)
	List() ([]models.LeadSource, error)
	Update(source *models.LeadSource) error
}

type LeadSourceService struct {
	sourceRepo leadSourceRepository
}

func NewLeadSourceService(sourceRepo leadSourceRepository) *LeadSourceService {
	return &LeadSourceService{sourceRepo: sourceRepo}
}

func (s *LeadSourceService) Create(input dto.CreateLeadSourceInput) (*dto.LeadSourceAdminResponse, error) {
	source := &models.LeadSource{
		Name:     input.Name,
		IsActive: true,
	}

	if err := s.sourceRepo.Create(source); err != nil {
		return nil, err
	}

	response := dto.ToLeadSourceAdminResponse(source)
	return &response, nil
}

func (s *LeadSourceService) List() ([]dto.LeadSourceAdminResponse, error) {
	sources, err := s.sourceRepo.List()
	if err != nil {
		return nil, err
	}
	return dto.ToLeadSourceAdminResponseList(sources), nil
}

// Update applies the given fields. Unlike TeamService.Update, deactivating a
// lead source is never blocked by existing lead references - a lead's
// lead_source_id stays a valid FK regardless of is_active, it just drops out
// of the /refs/lead-sources dropdown used by future create/edit-lead forms.
func (s *LeadSourceService) Update(id uuid.UUID, input dto.UpdateLeadSourceInput) (*dto.LeadSourceAdminResponse, error) {
	source, err := s.sourceRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		source.Name = *input.Name
	}
	if input.IsActive != nil {
		source.IsActive = *input.IsActive
	}

	if err := s.sourceRepo.Update(source); err != nil {
		return nil, err
	}

	response := dto.ToLeadSourceAdminResponse(source)
	return &response, nil
}
