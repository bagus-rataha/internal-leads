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

func (m *MockUserService) ListUsers(role, teamID string) ([]dto.UserResponse, error) {
	args := m.Called(role, teamID)
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

func (m *MockUserService) DeactivateUser(userID uuid.UUID, input dto.DeactivateUserInput) (int64, error) {
	args := m.Called(userID, input)
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
