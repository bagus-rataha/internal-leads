package services

import (
	"fiber-api-boilerplate/internal/models"
	"fiber-api-boilerplate/internal/repository"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockLeadRepository is a manual testify mock satisfying leadRepositoryForLead.
type MockLeadRepository struct {
	mock.Mock
}

func (m *MockLeadRepository) NextCode(month time.Time) (string, error) {
	args := m.Called(month)
	return args.String(0), args.Error(1)
}

func (m *MockLeadRepository) Create(lead *models.Lead) error {
	args := m.Called(lead)
	return args.Error(0)
}

func (m *MockLeadRepository) List(scope repository.LeadScope, filter repository.LeadFilter) ([]models.Lead, int64, error) {
	args := m.Called(scope, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]models.Lead), args.Get(1).(int64), args.Error(2)
}

func (m *MockLeadRepository) FindByCode(scope repository.LeadScope, code string) (*models.Lead, error) {
	args := m.Called(scope, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Lead), args.Error(1)
}

func (m *MockLeadRepository) Update(lead *models.Lead) error {
	args := m.Called(lead)
	return args.Error(0)
}

// userRepositoryForLead is satisfied by the existing MockUserRepository
// (mock_repository_test.go) - no new mock needed for it.
