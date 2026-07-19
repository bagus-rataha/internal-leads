package handlers

import (
	"fiber-api-boilerplate/internal/dto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockDashboardService is a manual testify mock satisfying dashboardService.
type MockDashboardService struct {
	mock.Mock
}

func (m *MockDashboardService) Summary(callerID uuid.UUID, role string, q dto.DashboardQuery) (*dto.DashboardSummaryResponse, error) {
	args := m.Called(callerID, role, q)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.DashboardSummaryResponse), args.Error(1)
}
