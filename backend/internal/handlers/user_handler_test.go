package handlers

import (
	"encoding/json"
	"errors"
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/utils"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// newTestUserApp wires the user routes behind a stub middleware that injects
// userID into Locals, mimicking what JWTProtected does after token validation.
func newTestUserApp(handler *UserHandler, userID uuid.UUID) *fiber.App {
	app := newTestApp()

	authMiddleware := func(c *fiber.Ctx) error {
		utils.SetUserID(c, userID)
		return c.Next()
	}

	app.Get("/users/me", authMiddleware, handler.GetProfile)
	app.Put("/users/me", authMiddleware, handler.UpdateProfile)
	app.Get("/users", authMiddleware, handler.ListUsers)
	return app
}

func TestUserHandler_GetProfile_Success(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewUserHandler(mockSvc)
	userID := uuid.Must(uuid.NewV7())
	app := newTestUserApp(handler, userID)

	userResponse := &dto.UserResponse{
		ID:        userID,
		Email:     "test@test.com",
		Name:      "Test",
		Role:      "user",
		CreatedAt: time.Now(),
	}

	mockSvc.On("GetProfile", userID).Return(userResponse, nil)

	req := httptest.NewRequest("GET", "/users/me", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result utils.Response
	json.NewDecoder(resp.Body).Decode(&result)
	assert.True(t, result.Success)
}

func TestUserHandler_GetProfile_NotFound(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewUserHandler(mockSvc)
	userID := uuid.Must(uuid.NewV7())
	app := newTestUserApp(handler, userID)

	mockSvc.On("GetProfile", userID).Return(nil, errors.New("user not found"))

	req := httptest.NewRequest("GET", "/users/me", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestUserHandler_UpdateProfile_Success(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewUserHandler(mockSvc)
	userID := uuid.Must(uuid.NewV7())
	app := newTestUserApp(handler, userID)

	input := dto.UpdateProfileInput{Name: "New Name"}
	userResponse := &dto.UserResponse{
		ID:    userID,
		Email: "test@test.com",
		Name:  "New Name",
	}

	mockSvc.On("UpdateProfile", userID, input).Return(userResponse, nil)

	body := `{"name":"New Name"}`
	req := httptest.NewRequest("PUT", "/users/me", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestUserHandler_ListUsers_Success(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewUserHandler(mockSvc)
	userID := uuid.Must(uuid.NewV7())
	app := newTestUserApp(handler, userID)

	users := []dto.UserResponse{
		{ID: uuid.Must(uuid.NewV7()), Email: "a@test.com"},
		{ID: uuid.Must(uuid.NewV7()), Email: "b@test.com"},
	}

	mockSvc.On("ListUsers").Return(users, nil)

	req := httptest.NewRequest("GET", "/users", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}
