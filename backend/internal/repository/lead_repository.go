package repository

import (
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

// applyLeadScope is the single function every lead-touching query goes
// through - list, detail, and later count (dashboard). Uses a subquery
// rather than a JOIN so it composes safely with the team_id query-param
// narrowing filter, which also needs to restrict by team membership without
// colliding on a shared "users" join alias.
func applyLeadScope(tx *gorm.DB, scope LeadScope) *gorm.DB {
	if scope.OwnerID != nil {
		return tx.Where("leads.owner_id = ?", *scope.OwnerID)
	}
	if scope.TeamID != nil {
		return tx.Where("leads.owner_id IN (SELECT id FROM users WHERE team_id = ?)", *scope.TeamID)
	}
	return tx
}
