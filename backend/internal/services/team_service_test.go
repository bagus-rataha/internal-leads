package services

import (
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/models"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func newTeamService(teamRepo *MockSalesTeamRepository) *TeamService {
	return NewTeamService(teamRepo)
}

func TestTeamCreate_Success(t *testing.T) {
	teamRepo := new(MockSalesTeamRepository)
	teamRepo.On("Create", mock.AnythingOfType("*models.SalesTeam")).Return(nil)

	svc := newTeamService(teamRepo)
	result, err := svc.Create(dto.CreateTeamInput{Name: "Team A"})

	assert.NoError(t, err)
	assert.Equal(t, "Team A", result.Name)
	assert.True(t, result.IsActive)
}

func TestTeamList_Success(t *testing.T) {
	teamRepo := new(MockSalesTeamRepository)
	teams := []models.SalesTeam{
		{BaseModel: models.BaseModel{ID: uuid.Must(uuid.NewV7())}, Name: "A", IsActive: true},
		{BaseModel: models.BaseModel{ID: uuid.Must(uuid.NewV7())}, Name: "B", IsActive: true},
	}
	teamRepo.On("List").Return(teams, nil)

	svc := newTeamService(teamRepo)
	result, err := svc.List()

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestTeamFindByID_Success(t *testing.T) {
	teamRepo := new(MockSalesTeamRepository)
	teamID := uuid.Must(uuid.NewV7())
	team := &models.SalesTeam{BaseModel: models.BaseModel{ID: teamID}, Name: "A", IsActive: true}
	teamRepo.On("FindByID", teamID).Return(team, nil)

	svc := newTeamService(teamRepo)
	result, err := svc.FindByID(teamID)

	assert.NoError(t, err)
	assert.Equal(t, "A", result.Name)
}

func TestTeamFindByID_NotFound(t *testing.T) {
	teamRepo := new(MockSalesTeamRepository)
	teamID := uuid.Must(uuid.NewV7())
	teamRepo.On("FindByID", teamID).Return(nil, gorm.ErrRecordNotFound)

	svc := newTeamService(teamRepo)
	_, err := svc.FindByID(teamID)

	assert.Error(t, err)
}

func TestTeamUpdate_NameOnly(t *testing.T) {
	teamRepo := new(MockSalesTeamRepository)
	teamID := uuid.Must(uuid.NewV7())
	team := &models.SalesTeam{BaseModel: models.BaseModel{ID: teamID}, Name: "Old", IsActive: true}
	teamRepo.On("FindByID", teamID).Return(team, nil)
	teamRepo.On("Update", team).Return(nil)

	newName := "New"
	svc := newTeamService(teamRepo)
	result, err := svc.Update(teamID, dto.UpdateTeamInput{Name: &newName})

	assert.NoError(t, err)
	assert.Equal(t, "New", result.Name)
	assert.True(t, result.IsActive) // untouched
	teamRepo.AssertNotCalled(t, "CountActiveMembers", teamID)
}

func TestTeamUpdate_DeactivateNoActiveMembers_Success(t *testing.T) {
	teamRepo := new(MockSalesTeamRepository)
	teamID := uuid.Must(uuid.NewV7())
	team := &models.SalesTeam{BaseModel: models.BaseModel{ID: teamID}, Name: "A", IsActive: true}
	teamRepo.On("FindByID", teamID).Return(team, nil)
	teamRepo.On("CountActiveMembers", teamID).Return(int64(0), nil)
	teamRepo.On("Update", team).Return(nil)

	inactive := false
	svc := newTeamService(teamRepo)
	result, err := svc.Update(teamID, dto.UpdateTeamInput{IsActive: &inactive})

	assert.NoError(t, err)
	assert.False(t, result.IsActive)
	teamRepo.AssertCalled(t, "Update", team)
}

func TestTeamUpdate_DeactivateWithActiveMembers_Error(t *testing.T) {
	teamRepo := new(MockSalesTeamRepository)
	teamID := uuid.Must(uuid.NewV7())
	team := &models.SalesTeam{BaseModel: models.BaseModel{ID: teamID}, Name: "A", IsActive: true}
	teamRepo.On("FindByID", teamID).Return(team, nil)
	teamRepo.On("CountActiveMembers", teamID).Return(int64(3), nil)

	inactive := false
	svc := newTeamService(teamRepo)
	_, err := svc.Update(teamID, dto.UpdateTeamInput{IsActive: &inactive})

	assert.ErrorIs(t, err, ErrTeamHasActiveMembers)
	teamRepo.AssertNotCalled(t, "Update", team)
}
