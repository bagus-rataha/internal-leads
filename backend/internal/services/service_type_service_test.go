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

func newServiceTypeService(typeRepo *MockServiceTypeRepository) *ServiceTypeService {
	return NewServiceTypeService(typeRepo)
}

func TestServiceTypeCreate_Success(t *testing.T) {
	typeRepo := new(MockServiceTypeRepository)
	typeRepo.On("Create", mock.AnythingOfType("*models.ServiceType")).Return(nil)

	svc := newServiceTypeService(typeRepo)
	result, err := svc.Create(dto.CreateServiceTypeInput{Name: "Dedicated"})

	assert.NoError(t, err)
	assert.Equal(t, "Dedicated", result.Name)
	assert.True(t, result.IsActive)
}

func TestServiceTypeList_Success(t *testing.T) {
	typeRepo := new(MockServiceTypeRepository)
	types := []models.ServiceType{
		{BaseModel: models.BaseModel{ID: uuid.Must(uuid.NewV7())}, Name: "Dedicated", IsActive: true},
		{BaseModel: models.BaseModel{ID: uuid.Must(uuid.NewV7())}, Name: "Broadband", IsActive: false},
	}
	typeRepo.On("List").Return(types, nil)

	svc := newServiceTypeService(typeRepo)
	result, err := svc.List()

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestServiceTypeUpdate_NotFound(t *testing.T) {
	typeRepo := new(MockServiceTypeRepository)
	id := uuid.Must(uuid.NewV7())
	typeRepo.On("FindByID", id).Return(nil, gorm.ErrRecordNotFound)

	svc := newServiceTypeService(typeRepo)
	_, err := svc.Update(id, dto.UpdateServiceTypeInput{})

	assert.Error(t, err)
}

func TestServiceTypeUpdate_NameAndActive(t *testing.T) {
	typeRepo := new(MockServiceTypeRepository)
	id := uuid.Must(uuid.NewV7())
	serviceType := &models.ServiceType{BaseModel: models.BaseModel{ID: id}, Name: "Old", IsActive: true}
	typeRepo.On("FindByID", id).Return(serviceType, nil)
	typeRepo.On("Update", serviceType).Return(nil)

	newName := "New"
	inactive := false
	svc := newServiceTypeService(typeRepo)
	result, err := svc.Update(id, dto.UpdateServiceTypeInput{Name: &newName, IsActive: &inactive})

	assert.NoError(t, err)
	assert.Equal(t, "New", result.Name)
	assert.False(t, result.IsActive)
}

func TestServiceTypeUpdate_NoActiveReferenceGuard(t *testing.T) {
	// Unlike TeamService.Update, deactivating a service type never checks
	// whether existing leads still reference it - confirmed project decision.
	typeRepo := new(MockServiceTypeRepository)
	id := uuid.Must(uuid.NewV7())
	serviceType := &models.ServiceType{BaseModel: models.BaseModel{ID: id}, Name: "Widely Used Type", IsActive: true}
	typeRepo.On("FindByID", id).Return(serviceType, nil)
	typeRepo.On("Update", serviceType).Return(nil)

	inactive := false
	svc := newServiceTypeService(typeRepo)
	result, err := svc.Update(id, dto.UpdateServiceTypeInput{IsActive: &inactive})

	assert.NoError(t, err)
	assert.False(t, result.IsActive)
}
