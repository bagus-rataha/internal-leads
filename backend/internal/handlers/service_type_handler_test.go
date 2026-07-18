package handlers

import (
	"encoding/json"
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/utils"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func newTestServiceTypeApp(handler *ServiceTypeHandler) *fiber.App {
	app := newTestApp()

	app.Get("/service-types", handler.ListServiceTypesAdmin)
	app.Post("/service-types", handler.CreateServiceType)
	app.Patch("/service-types/:id", handler.UpdateServiceType)
	return app
}

func TestServiceTypeHandler_CreateServiceType_Success(t *testing.T) {
	mockSvc := new(MockServiceTypeService)
	handler := NewServiceTypeHandler(mockSvc)
	app := newTestServiceTypeApp(handler)

	input := dto.CreateServiceTypeInput{Name: "Dedicated"}
	response := &dto.ServiceTypeAdminResponse{
		ID: uuid.Must(uuid.NewV7()), Name: "Dedicated", IsActive: true, CreatedAt: time.Now(),
	}
	mockSvc.On("Create", input).Return(response, nil)

	req := httptest.NewRequest("POST", "/service-types", strings.NewReader(`{"name":"Dedicated"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

func TestServiceTypeHandler_ListServiceTypesAdmin_Success(t *testing.T) {
	mockSvc := new(MockServiceTypeService)
	handler := NewServiceTypeHandler(mockSvc)
	app := newTestServiceTypeApp(handler)

	types := []dto.ServiceTypeAdminResponse{
		{ID: uuid.Must(uuid.NewV7()), Name: "Dedicated", IsActive: true},
		{ID: uuid.Must(uuid.NewV7()), Name: "Broadband", IsActive: false},
	}
	mockSvc.On("List").Return(types, nil)

	req := httptest.NewRequest("GET", "/service-types", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result utils.Response
	json.NewDecoder(resp.Body).Decode(&result)
	assert.True(t, result.Success)
}

func TestServiceTypeHandler_UpdateServiceType_Success(t *testing.T) {
	mockSvc := new(MockServiceTypeService)
	handler := NewServiceTypeHandler(mockSvc)
	app := newTestServiceTypeApp(handler)

	id := uuid.Must(uuid.NewV7())
	newName := "New Name"
	input := dto.UpdateServiceTypeInput{Name: &newName}
	response := &dto.ServiceTypeAdminResponse{ID: id, Name: "New Name", IsActive: true}

	mockSvc.On("Update", id, input).Return(response, nil)

	req := httptest.NewRequest("PATCH", "/service-types/"+id.String(), strings.NewReader(`{"name":"New Name"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestServiceTypeHandler_UpdateServiceType_InvalidID(t *testing.T) {
	mockSvc := new(MockServiceTypeService)
	handler := NewServiceTypeHandler(mockSvc)
	app := newTestServiceTypeApp(handler)

	req := httptest.NewRequest("PATCH", "/service-types/not-a-uuid", strings.NewReader(`{"name":"X"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
