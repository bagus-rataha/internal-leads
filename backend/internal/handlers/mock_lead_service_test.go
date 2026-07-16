package handlers

import (
	"fiber-api-boilerplate/internal/dto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockLeadService is a manual testify mock for the leadService interface.
type MockLeadService struct {
	mock.Mock
}

func (m *MockLeadService) Create(callerID uuid.UUID, role string, input dto.CreateLeadInput) (*dto.LeadResponse, error) {
	args := m.Called(callerID, role, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.LeadResponse), args.Error(1)
}

func (m *MockLeadService) List(callerID uuid.UUID, role string, query dto.LeadListQuery) (*dto.PaginatedLeadResponse, error) {
	args := m.Called(callerID, role, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PaginatedLeadResponse), args.Error(1)
}

func (m *MockLeadService) FindByCode(callerID uuid.UUID, role, code string) (*dto.LeadResponse, error) {
	args := m.Called(callerID, role, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.LeadResponse), args.Error(1)
}

func (m *MockLeadService) Update(callerID uuid.UUID, role, code string, input dto.UpdateLeadInput) (*dto.LeadResponse, error) {
	args := m.Called(callerID, role, code, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.LeadResponse), args.Error(1)
}

func (m *MockLeadService) UpdateStatus(callerID uuid.UUID, role, code string, input dto.UpdateLeadStatusInput) (*dto.LeadResponse, error) {
	args := m.Called(callerID, role, code, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.LeadResponse), args.Error(1)
}
