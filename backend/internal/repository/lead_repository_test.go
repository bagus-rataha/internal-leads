//go:build integration

package repository

import (
	"fiber-api-boilerplate/internal/models"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
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

// seedTestUser creates a minimal active user and returns its ID, for tests
// that need a valid owner_id/created_by_id (both columns are FK-constrained
// to users) but don't care about role or team. A fresh random-UUID email
// keeps every call unique within a test's truncated-table state.
func seedTestUser(t *testing.T, db *gorm.DB) uuid.UUID {
	user := &models.User{
		Email:    fmt.Sprintf("user-%s@test.local", uuid.Must(uuid.NewV7())),
		Password: "h",
		Name:     "Test User",
		Role:     "ADMIN_SALES",
		IsActive: true,
	}
	require.NoError(t, db.Create(user).Error)
	return user.ID
}

func TestLeadRepository_NextCode_FormatAndSequence(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLeadRepository(db)

	month := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	code1, err := repo.NextCode(month)
	require.NoError(t, err)
	assert.Regexp(t, `^LD-2607-\d{4}$`, code1)

	code2, err := repo.NextCode(month)
	require.NoError(t, err)
	assert.NotEqual(t, code1, code2, "sequence must advance on every call")
}

func TestLeadRepository_Create_FindByCode(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLeadRepository(db)

	userID := seedTestUser(t, db)
	lead := &models.Lead{Code: "LD-2607-0100", OwnerID: userID, CreatedByID: userID, CompanyName: "Acme"}
	require.NoError(t, repo.Create(lead))
	assert.Equal(t, "BARU", lead.Status, "DB default applies")

	found, err := repo.FindByCode(LeadScope{}, "LD-2607-0100")
	require.NoError(t, err)
	assert.Equal(t, "Acme", found.CompanyName)
}

func TestLeadRepository_FindByCode_OutOfScope_NotFound(t *testing.T) {
	db := setupTestDB(t)
	_, _, _, salesB, leadA, _ := seedScopeFixture(t, db)
	repo := NewLeadRepository(db)

	_, err := repo.FindByCode(LeadScope{OwnerID: &salesB.ID}, leadA.Code)
	assert.Error(t, err)
}

func TestLeadRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLeadRepository(db)
	userID := seedTestUser(t, db)
	lead := &models.Lead{Code: "LD-2607-0200", OwnerID: userID, CreatedByID: userID, CompanyName: "Old Name"}
	require.NoError(t, repo.Create(lead))

	lead.CompanyName = "New Name"
	require.NoError(t, repo.Update(lead))

	found, _ := repo.FindByCode(LeadScope{}, lead.Code)
	assert.Equal(t, "New Name", found.CompanyName)
}

func TestLeadRepository_List_FilterByStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLeadRepository(db)
	userID := seedTestUser(t, db)
	require.NoError(t, db.Create(&models.Lead{Code: "LD-2607-0301", OwnerID: userID, CreatedByID: userID, CompanyName: "A", Status: "BARU"}).Error)
	require.NoError(t, db.Create(&models.Lead{Code: "LD-2607-0302", OwnerID: userID, CreatedByID: userID, CompanyName: "B", Status: "LOST", LostReason: strPtr("no budget")}).Error)

	results, total, err := repo.List(LeadScope{}, LeadFilter{Status: "LOST", Page: 1, Limit: 20})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, results, 1)
	assert.Equal(t, "B", results[0].CompanyName)
}

