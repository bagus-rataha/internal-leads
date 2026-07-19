//go:build integration

package services

import (
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/models"
	"fiber-api-boilerplate/internal/repository"
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
