package handlers

import (
	"encoding/json"
	"errors"
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/services"
	"fiber-api-boilerplate/internal/utils"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// newTestTeamApp wires the team routes directly against the handler, mirroring
// how the real router mounts them (JWTProtected + RequireRole run before the
// handler in production; here we test the handler's own behavior in isolation).
func newTestTeamApp(handler *TeamHandler) *fiber.App {
	app := newTestApp()

	app.Get("/teams", handler.ListTeams)
	app.Post("/teams", handler.CreateTeam)
	app.Get("/teams/:id", handler.GetTeam)
	app.Patch("/teams/:id", handler.UpdateTeam)
	return app
}

func TestTeamHandler_CreateTeam_Success(t *testing.T) {
	mockSvc := new(MockTeamService)
	handler := NewTeamHandler(mockSvc)
	app := newTestTeamApp(handler)

	input := dto.CreateTeamInput{Name: "Team A"}
	teamResponse := &dto.TeamResponse{
		ID:        uuid.Must(uuid.NewV7()),
		Name:      "Team A",
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	mockSvc.On("Create", input).Return(teamResponse, nil)

	body := `{"name":"Team A"}`
	req := httptest.NewRequest("POST", "/teams", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

func TestTeamHandler_ListTeams_Success(t *testing.T) {
	mockSvc := new(MockTeamService)
	handler := NewTeamHandler(mockSvc)
	app := newTestTeamApp(handler)

	teams := []dto.TeamResponse{
		{ID: uuid.Must(uuid.NewV7()), Name: "A", IsActive: true},
		{ID: uuid.Must(uuid.NewV7()), Name: "B", IsActive: true},
	}
	mockSvc.On("List").Return(teams, nil)

	req := httptest.NewRequest("GET", "/teams", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result utils.Response
	json.NewDecoder(resp.Body).Decode(&result)
	assert.True(t, result.Success)
}

func TestTeamHandler_GetTeam_Success(t *testing.T) {
	mockSvc := new(MockTeamService)
	handler := NewTeamHandler(mockSvc)
	app := newTestTeamApp(handler)

	teamID := uuid.Must(uuid.NewV7())
	teamResponse := &dto.TeamResponse{ID: teamID, Name: "A", IsActive: true}
	mockSvc.On("FindByID", teamID).Return(teamResponse, nil)

	req := httptest.NewRequest("GET", "/teams/"+teamID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestTeamHandler_GetTeam_InvalidID(t *testing.T) {
	mockSvc := new(MockTeamService)
	handler := NewTeamHandler(mockSvc)
	app := newTestTeamApp(handler)

	req := httptest.NewRequest("GET", "/teams/not-a-uuid", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestTeamHandler_UpdateTeam_Success(t *testing.T) {
	mockSvc := new(MockTeamService)
	handler := NewTeamHandler(mockSvc)
	app := newTestTeamApp(handler)

	teamID := uuid.Must(uuid.NewV7())
	newName := "New Name"
	input := dto.UpdateTeamInput{Name: &newName}
	teamResponse := &dto.TeamResponse{ID: teamID, Name: "New Name", IsActive: true}

	mockSvc.On("Update", teamID, input).Return(teamResponse, nil)

	body := `{"name":"New Name"}`
	req := httptest.NewRequest("PATCH", "/teams/"+teamID.String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestTeamHandler_UpdateTeam_InvalidID(t *testing.T) {
	mockSvc := new(MockTeamService)
	handler := NewTeamHandler(mockSvc)
	app := newTestTeamApp(handler)

	body := `{"name":"New Name"}`
	req := httptest.NewRequest("PATCH", "/teams/not-a-uuid", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestTeamHandler_UpdateTeam_DeactivateWithActiveMembers_422(t *testing.T) {
	mockSvc := new(MockTeamService)
	handler := NewTeamHandler(mockSvc)
	app := newTestTeamApp(handler)

	teamID := uuid.Must(uuid.NewV7())
	inactive := false
	input := dto.UpdateTeamInput{IsActive: &inactive}

	mockSvc.On("Update", teamID, input).Return(nil, services.ErrTeamHasActiveMembers)

	body := `{"is_active":false}`
	req := httptest.NewRequest("PATCH", "/teams/"+teamID.String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)
}

func TestTeamHandler_UpdateTeam_OtherServiceError_400(t *testing.T) {
	mockSvc := new(MockTeamService)
	handler := NewTeamHandler(mockSvc)
	app := newTestTeamApp(handler)

	teamID := uuid.Must(uuid.NewV7())
	newName := "New Name"
	input := dto.UpdateTeamInput{Name: &newName}

	mockSvc.On("Update", teamID, input).Return(nil, errors.New("team not found"))

	body := `{"name":"New Name"}`
	req := httptest.NewRequest("PATCH", "/teams/"+teamID.String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
