//go:build integration

package services

import (
	"testing"

	"fiber-api-boilerplate/internal/models"
	"fiber-api-boilerplate/internal/repository"
	"fiber-api-boilerplate/internal/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportService_Export_WrongPassword_ReturnsErrInvalidPassword(t *testing.T) {
	db := setupServiceTestDB(t)
	userRepo := repository.NewUserRepository(db)
	svc := NewExportService(db, userRepo)

	hash, err := utils.HashPassword("correct-password")
	require.NoError(t, err)
	owner := &models.User{Email: "export-owner@test.local", Password: hash, Name: "Owner", Role: "ADMIN_SALES", IsActive: true}
	require.NoError(t, db.Create(owner).Error)

	_, err = svc.Export(owner.ID, "ADMIN_SALES", "wrong-password", repository.LeadFilter{})
	assert.ErrorIs(t, err, ErrExportInvalidPassword)
}

func TestExportService_Export_ScopesToOwnLeadsForSales(t *testing.T) {
	db := setupServiceTestDB(t)
	userRepo := repository.NewUserRepository(db)
	svc := NewExportService(db, userRepo)

	team := &models.SalesTeam{Name: "Export Team", IsActive: true}
	require.NoError(t, db.Create(team).Error)

	hash, err := utils.HashPassword("secret123")
	require.NoError(t, err)
	salesA := &models.User{Email: "export-salesa@test.local", Password: hash, Name: "Sales A", Role: "SALES", TeamID: &team.ID, IsActive: true}
	salesB := &models.User{Email: "export-salesb@test.local", Password: hash, Name: "Sales B", Role: "SALES", TeamID: &team.ID, IsActive: true}
	require.NoError(t, db.Create(salesA).Error)
	require.NoError(t, db.Create(salesB).Error)

	leadA := &models.Lead{Code: "LD-2607-0101", OwnerID: salesA.ID, CreatedByID: salesA.ID, CompanyName: "Company A"}
	leadB := &models.Lead{Code: "LD-2607-0102", OwnerID: salesB.ID, CreatedByID: salesB.ID, CompanyName: "Company B"}
	require.NoError(t, db.Create(leadA).Error)
	require.NoError(t, db.Create(leadB).Error)

	f, err := svc.Export(salesA.ID, "SALES", "secret123", repository.LeadFilter{})
	require.NoError(t, err)

	rows, err := f.GetRows(leadsSheetName)
	require.NoError(t, err)
	require.Len(t, rows, 2) // header + leadA only
	assert.Equal(t, "Company A", rows[1][2])
}

func TestExportService_Export_OverRowCap_ReturnsErrTooManyRows(t *testing.T) {
	db := setupServiceTestDB(t)
	userRepo := repository.NewUserRepository(db)
	svc := NewExportService(db, userRepo)

	original := MaxExportRows
	MaxExportRows = 1
	t.Cleanup(func() { MaxExportRows = original })

	hash, err := utils.HashPassword("secret123")
	require.NoError(t, err)
	admin := &models.User{Email: "export-admin@test.local", Password: hash, Name: "Admin", Role: "ADMIN_SALES", IsActive: true}
	require.NoError(t, db.Create(admin).Error)

	require.NoError(t, db.Create(&models.Lead{Code: "LD-2607-0103", OwnerID: admin.ID, CreatedByID: admin.ID, CompanyName: "A"}).Error)
	require.NoError(t, db.Create(&models.Lead{Code: "LD-2607-0104", OwnerID: admin.ID, CreatedByID: admin.ID, CompanyName: "B"}).Error)

	_, err = svc.Export(admin.ID, "ADMIN_SALES", "secret123", repository.LeadFilter{})
	var tooMany *ErrExportTooManyRows
	require.ErrorAs(t, err, &tooMany)
	assert.Equal(t, int64(2), tooMany.Count)
}

func TestExportService_Export_IncludesFollowUps(t *testing.T) {
	db := setupServiceTestDB(t)
	userRepo := repository.NewUserRepository(db)
	svc := NewExportService(db, userRepo)

	hash, err := utils.HashPassword("secret123")
	require.NoError(t, err)
	admin := &models.User{Email: "export-admin2@test.local", Password: hash, Name: "Admin", Role: "ADMIN_SALES", IsActive: true}
	require.NoError(t, db.Create(admin).Error)

	lead := &models.Lead{Code: "LD-2607-0105", OwnerID: admin.ID, CreatedByID: admin.ID, CompanyName: "With Followup"}
	require.NoError(t, db.Create(lead).Error)
	require.NoError(t, db.Create(&models.FollowUp{LeadID: lead.ID, Note: "Sudah dihubungi", CreatedByID: admin.ID}).Error)

	f, err := svc.Export(admin.ID, "ADMIN_SALES", "secret123", repository.LeadFilter{})
	require.NoError(t, err)

	rows, err := f.GetRows(followUpsSheetName)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, "LD-2607-0105", rows[1][0])
	assert.Equal(t, "Sudah dihubungi", rows[1][3])
}
