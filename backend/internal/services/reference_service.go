package services

import (
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/models"
)

// referenceRepository is the subset of repository methods ReferenceService
// needs. Defined consumer-side for testability (satisfied by the real repo
// or a mock).
type referenceRepository interface {
	ListActiveLeadSources() ([]models.LeadSource, error)
	ListActiveServiceTypes() ([]models.ServiceType, error)
	ListProvinces() ([]models.Province, error)
	ListCitiesByProvince(provinceID int) ([]models.City, error)
	ListDistrictsByCity(cityID int) ([]models.District, error)
	ListVillagesByDistrict(districtID int, q string, limit int) ([]models.Village, error)
}

// villageSearchLimit caps how many villages a single /refs/villages call can
// return. Without a cap, a broad or empty q against a large district would
// dump thousands of rows into a dropdown.
const villageSearchLimit = 50

type ReferenceService struct {
	refRepo referenceRepository
}

func NewReferenceService(refRepo referenceRepository) *ReferenceService {
	return &ReferenceService{refRepo: refRepo}
}

func (s *ReferenceService) ListLeadSources() ([]dto.LeadSourceResponse, error) {
	sources, err := s.refRepo.ListActiveLeadSources()
	if err != nil {
		return nil, err
	}
	return dto.ToLeadSourceResponseList(sources), nil
}

func (s *ReferenceService) ListServiceTypes() ([]dto.ServiceTypeResponse, error) {
	types, err := s.refRepo.ListActiveServiceTypes()
	if err != nil {
		return nil, err
	}
	return dto.ToServiceTypeResponseList(types), nil
}

func (s *ReferenceService) ListProvinces() ([]dto.ProvinceResponse, error) {
	provinces, err := s.refRepo.ListProvinces()
	if err != nil {
		return nil, err
	}
	return dto.ToProvinceResponseList(provinces), nil
}

func (s *ReferenceService) ListCities(provinceID int) ([]dto.CityResponse, error) {
	cities, err := s.refRepo.ListCitiesByProvince(provinceID)
	if err != nil {
		return nil, err
	}
	return dto.ToCityResponseList(cities), nil
}

func (s *ReferenceService) ListDistricts(cityID int) ([]dto.DistrictResponse, error) {
	districts, err := s.refRepo.ListDistrictsByCity(cityID)
	if err != nil {
		return nil, err
	}
	return dto.ToDistrictResponseList(districts), nil
}

func (s *ReferenceService) ListVillages(districtID int, q string) ([]dto.VillageResponse, error) {
	villages, err := s.refRepo.ListVillagesByDistrict(districtID, q, villageSearchLimit)
	if err != nil {
		return nil, err
	}
	return dto.ToVillageResponseList(villages), nil
}
