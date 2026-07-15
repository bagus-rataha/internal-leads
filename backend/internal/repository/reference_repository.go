package repository

import (
	"fiber-api-boilerplate/internal/models"

	"gorm.io/gorm"
)

// ReferenceRepository handles read-only database access for lead sources,
// service types, and wilayah (region) reference data.
type ReferenceRepository struct {
	db *gorm.DB
}

// NewReferenceRepository creates new reference repository
func NewReferenceRepository(db *gorm.DB) *ReferenceRepository {
	return &ReferenceRepository{db: db}
}

// ListActiveLeadSources returns lead sources available for new leads.
func (r *ReferenceRepository) ListActiveLeadSources() ([]models.LeadSource, error) {
	var sources []models.LeadSource
	err := r.db.Where("is_active = true").Order("name").Find(&sources).Error
	return sources, err
}

// ListActiveServiceTypes returns service types available for new leads.
func (r *ReferenceRepository) ListActiveServiceTypes() ([]models.ServiceType, error) {
	var types []models.ServiceType
	err := r.db.Where("is_active = true").Order("name").Find(&types).Error
	return types, err
}

// ListProvinces returns all provinces.
func (r *ReferenceRepository) ListProvinces() ([]models.Province, error) {
	var provinces []models.Province
	err := r.db.Order("name").Find(&provinces).Error
	return provinces, err
}

// ListCitiesByProvince returns cities belonging to the given province.
func (r *ReferenceRepository) ListCitiesByProvince(provinceID int) ([]models.City, error) {
	var cities []models.City
	err := r.db.Where("province_id = ?", provinceID).Order("name").Find(&cities).Error
	return cities, err
}

// ListDistrictsByCity returns districts belonging to the given city.
func (r *ReferenceRepository) ListDistrictsByCity(cityID int) ([]models.District, error) {
	var districts []models.District
	err := r.db.Where("city_id = ?", cityID).Order("name").Find(&districts).Error
	return districts, err
}

// ListVillagesByDistrict returns villages belonging to the given district,
// optionally narrowed by a name search, capped at limit rows. Zip is
// preloaded so the caller gets the postal code without a second query.
func (r *ReferenceRepository) ListVillagesByDistrict(districtID int, q string, limit int) ([]models.Village, error) {
	var villages []models.Village
	tx := r.db.Preload("Zip").Where("district_id = ?", districtID)
	if q != "" {
		tx = tx.Where("name ILIKE ?", "%"+q+"%")
	}
	err := tx.Order("name").Limit(limit).Find(&villages).Error
	return villages, err
}
