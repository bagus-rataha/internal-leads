package handlers

import (
	"errors"
	"fiber-api-boilerplate/internal/config"
	"fiber-api-boilerplate/internal/dto"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func newTestConfig() *config.Config {
	return &config.Config{
		JWTAccessSecret:  "test-access-secret",
		JWTRefreshSecret: "test-refresh-secret",
		JWTAccessExpire:  time.Hour,
		JWTRefreshExpire: 24 * time.Hour,
		AppEnv:           "development",
	}
}

// newTestApp builds a Fiber app whose ErrorHandler leaves an already-written
// non-200 response untouched (ParseAndValidate writes its own 400 body and
// returns fiber.ErrBadRequest).
func newTestApp() *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if c.Response().StatusCode() != fiber.StatusOK {
				return nil
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		},
	})
}

func newTestAuthApp(handler *AuthHandler) *fiber.App {
	app := newTestApp()
	app.Post("/auth/login", handler.Login)
	app.Post("/auth/refresh", handler.RefreshToken)
	app.Post("/auth/logout", handler.Logout)
	return app
}

func TestAuthHandler_Login_Success(t *testing.T) {
	mockSvc := new(MockAuthService)
	handler := NewAuthHandler(mockSvc, newTestConfig())
	app := newTestAuthApp(handler)

	tokenResponse := &dto.TokenResponse{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}

	mockSvc.On("Login", dto.LoginInput{
		Email:    "test@test.com",
		Password: "123456",
	}).Return(tokenResponse, nil)

	body := `{"email":"test@test.com","password":"123456"}`
	req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	mockSvc := new(MockAuthService)
	handler := NewAuthHandler(mockSvc, newTestConfig())
	app := newTestAuthApp(handler)

	mockSvc.On("Login", dto.LoginInput{
		Email:    "test@test.com",
		Password: "wrong",
	}).Return(nil, errors.New("invalid credentials"))

	body := `{"email":"test@test.com","password":"wrong"}`
	req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
