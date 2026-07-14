package services

import (
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/models"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func newUserService(userRepo *MockUserRepository) *UserService {
	return NewUserService(userRepo)
}

func TestGetProfile_Success(t *testing.T) {
	userRepo := new(MockUserRepository)
	userID := uuid.Must(uuid.NewV7())

	user := &models.User{
		BaseModel: models.BaseModel{ID: userID},
		Email:     "test@test.com",
		Name:      "Test",
		Role:      "user",
	}

	userRepo.On("FindByID", userID).Return(user, nil)

	svc := newUserService(userRepo)
	result, err := svc.GetProfile(userID)

	assert.NoError(t, err)
	assert.Equal(t, "test@test.com", result.Email)
	assert.Equal(t, "Test", result.Name)
}

func TestGetProfile_NotFound(t *testing.T) {
	userRepo := new(MockUserRepository)
	userID := uuid.Must(uuid.NewV7())

	userRepo.On("FindByID", userID).Return(nil, gorm.ErrRecordNotFound)

	svc := newUserService(userRepo)
	_, err := svc.GetProfile(userID)

	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
}

func TestUpdateProfile_Success(t *testing.T) {
	userRepo := new(MockUserRepository)
	userID := uuid.Must(uuid.NewV7())

	user := &models.User{
		BaseModel: models.BaseModel{ID: userID},
		Email:     "test@test.com",
		Name:      "Old Name",
		Role:      "user",
	}

	userRepo.On("FindByID", userID).Return(user, nil)
	userRepo.On("Update", user).Return(nil)

	svc := newUserService(userRepo)
	result, err := svc.UpdateProfile(userID, dto.UpdateProfileInput{Name: "New Name"})

	assert.NoError(t, err)
	assert.Equal(t, "New Name", result.Name)
}

func TestUpdateProfile_UserNotFound(t *testing.T) {
	userRepo := new(MockUserRepository)
	userID := uuid.Must(uuid.NewV7())

	userRepo.On("FindByID", userID).Return(nil, gorm.ErrRecordNotFound)

	svc := newUserService(userRepo)
	_, err := svc.UpdateProfile(userID, dto.UpdateProfileInput{Name: "New Name"})

	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
}

func TestListUsers_Success(t *testing.T) {
	userRepo := new(MockUserRepository)

	users := []models.User{
		{BaseModel: models.BaseModel{ID: uuid.Must(uuid.NewV7())}, Email: "a@test.com", Name: "A"},
		{BaseModel: models.BaseModel{ID: uuid.Must(uuid.NewV7())}, Email: "b@test.com", Name: "B"},
	}

	userRepo.On("List").Return(users, nil)

	svc := newUserService(userRepo)
	result, err := svc.ListUsers()

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}
