package handlers

import (
	"fiber-api-boilerplate/internal/dto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockFollowUpService is a manual testify mock for the followUpService interface.
type MockFollowUpService struct {
	mock.Mock
}

func (m *MockFollowUpService) Create(callerID uuid.UUID, role, leadCode string, input dto.CreateFollowUpInput) (*dto.FollowUpResponse, error) {
	args := m.Called(callerID, role, leadCode, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.FollowUpResponse), args.Error(1)
}

func (m *MockFollowUpService) List(callerID uuid.UUID, role, leadCode string) ([]dto.FollowUpResponse, error) {
	args := m.Called(callerID, role, leadCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.FollowUpResponse), args.Error(1)
}
