package handlers

import (
	"fiber-api-boilerplate/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/xuri/excelize/v2"
)

// MockExportService is a manual testify mock for the exportService interface.
type MockExportService struct {
	mock.Mock
}

func (m *MockExportService) Export(callerID uuid.UUID, role, password string, filter repository.LeadFilter) (*excelize.File, error) {
	args := m.Called(callerID, role, password, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*excelize.File), args.Error(1)
}
