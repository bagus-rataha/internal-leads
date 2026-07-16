package services

import (
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/models"
	"fiber-api-boilerplate/internal/repository"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestLeadCreate_Sales_OwnerForcedToSelf(t *testing.T) {
	leadRepo := new(MockLeadRepository)
	userRepo := new(MockUserRepository)
	salesID := uuid.Must(uuid.NewV7())
	otherID := uuid.Must(uuid.NewV7())

	leadRepo.On("NextCode", mock.AnythingOfType("time.Time")).Return("LD-2607-0001", nil)
	leadRepo.On("Create", mock.AnythingOfType("*models.Lead")).Run(func(args mock.Arguments) {
		lead := args.Get(0).(*models.Lead)
		assert.Equal(t, salesID, lead.OwnerID, "owner_id forced to self for SALES even though otherID was submitted")
	}).Return(nil)

	svc := NewLeadService(nil, leadRepo, userRepo)
	input := dto.CreateLeadInput{CompanyName: "Acme", OwnerID: &otherID}
	result, err := svc.createLead(leadRepo, userRepo, salesID, "SALES", input)

	assert.NoError(t, err)
	assert.Equal(t, salesID, result.OwnerID)
	userRepo.AssertNotCalled(t, "FindByID", mock.Anything)
}

func TestLeadCreate_Leader_ValidatesOwnerIsTeamMember(t *testing.T) {
	leadRepo := new(MockLeadRepository)
	userRepo := new(MockUserRepository)
	leaderID := uuid.Must(uuid.NewV7())
	teamID := uuid.Must(uuid.NewV7())
	memberID := uuid.Must(uuid.NewV7())

	userRepo.On("FindByID", leaderID).Return(&models.User{TeamID: &teamID}, nil)
	userRepo.On("FindByID", memberID).Return(&models.User{TeamID: &teamID}, nil)
	leadRepo.On("NextCode", mock.AnythingOfType("time.Time")).Return("LD-2607-0002", nil)
	leadRepo.On("Create", mock.AnythingOfType("*models.Lead")).Return(nil)

	svc := NewLeadService(nil, leadRepo, userRepo)
	input := dto.CreateLeadInput{CompanyName: "Acme", OwnerID: &memberID}
	_, err := svc.createLead(leadRepo, userRepo, leaderID, "LEADER", input)

	assert.NoError(t, err)
}

func TestLeadCreate_Leader_RejectsOwnerOutsideTeam(t *testing.T) {
	leadRepo := new(MockLeadRepository)
	userRepo := new(MockUserRepository)
	leaderID := uuid.Must(uuid.NewV7())
	leaderTeamID := uuid.Must(uuid.NewV7())
	otherTeamID := uuid.Must(uuid.NewV7())
	outsiderID := uuid.Must(uuid.NewV7())

	userRepo.On("FindByID", leaderID).Return(&models.User{TeamID: &leaderTeamID}, nil)
	userRepo.On("FindByID", outsiderID).Return(&models.User{TeamID: &otherTeamID}, nil)

	svc := NewLeadService(nil, leadRepo, userRepo)
	input := dto.CreateLeadInput{CompanyName: "Acme", OwnerID: &outsiderID}
	_, err := svc.createLead(leadRepo, userRepo, leaderID, "LEADER", input)

	assert.ErrorIs(t, err, ErrOwnerNotTeamMember)
	leadRepo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestLeadCreate_Admin_OwnerIDOmitted_DefaultsToSelf(t *testing.T) {
	leadRepo := new(MockLeadRepository)
	userRepo := new(MockUserRepository)
	adminID := uuid.Must(uuid.NewV7())

	leadRepo.On("NextCode", mock.AnythingOfType("time.Time")).Return("LD-2607-0003", nil)
	leadRepo.On("Create", mock.AnythingOfType("*models.Lead")).Return(nil)

	svc := NewLeadService(nil, leadRepo, userRepo)
	input := dto.CreateLeadInput{CompanyName: "Acme"}
	result, err := svc.createLead(leadRepo, userRepo, adminID, "ADMIN_SALES", input)

	assert.NoError(t, err)
	assert.Equal(t, adminID, result.OwnerID)
}

func TestLeadUpdateStatus_BaruToFollowUp_Rejected(t *testing.T) {
	leadRepo := new(MockLeadRepository)
	userRepo := new(MockUserRepository)
	adminID := uuid.Must(uuid.NewV7())
	lead := &models.Lead{Code: "LD-2607-0004", Status: "BARU"}
	leadRepo.On("FindByCode", repository.LeadScope{}, "LD-2607-0004").Return(lead, nil)

	svc := NewLeadService(nil, leadRepo, userRepo)
	_, err := svc.UpdateStatus(adminID, "ADMIN_SALES", "LD-2607-0004", dto.UpdateLeadStatusInput{Status: "FOLLOW_UP"})

	assert.ErrorIs(t, err, ErrInvalidStatusTransition)
	leadRepo.AssertNotCalled(t, "Update", mock.Anything)
}

func TestLeadUpdateStatus_FollowUpToHandoff_Allowed(t *testing.T) {
	leadRepo := new(MockLeadRepository)
	userRepo := new(MockUserRepository)
	adminID := uuid.Must(uuid.NewV7())
	lead := &models.Lead{Code: "LD-2607-0005", Status: "FOLLOW_UP"}
	leadRepo.On("FindByCode", repository.LeadScope{}, "LD-2607-0005").Return(lead, nil)
	leadRepo.On("Update", mock.AnythingOfType("*models.Lead")).Return(nil)

	svc := NewLeadService(nil, leadRepo, userRepo)
	result, err := svc.UpdateStatus(adminID, "ADMIN_SALES", "LD-2607-0005", dto.UpdateLeadStatusInput{Status: "HANDOFF_ODOO"})

	assert.NoError(t, err)
	assert.Equal(t, "HANDOFF_ODOO", result.Status)
}

func TestLeadUpdateStatus_Terminal_Rejected(t *testing.T) {
	leadRepo := new(MockLeadRepository)
	userRepo := new(MockUserRepository)
	adminID := uuid.Must(uuid.NewV7())
	lead := &models.Lead{Code: "LD-2607-0006", Status: "HANDOFF_ODOO"}
	leadRepo.On("FindByCode", repository.LeadScope{}, "LD-2607-0006").Return(lead, nil)

	svc := NewLeadService(nil, leadRepo, userRepo)
	_, err := svc.UpdateStatus(adminID, "ADMIN_SALES", "LD-2607-0006", dto.UpdateLeadStatusInput{Status: "LOST", LostReason: strPtrSvc("x")})

	assert.ErrorIs(t, err, ErrInvalidStatusTransition)
}

func TestLeadUpdateStatus_ToLost_RequiresReason(t *testing.T) {
	leadRepo := new(MockLeadRepository)
	userRepo := new(MockUserRepository)
	adminID := uuid.Must(uuid.NewV7())
	lead := &models.Lead{Code: "LD-2607-0007", Status: "BARU"}
	leadRepo.On("FindByCode", repository.LeadScope{}, "LD-2607-0007").Return(lead, nil)

	svc := NewLeadService(nil, leadRepo, userRepo)
	_, err := svc.UpdateStatus(adminID, "ADMIN_SALES", "LD-2607-0007", dto.UpdateLeadStatusInput{Status: "LOST"})

	assert.Error(t, err)
	leadRepo.AssertNotCalled(t, "Update", mock.Anything)
}

func TestLeadFindByCode_OutOfScope_NotFound(t *testing.T) {
	leadRepo := new(MockLeadRepository)
	userRepo := new(MockUserRepository)
	salesID := uuid.Must(uuid.NewV7())
	leadRepo.On("FindByCode", repository.LeadScope{OwnerID: &salesID}, "LD-2607-0008").
		Return(nil, gorm.ErrRecordNotFound)

	svc := NewLeadService(nil, leadRepo, userRepo)
	_, err := svc.FindByCode(salesID, "SALES", "LD-2607-0008")

	assert.ErrorIs(t, err, ErrLeadNotFound)
}

func TestLeadList_Leader_ScopesByOwnTeam(t *testing.T) {
	leadRepo := new(MockLeadRepository)
	userRepo := new(MockUserRepository)
	leaderID := uuid.Must(uuid.NewV7())
	teamID := uuid.Must(uuid.NewV7())
	userRepo.On("FindByID", leaderID).Return(&models.User{TeamID: &teamID}, nil)
	leadRepo.On("List", repository.LeadScope{TeamID: &teamID}, mock.AnythingOfType("repository.LeadFilter")).
		Return([]models.Lead{}, int64(0), nil)

	svc := NewLeadService(nil, leadRepo, userRepo)
	_, err := svc.List(leaderID, "LEADER", dto.LeadListQuery{Page: 1, Limit: 20})

	assert.NoError(t, err)
	leadRepo.AssertExpectations(t)
}

func strPtrSvc(s string) *string { return &s }
