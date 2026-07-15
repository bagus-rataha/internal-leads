package handlers

import (
	"encoding/json"
	"errors"
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/middleware"
	"fiber-api-boilerplate/internal/services"
	"fiber-api-boilerplate/internal/utils"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	return app
}

// newTestUserAdminApp mirrors routes/api.go's admin sub-group: a stub role
// middleware (standing in for JWTProtected, which normally sets Locals
// "role") followed by the real RequireRole gate, so 403 behavior is covered
// the same way production requests hit it.
func newTestUserAdminApp(handler *UserHandler, role string) *fiber.App {
	app := newTestApp()

	roleMiddleware := func(c *fiber.Ctx) error {
		c.Locals("role", role)
		return c.Next()
	}

	admin := app.Group("", roleMiddleware, middleware.RequireRole("ADMIN_SALES", "SU"))
	admin.Get("/users", handler.ListUsers)
	admin.Post("/users", handler.CreateUser)
	admin.Get("/users/:id", handler.GetUser)
	admin.Patch("/users/:id", handler.UpdateUser)
	admin.Post("/users/:id/reset-password", handler.ResetPassword)
	admin.Post("/users/:id/deactivate", handler.DeactivateUser)
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
	app := newTestUserAdminApp(handler, "SU")

	users := []dto.UserResponse{
		{ID: uuid.Must(uuid.NewV7()), Email: "a@test.com"},
		{ID: uuid.Must(uuid.NewV7()), Email: "b@test.com"},
	}

	mockSvc.On("ListUsers", "", "").Return(users, nil)

	req := httptest.NewRequest("GET", "/users", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestUserHandler_ListUsers_WrongRole_403(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewUserHandler(mockSvc)
	app := newTestUserAdminApp(handler, "SALES")

	req := httptest.NewRequest("GET", "/users", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	mockSvc.AssertNotCalled(t, "ListUsers", mock.Anything, mock.Anything)
}

func TestUserHandler_CreateUser_Success(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewUserHandler(mockSvc)
	app := newTestUserAdminApp(handler, "ADMIN_SALES")

	teamID := uuid.Must(uuid.NewV7())
	input := dto.CreateUserInput{Name: "New Sales", Email: "new@test.com", Password: "secret1", Role: "SALES", TeamID: &teamID}
	userResponse := &dto.UserResponse{ID: uuid.Must(uuid.NewV7()), Email: "new@test.com", Role: "SALES"}

	mockSvc.On("CreateUser", input).Return(userResponse, nil)

	body := `{"name":"New Sales","email":"new@test.com","password":"secret1","role":"SALES","team_id":"` + teamID.String() + `"}`
	req := httptest.NewRequest("POST", "/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

func TestUserHandler_CreateUser_WrongRole_403(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewUserHandler(mockSvc)
	app := newTestUserAdminApp(handler, "LEADER")

	body := `{"name":"New Sales","email":"new@test.com","password":"secret1","role":"SALES"}`
	req := httptest.NewRequest("POST", "/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestUserHandler_GetUser_Success(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewUserHandler(mockSvc)
	app := newTestUserAdminApp(handler, "SU")

	userID := uuid.Must(uuid.NewV7())
	userResponse := &dto.UserResponse{ID: userID, Email: "test@test.com"}
	mockSvc.On("GetUser", userID).Return(userResponse, nil)

	req := httptest.NewRequest("GET", "/users/"+userID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestUserHandler_UpdateUser_Success(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewUserHandler(mockSvc)
	app := newTestUserAdminApp(handler, "SU")

	userID := uuid.Must(uuid.NewV7())
	newName := "Renamed"
	input := dto.UpdateUserInput{Name: &newName}
	userResponse := &dto.UserResponse{ID: userID, Name: "Renamed"}

	mockSvc.On("UpdateUser", userID, input).Return(userResponse, nil)

	body := `{"name":"Renamed"}`
	req := httptest.NewRequest("PATCH", "/users/"+userID.String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestUserHandler_ResetPassword_Success(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewUserHandler(mockSvc)
	app := newTestUserAdminApp(handler, "SU")

	userID := uuid.Must(uuid.NewV7())
	input := dto.ResetPasswordInput{NewPassword: "newpass1"}
	mockSvc.On("ResetPassword", userID, input).Return(nil)

	body := `{"new_password":"newpass1"}`
	req := httptest.NewRequest("POST", "/users/"+userID.String()+"/reset-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestUserHandler_DeactivateUser_Success(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewUserHandler(mockSvc)
	app := newTestUserAdminApp(handler, "SU")

	userID := uuid.Must(uuid.NewV7())
	mockSvc.On("DeactivateUser", userID, dto.DeactivateUserInput{}).Return(int64(0), nil)

	req := httptest.NewRequest("POST", "/users/"+userID.String()+"/deactivate", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestUserHandler_DeactivateUser_ActiveLeadsNoReassign_422(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewUserHandler(mockSvc)
	app := newTestUserAdminApp(handler, "SU")

	userID := uuid.Must(uuid.NewV7())
	mockSvc.On("DeactivateUser", userID, dto.DeactivateUserInput{}).Return(int64(5), services.ErrActiveLeadsExist)

	req := httptest.NewRequest("POST", "/users/"+userID.String()+"/deactivate", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)

	var result utils.Response
	json.NewDecoder(resp.Body).Decode(&result)
	assert.False(t, result.Success)
	data, ok := result.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(5), data["active_lead_count"])
}

func TestUserHandler_DeactivateUser_WrongRole_403(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewUserHandler(mockSvc)
	app := newTestUserAdminApp(handler, "SALES")

	userID := uuid.Must(uuid.NewV7())
	req := httptest.NewRequest("POST", "/users/"+userID.String()+"/deactivate", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
