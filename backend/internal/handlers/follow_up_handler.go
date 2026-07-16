package handlers

import (
	"errors"
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/services"
	"fiber-api-boilerplate/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// followUpService is the business-logic contract FollowUpHandler depends
// on. Defined consumer-side so it can be satisfied by the real service or a
// mock.
type followUpService interface {
	Create(callerID uuid.UUID, role, leadCode string, input dto.CreateFollowUpInput) (*dto.FollowUpResponse, error)
	List(callerID uuid.UUID, role, leadCode string) ([]dto.FollowUpResponse, error)
}

type FollowUpHandler struct {
	followUpService followUpService
}

func NewFollowUpHandler(followUpService followUpService) *FollowUpHandler {
	return &FollowUpHandler{followUpService: followUpService}
}

// CreateFollowUp godoc
// @Summary Create follow-up
// @Tags followups
// @Security BearerAuth
// @Param code path string true "Lead code"
// @Param request body dto.CreateFollowUpInput true "Create follow-up"
// @Success 201 {object} utils.Response{data=dto.FollowUpResponse}
// @Router /leads/{code}/followups [post]
func (h *FollowUpHandler) CreateFollowUp(c *fiber.Ctx) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid session")
	}
	role, _ := c.Locals("role").(string)

	var input dto.CreateFollowUpInput
	if err := utils.ParseAndValidate(c, &input); err != nil {
		return err
	}

	followUp, err := h.followUpService.Create(userID, role, c.Params("code"), input)
	if err != nil {
		status := fiber.StatusBadRequest
		if errors.Is(err, services.ErrLeadNotFound) {
			status = fiber.StatusNotFound
		}
		return utils.ErrorResponse(c, status, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Follow-up created successfully", followUp)
}

// ListFollowUps godoc
// @Summary List follow-ups
// @Tags followups
// @Security BearerAuth
// @Param code path string true "Lead code"
// @Success 200 {object} utils.Response{data=[]dto.FollowUpResponse}
// @Router /leads/{code}/followups [get]
func (h *FollowUpHandler) ListFollowUps(c *fiber.Ctx) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid session")
	}
	role, _ := c.Locals("role").(string)

	followUps, err := h.followUpService.List(userID, role, c.Params("code"))
	if err != nil {
		status := fiber.StatusBadRequest
		if errors.Is(err, services.ErrLeadNotFound) {
			status = fiber.StatusNotFound
		}
		return utils.ErrorResponse(c, status, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Follow-ups retrieved successfully", followUps)
}
