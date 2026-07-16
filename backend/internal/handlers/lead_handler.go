package handlers

import (
	"errors"
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/services"
	"fiber-api-boilerplate/internal/utils"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// leadService is the business-logic contract LeadHandler depends on.
// Defined consumer-side so it can be satisfied by the real service or a
// mock.
type leadService interface {
	Create(callerID uuid.UUID, role string, input dto.CreateLeadInput) (*dto.LeadResponse, error)
	List(callerID uuid.UUID, role string, query dto.LeadListQuery) (*dto.PaginatedLeadResponse, error)
	FindByCode(callerID uuid.UUID, role, code string) (*dto.LeadResponse, error)
	Update(callerID uuid.UUID, role, code string, input dto.UpdateLeadInput) (*dto.LeadResponse, error)
	UpdateStatus(callerID uuid.UUID, role, code string, input dto.UpdateLeadStatusInput) (*dto.LeadResponse, error)
}

type LeadHandler struct {
	leadService leadService
}

func NewLeadHandler(leadService leadService) *LeadHandler {
	return &LeadHandler{leadService: leadService}
}

// leadServiceErrorStatus maps a service error to an HTTP status. Falls back
// to 400 for anything unrecognized (e.g. an FK constraint violation
// translated to a plain error by the service).
func leadServiceErrorStatus(err error) int {
	switch {
	case errors.Is(err, services.ErrLeadNotFound):
		return fiber.StatusNotFound
	case errors.Is(err, services.ErrInvalidStatusTransition):
		return fiber.StatusUnprocessableEntity
	default:
		return fiber.StatusBadRequest
	}
}

// CreateLead godoc
// @Summary Create lead
// @Tags leads
// @Security BearerAuth
// @Param request body dto.CreateLeadInput true "Create lead"
// @Success 201 {object} utils.Response{data=dto.LeadResponse}
// @Router /leads [post]
func (h *LeadHandler) CreateLead(c *fiber.Ctx) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid session")
	}
	role, _ := c.Locals("role").(string)

	var input dto.CreateLeadInput
	if err := utils.ParseAndValidate(c, &input); err != nil {
		return err
	}

	lead, err := h.leadService.Create(userID, role, input)
	if err != nil {
		return utils.ErrorResponse(c, leadServiceErrorStatus(err), err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Lead created successfully", lead)
}

// ListLeads godoc
// @Summary List leads
// @Tags leads
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=dto.PaginatedLeadResponse}
// @Router /leads [get]
func (h *LeadHandler) ListLeads(c *fiber.Ctx) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid session")
	}
	role, _ := c.Locals("role").(string)

	query, err := parseLeadListQuery(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	result, err := h.leadService.List(userID, role, query)
	if err != nil {
		return utils.ErrorResponse(c, leadServiceErrorStatus(err), err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Leads retrieved successfully", result)
}

// parseLeadListQuery builds a typed dto.LeadListQuery from raw query
// strings. Every filter is strict (malformed value -> error) except
// page/limit, which clamp per ARCHITECTURE.md §9 rather than reject, and
// an empty string, which always means "no filter" rather than an error.
func parseLeadListQuery(c *fiber.Ctx) (dto.LeadListQuery, error) {
	q := dto.LeadListQuery{
		Q:      c.Query("q"),
		Status: c.Query("status"),
		Sort:   c.Query("sort"),
	}

	if v := c.Query("status"); v != "" {
		switch v {
		case "BARU", "FOLLOW_UP", "HANDOFF_ODOO", "LOST":
		default:
			return q, errors.New("status must be one of BARU, FOLLOW_UP, HANDOFF_ODOO, LOST")
		}
	}

	if v := c.Query("sort"); v != "" {
		switch v {
		case "code", "-code", "company_name", "-company_name":
		default:
			return q, errors.New("sort must be one of code, -code, company_name, -company_name")
		}
	}

	var err error
	if q.SourceID, err = parseOptionalUUID(c.Query("source_id")); err != nil {
		return q, errors.New("source_id must be a valid UUID")
	}
	if q.TeamID, err = parseOptionalUUID(c.Query("team_id")); err != nil {
		return q, errors.New("team_id must be a valid UUID")
	}
	if q.OwnerID, err = parseOptionalUUID(c.Query("owner_id")); err != nil {
		return q, errors.New("owner_id must be a valid UUID")
	}
	if q.ProvinceID, err = parseOptionalInt(c.Query("province_id")); err != nil {
		return q, errors.New("province_id must be a valid integer")
	}
	if q.CityID, err = parseOptionalInt(c.Query("city_id")); err != nil {
		return q, errors.New("city_id must be a valid integer")
	}
	if q.DateFrom, err = parseOptionalDate(c.Query("date_from")); err != nil {
		return q, errors.New("date_from must be YYYY-MM-DD")
	}
	if q.DateTo, err = parseOptionalDate(c.Query("date_to")); err != nil {
		return q, errors.New("date_to must be YYYY-MM-DD")
	}
	if q.FollowUpFrom, err = parseOptionalDate(c.Query("follow_up_from")); err != nil {
		return q, errors.New("follow_up_from must be YYYY-MM-DD")
	}
	if q.FollowUpTo, err = parseOptionalDate(c.Query("follow_up_to")); err != nil {
		return q, errors.New("follow_up_to must be YYYY-MM-DD")
	}
	if v := c.Query("stale"); v != "" {
		q.Stale, err = strconv.ParseBool(v)
		if err != nil {
			return q, errors.New("stale must be true or false")
		}
	}

	q.Page = 1
	if v := c.Query("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			q.Page = p
		}
	}

	q.Limit = 20
	if v := c.Query("limit"); v != "" {
		if l, err := strconv.Atoi(v); err == nil && l > 0 {
			q.Limit = l
		}
	}
	if q.Limit > 100 {
		q.Limit = 100
	}

	return q, nil
}

func parseOptionalUUID(v string) (*uuid.UUID, error) {
	if v == "" {
		return nil, nil
	}
	id, err := uuid.Parse(v)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func parseOptionalInt(v string) (*int, error) {
	if v == "" {
		return nil, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func parseOptionalDate(v string) (*time.Time, error) {
	if v == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// GetLead godoc
// @Summary Get lead
// @Tags leads
// @Security BearerAuth
// @Param code path string true "Lead code"
// @Success 200 {object} utils.Response{data=dto.LeadResponse}
// @Router /leads/{code} [get]
func (h *LeadHandler) GetLead(c *fiber.Ctx) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid session")
	}
	role, _ := c.Locals("role").(string)

	lead, err := h.leadService.FindByCode(userID, role, c.Params("code"))
	if err != nil {
		return utils.ErrorResponse(c, leadServiceErrorStatus(err), err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Lead retrieved successfully", lead)
}

// UpdateLead godoc
// @Summary Update lead
// @Tags leads
// @Security BearerAuth
// @Param code path string true "Lead code"
// @Param request body dto.UpdateLeadInput true "Update lead"
// @Success 200 {object} utils.Response{data=dto.LeadResponse}
// @Router /leads/{code} [patch]
func (h *LeadHandler) UpdateLead(c *fiber.Ctx) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid session")
	}
	role, _ := c.Locals("role").(string)

	var input dto.UpdateLeadInput
	if err := utils.ParseAndValidate(c, &input); err != nil {
		return err
	}

	lead, err := h.leadService.Update(userID, role, c.Params("code"), input)
	if err != nil {
		return utils.ErrorResponse(c, leadServiceErrorStatus(err), err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Lead updated successfully", lead)
}

// UpdateLeadStatus godoc
// @Summary Update lead status
// @Tags leads
// @Security BearerAuth
// @Param code path string true "Lead code"
// @Param request body dto.UpdateLeadStatusInput true "Update lead status"
// @Success 200 {object} utils.Response{data=dto.LeadResponse}
// @Router /leads/{code}/status [patch]
func (h *LeadHandler) UpdateLeadStatus(c *fiber.Ctx) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid session")
	}
	role, _ := c.Locals("role").(string)

	var input dto.UpdateLeadStatusInput
	if err := utils.ParseAndValidate(c, &input); err != nil {
		return err
	}

	lead, err := h.leadService.UpdateStatus(userID, role, c.Params("code"), input)
	if err != nil {
		return utils.ErrorResponse(c, leadServiceErrorStatus(err), err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Lead status updated successfully", lead)
}
