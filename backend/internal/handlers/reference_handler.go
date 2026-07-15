package handlers

import (
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// referenceService is the business-logic contract ReferenceHandler depends
// on. Defined consumer-side so it can be satisfied by the real service or a
// mock.
type referenceService interface {
	ListLeadSources() ([]dto.LeadSourceResponse, error)
	ListServiceTypes() ([]dto.ServiceTypeResponse, error)
	ListProvinces() ([]dto.ProvinceResponse, error)
	ListCities(provinceID int) ([]dto.CityResponse, error)
	ListDistricts(cityID int) ([]dto.DistrictResponse, error)
	ListVillages(districtID int, q string) ([]dto.VillageResponse, error)
}

type ReferenceHandler struct {
	referenceService referenceService
}

func NewReferenceHandler(referenceService referenceService) *ReferenceHandler {
	return &ReferenceHandler{referenceService: referenceService}
}

// ListLeadSources godoc
// @Summary List active lead sources
// @Tags refs
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=[]dto.LeadSourceResponse}
// @Router /refs/lead-sources [get]
func (h *ReferenceHandler) ListLeadSources(c *fiber.Ctx) error {
	sources, err := h.referenceService.ListLeadSources()
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Lead sources retrieved successfully", sources)
}

// ListServiceTypes godoc
// @Summary List active service types
// @Tags refs
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=[]dto.ServiceTypeResponse}
// @Router /refs/service-types [get]
func (h *ReferenceHandler) ListServiceTypes(c *fiber.Ctx) error {
	types, err := h.referenceService.ListServiceTypes()
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Service types retrieved successfully", types)
}

// ListProvinces godoc
// @Summary List provinces
// @Tags refs
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=[]dto.ProvinceResponse}
// @Router /refs/provinces [get]
func (h *ReferenceHandler) ListProvinces(c *fiber.Ctx) error {
	provinces, err := h.referenceService.ListProvinces()
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Provinces retrieved successfully", provinces)
}

// ListCities godoc
// @Summary List cities by province
// @Tags refs
// @Security BearerAuth
// @Param province_id query int true "Province ID"
// @Success 200 {object} utils.Response{data=[]dto.CityResponse}
// @Router /refs/cities [get]
func (h *ReferenceHandler) ListCities(c *fiber.Ctx) error {
	provinceID, err := strconv.Atoi(c.Query("province_id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "province_id is required")
	}

	cities, err := h.referenceService.ListCities(provinceID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Cities retrieved successfully", cities)
}

// ListDistricts godoc
// @Summary List districts by city
// @Tags refs
// @Security BearerAuth
// @Param city_id query int true "City ID"
// @Success 200 {object} utils.Response{data=[]dto.DistrictResponse}
// @Router /refs/districts [get]
func (h *ReferenceHandler) ListDistricts(c *fiber.Ctx) error {
	cityID, err := strconv.Atoi(c.Query("city_id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "city_id is required")
	}

	districts, err := h.referenceService.ListDistricts(cityID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Districts retrieved successfully", districts)
}

// ListVillages godoc
// @Summary List villages by district
// @Tags refs
// @Security BearerAuth
// @Param district_id query int true "District ID"
// @Param q query string false "Name search"
// @Success 200 {object} utils.Response{data=[]dto.VillageResponse}
// @Router /refs/villages [get]
func (h *ReferenceHandler) ListVillages(c *fiber.Ctx) error {
	districtID, err := strconv.Atoi(c.Query("district_id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "district_id is required")
	}

	villages, err := h.referenceService.ListVillages(districtID, c.Query("q"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Villages retrieved successfully", villages)
}
