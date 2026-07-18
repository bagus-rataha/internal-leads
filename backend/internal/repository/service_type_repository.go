package repository

import (
	"fiber-api-boilerplate/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ServiceTypeRepository handles database operations for service type master data.
type ServiceTypeRepository struct {
	db *gorm.DB
}

// NewServiceTypeRepository creates new service type repository
func NewServiceTypeRepository(db *gorm.DB) *ServiceTypeRepository {
	return &ServiceTypeRepository{db: db}
}

// Create creates a new service type
func (r *ServiceTypeRepository) Create(serviceType *models.ServiceType) error {
	return r.db.Create(serviceType).Error
}

// FindByID finds a service type by ID
func (r *ServiceTypeRepository) FindByID(id uuid.UUID) (*models.ServiceType, error) {
	var serviceType models.ServiceType
	err := r.db.Where("id = ?", id).First(&serviceType).Error
	if err != nil {
		return nil, err
	}
	return &serviceType, nil
}

// List returns all service types, active and inactive
func (r *ServiceTypeRepository) List() ([]models.ServiceType, error) {
	var types []models.ServiceType
	err := r.db.Order("name").Find(&types).Error
	return types, err
}

// Update updates service type data
func (r *ServiceTypeRepository) Update(serviceType *models.ServiceType) error {
	return r.db.Save(serviceType).Error
}
