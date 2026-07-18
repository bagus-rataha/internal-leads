package repository

import (
	"fiber-api-boilerplate/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LeadSourceRepository handles database operations for lead source master data.
type LeadSourceRepository struct {
	db *gorm.DB
}

// NewLeadSourceRepository creates new lead source repository
func NewLeadSourceRepository(db *gorm.DB) *LeadSourceRepository {
	return &LeadSourceRepository{db: db}
}

// Create creates a new lead source
func (r *LeadSourceRepository) Create(source *models.LeadSource) error {
	return r.db.Create(source).Error
}

// FindByID finds a lead source by ID
func (r *LeadSourceRepository) FindByID(id uuid.UUID) (*models.LeadSource, error) {
	var source models.LeadSource
	err := r.db.Where("id = ?", id).First(&source).Error
	if err != nil {
		return nil, err
	}
	return &source, nil
}

// List returns all lead sources, active and inactive
func (r *LeadSourceRepository) List() ([]models.LeadSource, error) {
	var sources []models.LeadSource
	err := r.db.Order("name").Find(&sources).Error
	return sources, err
}

// Update updates lead source data
func (r *LeadSourceRepository) Update(source *models.LeadSource) error {
	return r.db.Save(source).Error
}
