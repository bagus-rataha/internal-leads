package services

import (
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/models"
	"fiber-api-boilerplate/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// followUpRepositoryForFollowUp is the subset of repository methods
// FollowUpService needs. Defined consumer-side for testability.
type followUpRepositoryForFollowUp interface {
	Create(f *models.FollowUp) error
	ListByLeadID(leadID uuid.UUID) ([]models.FollowUp, error)
}

// leadRepositoryForFollowUp is the subset of lead-repository methods
// FollowUpService needs - narrower than leadRepositoryForLead (no NextCode,
// no List, no plain Update), since follow-up only ever fetches a lead to
// scope-check it and applies the two follow-up-specific updates.
type leadRepositoryForFollowUp interface {
	FindByCode(scope repository.LeadScope, code string) (*models.Lead, error)
	RecordFollowUp(leadID uuid.UUID) error
	MaybeTransitionToFollowUp(leadID uuid.UUID) error
}

type FollowUpService struct {
	db           *gorm.DB
	followUpRepo followUpRepositoryForFollowUp
	leadRepo     leadRepositoryForFollowUp
	userRepo     userRepositoryForLead
}

func NewFollowUpService(
	db *gorm.DB,
	followUpRepo followUpRepositoryForFollowUp,
	leadRepo leadRepositoryForFollowUp,
	userRepo userRepositoryForLead,
) *FollowUpService {
	return &FollowUpService{db: db, followUpRepo: followUpRepo, leadRepo: leadRepo, userRepo: userRepo}
}

// Create writes a follow-up and its side effects on the parent lead in one
// transaction: insert follow_up, record last_follow_up_at/follow_up_count,
// and conditionally flip BARU->FOLLOW_UP. Access is scope-based: a lead
// outside the caller's scope returns ErrLeadNotFound, same as LeadService.
func (s *FollowUpService) Create(callerID uuid.UUID, role, leadCode string, input dto.CreateFollowUpInput) (*dto.FollowUpResponse, error) {
	var result *dto.FollowUpResponse
	err := s.db.Transaction(func(tx *gorm.DB) error {
		txLeadRepo := repository.NewLeadRepository(tx)
		txFollowUpRepo := repository.NewFollowUpRepository(tx)

		response, err := s.createFollowUp(txFollowUpRepo, txLeadRepo, callerID, role, leadCode, input)
		if err != nil {
			return err
		}
		result = response
		return nil
	})
	return result, err
}

// createFollowUp holds the actual logic, taking the repos as parameters
// rather than reading them off s - in production s.Create supplies
// tx-scoped repos so the four writes commit together; in tests this can be
// called directly with mocks.
func (s *FollowUpService) createFollowUp(
	followUpRepo followUpRepositoryForFollowUp,
	leadRepo leadRepositoryForFollowUp,
	callerID uuid.UUID,
	role, leadCode string,
	input dto.CreateFollowUpInput,
) (*dto.FollowUpResponse, error) {
	scope, err := buildLeadScope(s.userRepo, callerID, role)
	if err != nil {
		return nil, err
	}

	lead, err := leadRepo.FindByCode(scope, leadCode)
	if err != nil {
		return nil, ErrLeadNotFound
	}

	followUp := &models.FollowUp{
		LeadID:      lead.ID,
		Note:        input.Note,
		CreatedByID: callerID,
	}
	if err := followUpRepo.Create(followUp); err != nil {
		return nil, err
	}
	if err := leadRepo.RecordFollowUp(lead.ID); err != nil {
		return nil, err
	}
	if err := leadRepo.MaybeTransitionToFollowUp(lead.ID); err != nil {
		return nil, err
	}

	// followUp is in-memory only (no Preload happened) - fetch the creator
	// so created_by_name isn't empty on the response, mirroring createLead's
	// post-Create owner-name fetch.
	creator, err := s.userRepo.FindByID(callerID)
	if err != nil {
		return nil, err
	}
	followUp.CreatedBy = creator

	response := dto.ToFollowUpResponse(followUp)
	return &response, nil
}

// List returns every follow-up for a lead, scoped through the parent lead.
func (s *FollowUpService) List(callerID uuid.UUID, role, leadCode string) ([]dto.FollowUpResponse, error) {
	scope, err := buildLeadScope(s.userRepo, callerID, role)
	if err != nil {
		return nil, err
	}

	lead, err := s.leadRepo.FindByCode(scope, leadCode)
	if err != nil {
		return nil, ErrLeadNotFound
	}

	followUps, err := s.followUpRepo.ListByLeadID(lead.ID)
	if err != nil {
		return nil, err
	}

	return dto.ToFollowUpResponseList(followUps), nil
}
