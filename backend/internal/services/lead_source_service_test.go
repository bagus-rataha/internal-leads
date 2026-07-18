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

func newLeadSourceService(sourceRepo *MockLeadSourceRepository) *LeadSourceService {
	return NewLeadSourceService(sourceRepo)
}

func TestLeadSourceCreate_Success(t *testing.T) {
	sourceRepo := new(MockLeadSourceRepository)
	sourceRepo.On("Create", mock.AnythingOfType("*models.LeadSource")).Return(nil)

	svc := newLeadSourceService(sourceRepo)
	result, err := svc.Create(dto.CreateLeadSourceInput{Name: "Website"})

	assert.NoError(t, err)
	assert.Equal(t, "Website", result.Name)
	assert.True(t, result.IsActive)
}

func TestLeadSourceList_Success(t *testing.T) {
	sourceRepo := new(MockLeadSourceRepository)
	sources := []models.LeadSource{
		{BaseModel: models.BaseModel{ID: uuid.Must(uuid.NewV7())}, Name: "A", IsActive: true},
		{BaseModel: models.BaseModel{ID: uuid.Must(uuid.NewV7())}, Name: "B", IsActive: false},
	}
	sourceRepo.On("List").Return(sources, nil)

	svc := newLeadSourceService(sourceRepo)
	result, err := svc.List()

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestLeadSourceUpdate_NotFound(t *testing.T) {
	sourceRepo := new(MockLeadSourceRepository)
	id := uuid.Must(uuid.NewV7())
	sourceRepo.On("FindByID", id).Return(nil, gorm.ErrRecordNotFound)

	svc := newLeadSourceService(sourceRepo)
	_, err := svc.Update(id, dto.UpdateLeadSourceInput{})

	assert.Error(t, err)
}

func TestLeadSourceUpdate_NameAndActive(t *testing.T) {
	sourceRepo := new(MockLeadSourceRepository)
	id := uuid.Must(uuid.NewV7())
	source := &models.LeadSource{BaseModel: models.BaseModel{ID: id}, Name: "Old", IsActive: true}
	sourceRepo.On("FindByID", id).Return(source, nil)
	sourceRepo.On("Update", source).Return(nil)

	newName := "New"
	inactive := false
	svc := newLeadSourceService(sourceRepo)
	result, err := svc.Update(id, dto.UpdateLeadSourceInput{Name: &newName, IsActive: &inactive})

	assert.NoError(t, err)
	assert.Equal(t, "New", result.Name)
	assert.False(t, result.IsActive)
}

func TestLeadSourceUpdate_NoActiveReferenceGuard(t *testing.T) {
	// Unlike TeamService.Update, deactivating a lead source never checks
	// whether existing leads still reference it - confirmed project decision.
	sourceRepo := new(MockLeadSourceRepository)
	id := uuid.Must(uuid.NewV7())
	source := &models.LeadSource{BaseModel: models.BaseModel{ID: id}, Name: "Widely Used", IsActive: true}
	sourceRepo.On("FindByID", id).Return(source, nil)
	sourceRepo.On("Update", source).Return(nil)

	inactive := false
	svc := newLeadSourceService(sourceRepo)
	result, err := svc.Update(id, dto.UpdateLeadSourceInput{IsActive: &inactive})

	assert.NoError(t, err)
	assert.False(t, result.IsActive)
}
