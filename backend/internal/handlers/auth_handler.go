package handlers

import (
	"fiber-api-boilerplate/internal/config"
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// authService is the business-logic contract AuthHandler depends on.
// Defined consumer-side so it can be satisfied by the real service or a mock.
type authService interface {
	Register(input dto.RegisterInput) (*dto.TokenResponse, error)
	Login(input dto.LoginInput) (*dto.TokenResponse, error)
	RefreshToken(refreshToken string) (*dto.TokenResponse, error)
	Logout(refreshToken string) error
	LogoutAll(userID uuid.UUID) error
}

type AuthHandler struct {
	authService authService
	config      *config.Config
}

func NewAuthHandler(authService authService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{authService: authService, config: cfg}
}

// extractRefreshToken reads the refresh token from the cookie first and
// falls back to the request body so non-browser clients still work.
func (h *AuthHandler) extractRefreshToken(c *fiber.Ctx) string {
	if token := c.Cookies(utils.RefreshTokenCookieName); token != "" {
		return token
	}
	var input dto.RefreshTokenInput
	_ = c.BodyParser(&input)
	return input.RefreshToken
}

// Register godoc
// @Summary Register new user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterInput true "Register request"
// @Success 201 {object} utils.Response{data=dto.TokenResponse}
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var input dto.RegisterInput

	if err := utils.ParseAndValidate(c, &input); err != nil {
		return err
	}

	result, err := h.authService.Register(input)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	utils.SetRefreshTokenCookie(c, result.RefreshToken, h.config.JWTRefreshExpire, h.config.IsProduction())
	return utils.SuccessResponse(c, fiber.StatusCreated, "User registered successfully", result)
}

// Login godoc
// @Summary Login user
// @Tags auth
// @Param request body dto.LoginInput true "Login request"
// @Success 200 {object} utils.Response{data=dto.TokenResponse}
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var input dto.LoginInput

	if err := utils.ParseAndValidate(c, &input); err != nil {
		return err
	}

	result, err := h.authService.Login(input)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	utils.SetRefreshTokenCookie(c, result.RefreshToken, h.config.JWTRefreshExpire, h.config.IsProduction())
	return utils.SuccessResponse(c, fiber.StatusOK, "Login successful", result)
}

// RefreshToken godoc
// @Summary Refresh token
// @Tags auth
// @Param request body dto.RefreshTokenInput false "Refresh token (optional if cookie present)"
// @Success 200 {object} utils.Response{data=dto.TokenResponse}
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	refreshToken := h.extractRefreshToken(c)
	if refreshToken == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "refresh token is required")
	}

	result, err := h.authService.RefreshToken(refreshToken)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	utils.SetRefreshTokenCookie(c, result.RefreshToken, h.config.JWTRefreshExpire, h.config.IsProduction())
	return utils.SuccessResponse(c, fiber.StatusOK, "Token refreshed successfully", result)
}

// Logout godoc
// @Summary Logout current session
// @Tags auth
// @Success 200 {object} utils.Response
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	refreshToken := h.extractRefreshToken(c)
	if refreshToken != "" {
		_ = h.authService.Logout(refreshToken)
	}

	utils.ClearRefreshTokenCookie(c, h.config.IsProduction())
	return utils.SuccessResponse(c, fiber.StatusOK, "Logged out successfully", nil)
}

// LogoutAll godoc
// @Summary Logout from all devices
// @Tags auth
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /auth/logout-all [post]
func (h *AuthHandler) LogoutAll(c *fiber.Ctx) error {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid session")
	}

	if err := h.authService.LogoutAll(userID); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	utils.ClearRefreshTokenCookie(c, h.config.IsProduction())
	return utils.SuccessResponse(c, fiber.StatusOK, "Logged out from all devices", nil)
}
