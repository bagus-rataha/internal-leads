package handlers

import (
	"fiber-api-boilerplate/internal/dto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockAuthService is a manual testify mock for the authService interface.
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(input dto.LoginInput) (*dto.TokenResponse, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TokenResponse), args.Error(1)
}

func (m *MockAuthService) RefreshToken(refreshToken string) (*dto.TokenResponse, error) {
	args := m.Called(refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TokenResponse), args.Error(1)
}

func (m *MockAuthService) Logout(refreshToken string) error {
	args := m.Called(refreshToken)
	return args.Error(0)
}

func (m *MockAuthService) LogoutAll(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

// MockUserService is a manual testify mock for the userService interface.
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetProfile(userID uuid.UUID) (*dto.UserResponse, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserResponse), args.Error(1)
}

func (m *MockUserService) UpdateProfile(userID uuid.UUID, input dto.UpdateProfileInput) (*dto.UserResponse, error) {
	args := m.Called(userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserResponse), args.Error(1)
}

func (m *MockUserService) ListUsers(callerID uuid.UUID, callerRole, role, teamID string) ([]dto.UserResponse, error) {
	args := m.Called(callerID, callerRole, role, teamID)
	return args.Get(0).([]dto.UserResponse), args.Error(1)
}

func (m *MockUserService) CreateUser(input dto.CreateUserInput) (*dto.UserResponse, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserResponse), args.Error(1)
}

func (m *MockUserService) GetUser(userID uuid.UUID) (*dto.UserResponse, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserResponse), args.Error(1)
}

func (m *MockUserService) UpdateUser(userID uuid.UUID, input dto.UpdateUserInput) (*dto.UserResponse, error) {
	args := m.Called(userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserResponse), args.Error(1)
}

func (m *MockUserService) ResetPassword(userID uuid.UUID, input dto.ResetPasswordInput) error {
	args := m.Called(userID, input)
	return args.Error(0)
}

func (m *MockUserService) DeactivateUser(callerID, userID uuid.UUID, input dto.DeactivateUserInput) (int64, error) {
	args := m.Called(callerID, userID, input)
	return args.Get(0).(int64), args.Error(1)
}

// MockTeamService is a manual testify mock for the teamService interface.
type MockTeamService struct {
	mock.Mock
}

func (m *MockTeamService) Create(input dto.CreateTeamInput) (*dto.TeamResponse, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TeamResponse), args.Error(1)
}

func (m *MockTeamService) List() ([]dto.TeamResponse, error) {
	args := m.Called()
	return args.Get(0).([]dto.TeamResponse), args.Error(1)
}

func (m *MockTeamService) FindByID(id uuid.UUID) (*dto.TeamResponse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TeamResponse), args.Error(1)
}

func (m *MockTeamService) Update(id uuid.UUID, input dto.UpdateTeamInput) (*dto.TeamResponse, error) {
	args := m.Called(id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TeamResponse), args.Error(1)
}

// MockReferenceService is a manual testify mock for the referenceService interface.
type MockReferenceService struct {
	mock.Mock
}

func (m *MockReferenceService) ListLeadSources() ([]dto.LeadSourceResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.LeadSourceResponse), args.Error(1)
}

func (m *MockReferenceService) ListServiceTypes() ([]dto.ServiceTypeResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.ServiceTypeResponse), args.Error(1)
}

func (m *MockReferenceService) ListProvinces() ([]dto.ProvinceResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.ProvinceResponse), args.Error(1)
}

func (m *MockReferenceService) ListCities(provinceID int) ([]dto.CityResponse, error) {
	args := m.Called(provinceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.CityResponse), args.Error(1)
}

func (m *MockReferenceService) ListDistricts(cityID int) ([]dto.DistrictResponse, error) {
	args := m.Called(cityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.DistrictResponse), args.Error(1)
}

func (m *MockReferenceService) ListVillages(districtID int, q string) ([]dto.VillageResponse, error) {
	args := m.Called(districtID, q)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.VillageResponse), args.Error(1)
}

// MockLeadSourceService is a manual testify mock for the leadSourceService interface.
type MockLeadSourceService struct {
	mock.Mock
}

func (m *MockLeadSourceService) Create(input dto.CreateLeadSourceInput) (*dto.LeadSourceAdminResponse, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.LeadSourceAdminResponse), args.Error(1)
}

func (m *MockLeadSourceService) List() ([]dto.LeadSourceAdminResponse, error) {
	args := m.Called()
	return args.Get(0).([]dto.LeadSourceAdminResponse), args.Error(1)
}

func (m *MockLeadSourceService) Update(id uuid.UUID, input dto.UpdateLeadSourceInput) (*dto.LeadSourceAdminResponse, error) {
	args := m.Called(id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.LeadSourceAdminResponse), args.Error(1)
}

// MockServiceTypeService is a manual testify mock for the serviceTypeService interface.
type MockServiceTypeService struct {
	mock.Mock
}

func (m *MockServiceTypeService) Create(input dto.CreateServiceTypeInput) (*dto.ServiceTypeAdminResponse, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ServiceTypeAdminResponse), args.Error(1)
}

func (m *MockServiceTypeService) List() ([]dto.ServiceTypeAdminResponse, error) {
	args := m.Called()
	return args.Get(0).([]dto.ServiceTypeAdminResponse), args.Error(1)
}

func (m *MockServiceTypeService) Update(id uuid.UUID, input dto.UpdateServiceTypeInput) (*dto.ServiceTypeAdminResponse, error) {
	args := m.Called(id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ServiceTypeAdminResponse), args.Error(1)
}
