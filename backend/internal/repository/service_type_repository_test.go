//go:build integration

package repository

import (
	"fiber-api-boilerplate/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceTypeRepository_Create_FindByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceTypeRepository(db)

	serviceType := &models.ServiceType{Name: "Dedicated"}
	err := repo.Create(serviceType)

	assert.NoError(t, err)
	assert.NotEmpty(t, serviceType.ID)

	found, err := repo.FindByID(serviceType.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Dedicated", found.Name)
	assert.True(t, found.IsActive)
}

func TestServiceTypeRepository_List_IncludesInactive(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceTypeRepository(db)

	active := &models.ServiceType{Name: "Dedicated"}
	db.Create(active)
	inactive := &models.ServiceType{Name: "Old Type"}
	db.Create(inactive)
	db.Model(inactive).Update("is_active", false)

	types, err := repo.List()

	assert.NoError(t, err)
	assert.Len(t, types, 2)
}

func TestServiceTypeRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceTypeRepository(db)

	serviceType := &models.ServiceType{Name: "Old Name"}
	db.Create(serviceType)

	serviceType.Name = "New Name"
	serviceType.IsActive = false
	err := repo.Update(serviceType)

	assert.NoError(t, err)

	found, _ := repo.FindByID(serviceType.ID)
	assert.Equal(t, "New Name", found.Name)
	assert.False(t, found.IsActive)
}
