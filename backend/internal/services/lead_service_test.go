package services

import (
	"errors"
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/models"
	"fiber-api-boilerplate/internal/repository"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestIsLeadStale_ActivePastThreshold_True(t *testing.T) {
	now := time.Now()
	old := now.AddDate(0, 0, -10)

	byLastFollowUp := &models.Lead{Status: "FOLLOW_UP", LastFollowUpAt: &old}
	assert.True(t, isLeadStale(byLastFollowUp, now))

	byCreatedAt := &models.Lead{
		BaseModel: models.BaseModel{CreatedAt: old},
		Status:    "BARU",
	}
	assert.True(t, isLeadStale(byCreatedAt, now), "falls back to created_at when never followed up")
}

func TestIsLeadStale_Terminal_AlwaysFalse(t *testing.T) {
	now := time.Now()
	old := now.AddDate(0, 0, -10)

	handoff := &models.Lead{Status: "HANDOFF_ODOO", LastFollowUpAt: &old}
	assert.False(t, isLeadStale(handoff, now), "terminal status is never stale regardless of dates")

	lost := &models.Lead{Status: "LOST", LastFollowUpAt: &old}
	assert.False(t, isLeadStale(lost, now), "terminal status is never stale regardless of dates")
}

func TestIsLeadStale_WithinWindow_False(t *testing.T) {
	now := time.Now()
	recent := now.AddDate(0, 0, -1)

	lead := &models.Lead{Status: "FOLLOW_UP", LastFollowUpAt: &recent}
	assert.False(t, isLeadStale(lead, now))
}

func TestLeadCreate_Sales_OwnerForcedToSelf(t *testing.T) {
	leadRepo := new(MockLeadRepository)
	userRepo := new(MockUserRepository)
	salesID := uuid.Must(uuid.NewV7())
	otherID := uuid.Must(uuid.NewV7())

	userRepo.On("FindByID", salesID).Return(&models.User{Name: "Sales Person"}, nil)
	leadRepo.On("NextCode", mock.AnythingOfType("time.Time")).Return("LD-2607-0001", nil)
	leadRepo.On("Create", mock.AnythingOfType("*models.Lead")).Run(func(args mock.Arguments) {
		lead := args.Get(0).(*models.Lead)
		assert.Equal(t, salesID, lead.OwnerID, "owner_id forced to self for SALES even though otherID was submitted")
	}).Return(nil)

	svc := NewLeadService(nil, leadRepo, userRepo, nil)
	input := dto.CreateLeadInput{CompanyName: "Acme", OwnerID: &otherID}
	result, err := svc.createLead(leadRepo, userRepo, salesID, "SALES", input)

	assert.NoError(t, err)
	assert.Equal(t, salesID, result.OwnerID)
	// Exactly one FindByID call (the post-create owner-name fetch) - no
	// team-membership validation query happens for SALES.
	userRepo.AssertNumberOfCalls(t, "FindByID", 1)
}

func TestLeadCreate_ForecastMrrPassedThrough(t *testing.T) {
	leadRepo := new(MockLeadRepository)
	userRepo := new(MockUserRepository)
	salesID := uuid.Must(uuid.NewV7())
	forecast := 5000000.0

	userRepo.On("FindByID", salesID).Return(&models.User{Name: "Sales Person"}, nil)
	leadRepo.On("NextCode", mock.AnythingOfType("time.Time")).Return("LD-2607-0002", nil)
	leadRepo.On("Create", mock.AnythingOfType("*models.Lead")).Run(func(args mock.Arguments) {
		lead := args.Get(0).(*models.Lead)
		require.NotNil(t, lead.ForecastMrr)
		assert.Equal(t, forecast, *lead.ForecastMrr)
	}).Return(nil)

	svc := NewLeadService(nil, leadRepo, userRepo, nil)
	input := dto.CreateLeadInput{CompanyName: "Acme", ForecastMrr: &forecast}
	result, err := svc.createLead(leadRepo, userRepo, salesID, "SALES", input)

	assert.NoError(t, err)
	require.NotNil(t, result.ForecastMrr)
	assert.Equal(t, forecast, *result.ForecastMrr)
}

func TestApplyLeadUpdate_ForecastMrrSubmitted_Applied(t *testing.T) {
	lead := &models.Lead{}
	forecast := 12000000.0
	applyLeadUpdate(lead, dto.UpdateLeadInput{ForecastMrr: &forecast}, "SALES")
	require.NotNil(t, lead.ForecastMrr)
	assert.Equal(t, forecast, *lead.ForecastMrr)
}

func TestApplyLeadUpdate_ForecastMrrOmitted_Unchanged(t *testing.T) {
	existing := 3000000.0
	lead := &models.Lead{ForecastMrr: &existing}
	applyLeadUpdate(lead, dto.UpdateLeadInput{}, "SALES")
	require.NotNil(t, lead.ForecastMrr)
	assert.Equal(t, existing, *lead.ForecastMrr)
}

// TestLeadCreate_OwnerNamePopulatedInResponse is the regression test for the
// create-response gap: createLead used to build the DTO straight from the
// in-memory lead without ever loading Owner, so owner_name always came back
// "" even though a real owner had just been assigned.
func TestLeadCreate_OwnerNamePopulatedInResponse(t *testing.T) {
	leadRepo := new(MockLeadRepository)
	userRepo := new(MockUserRepository)
	salesID := uuid.Must(uuid.NewV7())

	userRepo.On("FindByID", salesID).Return(&models.User{Name: "Budi Santoso"}, nil)
	leadRepo.On("NextCode", mock.AnythingOfType("time.Time")).Return("LD-2607-0013", nil)
	leadRepo.On("Create", mock.AnythingOfType("*models.Lead")).Return(nil)

	svc := NewLeadService(nil, leadRepo, userRepo, nil)
	input := dto.CreateLeadInput{CompanyName: "Acme"}
	result, err := svc.createLead(leadRepo, userRepo, salesID, "SALES", input)

	assert.NoError(t, err)
	assert.Equal(t, "Budi Santoso", result.OwnerName, "owner_name must be populated on the create response, not left empty")
}

// TestLeadCreate_CityIDSet_CityNamePopulatedInResponse mirrors the owner_name
// regression test above: a freshly created lead is in-memory only, so
// city_name must come from an explicit post-Create fetch, not a Preload.
func TestLeadCreate_CityIDSet_CityNamePopulatedInResponse(t *testing.T) {
	leadRepo := new(MockLeadRepository)
	userRepo := new(MockUserRepository)
	cityRepo := new(MockCityRepository)
	salesID := uuid.Must(uuid.NewV7())

	userRepo.On("FindByID", salesID).Return(&models.User{Name: "Budi Santoso"}, nil)
	cityRepo.On("FindCityByID", 1).Return(&models.City{ID: 1, Name: "Jakarta Selatan"}, nil)
	leadRepo.On("NextCode", mock.AnythingOfType("time.Time")).Return("LD-2607-0014", nil)
	leadRepo.On("Create", mock.AnythingOfType("*models.Lead")).Return(nil)

	svc := NewLeadService(nil, leadRepo, userRepo, cityRepo)
	cityID := 1
	input := dto.CreateLeadInput{CompanyName: "Acme", CityID: &cityID}
	result, err := svc.createLead(leadRepo, userRepo, salesID, "SALES", input)

	assert.NoError(t, err)
	if assert.NotNil(t, result.CityName) {
		assert.Equal(t, "Jakarta Selatan", *result.CityName)
	}
}

// TestLeadCreate_NoCityID_CityNameNilWithoutFetch confirms the city lookup is
// skipped entirely (not just tolerated) when city_id wasn't submitted - a
// lead with no address filled in legitimately has no city.
func TestLeadCreate_NoCityID_CityNameNilWithoutFetch(t *testing.T) {
	leadRepo := new(MockLeadRepository)
	userRepo := new(MockUserRepository)
	cityRepo := new(MockCityRepository)
	salesID := uuid.Must(uuid.NewV7())

	userRepo.On("FindByID", salesID).Return(&models.User{Name: "Budi Santoso"}, nil)
	leadRepo.On("NextCode", mock.AnythingOfType("time.Time")).Return("LD-2607-0015", nil)
	leadRepo.On("Create", mock.AnythingOfType("*models.Lead")).Return(nil)

	svc := NewLeadService(nil, leadRepo, userRepo, cityRepo)
	input := dto.CreateLeadInput{CompanyName: "Acme"}
	result, err := svc.createLead(leadRepo, userRepo, salesID, "SALES", input)

	assert.NoError(t, err)
	assert.Nil(t, result.CityName)
	cityRepo.AssertNotCalled(t, "FindCityByID", mock.Anything)
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

	svc := NewLeadService(nil, leadRepo, userRepo, nil)
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

	svc := NewLeadService(nil, leadRepo, userRepo, nil)
	input := dto.CreateLeadInput{CompanyName: "Acme", OwnerID: &outsiderID}
	_, err := svc.createLead(leadRepo, userRepo, leaderID, "LEADER", input)

	assert.ErrorIs(t, err, ErrOwnerNotTeamMember)
	leadRepo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestLeadCreate_Admin_OwnerIDOmitted_DefaultsToSelf(t *testing.T) {
	leadRepo := new(MockLeadRepository)
	userRepo := new(MockUserRepository)
	adminID := uuid.Must(uuid.NewV7())

	userRepo.On("FindByID", adminID).Return(&models.User{Name: "Admin"}, nil)
	leadRepo.On("NextCode", mock.AnythingOfType("time.Time")).Return("LD-2607-0003", nil)
	leadRepo.On("Create", mock.AnythingOfType("*models.Lead")).Return(nil)

	svc := NewLeadService(nil, leadRepo, userRepo, nil)
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

	svc := NewLeadService(nil, leadRepo, userRepo, nil)
	_, err := svc.UpdateStatus(adminID, "ADMIN_SALES", "LD-2607-0004", dto.UpdateLeadStatusInput{Status: "FOLLOW_UP"})

	assert.ErrorIs(t, err, ErrInvalidStatusTransition)
	leadRepo.AssertNotCalled(t, "Update", mock.Anything)
}

func TestLeadUpdateStatus_FollowUpToSurvey_Allowed_SetsSurveyAt(t *testing.T) {
	leadRepo := new(MockLeadRepository)
	userRepo := new(MockUserRepository)
	adminID := uuid.Must(uuid.NewV7())
	lead := &models.Lead{Code: "LD-2607-0005", Status: "FOLLOW_UP"}
	leadRepo.On("FindByCode", repository.LeadScope{}, "LD-2607-0005").Return(lead, nil)
	leadRepo.On("Update", mock.AnythingOfType("*models.Lead")).Run(func(args mock.Arguments) {
		updated := args.Get(0).(*models.Lead)
		assert.NotNil(t, updated.SurveyAt, "survey_at set on first SURVEY entry")
	}).Return(nil)

	svc := NewLeadService(nil, leadRepo, userRepo, nil)
	result, err := svc.UpdateStatus(adminID, "ADMIN_SALES", "LD-2607-0005", dto.UpdateLeadStatusInput{Status: "SURVEY"})

	assert.NoError(t, err)
	assert.Equal(t, "SURVEY", result.Status)
}

// TestLeadUpdateStatus_AdminPullBackToFollowUp_ClearsSurveyAt: only an admin
// moving a lead from >=SURVEY back to FOLLOW_UP nulls survey_at.
func TestLeadUpdateStatus_AdminPullBackToFollowUp_ClearsSurveyAt(t *testing.T) {
	leadRepo := new(MockLeadRepository)
	userRepo := new(MockUserRepository)
	adminID := uuid.Must(uuid.NewV7())
	surveyed := time.Now().AddDate(0, 0, -3)
	lead := &models.Lead{Code: "LD-2607-0016", Status: "SURVEY", SurveyAt: &surveyed}
	leadRepo.On("FindByCode", repository.LeadScope{}, "LD-2607-0016").Return(lead, nil)
	leadRepo.On("Update", mock.AnythingOfType("*models.Lead")).Run(func(args mock.Arguments) {
		updated := args.Get(0).(*models.Lead)
		assert.Nil(t, updated.SurveyAt, "survey_at cleared on admin pull-back to FOLLOW_UP")
	}).Return(nil)

	svc := NewLeadService(nil, leadRepo, userRepo, nil)
	_, err := svc.UpdateStatus(adminID, "ADMIN_SALES", "LD-2607-0016", dto.UpdateLeadStatusInput{Status: "FOLLOW_UP"})

	assert.NoError(t, err)
}

// TestLeadUpdateStatus_LegacyHandoffAdvance_SetsSurveyAt: a legacy
// HANDOFF_ODOO lead advancing one step past SURVEY's slot must still stamp
// survey_at, or it drops out of every conversion metric (survey_at NULL and
// status no longer 'HANDOFF_ODOO' for the reachedSurveyExpr fallback).
func TestLeadUpdateStatus_LegacyHandoffAdvance_SetsSurveyAt(t *testing.T) {
	leadRepo := new(MockLeadRepository)
	userRepo := new(MockUserRepository)
	adminID := uuid.Must(uuid.NewV7())
	lead := &models.Lead{Code: "LD-2607-0017", Status: "HANDOFF_ODOO"}
	leadRepo.On("FindByCode", repository.LeadScope{}, "LD-2607-0017").Return(lead, nil)
	leadRepo.On("Update", mock.AnythingOfType("*models.Lead")).Run(func(args mock.Arguments) {
		updated := args.Get(0).(*models.Lead)
		assert.NotNil(t, updated.SurveyAt, "survey_at stamped when a legacy handoff lead advances")
	}).Return(nil)

	svc := NewLeadService(nil, leadRepo, userRepo, nil)
	_, err := svc.UpdateStatus(adminID, "ADMIN_SALES", "LD-2607-0017", dto.UpdateLeadStatusInput{Status: "SALES_CONFIRMATION"})

	assert.NoError(t, err)
	assert.NotNil(t, lead.SurveyAt)
}

// TestLeadUpdateStatus_ForwardPastSurvey_KeepsSurveyAt: a forward move whose
// target is beyond SURVEY must not overwrite an already-set survey_at.
func TestLeadUpdateStatus_ForwardPastSurvey_KeepsSurveyAt(t *testing.T) {
	leadRepo := new(MockLeadRepository)
	userRepo := new(MockUserRepository)
	adminID := uuid.Must(uuid.NewV7())
	surveyed := time.Now().AddDate(0, 0, -5)
	lead := &models.Lead{Code: "LD-2607-0018", Status: "SURVEY", SurveyAt: &surveyed}
	leadRepo.On("FindByCode", repository.LeadScope{}, "LD-2607-0018").Return(lead, nil)
	leadRepo.On("Update", mock.AnythingOfType("*models.Lead")).Run(func(args mock.Arguments) {
		updated := args.Get(0).(*models.Lead)
		assert.Equal(t, surveyed, *updated.SurveyAt, "survey_at not overwritten on a forward move past SURVEY")
	}).Return(nil)

	svc := NewLeadService(nil, leadRepo, userRepo, nil)
	_, err := svc.UpdateStatus(adminID, "ADMIN_SALES", "LD-2607-0018", dto.UpdateLeadStatusInput{Status: "SALES_CONFIRMATION"})

	assert.NoError(t, err)
	assert.Equal(t, surveyed, *lead.SurveyAt)
}

func TestLeadUpdateStatus_Advance_IgnoresStrayLostReason(t *testing.T) {
	leadRepo := new(MockLeadRepository)
	userRepo := new(MockUserRepository)
	adminID := uuid.Must(uuid.NewV7())
	lead := &models.Lead{Code: "LD-2607-0009", Status: "FOLLOW_UP"}
	leadRepo.On("FindByCode", repository.LeadScope{}, "LD-2607-0009").Return(lead, nil)
	leadRepo.On("Update", mock.AnythingOfType("*models.Lead")).Run(func(args mock.Arguments) {
		updated := args.Get(0).(*models.Lead)
		assert.Nil(t, updated.LostReason, "lost_reason must not be persisted on a non-LOST transition")
	}).Return(nil)

	svc := NewLeadService(nil, leadRepo, userRepo, nil)
	_, err := svc.UpdateStatus(adminID, "ADMIN_SALES", "LD-2607-0009", dto.UpdateLeadStatusInput{Status: "SURVEY", LostReason: strPtrSvc("stray")})

	assert.NoError(t, err)
	assert.Nil(t, lead.LostReason)
}

func TestLeadUpdateStatus_InvoiceToLost_Rejected(t *testing.T) {
	leadRepo := new(MockLeadRepository)
	userRepo := new(MockUserRepository)
	adminID := uuid.Must(uuid.NewV7())
	lead := &models.Lead{Code: "LD-2607-0006", Status: "INVOICE_BULANAN"}
	leadRepo.On("FindByCode", repository.LeadScope{}, "LD-2607-0006").Return(lead, nil)

	svc := NewLeadService(nil, leadRepo, userRepo, nil)
	_, err := svc.UpdateStatus(adminID, "ADMIN_SALES", "LD-2607-0006", dto.UpdateLeadStatusInput{Status: "LOST", LostReason: strPtrSvc("x")})

	assert.ErrorIs(t, err, ErrInvalidStatusTransition)
}

func TestLeadUpdateStatus_ToLost_RequiresReason(t *testing.T) {
	leadRepo := new(MockLeadRepository)
	userRepo := new(MockUserRepository)
	adminID := uuid.Must(uuid.NewV7())
	lead := &models.Lead{Code: "LD-2607-0007", Status: "BARU"}
	leadRepo.On("FindByCode", repository.LeadScope{}, "LD-2607-0007").Return(lead, nil)

	svc := NewLeadService(nil, leadRepo, userRepo, nil)
	_, err := svc.UpdateStatus(adminID, "ADMIN_SALES", "LD-2607-0007", dto.UpdateLeadStatusInput{Status: "LOST"})

	assert.Error(t, err)
	leadRepo.AssertNotCalled(t, "Update", mock.Anything)
}

func TestLeadFindByCode_OutOfScope_NotFound(t *testing.T) {
	leadRepo := new(MockLeadRepository)
	userRepo := new(MockUserRepository)
	salesID := uuid.Must(uuid.NewV7())
	leadRepo.On("FindDetailByCode", repository.LeadScope{OwnerID: &salesID}, "LD-2607-0008").
		Return(nil, gorm.ErrRecordNotFound)

	svc := NewLeadService(nil, leadRepo, userRepo, nil)
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

	svc := NewLeadService(nil, leadRepo, userRepo, nil)
	_, err := svc.List(leaderID, "LEADER", dto.LeadListQuery{Page: 1, Limit: 20})

	assert.NoError(t, err)
	leadRepo.AssertExpectations(t)
}

func TestLeadFindByCode_Leader_NilTeamID_FailsClosed(t *testing.T) {
	leadRepo := new(MockLeadRepository)
	userRepo := new(MockUserRepository)
	leaderID := uuid.Must(uuid.NewV7())
	userRepo.On("FindByID", leaderID).Return(&models.User{TeamID: nil}, nil)

	svc := NewLeadService(nil, leadRepo, userRepo, nil)
	_, err := svc.FindByCode(leaderID, "LEADER", "LD-2607-0011")

	assert.Error(t, err, "a LEADER with nil team_id must never get an unrestricted scope")
	leadRepo.AssertNotCalled(t, "FindDetailByCode", mock.Anything, mock.Anything)
}

func TestTranslateWriteError_FKViolation_MapsToErrInvalidReference(t *testing.T) {
	pgErr := &pgconn.PgError{Code: "23503", ConstraintName: "fk_leads_province"}

	got := translateWriteError(pgErr)

	assert.ErrorIs(t, got, ErrInvalidReference)
}

func TestTranslateWriteError_OtherError_PassesThroughUnchanged(t *testing.T) {
	other := errors.New("some other db error")

	got := translateWriteError(other)

	assert.Same(t, other, got)
}

func TestTranslateWriteError_Nil_ReturnsNil(t *testing.T) {
	assert.NoError(t, translateWriteError(nil))
}

func TestLeadUpdateStatus_FollowUpToLost_Allowed(t *testing.T) {
	leadRepo := new(MockLeadRepository)
	userRepo := new(MockUserRepository)
	adminID := uuid.Must(uuid.NewV7())
	lead := &models.Lead{Code: "LD-2607-0012", Status: "FOLLOW_UP"}
	leadRepo.On("FindByCode", repository.LeadScope{}, "LD-2607-0012").Return(lead, nil)
	leadRepo.On("Update", mock.AnythingOfType("*models.Lead")).Run(func(args mock.Arguments) {
		updated := args.Get(0).(*models.Lead)
		assert.Equal(t, "LOST", updated.Status)
		if assert.NotNil(t, updated.LostReason) {
			assert.Equal(t, "some valid reason", *updated.LostReason)
		}
	}).Return(nil)

	svc := NewLeadService(nil, leadRepo, userRepo, nil)
	result, err := svc.UpdateStatus(adminID, "ADMIN_SALES", "LD-2607-0012", dto.UpdateLeadStatusInput{Status: "LOST", LostReason: strPtrSvc("some valid reason")})

	assert.NoError(t, err)
	assert.Equal(t, "LOST", result.Status)
	if assert.NotNil(t, result.LostReason) {
		assert.Equal(t, "some valid reason", *result.LostReason)
	}
}

func TestLeadFindByCode_UnrecognizedRole_FailsClosed(t *testing.T) {
	leadRepo := new(MockLeadRepository)
	userRepo := new(MockUserRepository)
	callerID := uuid.Must(uuid.NewV7())

	svc := NewLeadService(nil, leadRepo, userRepo, nil)
	_, err := svc.FindByCode(callerID, "BOGUS", "LD-2607-0010")

	assert.Error(t, err)
	leadRepo.AssertNotCalled(t, "FindDetailByCode", mock.Anything, mock.Anything)
}

func TestLeadFindByCode_ReturnsDetailResponseWithResolvedNames(t *testing.T) {
	leadRepo := new(MockLeadRepository)
	userRepo := new(MockUserRepository)
	callerID := uuid.Must(uuid.NewV7())
	serviceTypeID := uuid.Must(uuid.NewV7())

	lead := &models.Lead{
		BaseModel: models.BaseModel{ID: uuid.Must(uuid.NewV7())},
		Code:      "LD-2607-0001", Status: "BARU", OwnerID: callerID, CreatedByID: callerID,
		ServiceTypeID: &serviceTypeID,
		Owner:         &models.User{Name: "Rani"},
		ServiceType:   &models.ServiceType{Name: "Dedicated"},
		CreatedBy:     &models.User{Name: "Rani"},
	}
	leadRepo.On("FindDetailByCode", mock.Anything, "LD-2607-0001").Return(lead, nil)

	svc := NewLeadService(nil, leadRepo, userRepo, nil)
	result, err := svc.FindByCode(callerID, "SALES", "LD-2607-0001")

	require.NoError(t, err)
	assert.Equal(t, "Dedicated", *result.ServiceTypeName)
	assert.Equal(t, "Rani", result.CreatedByName)
	assert.Nil(t, result.ProvinceName, "unset province_id must produce a nil name, not a panic or empty string")
}

func TestValidateStatusTransition(t *testing.T) {
	cases := []struct {
		name           string
		from, to, role string
		want           bool
	}{
		{"forward one step, sales", "SURVEY", "SALES_CONFIRMATION", "SALES", true},
		{"forward skip rejected", "SURVEY", "REGISTRASI", "SALES", false},
		{"forward skip rejected even for admin", "SURVEY", "REGISTRASI", "SU", false},
		{"backward rejected for sales", "INSTALASI", "SURVEY", "SALES", false},
		{"backward rejected for leader", "INSTALASI", "SURVEY", "LEADER", false},
		{"backward multi-step ok for admin", "INSTALASI", "SURVEY", "ADMIN_SALES", true},
		{"backward one step ok for su", "TRIAL", "INSTALASI", "SU", true},
		{"backward to follow_up ok for admin", "SURVEY", "FOLLOW_UP", "ADMIN_SALES", true},
		{"lost from follow_up", "FOLLOW_UP", "LOST", "SALES", true},
		{"lost from trial", "TRIAL", "LOST", "SALES", true},
		{"lost from invoice rejected", "INVOICE_BULANAN", "LOST", "SU", false},
		{"forward from invoice rejected", "INVOICE_BULANAN", "SURVEY", "SALES", false},
		{"invoice backward ok for admin", "INVOICE_BULANAN", "TRIAL", "ADMIN_SALES", true},
		{"baru to lost", "BARU", "LOST", "SALES", true},
		{"baru to survey rejected", "BARU", "SURVEY", "SU", false},
		{"baru to follow_up rejected via endpoint", "BARU", "FOLLOW_UP", "SU", false},
		{"legacy handoff forward", "HANDOFF_ODOO", "SALES_CONFIRMATION", "SALES", true},
		{"legacy handoff to lost", "HANDOFF_ODOO", "LOST", "SALES", true},
		{"from lost rejected", "LOST", "FOLLOW_UP", "SU", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, validateStatusTransition(c.from, c.to, c.role))
		})
	}
}

func strPtrSvc(s string) *string { return &s }
