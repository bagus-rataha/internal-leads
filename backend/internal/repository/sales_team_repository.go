package repository

import (
	"fiber-api-boilerplate/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SalesTeamRepository handles database operations for sales teams
type SalesTeamRepository struct {
	db *gorm.DB
}

// NewSalesTeamRepository creates new sales team repository
func NewSalesTeamRepository(db *gorm.DB) *SalesTeamRepository {
	return &SalesTeamRepository{db: db}
}

// Create creates new sales team
func (r *SalesTeamRepository) Create(team *models.SalesTeam) error {
	return r.db.Create(team).Error
}

// FindByID finds sales team by ID
func (r *SalesTeamRepository) FindByID(id uuid.UUID) (*models.SalesTeam, error) {
	var team models.SalesTeam
	err := r.db.Where("id = ?", id).First(&team).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// Update updates sales team data
func (r *SalesTeamRepository) Update(team *models.SalesTeam) error {
	return r.db.Save(team).Error
}

// List returns all sales teams
func (r *SalesTeamRepository) List() ([]models.SalesTeam, error) {
	var teams []models.SalesTeam
	err := r.db.Find(&teams).Error
	return teams, err
}

// CountActiveMembers counts active users belonging to the given team.
func (r *SalesTeamRepository) CountActiveMembers(teamID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Table("users").
		Where("team_id = ? AND is_active = true", teamID).
		Count(&count).Error
	return count, err
}
