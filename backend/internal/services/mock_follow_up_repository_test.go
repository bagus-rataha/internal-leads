package services

import (
	"fiber-api-boilerplate/internal/models"
	"fiber-api-boilerplate/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockFollowUpRepository is a manual testify mock satisfying
// followUpRepositoryForFollowUp.
type MockFollowUpRepository struct {
	mock.Mock
}

func (m *MockFollowUpRepository) Create(f *models.FollowUp) error {
	args := m.Called(f)
	return args.Error(0)
}

func (m *MockFollowUpRepository) ListByLeadID(leadID uuid.UUID) ([]models.FollowUp, error) {
	args := m.Called(leadID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.FollowUp), args.Error(1)
}

// MockLeadRepositoryForFollowUp is a manual testify mock satisfying
// leadRepositoryForFollowUp - separate from MockLeadRepository
// (lead_service's mock) because the method sets differ.
type MockLeadRepositoryForFollowUp struct {
	mock.Mock
}

func (m *MockLeadRepositoryForFollowUp) FindByCode(scope repository.LeadScope, code string) (*models.Lead, error) {
	args := m.Called(scope, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Lead), args.Error(1)
}

func (m *MockLeadRepositoryForFollowUp) RecordFollowUp(leadID uuid.UUID) error {
	args := m.Called(leadID)
	return args.Error(0)
}

func (m *MockLeadRepositoryForFollowUp) MaybeTransitionToFollowUp(leadID uuid.UUID) error {
	args := m.Called(leadID)
	return args.Error(0)
}
