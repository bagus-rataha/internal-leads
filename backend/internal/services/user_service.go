package services

import (
	"errors"
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/models"

	"github.com/google/uuid"
)

// userRepositoryForUser is the subset of repository methods UserService needs.
// Defined consumer-side for testability (satisfied by the real repo or a mock).
type userRepositoryForUser interface {
	FindByID(id uuid.UUID) (*models.User, error)
	Update(user *models.User) error
	List() ([]models.User, error)
}

type UserService struct {
	userRepo userRepositoryForUser
}

func NewUserService(userRepo userRepositoryForUser) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetProfile(userID uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	response := dto.ToUserResponse(user)
	return &response, nil
}

func (s *UserService) UpdateProfile(userID uuid.UUID, input dto.UpdateProfileInput) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	user.Name = input.Name

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	response := dto.ToUserResponse(user)
	return &response, nil
}

func (s *UserService) ListUsers() ([]dto.UserResponse, error) {
	users, err := s.userRepo.List()
	if err != nil {
		return nil, err
	}

	return dto.ToUserResponseList(users), nil
}