func TestLeadRepository_List_QSearchesCodeCompanyPic(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLeadRepository(db)
	userID := seedTestUser(t, db)
	require.NoError(t, db.Create(&models.Lead{Code: "LD-2607-0401", OwnerID: userID, CreatedByID: userID, CompanyName: "Acme Corp"}).Error)
	require.NoError(t, db.Create(&models.Lead{Code: "LD-2607-0402", OwnerID: userID, CreatedByID: userID, CompanyName: "Other", PicName: strPtr("Budi Santoso")}).Error)
	require.NoError(t, db.Create(&models.Lead{Code: "LD-2607-0403", OwnerID: userID, CreatedByID: userID, CompanyName: "Unrelated"}).Error)

	byCompany, _, err := repo.List(LeadScope{}, LeadFilter{Q: "acme", Page: 1, Limit: 20})
	require.NoError(t, err)
	assert.Len(t, byCompany, 1)

	byCode, _, err := repo.List(LeadScope{}, LeadFilter{Q: "0402", Page: 1, Limit: 20})
	require.NoError(t, err)
	assert.Len(t, byCode, 1)

	byPic, _, err := repo.List(LeadScope{}, LeadFilter{Q: "budi", Page: 1, Limit: 20})
	require.NoError(t, err)
	assert.Len(t, byPic, 1)
}

func TestLeadRepository_List_DefaultSort_NullsFirst(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLeadRepository(db)
	userID := seedTestUser(t, db)
	followedUp := time.Now().Add(-time.Hour)
	require.NoError(t, db.Create(&models.Lead{Code: "LD-2607-0501", OwnerID: userID, CreatedByID: userID, CompanyName: "Followed", LastFollowUpAt: &followedUp}).Error)
	require.NoError(t, db.Create(&models.Lead{Code: "LD-2607-0502", OwnerID: userID, CreatedByID: userID, CompanyName: "Never followed"}).Error)

	results, _, err := repo.List(LeadScope{}, LeadFilter{Page: 1, Limit: 20})
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "Never followed", results[0].CompanyName, "null last_follow_up_at sorts first")
}

func TestLeadRepository_List_SortByCompanyName(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLeadRepository(db)
	userID := seedTestUser(t, db)
	require.NoError(t, db.Create(&models.Lead{Code: "LD-2607-0601", OwnerID: userID, CreatedByID: userID, CompanyName: "Zebra"}).Error)
	require.NoError(t, db.Create(&models.Lead{Code: "LD-2607-0602", OwnerID: userID, CreatedByID: userID, CompanyName: "Apple"}).Error)

	results, _, err := repo.List(LeadScope{}, LeadFilter{Sort: "company_name", Page: 1, Limit: 20})
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "Apple", results[0].CompanyName)
}

func TestLeadRepository_List_Pagination(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLeadRepository(db)
	userID := seedTestUser(t, db)
	for i := 0; i < 5; i++ {
		require.NoError(t, db.Create(&models.Lead{Code: fmt.Sprintf("LD-2607-07%02d", i), OwnerID: userID, CreatedByID: userID, CompanyName: fmt.Sprintf("Company %d", i)}).Error)
	}

	page1, total, err := repo.List(LeadScope{}, LeadFilter{Sort: "company_name", Page: 1, Limit: 2})
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, page1, 2)

	page3, _, err := repo.List(LeadScope{}, LeadFilter{Sort: "company_name", Page: 3, Limit: 2})
	require.NoError(t, err)
	assert.Len(t, page3, 1, "5 rows, limit 2 -> last page has 1")
}

func TestLeadRepository_List_StaleFilter(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLeadRepository(db)
	userID := seedTestUser(t, db)
	oldFollowUp := time.Now().AddDate(0, 0, -10)
	recentFollowUp := time.Now().AddDate(0, 0, -1)
	require.NoError(t, db.Create(&models.Lead{Code: "LD-2607-0801", OwnerID: userID, CreatedByID: userID, CompanyName: "Stale", Status: "FOLLOW_UP", LastFollowUpAt: &oldFollowUp}).Error)
	require.NoError(t, db.Create(&models.Lead{Code: "LD-2607-0802", OwnerID: userID, CreatedByID: userID, CompanyName: "Fresh", Status: "FOLLOW_UP", LastFollowUpAt: &recentFollowUp}).Error)
	require.NoError(t, db.Create(&models.Lead{Code: "LD-2607-0803", OwnerID: userID, CreatedByID: userID, CompanyName: "Won, old but terminal", Status: "HANDOFF_ODOO", LastFollowUpAt: &oldFollowUp}).Error)

	results, _, err := repo.List(LeadScope{}, LeadFilter{Stale: true, Page: 1, Limit: 20})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Stale", results[0].CompanyName)
}

