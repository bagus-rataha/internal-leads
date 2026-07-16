//go:build integration

package repository

import (
	"fiber-api-boilerplate/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// seedScopeFixture creates two teams, one sales user per team, and one lead
// owned by each sales user. Returns everything a scoping test needs to
// assert against.
func seedScopeFixture(t *testing.T, db *gorm.DB) (teamA, teamB *models.SalesTeam, salesA, salesB *models.User, leadA, leadB *models.Lead) {
	teamA = &models.SalesTeam{Name: "Team A", IsActive: true}
	teamB = &models.SalesTeam{Name: "Team B", IsActive: true}
	require.NoError(t, db.Create(teamA).Error)
	require.NoError(t, db.Create(teamB).Error)

	salesA = &models.User{Email: "salesa@test.local", Password: "h", Name: "Sales A", Role: "SALES", TeamID: &teamA.ID, IsActive: true}
	salesB = &models.User{Email: "salesb@test.local", Password: "h", Name: "Sales B", Role: "SALES", TeamID: &teamB.ID, IsActive: true}
	require.NoError(t, db.Create(salesA).Error)
	require.NoError(t, db.Create(salesB).Error)

	leadA = &models.Lead{Code: "LD-2607-0001", OwnerID: salesA.ID, CreatedByID: salesA.ID, CompanyName: "Company A"}
	leadB = &models.Lead{Code: "LD-2607-0002", OwnerID: salesB.ID, CreatedByID: salesB.ID, CompanyName: "Company B"}
	require.NoError(t, db.Create(leadA).Error)
	require.NoError(t, db.Create(leadB).Error)

	return
}

func TestApplyLeadScope_OwnerScope_OnlyOwnLeads(t *testing.T) {
	db := setupTestDB(t)
	_, _, salesA, _, leadA, _ := seedScopeFixture(t, db)

	var results []models.Lead
	err := applyLeadScope(db.Model(&models.Lead{}), LeadScope{OwnerID: &salesA.ID}).Find(&results).Error

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, leadA.ID, results[0].ID)
}

func TestApplyLeadScope_TeamScope_OnlyTeamLeads(t *testing.T) {
	db := setupTestDB(t)
	teamA, _, _, _, leadA, _ := seedScopeFixture(t, db)

	var results []models.Lead
	err := applyLeadScope(db.Model(&models.Lead{}), LeadScope{TeamID: &teamA.ID}).Find(&results).Error

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, leadA.ID, results[0].ID)
}

func TestApplyLeadScope_EmptyScope_AllLeads(t *testing.T) {
	db := setupTestDB(t)
	seedScopeFixture(t, db)

	var results []models.Lead
	err := applyLeadScope(db.Model(&models.Lead{}), LeadScope{}).Find(&results).Error

	require.NoError(t, err)
	assert.Len(t, results, 2)
}
