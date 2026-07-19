package handlers

import (
	"net/http/httptest"
	"testing"

	"fiber-api-boilerplate/internal/repository"
	"fiber-api-boilerplate/internal/services"
	"fiber-api-boilerplate/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

func newTestExportApp(handler *ExportHandler, userID uuid.UUID, role string) *fiber.App {
	app := newTestApp()
	app.Use(func(c *fiber.Ctx) error {
		utils.SetUserID(c, userID)
		c.Locals("role", role)
		return c.Next()
	})
	app.Get("/leads/export", handler.ExportLeads)
	return app
}

func TestExportHandler_ExportLeads_MissingPassword_401(t *testing.T) {
	mockSvc := new(MockExportService)
	handler := NewExportHandler(mockSvc)
	app := newTestExportApp(handler, uuid.Must(uuid.NewV7()), "SALES")

	req := httptest.NewRequest("GET", "/leads/export", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	mockSvc.AssertNotCalled(t, "Export")
}

func TestExportHandler_ExportLeads_WrongPassword_401(t *testing.T) {
	mockSvc := new(MockExportService)
	userID := uuid.Must(uuid.NewV7())
	handler := NewExportHandler(mockSvc)
	app := newTestExportApp(handler, userID, "SALES")

	mockSvc.On("Export", userID, "SALES", "wrong", repository.LeadFilter{}).
		Return(nil, services.ErrExportInvalidPassword)

	req := httptest.NewRequest("GET", "/leads/export", nil)
	req.Header.Set("X-Export-Password", "wrong")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestExportHandler_ExportLeads_TooManyRows_422(t *testing.T) {
	mockSvc := new(MockExportService)
	userID := uuid.Must(uuid.NewV7())
	handler := NewExportHandler(mockSvc)
	app := newTestExportApp(handler, userID, "ADMIN_SALES")

	mockSvc.On("Export", userID, "ADMIN_SALES", "correct", repository.LeadFilter{}).
		Return(nil, &services.ErrExportTooManyRows{Count: 60000})

	req := httptest.NewRequest("GET", "/leads/export", nil)
	req.Header.Set("X-Export-Password", "correct")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)
}

func TestExportHandler_ExportLeads_Success_ReturnsXlsxContentType(t *testing.T) {
	mockSvc := new(MockExportService)
	userID := uuid.Must(uuid.NewV7())
	handler := NewExportHandler(mockSvc)
	app := newTestExportApp(handler, userID, "ADMIN_SALES")

	mockSvc.On("Export", userID, "ADMIN_SALES", "correct", repository.LeadFilter{}).
		Return(excelize.NewFile(), nil)

	req := httptest.NewRequest("GET", "/leads/export", nil)
	req.Header.Set("X-Export-Password", "correct")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", resp.Header.Get("Content-Type"))
	assert.Contains(t, resp.Header.Get("Content-Disposition"), "leads-export-")
}

func TestExportHandler_ExportLeads_QueryParsedIntoFilter(t *testing.T) {
	mockSvc := new(MockExportService)
	userID := uuid.Must(uuid.NewV7())
	handler := NewExportHandler(mockSvc)
	app := newTestExportApp(handler, userID, "ADMIN_SALES")

	mockSvc.On("Export", userID, "ADMIN_SALES", "correct", repository.LeadFilter{Status: "FOLLOW_UP", Q: "acme"}).
		Return(excelize.NewFile(), nil)

	req := httptest.NewRequest("GET", "/leads/export?status=FOLLOW_UP&q=acme", nil)
	req.Header.Set("X-Export-Password", "correct")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}
