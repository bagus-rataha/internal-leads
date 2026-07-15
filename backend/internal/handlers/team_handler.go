package handlers

import (
	"errors"
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/services"
	"fiber-api-boilerplate/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// teamService is the business-logic contract TeamHandler depends on.
// Defined consumer-side so it can be satisfied by the real service or a mock.
type teamService interface {
	Create(input dto.CreateTeamInput) (*dto.TeamResponse, error)
	List() ([]dto.TeamResponse, error)
	FindByID(id uuid.UUID) (*dto.TeamResponse, error)
	Update(id uuid.UUID, input dto.UpdateTeamInput) (*dto.TeamResponse, error)
}

type TeamHandler struct {
	teamService teamService
}

func NewTeamHandler(teamService teamService) *TeamHandler {
	return &TeamHandler{teamService: teamService}
}

// CreateTeam godoc
// @Summary Create sales team
// @Tags teams
// @Security BearerAuth
// @Param request body dto.CreateTeamInput true "Create team"
// @Success 201 {object} utils.Response{data=dto.TeamResponse}
// @Router /teams [post]
func (h *TeamHandler) CreateTeam(c *fiber.Ctx) error {
	var input dto.CreateTeamInput
	if err := utils.ParseAndValidate(c, &input); err != nil {
		return err
	}

	team, err := h.teamService.Create(input)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Team created successfully", team)
}

// ListTeams godoc
// @Summary List sales teams
// @Tags teams
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=[]dto.TeamResponse}
// @Router /teams [get]
func (h *TeamHandler) ListTeams(c *fiber.Ctx) error {
	teams, err := h.teamService.List()
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Teams retrieved successfully", teams)
}

// GetTeam godoc
// @Summary Get sales team
// @Tags teams
// @Security BearerAuth
// @Param id path string true "Team ID"
// @Success 200 {object} utils.Response{data=dto.TeamResponse}
// @Router /teams/{id} [get]
func (h *TeamHandler) GetTeam(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid team id")
	}

	team, err := h.teamService.FindByID(id)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Team retrieved successfully", team)
}

// UpdateTeam godoc
// @Summary Update sales team
// @Tags teams
// @Security BearerAuth
// @Param id path string true "Team ID"
// @Param request body dto.UpdateTeamInput true "Update team"
// @Success 200 {object} utils.Response{data=dto.TeamResponse}
// @Router /teams/{id} [patch]
func (h *TeamHandler) UpdateTeam(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid team id")
	}

	var input dto.UpdateTeamInput
	if err := utils.ParseAndValidate(c, &input); err != nil {
		return err
	}

	team, err := h.teamService.Update(id, input)
	if err != nil {
		if errors.Is(err, services.ErrTeamHasActiveMembers) {
			return utils.ErrorResponse(c, fiber.StatusUnprocessableEntity, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Team updated successfully", team)
}