func TestLeadRepository_List_DateTo_IncludesWholeDay(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLeadRepository(db)
	userID := seedTestUser(t, db)

	day := time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC)
	midDay := time.Date(2026, 7, 16, 14, 0, 0, 0, time.UTC)
	lead := &models.Lead{Code: "LD-2607-1101", OwnerID: userID, CreatedByID: userID, CompanyName: "Mid-day"}
	require.NoError(t, db.Create(lead).Error)
	// created_at has a DB default (now()); overwrite it directly to a known
	// mid-day timestamp on the filtered day.
	require.NoError(t, db.Model(lead).Update("created_at", midDay).Error)

	results, _, err := repo.List(LeadScope{}, LeadFilter{DateTo: &day, Page: 1, Limit: 20})
	require.NoError(t, err)
	require.Len(t, results, 1, "a lead created during the date_to day itself must be included")
	assert.Equal(t, "Mid-day", results[0].CompanyName)
}

func TestLeadRepository_List_TeamIDFilter_ComposesWithScope(t *testing.T) {
	db := setupTestDB(t)
	teamA, _, _, _, _, _ := seedScopeFixture(t, db)
	repo := NewLeadRepository(db)

	// ADMIN scope (empty) narrowed by team_id query param must equal the
	// team-scoped result directly - proves the subquery narrowing filter
	// doesn't collide with (or duplicate) the scope's own subquery.
	results, _, err := repo.List(LeadScope{}, LeadFilter{TeamID: &teamA.ID, Page: 1, Limit: 20})
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestLeadRepository_RecordFollowUp(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLeadRepository(db)
	userID := seedTestUser(t, db)
	lead := &models.Lead{Code: "LD-2607-0901", OwnerID: userID, CreatedByID: userID, CompanyName: "A"}
	require.NoError(t, repo.Create(lead))

	require.NoError(t, repo.RecordFollowUp(lead.ID))

	found, _ := repo.FindByCode(LeadScope{}, lead.Code)
	assert.NotNil(t, found.LastFollowUpAt)
	assert.Equal(t, 1, found.FollowUpCount)

	require.NoError(t, repo.RecordFollowUp(lead.ID))
	found, _ = repo.FindByCode(LeadScope{}, lead.Code)
	assert.Equal(t, 2, found.FollowUpCount, "increments, does not reset")
}

func TestLeadRepository_MaybeTransitionToFollowUp(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLeadRepository(db)
	userID := seedTestUser(t, db)

	baruLead := &models.Lead{Code: "LD-2607-1001", OwnerID: userID, CreatedByID: userID, CompanyName: "A"}
	require.NoError(t, repo.Create(baruLead))
	require.NoError(t, repo.MaybeTransitionToFollowUp(baruLead.ID))
	found, _ := repo.FindByCode(LeadScope{}, baruLead.Code)
	assert.Equal(t, "FOLLOW_UP", found.Status)

	lostLead := &models.Lead{Code: "LD-2607-1002", OwnerID: userID, CreatedByID: userID, CompanyName: "B", Status: "LOST", LostReason: strPtr("x")}
	require.NoError(t, db.Create(lostLead).Error)
	require.NoError(t, repo.MaybeTransitionToFollowUp(lostLead.ID))
	found, _ = repo.FindByCode(LeadScope{}, lostLead.Code)
	assert.Equal(t, "LOST", found.Status, "already-terminal status must not be pulled back")
}

func strPtr(s string) *string { return &s }
