package handlers

import (
	"errors"
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/services"
	"fiber-api-boilerplate/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// userService is the business-logic contract UserHandler depends on.
// Defined consumer-side so it can be satisfied by the real service or a mock.
type userService interface {
	GetProfile(userID uuid.UUID) (*dto.UserResponse, error)
	UpdateProfile(userID uuid.UUID, input dto.UpdateProfileInput) (*dto.UserResponse, error)
	ListUsers(callerID uuid.UUID, callerRole, role, teamID string) ([]dto.UserResponse, error)
	CreateUser(input dto.CreateUserInput) (*dto.UserResponse, error)
	GetUser(userID uuid.UUID) (*dto.UserResponse, error)
	UpdateUser(callerRole string, userID uuid.UUID, input dto.UpdateUserInput) (*dto.UserResponse, error)
	ResetPassword(callerRole string, userID uuid.UUID, input dto.ResetPasswordInput) error
	DeactivateUser(callerID uuid.UUID, callerRole string, userID uuid.UUID, input dto.DeactivateUserInput) (int64, error)
}

type UserHandler struct {
	userService userService
}

func NewUserHandler(userService userService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetProfile godoc
// @Summary Get profile
// @Tags users
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=dto.UserResponse}
// @Router /users/me [get]
func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid session")
	}

	user, err := h.userService.GetProfile(userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Profile retrieved successfully", user)
}

// UpdateProfile godoc
// @Summary Update profile
// @Tags users
// @Security BearerAuth
// @Param request body dto.UpdateProfileInput true "Update profile"
// @Success 200 {object} utils.Response{data=dto.UserResponse}
// @Router /users/me [put]
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid session")
	}

	var input dto.UpdateProfileInput
	if err := utils.ParseAndValidate(c, &input); err != nil {
		return err
	}

	user, err := h.userService.UpdateProfile(userID, input)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Profile updated successfully", user)
}

// ListUsers godoc
// @Summary List users (LEADER: own team roster; ADMIN_SALES/SU: unrestricted)
// @Tags users
// @Security BearerAuth
// @Param role query string false "Filter by role"
// @Param team_id query string false "Filter by team id (ignored for LEADER callers, who are always scoped to their own team)"
// @Success 200 {object} utils.Response{data=[]dto.UserResponse}
// @Router /users [get]
func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	callerID, ok := utils.GetUserID(c)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid session")
	}
	callerRole, _ := c.Locals("role").(string)

	role := c.Query("role")
	teamID := c.Query("team_id")

	users, err := h.userService.ListUsers(callerID, callerRole, role, teamID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Users retrieved successfully", users)
}

// CreateUser godoc
// @Summary Create user (admin only)
// @Tags users
// @Security BearerAuth
// @Param request body dto.CreateUserInput true "Create user"
// @Success 201 {object} utils.Response{data=dto.UserResponse}
// @Router /users [post]
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var input dto.CreateUserInput
	if err := utils.ParseAndValidate(c, &input); err != nil {
		return err
	}

	user, err := h.userService.CreateUser(input)
	if err != nil {
		var emailErr *services.EmailTakenError
		if errors.As(err, &emailErr) {
			return utils.ErrorResponseWithData(c, fiber.StatusConflict, "Email already registered", fiber.Map{
				"existing_user_id":     emailErr.ExistingUserID,
				"existing_user_name":   emailErr.ExistingName,
				"existing_user_active": emailErr.ExistingActive,
			})
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "User created successfully", user)
}

// GetUser godoc
// @Summary Get user by id (admin only)
// @Tags users
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} utils.Response{data=dto.UserResponse}
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid user id")
	}

	user, err := h.userService.GetUser(id)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "User retrieved successfully", user)
}

// UpdateUser godoc
// @Summary Update user (admin only)
// @Tags users
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body dto.UpdateUserInput true "Update user"
// @Success 200 {object} utils.Response{data=dto.UserResponse}
// @Router /users/{id} [patch]
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid user id")
	}
	callerRole, _ := c.Locals("role").(string)

	var input dto.UpdateUserInput
	if err := utils.ParseAndValidate(c, &input); err != nil {
		return err
	}

	user, err := h.userService.UpdateUser(callerRole, id, input)
	if err != nil {
		if errors.Is(err, services.ErrForbiddenTarget) {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "User updated successfully", user)
}

// ResetPassword godoc
// @Summary Reset a user's password (admin only)
// @Tags users
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body dto.ResetPasswordInput true "Reset password"
// @Success 200 {object} utils.Response
// @Router /users/{id}/reset-password [post]
func (h *UserHandler) ResetPassword(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid user id")
	}
	callerRole, _ := c.Locals("role").(string)

	var input dto.ResetPasswordInput
	if err := utils.ParseAndValidate(c, &input); err != nil {
		return err
	}

	if err := h.userService.ResetPassword(callerRole, id, input); err != nil {
		if errors.Is(err, services.ErrForbiddenTarget) {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Password reset successfully", nil)
}

// DeactivateUser godoc
// @Summary Deactivate a user (admin only)
// @Tags users
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body dto.DeactivateUserInput true "Deactivate user"
// @Success 200 {object} utils.Response
// @Failure 422 {object} utils.Response "user still has active leads, reassign_to_user_id required"
// @Router /users/{id}/deactivate [post]
func (h *UserHandler) DeactivateUser(c *fiber.Ctx) error {
	callerID, ok := utils.GetUserID(c)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid session")
	}
	callerRole, _ := c.Locals("role").(string)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid user id")
	}

	var input dto.DeactivateUserInput
	if err := utils.ParseAndValidate(c, &input); err != nil {
		return err
	}

	activeCount, err := h.userService.DeactivateUser(callerID, callerRole, id, input)
	if err != nil {
		if errors.Is(err, services.ErrActiveLeadsExist) {
			return utils.ErrorResponseWithData(c, fiber.StatusUnprocessableEntity, err.Error(),
				fiber.Map{"active_lead_count": activeCount})
		}
		if errors.Is(err, services.ErrForbiddenTarget) {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "User deactivated successfully", nil)
}
