package handlers

import (
	"errors"
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// dashboardService is the business-logic contract DashboardHandler depends
// on. Defined consumer-side so it can be satisfied by the real service or a
// mock. Grown one method per dashboard endpoint.
type dashboardService interface {
	Summary(callerID uuid.UUID, role string, q dto.DashboardQuery) (*dto.DashboardSummaryResponse, error)
	Activity(callerID uuid.UUID, role string, q dto.DashboardQuery) (*dto.DashboardActivityResponse, error)
	StaleLeads(callerID uuid.UUID, role string, q dto.DashboardQuery) (*dto.DashboardStaleLeadsResponse, error)
	SalesActivity(callerID uuid.UUID, role string, q dto.DashboardQuery) (*dto.DashboardSalesActivityResponse, error)
	Segments(callerID uuid.UUID, role string, q dto.DashboardQuery) (*dto.DashboardSegmentsResponse, error)
}

type DashboardHandler struct {
	dashboardService dashboardService
}

func NewDashboardHandler(dashboardService dashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService}
}

// jakartaLoc is this app's single fixed business timezone (Indonesia-only
// ISP sales tool, no multi-tenancy). Loaded once at package init; falls back
// to UTC if the container has no tzdata rather than panicking — a wrong
// "today" for the default dashboard range is a degraded convenience, not a
// security issue.
var jakartaLoc = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.UTC
	}
	return loc
}()

// parseDashboardQuery builds dto.DashboardQuery from raw query strings,
// shared by every dashboard handler method. date_from/date_to must be
// provided together or not at all (a lone one is ambiguous, rejected rather
// than guessed); omitting both defaults to a trailing 30-day window per
// ARCHITECTURE.md §9.
func parseDashboardQuery(c *fiber.Ctx) (dto.DashboardQuery, error) {
	var q dto.DashboardQuery

	fromStr, toStr := c.Query("date_from"), c.Query("date_to")
	switch {
	case fromStr == "" && toStr == "":
		now := time.Now().In(jakartaLoc)
		y, m, d := now.Date()
		q.DateTo = time.Date(y, m, d, 0, 0, 0, 0, jakartaLoc)
		q.DateFrom = q.DateTo.AddDate(0, 0, -29)
	case fromStr == "" || toStr == "":
		return q, errors.New("date_from and date_to must be provided together")
	default:
		from, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			return q, errors.New("date_from must be YYYY-MM-DD")
		}
		to, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			return q, errors.New("date_to must be YYYY-MM-DD")
		}
		if to.Before(from) {
			return q, errors.New("date_to must not be before date_from")
		}
		q.DateFrom, q.DateTo = from, to
	}

	if v := c.Query("team_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return q, errors.New("team_id must be a valid UUID")
		}
		q.TeamID = &id
	}
	if v := c.Query("owner_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return q, errors.New("owner_id must be a valid UUID")
		}
		q.OwnerID = &id
	}
	if v := c.Query("status"); v != "" {
		switch v {
		case "BARU", "FOLLOW_UP", "HANDOFF_ODOO", "LOST":
			q.Status = &v
		default:
			return q, errors.New("status must be one of BARU, FOLLOW_UP, HANDOFF_ODOO, LOST")
		}
	}

	return q, nil
}

// Summary godoc
// @Summary Dashboard summary (metric cards + funnel)
// @Tags dashboard
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=dto.DashboardSummaryResponse}
// @Router /dashboard/summary [get]
func (h *DashboardHandler) Summary(c *fiber.Ctx) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid session")
	}
	role, _ := c.Locals("role").(string)

	q, err := parseDashboardQuery(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	result, err := h.dashboardService.Summary(userID, role, q)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Dashboard summary retrieved successfully", result)
}

// Activity godoc
// @Summary Dashboard activity trend (lead baru vs follow-up per day)
// @Tags dashboard
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=dto.DashboardActivityResponse}
// @Router /dashboard/activity [get]
func (h *DashboardHandler) Activity(c *fiber.Ctx) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid session")
	}
	role, _ := c.Locals("role").(string)

	q, err := parseDashboardQuery(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	result, err := h.dashboardService.Activity(userID, role, q)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Dashboard activity retrieved successfully", result)
}

// StaleLeads godoc
// @Summary Dashboard stale (terlantar) leads
// @Tags dashboard
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=dto.DashboardStaleLeadsResponse}
// @Router /dashboard/stale-leads [get]
func (h *DashboardHandler) StaleLeads(c *fiber.Ctx) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid session")
	}
	role, _ := c.Locals("role").(string)

	q, err := parseDashboardQuery(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	result, err := h.dashboardService.StaleLeads(userID, role, q)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Stale leads retrieved successfully", result)
}

// SalesActivity godoc
// @Summary Dashboard sales activity ("Keaktifan Sales") - LEADER/ADMIN_SALES/SU only
// @Tags dashboard
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=dto.DashboardSalesActivityResponse}
// @Router /dashboard/sales-activity [get]
func (h *DashboardHandler) SalesActivity(c *fiber.Ctx) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid session")
	}
	role, _ := c.Locals("role").(string)

	q, err := parseDashboardQuery(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	result, err := h.dashboardService.SalesActivity(userID, role, q)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Sales activity retrieved successfully", result)
}

// Segments godoc
// @Summary Dashboard segments (sumber, wilayah, kompetitor, bidang usaha)
// @Tags dashboard
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=dto.DashboardSegmentsResponse}
// @Router /dashboard/segments [get]
func (h *DashboardHandler) Segments(c *fiber.Ctx) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid session")
	}
	role, _ := c.Locals("role").(string)

	q, err := parseDashboardQuery(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	result, err := h.dashboardService.Segments(userID, role, q)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Dashboard segments retrieved successfully", result)
}
