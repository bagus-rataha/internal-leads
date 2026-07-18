package handlers

import (
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// leadSourceService is the business-logic contract LeadSourceHandler depends
// on. Defined consumer-side so it can be satisfied by the real service or a
// mock.
type leadSourceService interface {
	Create(input dto.CreateLeadSourceInput) (*dto.LeadSourceAdminResponse, error)
	List() ([]dto.LeadSourceAdminResponse, error)
	Update(id uuid.UUID, input dto.UpdateLeadSourceInput) (*dto.LeadSourceAdminResponse, error)
}

type LeadSourceHandler struct {
	leadSourceService leadSourceService
}

func NewLeadSourceHandler(leadSourceService leadSourceService) *LeadSourceHandler {
	return &LeadSourceHandler{leadSourceService: leadSourceService}
}

// CreateLeadSource godoc
// @Summary Create lead source (admin only)
// @Tags lead-sources
// @Security BearerAuth
// @Param request body dto.CreateLeadSourceInput true "Create lead source"
// @Success 201 {object} utils.Response{data=dto.LeadSourceAdminResponse}
// @Router /lead-sources [post]
func (h *LeadSourceHandler) CreateLeadSource(c *fiber.Ctx) error {
	var input dto.CreateLeadSourceInput
	if err := utils.ParseAndValidate(c, &input); err != nil {
		return err
	}

	source, err := h.leadSourceService.Create(input)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Lead source created successfully", source)
}

// ListLeadSourcesAdmin godoc
// @Summary List all lead sources, active and inactive (admin only)
// @Tags lead-sources
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=[]dto.LeadSourceAdminResponse}
// @Router /lead-sources [get]
func (h *LeadSourceHandler) ListLeadSourcesAdmin(c *fiber.Ctx) error {
	sources, err := h.leadSourceService.List()
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Lead sources retrieved successfully", sources)
}

// UpdateLeadSource godoc
// @Summary Update lead source (admin only)
// @Tags lead-sources
// @Security BearerAuth
// @Param id path string true "Lead source ID"
// @Param request body dto.UpdateLeadSourceInput true "Update lead source"
// @Success 200 {object} utils.Response{data=dto.LeadSourceAdminResponse}
// @Router /lead-sources/{id} [patch]
func (h *LeadSourceHandler) UpdateLeadSource(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid lead source id")
	}

	var input dto.UpdateLeadSourceInput
	if err := utils.ParseAndValidate(c, &input); err != nil {
		return err
	}

	source, err := h.leadSourceService.Update(id, input)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Lead source updated successfully", source)
}
