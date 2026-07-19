package repository

import (
	"fmt"
	"time"

	"fiber-api-boilerplate/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LeadScope encodes the caller's role-based access to leads. Built by the
// service from the caller's identity, never by a handler or repository.
type LeadScope struct {
	OwnerID *uuid.UUID // set -> SALES: only their own leads
	TeamID  *uuid.UUID // set -> LEADER: leads owned by their team's members
	// both nil -> ADMIN_SALES/SU: unrestricted
}

// ApplyLeadScope is the single function every lead-touching query goes
// through - list, detail, dashboard, export. Uses a subquery rather than a
// JOIN so it composes safely with the team_id query-param narrowing filter,
// which also needs to restrict by team membership without colliding on a
// shared "users" join alias. Exported so the dashboard module (which has no
// repository per ARCHITECTURE.md's explicit exception) can reuse it instead
// of duplicating the scoping rule.
func ApplyLeadScope(tx *gorm.DB, scope LeadScope) *gorm.DB {
	if scope.OwnerID != nil {
		return tx.Where("leads.owner_id = ?", *scope.OwnerID)
	}
	if scope.TeamID != nil {
		return tx.Where("leads.owner_id IN (SELECT id FROM users WHERE team_id = ?)", *scope.TeamID)
	}
	return tx
}

// StaleLeadThresholdDays is the one definition of "terlantar" (stale) per
// ARCHITECTURE.md §10 - exported so it's reused verbatim wherever this
// concept appears (dashboard, in a later PR, must import and reuse this same
// constant).
const StaleLeadThresholdDays = 7

// LeadFilter holds every GET /leads query param except pagination cursor
// state, which List takes via Page/Limit directly.
type LeadFilter struct {
	Q            string
	Status       string
	SourceID     *uuid.UUID
	TeamID       *uuid.UUID
	OwnerID      *uuid.UUID
	ProvinceID   *int
	CityID       *int
	DateFrom     *time.Time
	DateTo       *time.Time
	FollowUpFrom *time.Time
	FollowUpTo   *time.Time
	Stale        bool
	Sort         string // "", "code", "-code", "company_name", "-company_name" - pre-validated by the caller
	Page         int    // >=1, pre-clamped by the caller
	Limit        int    // 1..100, pre-clamped by the caller
}

// LeadRepository handles database operations for leads
type LeadRepository struct {
	db *gorm.DB
}

// NewLeadRepository creates new lead repository
func NewLeadRepository(db *gorm.DB) *LeadRepository {
	return &LeadRepository{db: db}
}

// NextCode generates the next LD-YYMM-NNNN code from lead_code_seq. Callers
// that need code generation atomic with the INSERT (LeadService.Create)
// construct this repository with a transaction-scoped *gorm.DB, exactly
// like UserService does for its multi-step writes.
func (r *LeadRepository) NextCode(month time.Time) (string, error) {
	var seq int64
	if err := r.db.Raw("SELECT nextval('lead_code_seq')").Scan(&seq).Error; err != nil {
		return "", err
	}
	return fmt.Sprintf("LD-%s-%04d", month.Format("0601"), seq), nil
}

// Create creates a new lead
func (r *LeadRepository) Create(lead *models.Lead) error {
	return r.db.Create(lead).Error
}

// FindByCode finds a lead by its code, scoped. A lead outside scope returns
// gorm.ErrRecordNotFound, identical to a lead that doesn't exist at all -
// existence never leaks to a caller who can't see it.
func (r *LeadRepository) FindByCode(scope LeadScope, code string) (*models.Lead, error) {
	var lead models.Lead
	err := ApplyLeadScope(r.db.Model(&models.Lead{}), scope).
		Preload("Owner.Team").
		Preload("City").
		Where("leads.code = ?", code).First(&lead).Error
	if err != nil {
		return nil, err
	}
	return &lead, nil
}

// FindDetailByCode is FindByCode plus every reference-name association
// GET /leads/:code's response needs and List/write-path FindByCode don't -
// kept as a separate method (not a shared helper) so List's and the write
// paths' Preload chains stay exactly as lean as they were.
func (r *LeadRepository) FindDetailByCode(scope LeadScope, code string) (*models.Lead, error) {
	var lead models.Lead
	err := ApplyLeadScope(r.db.Model(&models.Lead{}), scope).
		Preload("Owner.Team").
		Preload("City").
		Preload("Province").
		Preload("District").
		Preload("Village").
		Preload("Zip").
		Preload("ServiceType").
		Preload("LeadSource").
		Preload("CreatedBy").
		Where("leads.code = ?", code).First(&lead).Error
	if err != nil {
		return nil, err
	}
	return &lead, nil
}

// Update saves an existing lead's fields
func (r *LeadRepository) Update(lead *models.Lead) error {
	return r.db.Save(lead).Error
}

// List returns leads matching scope and filter, plus the total count before
// pagination, for GET /leads.
func (r *LeadRepository) List(scope LeadScope, filter LeadFilter) ([]models.Lead, int64, error) {
	tx := ApplyLeadScope(r.db.Model(&models.Lead{}), scope)
	tx = applyLeadFilter(tx, filter)

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var leads []models.Lead
	offset := (filter.Page - 1) * filter.Limit
	err := applyLeadSort(tx, filter.Sort).Preload("Owner.Team").Preload("City").Offset(offset).Limit(filter.Limit).Find(&leads).Error
	return leads, total, err
}

// applyLeadFilter narrows a scoped query by every GET /leads filter param.
func applyLeadFilter(tx *gorm.DB, filter LeadFilter) *gorm.DB {
	if filter.Q != "" {
		like := "%" + filter.Q + "%"
		tx = tx.Where("leads.code ILIKE ? OR leads.company_name ILIKE ? OR leads.pic_name ILIKE ?", like, like, like)
	}
	if filter.Status != "" {
		tx = tx.Where("leads.status = ?", filter.Status)
	}
	if filter.SourceID != nil {
		tx = tx.Where("leads.lead_source_id = ?", *filter.SourceID)
	}
	if filter.TeamID != nil {
		tx = tx.Where("leads.owner_id IN (SELECT id FROM users WHERE team_id = ?)", *filter.TeamID)
	}
	if filter.OwnerID != nil {
		tx = tx.Where("leads.owner_id = ?", *filter.OwnerID)
	}
	if filter.ProvinceID != nil {
		tx = tx.Where("leads.province_id = ?", *filter.ProvinceID)
	}
	if filter.CityID != nil {
		tx = tx.Where("leads.city_id = ?", *filter.CityID)
	}
	if filter.DateFrom != nil {
		tx = tx.Where("leads.created_at >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		// DateTo is parsed as midnight on the given day; using it directly
		// with <= would exclude the entire day it names. Advance to the
		// next day's midnight and use < so the whole day is included.
		tx = tx.Where("leads.created_at < ?", filter.DateTo.AddDate(0, 0, 1))
	}
	if filter.FollowUpFrom != nil {
		tx = tx.Where("leads.last_follow_up_at >= ?", *filter.FollowUpFrom)
	}
	if filter.FollowUpTo != nil {
		tx = tx.Where("leads.last_follow_up_at < ?", filter.FollowUpTo.AddDate(0, 0, 1))
	}
	if filter.Stale {
		tx = tx.Where("leads.status IN ('BARU','FOLLOW_UP') AND COALESCE(leads.last_follow_up_at, leads.created_at) < ?",
			time.Now().AddDate(0, 0, -StaleLeadThresholdDays))
	}
	return tx
}

// applyLeadSort applies the whitelisted sort, defaulting to
// last_follow_up_at ASC NULLS FIRST - a lead never followed up is the most
// urgent, matching the stale-lead definition's created_at fallback.
func applyLeadSort(tx *gorm.DB, sort string) *gorm.DB {
	switch sort {
	case "code":
		return tx.Order("leads.code ASC")
	case "-code":
		return tx.Order("leads.code DESC")
	case "company_name":
		return tx.Order("leads.company_name ASC")
	case "-company_name":
		return tx.Order("leads.company_name DESC")
	default:
		return tx.Order("leads.last_follow_up_at ASC NULLS FIRST")
	}
}

// RecordFollowUp updates last_follow_up_at and increments follow_up_count in
// one statement - steps 2+3 of the follow-up write transaction. The
// increment is a SQL expression, not a Go read-modify-write, so concurrent
// follow-ups on the same lead can't lose an increment.
func (r *LeadRepository) RecordFollowUp(leadID uuid.UUID) error {
	return r.db.Model(&models.Lead{}).Where("id = ?", leadID).
		Updates(map[string]interface{}{
			"last_follow_up_at": time.Now(),
			"follow_up_count":   gorm.Expr("follow_up_count + 1"),
		}).Error
}

// MaybeTransitionToFollowUp flips status to FOLLOW_UP only if it's currently
// BARU - step 4 of the follow-up write transaction. The WHERE condition
// makes this a no-op for HANDOFF_ODOO/LOST leads, preventing them from
// being pulled backward by a follow-up.
func (r *LeadRepository) MaybeTransitionToFollowUp(leadID uuid.UUID) error {
	return r.db.Model(&models.Lead{}).
		Where("id = ? AND status = ?", leadID, "BARU").
		Update("status", "FOLLOW_UP").Error
}
