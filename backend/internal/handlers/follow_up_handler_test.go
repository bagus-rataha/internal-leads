package handlers

import (
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/services"
	"fiber-api-boilerplate/internal/utils"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func newTestFollowUpApp(handler *FollowUpHandler, userID uuid.UUID, role string) *fiber.App {
	app := newTestApp()
	app.Use(func(c *fiber.Ctx) error {
		utils.SetUserID(c, userID)
		c.Locals("role", role)
		return c.Next()
	})

	app.Post("/leads/:code/followups", handler.CreateFollowUp)
	app.Get("/leads/:code/followups", handler.ListFollowUps)
	return app
}

func TestFollowUpHandler_CreateFollowUp_Success(t *testing.T) {
	mockSvc := new(MockFollowUpService)
	userID := uuid.Must(uuid.NewV7())
	handler := NewFollowUpHandler(mockSvc)
	app := newTestFollowUpApp(handler, userID, "SALES")

	response := &dto.FollowUpResponse{ID: uuid.Must(uuid.NewV7()), Note: "called today"}
	mockSvc.On("Create", userID, "SALES", "LD-2607-0001", dto.CreateFollowUpInput{Note: "called today"}).
		Return(response, nil)

	body := `{"note":"called today"}`
	req := httptest.NewRequest("POST", "/leads/LD-2607-0001/followups", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

func TestFollowUpHandler_CreateFollowUp_ValidationError(t *testing.T) {
	mockSvc := new(MockFollowUpService)
	handler := NewFollowUpHandler(mockSvc)
	app := newTestFollowUpApp(handler, uuid.Must(uuid.NewV7()), "SALES")

	body := `{"note":""}`
	req := httptest.NewRequest("POST", "/leads/LD-2607-0001/followups", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestFollowUpHandler_CreateFollowUp_LeadOutOfScope_404(t *testing.T) {
	mockSvc := new(MockFollowUpService)
	userID := uuid.Must(uuid.NewV7())
	handler := NewFollowUpHandler(mockSvc)
	app := newTestFollowUpApp(handler, userID, "SALES")

	mockSvc.On("Create", userID, "SALES", "LD-2607-0002", dto.CreateFollowUpInput{Note: "x"}).
		Return(nil, services.ErrLeadNotFound)

	body := `{"note":"x"}`
	req := httptest.NewRequest("POST", "/leads/LD-2607-0002/followups", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestFollowUpHandler_ListFollowUps_Success(t *testing.T) {
	mockSvc := new(MockFollowUpService)
	userID := uuid.Must(uuid.NewV7())
	handler := NewFollowUpHandler(mockSvc)
	app := newTestFollowUpApp(handler, userID, "ADMIN_SALES")

	mockSvc.On("List", userID, "ADMIN_SALES", "LD-2607-0003").Return([]dto.FollowUpResponse{}, nil)

	req := httptest.NewRequest("GET", "/leads/LD-2607-0003/followups", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestFollowUpHandler_ListFollowUps_LeadOutOfScope_404(t *testing.T) {
	mockSvc := new(MockFollowUpService)
	userID := uuid.Must(uuid.NewV7())
	handler := NewFollowUpHandler(mockSvc)
	app := newTestFollowUpApp(handler, userID, "SALES")

	mockSvc.On("List", userID, "SALES", "LD-2607-0004").Return(nil, services.ErrLeadNotFound)

	req := httptest.NewRequest("GET", "/leads/LD-2607-0004/followups", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}
