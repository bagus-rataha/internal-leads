package repository

import (
	"fiber-api-boilerplate/internal/models"
	"strings"

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

// FindByEmail finds user by email, with Team preloaded so callers that
// build a UserResponse (e.g. login) get a populated team name instead of
// silently nil-ing it — see models.User.Team's comment.
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Team").Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID finds user by ID, with Team preloaded for display purposes
// (e.g. the profile endpoint's team name) — see models.User.Team's comment.
func (r *UserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Team").Where("id = ?", id).First(&user).Error
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
// Empty string means "no filter" for that field.
// The role parameter accepts comma-separated values (e.g. "SALES,LEADER"),
// which are split, trimmed, and filtered empty — if any remain, they're used
// as an IN clause. Empty strings or all-whitespace input results in no role filter.
// This is query-param narrowing for the admin user list, not the lead ownership
// scoping that belongs to the Lead module.
func (r *UserRepository) ListWithFilter(role, teamID string) ([]models.User, error) {
	var users []models.User
	q := r.db.Model(&models.User{})
	if role != "" {
		var roles []string
		for _, part := range strings.Split(role, ",") {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				roles = append(roles, trimmed)
			}
		}
		if len(roles) > 0 {
			q = q.Where("role IN ?", roles)
		}
	}
	if teamID != "" {
		q = q.Where("team_id = ?", teamID)
	}
	err := q.Find(&users).Error
	return users, err
}

// CountActiveByOwner counts leads owned by the given user that are still
// being worked by sales. "Active" means the lead is still being worked by
// sales - anything before INVOICE_BULANAN and not LOST (legacy HANDOFF_ODOO
// excluded, matching its pre-pipeline behavior). This is a narrower concept
// than the "terlantar" (stale) lead definition, which additionally requires
// 7 days since the last follow-up. Used by DeactivateUser to decide whether
// reassignment is required.
func (r *UserRepository) CountActiveByOwner(ownerID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Table("leads").
		Where("owner_id = ? AND status IN ('BARU','FOLLOW_UP','SURVEY','SALES_CONFIRMATION','REGISTRASI','INSTALASI','TRIAL')", ownerID).
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

// CountActiveAdmins counts active users who can manage other users
// (ADMIN_SALES or SU). Used to refuse deactivating the last one, which would
// lock the organization out of user administration.
func (r *UserRepository) CountActiveAdmins() (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).
		Where("role IN ? AND is_active = ?", []string{"ADMIN_SALES", "SU"}, true).
		Count(&count).Error
	return count, err
}
