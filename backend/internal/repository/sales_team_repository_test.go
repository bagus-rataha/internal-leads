//go:build integration

package repository

import (
	"fiber-api-boilerplate/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSalesTeamRepository_Create_FindByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSalesTeamRepository(db)

	team := &models.SalesTeam{Name: "Team A"}
	err := repo.Create(team)

	assert.NoError(t, err)
	assert.NotEmpty(t, team.ID)

	found, err := repo.FindByID(team.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Team A", found.Name)
	assert.True(t, found.IsActive)
}

func TestSalesTeamRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSalesTeamRepository(db)

	db.Create(&models.SalesTeam{Name: "Team A"})
	db.Create(&models.SalesTeam{Name: "Team B"})

	teams, err := repo.List()

	assert.NoError(t, err)
	assert.Len(t, teams, 2)
}

func TestSalesTeamRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSalesTeamRepository(db)

	team := &models.SalesTeam{Name: "Old Name"}
	db.Create(team)

	team.Name = "New Name"
	team.IsActive = false
	err := repo.Update(team)

	assert.NoError(t, err)

	found, _ := repo.FindByID(team.ID)
	assert.Equal(t, "New Name", found.Name)
	assert.False(t, found.IsActive)
}

func TestSalesTeamRepository_CountActiveMembers(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSalesTeamRepository(db)

	team := &models.SalesTeam{Name: "Team A"}
	db.Create(team)

	otherTeam := &models.SalesTeam{Name: "Team B"}
	db.Create(otherTeam)

	// 2 active members of team
	db.Create(&models.User{Email: "active1@test.com", Password: "h", Name: "Active 1", Role: "SALES", TeamID: &team.ID, IsActive: true})
	db.Create(&models.User{Email: "active2@test.com", Password: "h", Name: "Active 2", Role: "SALES", TeamID: &team.ID, IsActive: true})
	// 1 inactive member of team - should not be counted.
	// IsActive:false is the Go zero value, and the column has a `default:true`
	// tag, so GORM omits it from the INSERT and lets the DB default apply -
	// force it false with an explicit UPDATE after create.
	inactiveUser := &models.User{Email: "inactive@test.com", Password: "h", Name: "Inactive", Role: "SALES", TeamID: &team.ID, IsActive: true}
	db.Create(inactiveUser)
	db.Model(inactiveUser).Update("is_active", false)
	// active member of a different team - should not be counted
	db.Create(&models.User{Email: "other@test.com", Password: "h", Name: "Other Team", Role: "SALES", TeamID: &otherTeam.ID, IsActive: true})

	count, err := repo.CountActiveMembers(team.ID)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}
