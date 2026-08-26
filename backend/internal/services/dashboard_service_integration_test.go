//go:build integration

package services

import (
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/models"
	"fiber-api-boilerplate/internal/repository"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDashboardSummary_SalesScope_OnlyOwnLeadsCounted(t *testing.T) {
	db := setupServiceTestDB(t)
	userRepo := repository.NewUserRepository(db)
	svc := NewDashboardService(db, userRepo)

	teamID := uuid.Must(uuid.NewV7())
	require.NoError(t, db.Create(&models.SalesTeam{BaseModel: models.BaseModel{ID: teamID}, Name: "Team A", IsActive: true}).Error)

	salesA := uuid.Must(uuid.NewV7())
	salesB := uuid.Must(uuid.NewV7())
	require.NoError(t, db.Create(&models.User{BaseModel: models.BaseModel{ID: salesA}, Email: "a@test.local", Password: "h", Name: "Sales A", Role: "SALES", TeamID: &teamID, IsActive: true}).Error)
	require.NoError(t, db.Create(&models.User{BaseModel: models.BaseModel{ID: salesB}, Email: "b@test.local", Password: "h", Name: "Sales B", Role: "SALES", TeamID: &teamID, IsActive: true}).Error)

	now := time.Now()
	require.NoError(t, db.Create(&models.Lead{Code: "LD-2607-0001", OwnerID: salesA, CreatedByID: salesA, CompanyName: "Owned by A"}).Error)
	require.NoError(t, db.Create(&models.Lead{Code: "LD-2607-0002", OwnerID: salesB, CreatedByID: salesB, CompanyName: "Owned by B"}).Error)

	q := dto.DashboardQuery{DateFrom: now.AddDate(0, 0, -1), DateTo: now.AddDate(0, 0, 1)}
	result, err := svc.Summary(salesA, "SALES", q)

	require.NoError(t, err)
	assert.Equal(t, int64(1), result.LeadBaru.Value, "SALES caller must only see their own lead, not both")
}

func TestDashboardSummary_ZeroBaseline_ChangePctNilNotHundred(t *testing.T) {
	db := setupServiceTestDB(t)
	userRepo := repository.NewUserRepository(db)
	svc := NewDashboardService(db, userRepo)

	adminID := uuid.Must(uuid.NewV7())
	require.NoError(t, db.Create(&models.User{BaseModel: models.BaseModel{ID: adminID}, Email: "admin@test.local", Password: "h", Name: "Admin", Role: "SU", IsActive: true}).Error)
	require.NoError(t, db.Create(&models.Lead{Code: "LD-2607-0003", OwnerID: adminID, CreatedByID: adminID, CompanyName: "Only lead"}).Error)

	now := time.Now()
	q := dto.DashboardQuery{DateFrom: now.AddDate(0, 0, -1), DateTo: now.AddDate(0, 0, 1)}
	result, err := svc.Summary(adminID, "SU", q)

	require.NoError(t, err)
	assert.Equal(t, int64(1), result.LeadBaru.Value)
	assert.Nil(t, result.LeadBaru.ChangePct, "comparison period has 0 leads -> nil, never a fake 100%")
}

func TestDashboardActivity_BucketsOneRowPerDay_ScopedCounts(t *testing.T) {
	db := setupServiceTestDB(t)
	userRepo := repository.NewUserRepository(db)
	svc := NewDashboardService(db, userRepo)

	adminID := uuid.Must(uuid.NewV7())
	require.NoError(t, db.Create(&models.User{BaseModel: models.BaseModel{ID: adminID}, Email: "admin2@test.local", Password: "h", Name: "Admin", Role: "SU", IsActive: true}).Error)

	today := time.Now().Truncate(24 * time.Hour)
	lead := &models.Lead{Code: "LD-2607-0004", OwnerID: adminID, CreatedByID: adminID, CompanyName: "Today's lead"}
	require.NoError(t, db.Create(lead).Error)

	q := dto.DashboardQuery{DateFrom: today.AddDate(0, 0, -2), DateTo: today}
	result, err := svc.Activity(adminID, "SU", q)

	require.NoError(t, err)
	assert.Len(t, result.Buckets, 3, "3-day range -> 3 daily buckets")
	last := result.Buckets[len(result.Buckets)-1]
	assert.Equal(t, today.Format("2006-01-02"), last.Date)
	assert.Equal(t, int64(1), last.LeadBaru, "the lead created today lands in today's bucket")
	assert.Equal(t, int64(0), result.Buckets[0].LeadBaru, "a day with no leads is a real zero, not an omitted bucket")
}

func TestDashboardStaleLeads_Top7MostOverdueFirst(t *testing.T) {
	db := setupServiceTestDB(t)
	userRepo := repository.NewUserRepository(db)
	svc := NewDashboardService(db, userRepo)

	adminID := uuid.Must(uuid.NewV7())
	require.NoError(t, db.Create(&models.User{BaseModel: models.BaseModel{ID: adminID}, Email: "admin3@test.local", Password: "h", Name: "Admin", Role: "SU", IsActive: true}).Error)

	oldTime := time.Now().AddDate(0, 0, -20)
	recentTime := time.Now().AddDate(0, 0, -10)
	require.NoError(t, db.Create(&models.Lead{Code: "LD-2607-0005", OwnerID: adminID, CreatedByID: adminID, CompanyName: "Very stale", Status: "BARU", BaseModel: models.BaseModel{CreatedAt: oldTime}}).Error)
	require.NoError(t, db.Create(&models.Lead{Code: "LD-2607-0006", OwnerID: adminID, CreatedByID: adminID, CompanyName: "Less stale", Status: "FOLLOW_UP", LastFollowUpAt: &recentTime}).Error)
	require.NoError(t, db.Create(&models.Lead{Code: "LD-2607-0007", OwnerID: adminID, CreatedByID: adminID, CompanyName: "Handoff, never stale", Status: "HANDOFF_ODOO", BaseModel: models.BaseModel{CreatedAt: oldTime}}).Error)

	q := dto.DashboardQuery{DateFrom: time.Now().AddDate(0, 0, -30), DateTo: time.Now()}
	result, err := svc.StaleLeads(adminID, "SU", q)

	require.NoError(t, err)
	require.Len(t, result.Items, 2, "HANDOFF_ODOO lead is terminal, never stale, excluded")
	assert.Equal(t, "LD-2607-0005", result.Items[0].Code, "most-overdue (created 20d ago) sorts first")
	assert.Equal(t, "LD-2607-0006", result.Items[1].Code)
}

func TestDashboardSalesActivity_InactiveRowsSortFirst(t *testing.T) {
	db := setupServiceTestDB(t)
	userRepo := repository.NewUserRepository(db)
	svc := NewDashboardService(db, userRepo)

	teamID := uuid.Must(uuid.NewV7())
	require.NoError(t, db.Create(&models.SalesTeam{BaseModel: models.BaseModel{ID: teamID}, Name: "Team X", IsActive: true}).Error)

	active := uuid.Must(uuid.NewV7())
	inactive := uuid.Must(uuid.NewV7())
	require.NoError(t, db.Create(&models.User{BaseModel: models.BaseModel{ID: active}, Email: "active@test.local", Password: "h", Name: "Active Sales", Role: "SALES", TeamID: &teamID, IsActive: true}).Error)
	require.NoError(t, db.Create(&models.User{BaseModel: models.BaseModel{ID: inactive}, Email: "inactive@test.local", Password: "h", Name: "Inactive Sales", Role: "SALES", TeamID: &teamID, IsActive: true}).Error)

	lead := &models.Lead{Code: "LD-2607-0008", OwnerID: active, CreatedByID: active, CompanyName: "Active's lead"}
	require.NoError(t, db.Create(lead).Error)
	require.NoError(t, db.Create(&models.FollowUp{LeadID: lead.ID, Note: "recent", CreatedByID: active}).Error)

	q := dto.DashboardQuery{DateFrom: time.Now().AddDate(0, 0, -30), DateTo: time.Now()}
	result, err := svc.SalesActivity(active, "SU", q)

	require.NoError(t, err)
	require.Len(t, result.Items, 2)
	assert.Equal(t, "Inactive Sales", result.Items[0].Name, "no follow-up in 7 days sorts first (needs attention)")
	if assert.NotNil(t, result.Items[0].AttentionTag) {
		assert.Equal(t, "Tanpa aktivitas 7 hari", *result.Items[0].AttentionTag)
	}
	assert.Nil(t, result.Items[1].AttentionTag, "the active seller has a recent follow-up, no tag")
}

func TestDashboardSalesActivity_ZeroOwnedLeads_ConvPctNil(t *testing.T) {
	db := setupServiceTestDB(t)
	userRepo := repository.NewUserRepository(db)
	svc := NewDashboardService(db, userRepo)

	teamID := uuid.Must(uuid.NewV7())
	require.NoError(t, db.Create(&models.SalesTeam{BaseModel: models.BaseModel{ID: teamID}, Name: "Team Y", IsActive: true}).Error)
	salesID := uuid.Must(uuid.NewV7())
	require.NoError(t, db.Create(&models.User{BaseModel: models.BaseModel{ID: salesID}, Email: "nolead@test.local", Password: "h", Name: "No Lead Sales", Role: "SALES", TeamID: &teamID, IsActive: true}).Error)

	q := dto.DashboardQuery{DateFrom: time.Now().AddDate(0, 0, -30), DateTo: time.Now()}
	result, err := svc.SalesActivity(salesID, "SU", q)

	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	assert.Nil(t, result.Items[0].ConvPct, "0 owned leads -> nil, not a divide-by-zero 0%")
}

func TestDashboardSegments_NoHandoffAnywhere_AllConversionsNil(t *testing.T) {
	db := setupServiceTestDB(t)
	userRepo := repository.NewUserRepository(db)
	svc := NewDashboardService(db, userRepo)

	adminID := uuid.Must(uuid.NewV7())
	require.NoError(t, db.Create(&models.User{BaseModel: models.BaseModel{ID: adminID}, Email: "admin4@test.local", Password: "h", Name: "Admin", Role: "SU", IsActive: true}).Error)
	source := &models.LeadSource{Name: "Google", IsActive: true}
	require.NoError(t, db.Create(source).Error)
	require.NoError(t, db.Create(&models.Lead{Code: "LD-2607-0009", OwnerID: adminID, CreatedByID: adminID, CompanyName: "No handoff yet", LeadSourceID: &source.ID, Status: "BARU"}).Error)

	q := dto.DashboardQuery{DateFrom: time.Now().AddDate(0, 0, -30), DateTo: time.Now()}
	result, err := svc.Segments(adminID, "SU", q)

	require.NoError(t, err)
	assert.False(t, result.AnyHandoff)
	require.Len(t, result.Sources, 1)
	assert.Nil(t, result.Sources[0].ConversionPct, "zero handoffs anywhere in scope -> null, not 0%")
}

func TestDashboardSegments_HighVolumeLowConversion_WarnFlagSet(t *testing.T) {
	db := setupServiceTestDB(t)
	userRepo := repository.NewUserRepository(db)
	svc := NewDashboardService(db, userRepo)

	adminID := uuid.Must(uuid.NewV7())
	require.NoError(t, db.Create(&models.User{BaseModel: models.BaseModel{ID: adminID}, Email: "admin5@test.local", Password: "h", Name: "Admin", Role: "SU", IsActive: true}).Error)
	source := &models.LeadSource{Name: "Cold Call", IsActive: true}
	require.NoError(t, db.Create(source).Error)
	otherSource := &models.LeadSource{Name: "Referral", IsActive: true}
	require.NoError(t, db.Create(otherSource).Error)

	// One handoff elsewhere so anyHandoff is true, then 4 low-converting
	// Cold Call leads (1/4 = 25%... use 5 leads/0 handoff = 0% to force <20%).
	require.NoError(t, db.Create(&models.Lead{Code: "LD-2607-0010", OwnerID: adminID, CreatedByID: adminID, CompanyName: "Converts", LeadSourceID: &otherSource.ID, Status: "HANDOFF_ODOO"}).Error)
	for i := 0; i < 4; i++ {
		require.NoError(t, db.Create(&models.Lead{Code: fmt.Sprintf("LD-2607-%04d", 11+i), OwnerID: adminID, CreatedByID: adminID, CompanyName: "Cold call lead", LeadSourceID: &source.ID, Status: "BARU"}).Error)
	}

	q := dto.DashboardQuery{DateFrom: time.Now().AddDate(0, 0, -30), DateTo: time.Now()}
	result, err := svc.Segments(adminID, "SU", q)

	require.NoError(t, err)
	assert.True(t, result.AnyHandoff)
	var coldCall dto.SegmentRow
	for _, s := range result.Sources {
		if s.Name == "Cold Call" {
			coldCall = s
		}
	}
	assert.True(t, coldCall.Warn, "4 leads (>=3), 0% conversion (<20%) -> warn")
	if assert.NotNil(t, coldCall.ConversionPct) {
		assert.Equal(t, 0, *coldCall.ConversionPct)
	}
}

func TestDashboardSummary_ForecastMrr_SummedAcrossAllStatusesIncludingLost(t *testing.T) {
	db := setupServiceTestDB(t)
	userRepo := repository.NewUserRepository(db)
	svc := NewDashboardService(db, userRepo)

	adminID := uuid.Must(uuid.NewV7())
	require.NoError(t, db.Create(&models.User{BaseModel: models.BaseModel{ID: adminID}, Email: "forecast-dash@test.local", Password: "h", Name: "Admin", Role: "SU", IsActive: true}).Error)

	f1, f2 := 1000000.0, 2000000.0
	require.NoError(t, db.Create(&models.Lead{Code: "LD-2607-0010", OwnerID: adminID, CreatedByID: adminID, CompanyName: "Live", Status: "FOLLOW_UP", ForecastMrr: &f1}).Error)
	require.NoError(t, db.Create(&models.Lead{Code: "LD-2607-0011", OwnerID: adminID, CreatedByID: adminID, CompanyName: "Lost deal", Status: "LOST", LostReason: strPtrSvc("no budget"), ForecastMrr: &f2}).Error)

	now := time.Now()
	q := dto.DashboardQuery{DateFrom: now.AddDate(0, 0, -1), DateTo: now.AddDate(0, 0, 1)}
	result, err := svc.Summary(adminID, "SU", q)

	require.NoError(t, err)
	assert.Equal(t, f1+f2, result.TotalForecastMrr, "LOST leads must still count toward the total per the user's explicit choice")
}

func TestDashboardSummary_ForecastMrr_StatusFilterNarrowsOnlyForecast(t *testing.T) {
	db := setupServiceTestDB(t)
	userRepo := repository.NewUserRepository(db)
	svc := NewDashboardService(db, userRepo)

	adminID := uuid.Must(uuid.NewV7())
	require.NoError(t, db.Create(&models.User{BaseModel: models.BaseModel{ID: adminID}, Email: "forecast-filter@test.local", Password: "h", Name: "Admin", Role: "SU", IsActive: true}).Error)

	f1, f2 := 1000000.0, 2000000.0
	require.NoError(t, db.Create(&models.Lead{Code: "LD-2607-0012", OwnerID: adminID, CreatedByID: adminID, CompanyName: "Live", Status: "FOLLOW_UP", ForecastMrr: &f1}).Error)
	require.NoError(t, db.Create(&models.Lead{Code: "LD-2607-0013", OwnerID: adminID, CreatedByID: adminID, CompanyName: "Lost deal", Status: "LOST", LostReason: strPtrSvc("no budget"), ForecastMrr: &f2}).Error)

	now := time.Now()
	status := "FOLLOW_UP"
	q := dto.DashboardQuery{DateFrom: now.AddDate(0, 0, -1), DateTo: now.AddDate(0, 0, 1), Status: &status}
	result, err := svc.Summary(adminID, "SU", q)

	require.NoError(t, err)
	assert.Equal(t, f1, result.TotalForecastMrr, "status filter narrows the forecast sum to FOLLOW_UP only")
	assert.Equal(t, int64(2), result.LeadBaru.Value, "the status filter must NOT narrow lead_baru - both leads were created in-window regardless of status")
}

func TestDashboardSalesActivity_ForecastMrr_SummedPerOwner(t *testing.T) {
	db := setupServiceTestDB(t)
	userRepo := repository.NewUserRepository(db)
	svc := NewDashboardService(db, userRepo)

	teamID := uuid.Must(uuid.NewV7())
	require.NoError(t, db.Create(&models.SalesTeam{BaseModel: models.BaseModel{ID: teamID}, Name: "Team Forecast", IsActive: true}).Error)
	salesID := uuid.Must(uuid.NewV7())
	require.NoError(t, db.Create(&models.User{BaseModel: models.BaseModel{ID: salesID}, Email: "sales-forecast@test.local", Password: "h", Name: "Sales F", Role: "SALES", TeamID: &teamID, IsActive: true}).Error)

	f1, f2 := 1500000.0, 2500000.0
	require.NoError(t, db.Create(&models.Lead{Code: "LD-2607-0014", OwnerID: salesID, CreatedByID: salesID, CompanyName: "A", Status: "FOLLOW_UP", ForecastMrr: &f1}).Error)
	require.NoError(t, db.Create(&models.Lead{Code: "LD-2607-0015", OwnerID: salesID, CreatedByID: salesID, CompanyName: "B", Status: "LOST", LostReason: strPtrSvc("no budget"), ForecastMrr: &f2}).Error)

	now := time.Now()
	q := dto.DashboardQuery{DateFrom: now.AddDate(0, 0, -1), DateTo: now.AddDate(0, 0, 1)}
	result, err := svc.SalesActivity(salesID, "SALES", q)

	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	assert.Equal(t, f1+f2, result.Items[0].ForecastMrr)
}

func TestDashboardSalesActivity_LeaderInRosterAndFilterable(t *testing.T) {
	db := setupServiceTestDB(t)
	userRepo := repository.NewUserRepository(db)
	svc := NewDashboardService(db, userRepo)

	teamID := uuid.Must(uuid.NewV7())
	require.NoError(t, db.Create(&models.SalesTeam{BaseModel: models.BaseModel{ID: teamID}, Name: "Team Leader", IsActive: true}).Error)

	salesID := uuid.Must(uuid.NewV7())
	leaderID := uuid.Must(uuid.NewV7())
	require.NoError(t, db.Create(&models.User{BaseModel: models.BaseModel{ID: salesID}, Email: "sales@test.local", Password: "h", Name: "Sales User", Role: "SALES", TeamID: &teamID, IsActive: true}).Error)
	require.NoError(t, db.Create(&models.User{BaseModel: models.BaseModel{ID: leaderID}, Email: "leader@test.local", Password: "h", Name: "Leader User", Role: "LEADER", TeamID: &teamID, IsActive: true}).Error)

	lead := &models.Lead{Code: "LD-2607-0015", OwnerID: leaderID, CreatedByID: leaderID, CompanyName: "Leader's lead"}
	require.NoError(t, db.Create(lead).Error)

	adminID := uuid.Must(uuid.NewV7())
	require.NoError(t, db.Create(&models.User{BaseModel: models.BaseModel{ID: adminID}, Email: "admin6@test.local", Password: "h", Name: "Admin", Role: "SU", IsActive: true}).Error)

	// Test 1: Full roster includes both SALES and LEADER
	q := dto.DashboardQuery{DateFrom: time.Now().AddDate(0, 0, -30), DateTo: time.Now()}
	result, err := svc.SalesActivity(adminID, "SU", q)
	require.NoError(t, err)
	require.Len(t, result.Items, 2, "roster must include both SALES and LEADER users")

	// Test 2: Filter by leader ID returns exactly 1 row (regression: used to return empty table)
	q.OwnerID = &leaderID
	result, err = svc.SalesActivity(adminID, "SU", q)
	require.NoError(t, err)
	require.Len(t, result.Items, 1, "filtering by leader ID must return exactly 1 row with the leader's data")
	assert.Equal(t, leaderID, result.Items[0].UserID)
	assert.Equal(t, "Leader User", result.Items[0].Name)
}
