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

func newTestLeadSourceApp(handler *LeadSourceHandler) *fiber.App {
	app := newTestApp()

	app.Get("/lead-sources", handler.ListLeadSourcesAdmin)
	app.Post("/lead-sources", handler.CreateLeadSource)
	app.Patch("/lead-sources/:id", handler.UpdateLeadSource)
	return app
}

func TestLeadSourceHandler_CreateLeadSource_Success(t *testing.T) {
	mockSvc := new(MockLeadSourceService)
	handler := NewLeadSourceHandler(mockSvc)
	app := newTestLeadSourceApp(handler)

	input := dto.CreateLeadSourceInput{Name: "Website"}
	response := &dto.LeadSourceAdminResponse{
		ID: uuid.Must(uuid.NewV7()), Name: "Website", IsActive: true, CreatedAt: time.Now(),
	}
	mockSvc.On("Create", input).Return(response, nil)

	req := httptest.NewRequest("POST", "/lead-sources", strings.NewReader(`{"name":"Website"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

func TestLeadSourceHandler_ListLeadSourcesAdmin_Success(t *testing.T) {
	mockSvc := new(MockLeadSourceService)
	handler := NewLeadSourceHandler(mockSvc)
	app := newTestLeadSourceApp(handler)

	sources := []dto.LeadSourceAdminResponse{
		{ID: uuid.Must(uuid.NewV7()), Name: "A", IsActive: true},
		{ID: uuid.Must(uuid.NewV7()), Name: "B", IsActive: false},
	}
	mockSvc.On("List").Return(sources, nil)

	req := httptest.NewRequest("GET", "/lead-sources", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result utils.Response
	json.NewDecoder(resp.Body).Decode(&result)
	assert.True(t, result.Success)
}

func TestLeadSourceHandler_UpdateLeadSource_Success(t *testing.T) {
	mockSvc := new(MockLeadSourceService)
	handler := NewLeadSourceHandler(mockSvc)
	app := newTestLeadSourceApp(handler)

	id := uuid.Must(uuid.NewV7())
	newName := "New Name"
	input := dto.UpdateLeadSourceInput{Name: &newName}
	response := &dto.LeadSourceAdminResponse{ID: id, Name: "New Name", IsActive: true}

	mockSvc.On("Update", id, input).Return(response, nil)

	req := httptest.NewRequest("PATCH", "/lead-sources/"+id.String(), strings.NewReader(`{"name":"New Name"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestLeadSourceHandler_UpdateLeadSource_InvalidID(t *testing.T) {
	mockSvc := new(MockLeadSourceService)
	handler := NewLeadSourceHandler(mockSvc)
	app := newTestLeadSourceApp(handler)

	req := httptest.NewRequest("PATCH", "/lead-sources/not-a-uuid", strings.NewReader(`{"name":"X"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
