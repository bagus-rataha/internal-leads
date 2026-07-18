//go:build integration

package repository

import (
	"fiber-api-boilerplate/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLeadSourceRepository_Create_FindByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLeadSourceRepository(db)

	source := &models.LeadSource{Name: "Website"}
	err := repo.Create(source)

	assert.NoError(t, err)
	assert.NotEmpty(t, source.ID)

	found, err := repo.FindByID(source.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Website", found.Name)
	assert.True(t, found.IsActive)
}

func TestLeadSourceRepository_List_IncludesInactive(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLeadSourceRepository(db)

	active := &models.LeadSource{Name: "Website"}
	db.Create(active)
	inactive := &models.LeadSource{Name: "Old Source"}
	db.Create(inactive)
	db.Model(inactive).Update("is_active", false)

	sources, err := repo.List()

	assert.NoError(t, err)
	assert.Len(t, sources, 2)
}

func TestLeadSourceRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLeadSourceRepository(db)

	source := &models.LeadSource{Name: "Old Name"}
	db.Create(source)

	source.Name = "New Name"
	source.IsActive = false
	err := repo.Update(source)

	assert.NoError(t, err)

	found, _ := repo.FindByID(source.ID)
	assert.Equal(t, "New Name", found.Name)
	assert.False(t, found.IsActive)
}
