package handlers

import (
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/utils"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// newDashboardTestApp wires the dashboard route directly against the
// handler and stubs identity/role into Locals the same way
// JWTProtected/RequireRole would in production - see newTestLeadApp for the
// same idea.
func newDashboardTestApp(svc *MockDashboardService, userID uuid.UUID, role string) *fiber.App {
	app := newTestApp()
	app.Use(func(c *fiber.Ctx) error {
		utils.SetUserID(c, userID)
		c.Locals("role", role)
		return c.Next()
	})
	h := NewDashboardHandler(svc)
	app.Get("/dashboard/summary", h.Summary)
	return app
}

func TestDashboardSummary_DateOnlyOneProvided_Rejected(t *testing.T) {
	svc := new(MockDashboardService)
	app := newDashboardTestApp(svc, uuid.Must(uuid.NewV7()), "SALES")

	req := httptest.NewRequest("GET", "/dashboard/summary?date_from=2026-07-01", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	svc.AssertNotCalled(t, "Summary")
}

func TestDashboardSummary_NoDateParams_DefaultsAndCallsService(t *testing.T) {
	svc := new(MockDashboardService)
	userID := uuid.Must(uuid.NewV7())
	svc.On("Summary", userID, "SALES", mock.AnythingOfType("dto.DashboardQuery")).
		Return(&dto.DashboardSummaryResponse{}, nil)
	app := newDashboardTestApp(svc, userID, "SALES")

	req := httptest.NewRequest("GET", "/dashboard/summary", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	svc.AssertCalled(t, "Summary", userID, "SALES", mock.AnythingOfType("dto.DashboardQuery"))
}

func TestDashboardSummary_InvalidStatus_Rejected(t *testing.T) {
	svc := new(MockDashboardService)
	app := newDashboardTestApp(svc, uuid.Must(uuid.NewV7()), "SALES")

	req := httptest.NewRequest("GET", "/dashboard/summary?status=NOT_A_STATUS", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	svc.AssertNotCalled(t, "Summary")
}

func TestDashboardSummary_ValidStatus_ParsedAndCallsService(t *testing.T) {
	svc := new(MockDashboardService)
	userID := uuid.Must(uuid.NewV7())
	svc.On("Summary", userID, "SALES", mock.MatchedBy(func(q dto.DashboardQuery) bool {
		return q.Status != nil && *q.Status == "FOLLOW_UP"
	})).Return(&dto.DashboardSummaryResponse{}, nil)
	app := newDashboardTestApp(svc, userID, "SALES")

	req := httptest.NewRequest("GET", "/dashboard/summary?status=FOLLOW_UP", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestDashboardSummary_DateRangeExceeds90Days_Rejected(t *testing.T) {
	svc := new(MockDashboardService)
	app := newDashboardTestApp(svc, uuid.Must(uuid.NewV7()), "SALES")

	req := httptest.NewRequest("GET", "/dashboard/summary?date_from=2026-01-01&date_to=2026-12-31", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	svc.AssertNotCalled(t, "Summary")
}

func TestDashboardSummary_DateRangeExactly90Days_Accepted(t *testing.T) {
	svc := new(MockDashboardService)
	userID := uuid.Must(uuid.NewV7())
	svc.On("Summary", userID, "SALES", mock.AnythingOfType("dto.DashboardQuery")).
		Return(&dto.DashboardSummaryResponse{}, nil)
	app := newDashboardTestApp(svc, userID, "SALES")

	// 2026-01-01 .. 2026-03-31 inclusive = 31 + 28 + 31 = 90 days, the exact cap.
	req := httptest.NewRequest("GET", "/dashboard/summary?date_from=2026-01-01&date_to=2026-03-31", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestDashboardSummary_DateRangeOneDayOver90_Rejected(t *testing.T) {
	svc := new(MockDashboardService)
	app := newDashboardTestApp(svc, uuid.Must(uuid.NewV7()), "SALES")

	// 2026-01-01 .. 2026-04-01 inclusive = 91 days, one over the 90-day cap.
	req := httptest.NewRequest("GET", "/dashboard/summary?date_from=2026-01-01&date_to=2026-04-01", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	svc.AssertNotCalled(t, "Summary")
}
