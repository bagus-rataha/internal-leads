//go:build integration

package repository

import (
	"fiber-api-boilerplate/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReferenceRepository_ListActiveLeadSources_ExcludesInactive(t *testing.T) {
	db := setupTestDB(t)
	repo := NewReferenceRepository(db)

	db.Create(&models.LeadSource{Name: "Google", IsActive: true})
	inactive := &models.LeadSource{Name: "Legacy Source", IsActive: true}
	db.Create(inactive)
	db.Model(inactive).Update("is_active", false)

	sources, err := repo.ListActiveLeadSources()

	assert.NoError(t, err)
	assert.Len(t, sources, 1)
	assert.Equal(t, "Google", sources[0].Name)
}

func TestReferenceRepository_ListActiveServiceTypes_ExcludesInactive(t *testing.T) {
	db := setupTestDB(t)
	repo := NewReferenceRepository(db)

	db.Create(&models.ServiceType{Name: "Dedicated", IsActive: true})
	inactive := &models.ServiceType{Name: "Discontinued", IsActive: true}
	db.Create(inactive)
	db.Model(inactive).Update("is_active", false)

	types, err := repo.ListActiveServiceTypes()

	assert.NoError(t, err)
	assert.Len(t, types, 1)
	assert.Equal(t, "Dedicated", types[0].Name)
}

func TestReferenceRepository_ListCitiesByProvince_FiltersByParent(t *testing.T) {
	db := setupTestDB(t)
	repo := NewReferenceRepository(db)

	db.Create(&models.Province{ID: 1, ExternalID: "p1", Name: "Prov A"})
	db.Create(&models.Province{ID: 2, ExternalID: "p2", Name: "Prov B"})
	db.Create(&models.City{ID: 1, ExternalID: "c1", ProvinceID: 1, Name: "City A1"})
	db.Create(&models.City{ID: 2, ExternalID: "c2", ProvinceID: 2, Name: "City B1"})

	cities, err := repo.ListCitiesByProvince(1)

	assert.NoError(t, err)
	assert.Len(t, cities, 1)
	assert.Equal(t, "City A1", cities[0].Name)
}

func TestReferenceRepository_ListDistrictsByCity_FiltersByParent(t *testing.T) {
	db := setupTestDB(t)
	repo := NewReferenceRepository(db)

	db.Create(&models.Province{ID: 1, ExternalID: "p1", Name: "Prov A"})
	db.Create(&models.City{ID: 1, ExternalID: "c1", ProvinceID: 1, Name: "City A"})
	db.Create(&models.City{ID: 2, ExternalID: "c2", ProvinceID: 1, Name: "City B"})
	db.Create(&models.District{ID: 1, ExternalID: "d1", CityID: 1, Name: "District A1"})
	db.Create(&models.District{ID: 2, ExternalID: "d2", CityID: 2, Name: "District B1"})

	districts, err := repo.ListDistrictsByCity(1)

	assert.NoError(t, err)
	assert.Len(t, districts, 1)
	assert.Equal(t, "District A1", districts[0].Name)
}

func TestReferenceRepository_ListVillagesByDistrict(t *testing.T) {
	db := setupTestDB(t)
	repo := NewReferenceRepository(db)

	db.Create(&models.Province{ID: 1, ExternalID: "p1", Name: "Prov A"})
	db.Create(&models.City{ID: 1, ExternalID: "c1", ProvinceID: 1, Name: "City A"})
	db.Create(&models.District{ID: 1, ExternalID: "d1", CityID: 1, Name: "District A"})
	db.Create(&models.District{ID: 2, ExternalID: "d2", CityID: 1, Name: "District B"})
	db.Create(&models.Zip{ID: 1, ExternalID: "z1", Code: "12120", CityID: 1, DistrictID: 1})

	// Village with a zip.
	db.Create(&models.Village{ID: 1, ExternalID: "v1", DistrictID: 1, ZipID: intPtr(1), Name: "Kebayoran"})
	// Village without a zip - zip_id nullable per ARCHITECTURE.md (not every
	// kelurahan has one in the Odoo export).
	db.Create(&models.Village{ID: 2, ExternalID: "v2", DistrictID: 1, Name: "Senayan"})
	// Village in a different district - must not appear.
	db.Create(&models.Village{ID: 3, ExternalID: "v3", DistrictID: 2, Name: "Other District Village"})

	t.Run("filters by parent district", func(t *testing.T) {
		villages, err := repo.ListVillagesByDistrict(1, "", 50)
		assert.NoError(t, err)
		assert.Len(t, villages, 2)
	})

	t.Run("preloads zip when present", func(t *testing.T) {
		villages, err := repo.ListVillagesByDistrict(1, "", 50)
		assert.NoError(t, err)

		var withZip, withoutZip *models.Village
		for i := range villages {
			if villages[i].Name == "Kebayoran" {
				withZip = &villages[i]
			}
			if villages[i].Name == "Senayan" {
				withoutZip = &villages[i]
			}
		}

		if assert.NotNil(t, withZip) {
			if assert.NotNil(t, withZip.Zip) {
				assert.Equal(t, "12120", withZip.Zip.Code)
			}
		}
		if assert.NotNil(t, withoutZip) {
			assert.Nil(t, withoutZip.Zip)
		}
	})

	t.Run("q narrows by name", func(t *testing.T) {
		villages, err := repo.ListVillagesByDistrict(1, "keba", 50)
		assert.NoError(t, err)
		assert.Len(t, villages, 1)
		assert.Equal(t, "Kebayoran", villages[0].Name)
	})

	t.Run("limit caps result count", func(t *testing.T) {
		villages, err := repo.ListVillagesByDistrict(1, "", 1)
		assert.NoError(t, err)
		assert.Len(t, villages, 1)
	})
}

func intPtr(i int) *int { return &i }
