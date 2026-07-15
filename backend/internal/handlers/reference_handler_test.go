package handlers

import (
	"fiber-api-boilerplate/internal/dto"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// newTestReferenceApp wires the reference routes directly against the
// handler, mirroring how the real router mounts them (JWTProtected runs
// before the handler in production; here we test the handler's own
// behavior in isolation).
func newTestReferenceApp(handler *ReferenceHandler) *fiber.App {
	app := newTestApp()

	app.Get("/refs/lead-sources", handler.ListLeadSources)
	app.Get("/refs/service-types", handler.ListServiceTypes)
	app.Get("/refs/provinces", handler.ListProvinces)
	app.Get("/refs/cities", handler.ListCities)
	app.Get("/refs/districts", handler.ListDistricts)
	app.Get("/refs/villages", handler.ListVillages)
	return app
}

func TestReferenceHandler_ListLeadSources_Success(t *testing.T) {
	mockSvc := new(MockReferenceService)
	handler := NewReferenceHandler(mockSvc)
	app := newTestReferenceApp(handler)

	mockSvc.On("ListLeadSources").Return([]dto.LeadSourceResponse{{Name: "Google"}}, nil)

	req := httptest.NewRequest("GET", "/refs/lead-sources", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestReferenceHandler_ListServiceTypes_Success(t *testing.T) {
	mockSvc := new(MockReferenceService)
	handler := NewReferenceHandler(mockSvc)
	app := newTestReferenceApp(handler)

	mockSvc.On("ListServiceTypes").Return([]dto.ServiceTypeResponse{{Name: "Dedicated"}}, nil)

	req := httptest.NewRequest("GET", "/refs/service-types", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestReferenceHandler_ListProvinces_Success(t *testing.T) {
	mockSvc := new(MockReferenceService)
	handler := NewReferenceHandler(mockSvc)
	app := newTestReferenceApp(handler)

	mockSvc.On("ListProvinces").Return([]dto.ProvinceResponse{{ID: 1, Name: "DKI Jakarta"}}, nil)

	req := httptest.NewRequest("GET", "/refs/provinces", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestReferenceHandler_ListCities_Success(t *testing.T) {
	mockSvc := new(MockReferenceService)
	handler := NewReferenceHandler(mockSvc)
	app := newTestReferenceApp(handler)

	mockSvc.On("ListCities", 1).Return([]dto.CityResponse{{ID: 1, Name: "Jakarta Selatan"}}, nil)

	req := httptest.NewRequest("GET", "/refs/cities?province_id=1", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestReferenceHandler_ListCities_MissingProvinceID_400(t *testing.T) {
	mockSvc := new(MockReferenceService)
	handler := NewReferenceHandler(mockSvc)
	app := newTestReferenceApp(handler)

	req := httptest.NewRequest("GET", "/refs/cities", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestReferenceHandler_ListCities_NonNumericProvinceID_400(t *testing.T) {
	mockSvc := new(MockReferenceService)
	handler := NewReferenceHandler(mockSvc)
	app := newTestReferenceApp(handler)

	req := httptest.NewRequest("GET", "/refs/cities?province_id=abc", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestReferenceHandler_ListDistricts_Success(t *testing.T) {
	mockSvc := new(MockReferenceService)
	handler := NewReferenceHandler(mockSvc)
	app := newTestReferenceApp(handler)

	mockSvc.On("ListDistricts", 1).Return([]dto.DistrictResponse{{ID: 1, Name: "Kebayoran Baru"}}, nil)

	req := httptest.NewRequest("GET", "/refs/districts?city_id=1", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestReferenceHandler_ListDistricts_MissingCityID_400(t *testing.T) {
	mockSvc := new(MockReferenceService)
	handler := NewReferenceHandler(mockSvc)
	app := newTestReferenceApp(handler)

	req := httptest.NewRequest("GET", "/refs/districts", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestReferenceHandler_ListVillages_Success(t *testing.T) {
	mockSvc := new(MockReferenceService)
	handler := NewReferenceHandler(mockSvc)
	app := newTestReferenceApp(handler)

	mockSvc.On("ListVillages", 1, "keb").Return(
		[]dto.VillageResponse{{ID: 1, Name: "Kebayoran", Zip: &dto.ZipResponse{ID: 1, Code: "12120"}}}, nil)

	req := httptest.NewRequest("GET", "/refs/villages?district_id=1&q=keb", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestReferenceHandler_ListVillages_MissingDistrictID_400(t *testing.T) {
	mockSvc := new(MockReferenceService)
	handler := NewReferenceHandler(mockSvc)
	app := newTestReferenceApp(handler)

	req := httptest.NewRequest("GET", "/refs/villages", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
