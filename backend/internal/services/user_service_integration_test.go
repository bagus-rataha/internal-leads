//go:build integration

package services

import (
	"os"
	"strings"
	"testing"
	"time"

	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/models"
	"fiber-api-boilerplate/internal/repository"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupServiceTestDB connects to the dedicated test database, applies the
// migrations, and truncates tables so each test starts clean. Mirrors the
// repository package's own integration helper; kept here because the services
// package can't import that package-private one.
func setupServiceTestDB(t *testing.T) *gorm.DB {
	_ = godotenv.Load("../../.env")

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Fatal("TEST_DATABASE_URL is not set; required for service integration tests")
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		t.Fatal("Failed to connect test DB:", err)
	}

	m, err := migrate.New("file://../../migrations", dbURL)
	if err != nil {
		t.Fatal("Failed to init migrate:", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatal("Failed to run migrations:", err)
	}

	var tables []string
	if err := db.Raw(`SELECT tablename FROM pg_tables
		WHERE schemaname = 'public' AND tablename <> 'schema_migrations'`).
		Scan(&tables).Error; err != nil {
		t.Fatal("Failed to list tables:", err)
	}
	if len(tables) > 0 {
		if err := db.Exec("TRUNCATE " + strings.Join(tables, ", ") + " RESTART IDENTITY CASCADE").Error; err != nil {
			t.Fatal("Failed to truncate tables:", err)
		}
	}
	return db
}

// TestUpdateUser_Transaction_CommitsMutationAndRevokeTogether drives the real
// public UpdateUser (not the private helper) against a live database, proving
// the tx-scoped repositories it constructs actually persist the user change
// and the refresh-token revoke together. The mock-based unit tests exercise
// the private helper only, so this is the sole coverage of the transactional
// path itself.
func TestUpdateUser_Transaction_CommitsMutationAndRevokeTogether(t *testing.T) {
	db := setupServiceTestDB(t)

	userRepo := repository.NewUserRepository(db)
	rtRepo := repository.NewRefreshTokenRepository(db)
	teamRepo := repository.NewSalesTeamRepository(db)
	svc := NewUserService(db, userRepo, rtRepo, teamRepo)

	team := &models.SalesTeam{Name: "Team A", IsActive: true}
	require.NoError(t, db.Create(team).Error)

	user := &models.User{
		Email:    "sales@test.local",
		Password: "hash",
		Name:     "Sales A",
		Role:     "SALES",
		TeamID:   &team.ID,
		IsActive: true,
	}
	require.NoError(t, db.Create(user).Error)

	// A live refresh token that the role change must revoke.
	require.NoError(t, rtRepo.Create(&models.RefreshToken{
		Token:     "live-token",
		UserID:    user.ID,
		ExpiredAt: time.Now().Add(time.Hour),
	}))

	newRole := "LEADER"
	_, err := svc.UpdateUser("SU", user.ID, dto.UpdateUserInput{Role: &newRole})
	require.NoError(t, err)

	// Both effects must have committed.
	var reloaded models.User
	require.NoError(t, db.First(&reloaded, "id = ?", user.ID).Error)
	assert.Equal(t, "LEADER", reloaded.Role)

	var tokenCount int64
	require.NoError(t, db.Model(&models.RefreshToken{}).
		Where("user_id = ?", user.ID).Count(&tokenCount).Error)
	assert.Equal(t, int64(0), tokenCount, "refresh tokens should be revoked in the same transaction")
}

// TestUpdateUser_NameOnly_KeepsTokens confirms the transactional path does not
// revoke when neither role nor team_id changes.
func TestUpdateUser_NameOnly_KeepsTokens(t *testing.T) {
	db := setupServiceTestDB(t)

	userRepo := repository.NewUserRepository(db)
	rtRepo := repository.NewRefreshTokenRepository(db)
	teamRepo := repository.NewSalesTeamRepository(db)
	svc := NewUserService(db, userRepo, rtRepo, teamRepo)

	team := &models.SalesTeam{Name: "Team A", IsActive: true}
	require.NoError(t, db.Create(team).Error)
	user := &models.User{
		Email: "s2@test.local", Password: "hash", Name: "Old", Role: "SALES",
		TeamID: &team.ID, IsActive: true,
	}
	require.NoError(t, db.Create(user).Error)
	require.NoError(t, rtRepo.Create(&models.RefreshToken{
		Token: "keep-token", UserID: user.ID, ExpiredAt: time.Now().Add(time.Hour),
	}))

	newName := "New Name"
	_, err := svc.UpdateUser("SU", user.ID, dto.UpdateUserInput{Name: &newName})
	require.NoError(t, err)

	var tokenCount int64
	require.NoError(t, db.Model(&models.RefreshToken{}).
		Where("user_id = ?", user.ID).Count(&tokenCount).Error)
	assert.Equal(t, int64(1), tokenCount, "a name-only change must not revoke tokens")
}
