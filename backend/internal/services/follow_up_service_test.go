package services

import (
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/models"
	"fiber-api-boilerplate/internal/repository"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestFollowUpCreate_Success_RecordsAndTransitions(t *testing.T) {
	followUpRepo := new(MockFollowUpRepository)
	leadRepo := new(MockLeadRepositoryForFollowUp)
	userRepo := new(MockUserRepository)

	callerID := uuid.Must(uuid.NewV7())
	leadID := uuid.Must(uuid.NewV7())
	lead := &models.Lead{Code: "LD-2607-3001", Status: "BARU"}
	lead.ID = leadID

	leadRepo.On("FindByCode", repository.LeadScope{}, "LD-2607-3001").Return(lead, nil)
	followUpRepo.On("Create", mock.AnythingOfType("*models.FollowUp")).Return(nil)
	leadRepo.On("RecordFollowUp", leadID).Return(nil)
	leadRepo.On("MaybeTransitionToFollowUp", leadID).Return(nil)
	userRepo.On("FindByID", callerID).Return(&models.User{Name: "Test Caller"}, nil)

	svc := NewFollowUpService(nil, followUpRepo, leadRepo, userRepo)
	result, err := svc.createFollowUp(followUpRepo, leadRepo, callerID, "ADMIN_SALES", "LD-2607-3001", dto.CreateFollowUpInput{Note: "called, will follow up next week"})

	assert.NoError(t, err)
	assert.Equal(t, "called, will follow up next week", result.Note)
	assert.Equal(t, callerID, result.CreatedByID)
	leadRepo.AssertCalled(t, "RecordFollowUp", leadID)
	leadRepo.AssertCalled(t, "MaybeTransitionToFollowUp", leadID)
}

func TestFollowUpCreate_LeadOutOfScope_NotFound(t *testing.T) {
	followUpRepo := new(MockFollowUpRepository)
	leadRepo := new(MockLeadRepositoryForFollowUp)
	userRepo := new(MockUserRepository)

	salesID := uuid.Must(uuid.NewV7())
	leadRepo.On("FindByCode", repository.LeadScope{OwnerID: &salesID}, "LD-2607-3002").
		Return(nil, gorm.ErrRecordNotFound)

	svc := NewFollowUpService(nil, followUpRepo, leadRepo, userRepo)
	_, err := svc.createFollowUp(followUpRepo, leadRepo, salesID, "SALES", "LD-2607-3002", dto.CreateFollowUpInput{Note: "x"})

	assert.ErrorIs(t, err, ErrLeadNotFound)
	followUpRepo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestFollowUpList_LeadOutOfScope_NotFound(t *testing.T) {
	followUpRepo := new(MockFollowUpRepository)
	leadRepo := new(MockLeadRepositoryForFollowUp)
	userRepo := new(MockUserRepository)

	salesID := uuid.Must(uuid.NewV7())
	leadRepo.On("FindByCode", repository.LeadScope{OwnerID: &salesID}, "LD-2607-3003").
		Return(nil, gorm.ErrRecordNotFound)

	svc := NewFollowUpService(nil, followUpRepo, leadRepo, userRepo)
	_, err := svc.List(salesID, "SALES", "LD-2607-3003")

	assert.ErrorIs(t, err, ErrLeadNotFound)
}

func TestFollowUpList_Success(t *testing.T) {
	followUpRepo := new(MockFollowUpRepository)
	leadRepo := new(MockLeadRepositoryForFollowUp)
	userRepo := new(MockUserRepository)

	leadID := uuid.Must(uuid.NewV7())
	lead := &models.Lead{Code: "LD-2607-3004"}
	lead.ID = leadID
	leadRepo.On("FindByCode", repository.LeadScope{}, "LD-2607-3004").Return(lead, nil)
	followUpRepo.On("ListByLeadID", leadID).Return([]models.FollowUp{{Note: "first"}}, nil)

	svc := NewFollowUpService(nil, followUpRepo, leadRepo, userRepo)
	result, err := svc.List(uuid.Must(uuid.NewV7()), "ADMIN_SALES", "LD-2607-3004")

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestFollowUpCreate_CreatedByNamePopulatedInResponse(t *testing.T) {
	followUpRepo := new(MockFollowUpRepository)
	leadRepo := new(MockLeadRepositoryForFollowUp)
	userRepo := new(MockUserRepository)
	callerID := uuid.Must(uuid.NewV7())
	leadID := uuid.Must(uuid.NewV7())

	leadRepo.On("FindByCode", mock.Anything, "LD-2607-0001").Return(&models.Lead{BaseModel: models.BaseModel{ID: leadID}, Status: "BARU"}, nil)
	followUpRepo.On("Create", mock.AnythingOfType("*models.FollowUp")).Return(nil)
	leadRepo.On("RecordFollowUp", leadID).Return(nil)
	leadRepo.On("MaybeTransitionToFollowUp", leadID).Return(nil)
	userRepo.On("FindByID", callerID).Return(&models.User{Name: "Budi Santoso"}, nil)

	svc := NewFollowUpService(nil, followUpRepo, leadRepo, userRepo)
	result, err := svc.createFollowUp(followUpRepo, leadRepo, callerID, "SALES", "LD-2607-0001", dto.CreateFollowUpInput{Note: "contacted"})

	require.NoError(t, err)
	assert.Equal(t, "Budi Santoso", result.CreatedByName, "created_by_name must be populated on the create response")
}
