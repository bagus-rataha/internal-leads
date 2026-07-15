package services

import (
	"errors"
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/models"

	"github.com/google/uuid"
)

// ErrTeamHasActiveMembers is returned when a caller tries to deactivate a
// team that still has active members. The handler maps this to 422.
var ErrTeamHasActiveMembers = errors.New("team still has active members")

// salesTeamRepository is the subset of repository methods TeamService needs.
// Defined consumer-side for testability (satisfied by the real repo or a mock).
type salesTeamRepository interface {
	Create(team *models.SalesTeam) error
	FindByID(id uuid.UUID) (*models.SalesTeam, error)
	List() ([]models.SalesTeam, error)
	Update(team *models.SalesTeam) error
	CountActiveMembers(teamID uuid.UUID) (int64, error)
}

type TeamService struct {
	teamRepo salesTeamRepository
}

func NewTeamService(teamRepo salesTeamRepository) *TeamService {
	return &TeamService{teamRepo: teamRepo}
}

func (s *TeamService) Create(input dto.CreateTeamInput) (*dto.TeamResponse, error) {
	team := &models.SalesTeam{
		Name:     input.Name,
		IsActive: true,
	}

	if err := s.teamRepo.Create(team); err != nil {
		return nil, err
	}

	response := dto.ToTeamResponse(team)
	return &response, nil
}

func (s *TeamService) List() ([]dto.TeamResponse, error) {
	teams, err := s.teamRepo.List()
	if err != nil {
		return nil, err
	}

	return dto.ToTeamResponseList(teams), nil
}

func (s *TeamService) FindByID(id uuid.UUID) (*dto.TeamResponse, error) {
	team, err := s.teamRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	response := dto.ToTeamResponse(team)
	return &response, nil
}

func (s *TeamService) Update(id uuid.UUID, input dto.UpdateTeamInput) (*dto.TeamResponse, error) {
	team, err := s.teamRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if input.IsActive != nil && !*input.IsActive {
		count, err := s.teamRepo.CountActiveMembers(id)
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, ErrTeamHasActiveMembers
		}
	}

	if input.Name != nil {
		team.Name = *input.Name
	}
	if input.IsActive != nil {
		team.IsActive = *input.IsActive
	}

	if err := s.teamRepo.Update(team); err != nil {
		return nil, err
	}

	response := dto.ToTeamResponse(team)
	return &response, nil
}
