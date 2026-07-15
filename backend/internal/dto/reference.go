package dto

import (
	"fiber-api-boilerplate/internal/models"

	"github.com/google/uuid"
)

// LeadSourceResponse for lead source reference API responses
type LeadSourceResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// ToLeadSourceResponse converts model to DTO
func ToLeadSourceResponse(s *models.LeadSource) LeadSourceResponse {
	return LeadSourceResponse{ID: s.ID, Name: s.Name}
}

// ToLeadSourceResponseList converts models to DTOs
func ToLeadSourceResponseList(sources []models.LeadSource) []LeadSourceResponse {
	responses := make([]LeadSourceResponse, len(sources))
	for i, s := range sources {
		responses[i] = ToLeadSourceResponse(&s)
	}
	return responses
}

// ServiceTypeResponse for service type reference API responses
type ServiceTypeResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// ToServiceTypeResponse converts model to DTO
func ToServiceTypeResponse(s *models.ServiceType) ServiceTypeResponse {
	return ServiceTypeResponse{ID: s.ID, Name: s.Name}
}

// ToServiceTypeResponseList converts models to DTOs
func ToServiceTypeResponseList(types []models.ServiceType) []ServiceTypeResponse {
	responses := make([]ServiceTypeResponse, len(types))
	for i, t := range types {
		responses[i] = ToServiceTypeResponse(&t)
	}
	return responses
}

// ProvinceResponse for province reference API responses
type ProvinceResponse struct {
	ID   int     `json:"id"`
	Code *string `json:"code"`
	Name string  `json:"name"`
}

// ToProvinceResponse converts model to DTO
func ToProvinceResponse(p *models.Province) ProvinceResponse {
	return ProvinceResponse{ID: p.ID, Code: p.Code, Name: p.Name}
}

// ToProvinceResponseList converts models to DTOs
func ToProvinceResponseList(provinces []models.Province) []ProvinceResponse {
	responses := make([]ProvinceResponse, len(provinces))
	for i, p := range provinces {
		responses[i] = ToProvinceResponse(&p)
	}
	return responses
}

// CityResponse for city reference API responses
type CityResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ToCityResponse converts model to DTO
func ToCityResponse(c *models.City) CityResponse {
	return CityResponse{ID: c.ID, Name: c.Name}
}

// ToCityResponseList converts models to DTOs
func ToCityResponseList(cities []models.City) []CityResponse {
	responses := make([]CityResponse, len(cities))
	for i, c := range cities {
		responses[i] = ToCityResponse(&c)
	}
	return responses
}

// DistrictResponse for district reference API responses
type DistrictResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ToDistrictResponse converts model to DTO
func ToDistrictResponse(d *models.District) DistrictResponse {
	return DistrictResponse{ID: d.ID, Name: d.Name}
}

// ToDistrictResponseList converts models to DTOs
func ToDistrictResponseList(districts []models.District) []DistrictResponse {
	responses := make([]DistrictResponse, len(districts))
	for i, d := range districts {
		responses[i] = ToDistrictResponse(&d)
	}
	return responses
}

// ZipResponse for zip reference API responses. Only ever appears embedded
// inside VillageResponse — zip has no standalone list endpoint (see
// VillageResponse doc).
type ZipResponse struct {
	ID   int    `json:"id"`
	Code string `json:"code"`
}

// VillageResponse for village reference API responses. Embeds Zip (nil if
// the village has none) so the create-lead form gets the postal code
// without a second request — a village maps to at most one zip, so there's
// nothing for the caller to choose between.
type VillageResponse struct {
	ID   int          `json:"id"`
	Name string       `json:"name"`
	Zip  *ZipResponse `json:"zip"`
}

// ToVillageResponse converts model to DTO
func ToVillageResponse(v *models.Village) VillageResponse {
	resp := VillageResponse{ID: v.ID, Name: v.Name}
	if v.Zip != nil {
		resp.Zip = &ZipResponse{ID: v.Zip.ID, Code: v.Zip.Code}
	}
	return resp
}

// ToVillageResponseList converts models to DTOs
func ToVillageResponseList(villages []models.Village) []VillageResponse {
	responses := make([]VillageResponse, len(villages))
	for i, v := range villages {
		responses[i] = ToVillageResponse(&v)
	}
	return responses
}
