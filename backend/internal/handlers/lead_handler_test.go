package handlers

import (
	"encoding/json"
	"errors"
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

// newTestLeadApp wires the lead routes directly against the handler and
// stubs identity/role into Locals the same way JWTProtected/RequireRole
// would in production - see rbac_test.go's newRoleApp for the same idea.
func newTestLeadApp(handler *LeadHandler, userID uuid.UUID, role string) *fiber.App {
	app := newTestApp()
	app.Use(func(c *fiber.Ctx) error {
		utils.SetUserID(c, userID)
		c.Locals("role", role)
		return c.Next()
	})

	app.Post("/leads", handler.CreateLead)
	app.Get("/leads", handler.ListLeads)
	app.Get("/leads/:code", handler.GetLead)
	app.Patch("/leads/:code", handler.UpdateLead)
	app.Patch("/leads/:code/status", handler.UpdateLeadStatus)
	return app
}

func TestLeadHandler_CreateLead_Success(t *testing.T) {
	mockSvc := new(MockLeadService)
	userID := uuid.Must(uuid.NewV7())
	handler := NewLeadHandler(mockSvc)
	app := newTestLeadApp(handler, userID, "SALES")

	leadResponse := &dto.LeadResponse{ID: uuid.Must(uuid.NewV7()), Code: "LD-2607-0001", CompanyName: "Acme"}
	mockSvc.On("Create", userID, "SALES", dto.CreateLeadInput{CompanyName: "Acme"}).Return(leadResponse, nil)

	body := `{"company_name":"Acme"}`
	req := httptest.NewRequest("POST", "/leads", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

func TestLeadHandler_CreateLead_ValidationError(t *testing.T) {
	mockSvc := new(MockLeadService)
	handler := NewLeadHandler(mockSvc)
	app := newTestLeadApp(handler, uuid.Must(uuid.NewV7()), "SALES")

	body := `{}`
	req := httptest.NewRequest("POST", "/leads", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestLeadHandler_CreateLead_OwnerNotTeamMember_400(t *testing.T) {
	mockSvc := new(MockLeadService)
	userID := uuid.Must(uuid.NewV7())
	handler := NewLeadHandler(mockSvc)
	app := newTestLeadApp(handler, userID, "LEADER")

	mockSvc.On("Create", userID, "LEADER", dto.CreateLeadInput{CompanyName: "Acme"}).
		Return(nil, services.ErrOwnerNotTeamMember)

	body := `{"company_name":"Acme"}`
	req := httptest.NewRequest("POST", "/leads", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestLeadHandler_GetLead_Success(t *testing.T) {
	mockSvc := new(MockLeadService)
	userID := uuid.Must(uuid.NewV7())
	handler := NewLeadHandler(mockSvc)
	app := newTestLeadApp(handler, userID, "ADMIN_SALES")

	leadResponse := &dto.LeadResponse{Code: "LD-2607-0002", CompanyName: "Acme"}
	mockSvc.On("FindByCode", userID, "ADMIN_SALES", "LD-2607-0002").Return(leadResponse, nil)

	req := httptest.NewRequest("GET", "/leads/LD-2607-0002", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestLeadHandler_GetLead_OutOfScope_404(t *testing.T) {
	mockSvc := new(MockLeadService)
	userID := uuid.Must(uuid.NewV7())
	handler := NewLeadHandler(mockSvc)
	app := newTestLeadApp(handler, userID, "SALES")

	mockSvc.On("FindByCode", userID, "SALES", "LD-2607-0003").Return(nil, services.ErrLeadNotFound)

	req := httptest.NewRequest("GET", "/leads/LD-2607-0003", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestLeadHandler_ListLeads_Success(t *testing.T) {
	mockSvc := new(MockLeadService)
	userID := uuid.Must(uuid.NewV7())
	handler := NewLeadHandler(mockSvc)
	app := newTestLeadApp(handler, userID, "ADMIN_SALES")

	paginated := &dto.PaginatedLeadResponse{Items: []dto.LeadResponse{}, Total: 0, Page: 1, Limit: 20}
	mockSvc.On("List", userID, "ADMIN_SALES", dto.LeadListQuery{Page: 1, Limit: 20}).Return(paginated, nil)

	req := httptest.NewRequest("GET", "/leads", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	var result utils.Response
	json.NewDecoder(resp.Body).Decode(&result)
	assert.True(t, result.Success)
}

func TestLeadHandler_ListLeads_InvalidProvinceID_400(t *testing.T) {
	mockSvc := new(MockLeadService)
	handler := NewLeadHandler(mockSvc)
	app := newTestLeadApp(handler, uuid.Must(uuid.NewV7()), "ADMIN_SALES")

	req := httptest.NewRequest("GET", "/leads?province_id=abc", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestLeadHandler_ListLeads_LimitClampedNotRejected(t *testing.T) {
	mockSvc := new(MockLeadService)
	userID := uuid.Must(uuid.NewV7())
	handler := NewLeadHandler(mockSvc)
	app := newTestLeadApp(handler, userID, "ADMIN_SALES")

	paginated := &dto.PaginatedLeadResponse{Items: []dto.LeadResponse{}, Total: 0, Page: 1, Limit: 100}
	mockSvc.On("List", userID, "ADMIN_SALES", dto.LeadListQuery{Page: 1, Limit: 100}).Return(paginated, nil)

	req := httptest.NewRequest("GET", "/leads?limit=9999", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode, "oversized limit clamps to 100 rather than erroring")
}

func TestLeadHandler_UpdateLeadStatus_InvalidTransition_422(t *testing.T) {
	mockSvc := new(MockLeadService)
	userID := uuid.Must(uuid.NewV7())
	handler := NewLeadHandler(mockSvc)
	app := newTestLeadApp(handler, userID, "ADMIN_SALES")

	mockSvc.On("UpdateStatus", userID, "ADMIN_SALES", "LD-2607-0004", dto.UpdateLeadStatusInput{Status: "FOLLOW_UP"}).
		Return(nil, services.ErrInvalidStatusTransition)

	body := `{"status":"FOLLOW_UP"}`
	req := httptest.NewRequest("PATCH", "/leads/LD-2607-0004/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)
}

func TestLeadHandler_UpdateLeadStatus_Success(t *testing.T) {
	mockSvc := new(MockLeadService)
	userID := uuid.Must(uuid.NewV7())
	handler := NewLeadHandler(mockSvc)
	app := newTestLeadApp(handler, userID, "ADMIN_SALES")

	leadResponse := &dto.LeadResponse{Code: "LD-2607-0005", Status: "HANDOFF_ODOO"}
	mockSvc.On("UpdateStatus", userID, "ADMIN_SALES", "LD-2607-0005", dto.UpdateLeadStatusInput{Status: "HANDOFF_ODOO"}).
		Return(leadResponse, nil)

	body := `{"status":"HANDOFF_ODOO"}`
	req := httptest.NewRequest("PATCH", "/leads/LD-2607-0005/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestLeadHandler_UpdateLead_OtherServiceError_400(t *testing.T) {
	mockSvc := new(MockLeadService)
	userID := uuid.Must(uuid.NewV7())
	handler := NewLeadHandler(mockSvc)
	app := newTestLeadApp(handler, userID, "ADMIN_SALES")

	mockSvc.On("Update", userID, "ADMIN_SALES", "LD-2607-0006", dto.UpdateLeadInput{}).
		Return(nil, errors.New("invalid reference id"))

	body := `{}`
	req := httptest.NewRequest("PATCH", "/leads/LD-2607-0006", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
