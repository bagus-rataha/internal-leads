package handlers

import (
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// userService is the business-logic contract UserHandler depends on.
// Defined consumer-side so it can be satisfied by the real service or a mock.
type userService interface {
	GetProfile(userID uuid.UUID) (*dto.UserResponse, error)
	UpdateProfile(userID uuid.UUID, input dto.UpdateProfileInput) (*dto.UserResponse, error)
	ListUsers() ([]dto.UserResponse, error)
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
// @Summary List users
// @Tags users
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=[]dto.UserResponse}
// @Router /users [get]
func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	users, err := h.userService.ListUsers()
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Users retrieved successfully", users)
}
