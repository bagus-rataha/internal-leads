package repository

import (
	"fiber-api-boilerplate/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates new user repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates new user
func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// FindByEmail finds user by email
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID finds user by ID
func (r *UserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates user data
func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// List returns all users (with pagination in real case)
func (r *UserRepository) List() ([]models.User, error) {
	var users []models.User
	err := r.db.Find(&users).Error
	return users, err
}

// ListWithFilter returns users optionally narrowed by role and/or team_id.
// Empty string means "no filter" for that field. This is query-param
// narrowing for the admin user list, not the lead ownership scoping that
// belongs to the Lead module.
func (r *UserRepository) ListWithFilter(role, teamID string) ([]models.User, error) {
	var users []models.User
	q := r.db.Model(&models.User{})
	if role != "" {
		q = q.Where("role = ?", role)
	}
	if teamID != "" {
		q = q.Where("team_id = ?", teamID)
	}
	err := q.Find(&users).Error
	return users, err
}

// CountActiveByOwner counts leads owned by the given user that are not yet
// in a terminal status. "Active" here means status is not HANDOFF_ODOO/LOST
// - a narrower concept than the "terlantar" (stale) lead definition, which
// additionally requires 7 days since the last follow-up. Used by
// DeactivateUser to decide whether reassignment is required.
func (r *UserRepository) CountActiveByOwner(ownerID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Table("leads").
		Where("owner_id = ? AND status IN ('BARU','FOLLOW_UP')", ownerID).
		Count(&count).Error
	return count, err
}

// ReassignOwner moves every lead owned by oldOwnerID to newOwnerID. Used by
// DeactivateUser when the outgoing user still owns active leads.
func (r *UserRepository) ReassignOwner(oldOwnerID, newOwnerID uuid.UUID) error {
	return r.db.Table("leads").
		Where("owner_id = ?", oldOwnerID).
		Update("owner_id", newOwnerID).Error
}
