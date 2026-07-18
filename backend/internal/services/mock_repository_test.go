package services

import (
	"fiber-api-boilerplate/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a manual testify mock satisfying both the
// userRepository and userRepositoryForUser interfaces.
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) List() ([]models.User, error) {
	args := m.Called()
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepository) ListWithFilter(role, teamID string) ([]models.User, error) {
	args := m.Called(role, teamID)
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepository) CountActiveByOwner(ownerID uuid.UUID) (int64, error) {
	args := m.Called(ownerID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) ReassignOwner(oldOwnerID, newOwnerID uuid.UUID) error {
	args := m.Called(oldOwnerID, newOwnerID)
	return args.Error(0)
}

func (m *MockUserRepository) CountActiveAdmins() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

// MockRefreshTokenRepository is a manual testify mock for refreshTokenRepository.
type MockRefreshTokenRepository struct {
	mock.Mock
}

func (m *MockRefreshTokenRepository) Create(rt *models.RefreshToken) error {
	args := m.Called(rt)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) FindByToken(token string) (*models.RefreshToken, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) DeleteByToken(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) DeleteAllByUserID(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

// MockSalesTeamRepository is a manual testify mock satisfying the
// salesTeamRepository interface.
type MockSalesTeamRepository struct {
	mock.Mock
}

func (m *MockSalesTeamRepository) Create(team *models.SalesTeam) error {
	args := m.Called(team)
	return args.Error(0)
}

func (m *MockSalesTeamRepository) FindByID(id uuid.UUID) (*models.SalesTeam, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SalesTeam), args.Error(1)
}

func (m *MockSalesTeamRepository) List() ([]models.SalesTeam, error) {
	args := m.Called()
	return args.Get(0).([]models.SalesTeam), args.Error(1)
}

func (m *MockSalesTeamRepository) Update(team *models.SalesTeam) error {
	args := m.Called(team)
	return args.Error(0)
}

func (m *MockSalesTeamRepository) CountActiveMembers(teamID uuid.UUID) (int64, error) {
	args := m.Called(teamID)
	return args.Get(0).(int64), args.Error(1)
}

// MockLeadSourceRepository is a manual testify mock satisfying the
// leadSourceRepository interface.
type MockLeadSourceRepository struct {
	mock.Mock
}

func (m *MockLeadSourceRepository) Create(source *models.LeadSource) error {
	args := m.Called(source)
	return args.Error(0)
}

func (m *MockLeadSourceRepository) FindByID(id uuid.UUID) (*models.LeadSource, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LeadSource), args.Error(1)
}

func (m *MockLeadSourceRepository) List() ([]models.LeadSource, error) {
	args := m.Called()
	return args.Get(0).([]models.LeadSource), args.Error(1)
}

func (m *MockLeadSourceRepository) Update(source *models.LeadSource) error {
	args := m.Called(source)
	return args.Error(0)
}

// MockServiceTypeRepository is a manual testify mock satisfying the
// serviceTypeRepository interface.
type MockServiceTypeRepository struct {
	mock.Mock
}

func (m *MockServiceTypeRepository) Create(serviceType *models.ServiceType) error {
	args := m.Called(serviceType)
	return args.Error(0)
}

func (m *MockServiceTypeRepository) FindByID(id uuid.UUID) (*models.ServiceType, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ServiceType), args.Error(1)
}

func (m *MockServiceTypeRepository) List() ([]models.ServiceType, error) {
	args := m.Called()
	return args.Get(0).([]models.ServiceType), args.Error(1)
}

func (m *MockServiceTypeRepository) Update(serviceType *models.ServiceType) error {
	args := m.Called(serviceType)
	return args.Error(0)
}
