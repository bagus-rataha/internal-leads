package routes

import (
	"fiber-api-boilerplate/internal/config"
	"fiber-api-boilerplate/internal/container"
	"fiber-api-boilerplate/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

// SetupAPIRoutes configures all API v1 routes
func SetupAPIRoutes(app *fiber.App, cnt interface{}, cfg *config.Config) {
	// Type assertion for container
	c, ok := cnt.(*container.Container)
	if !ok {
		return
	}

	// API v1 group
	api := app.Group("/api/v1")

	// Setup module routes
	setupAuthRoutes(api, c, cfg)
	setupUserRoutes(api, c, cfg)
	setupTeamRoutes(api, c, cfg)
}

// setupAuthRoutes configures authentication routes
func setupAuthRoutes(api fiber.Router, c *container.Container, cfg *config.Config) {
	auth := api.Group("/auth")
	auth.Post("/login", c.AuthHandler.Login)
	auth.Post("/refresh", c.AuthHandler.RefreshToken)
	auth.Post("/logout", c.AuthHandler.Logout)
	auth.Post("/logout-all", middleware.JWTProtected(cfg.JWTAccessSecret), c.AuthHandler.LogoutAll)
}

// setupUserRoutes configures user routes. /me is open to every authenticated
// role; the admin endpoints below it require ADMIN_SALES or SU.
func setupUserRoutes(api fiber.Router, c *container.Container, cfg *config.Config) {
	users := api.Group("/users")
	users.Use(middleware.JWTProtected(cfg.JWTAccessSecret))

	users.Get("/me", c.UserHandler.GetProfile)
	users.Put("/me", c.UserHandler.UpdateProfile)

	admin := users.Group("", middleware.RequireRole("ADMIN_SALES", "SU"))
	admin.Get("/", c.UserHandler.ListUsers)
	admin.Post("/", c.UserHandler.CreateUser)
	admin.Get("/:id", c.UserHandler.GetUser)
	admin.Patch("/:id", c.UserHandler.UpdateUser)
	admin.Post("/:id/reset-password", c.UserHandler.ResetPassword)
	admin.Post("/:id/deactivate", c.UserHandler.DeactivateUser)
}

// setupTeamRoutes configures sales team routes (admin-only)
func setupTeamRoutes(api fiber.Router, c *container.Container, cfg *config.Config) {
	teams := api.Group("/teams")
	teams.Use(middleware.JWTProtected(cfg.JWTAccessSecret))
	teams.Use(middleware.RequireRole("ADMIN_SALES", "SU"))

	teams.Get("/", c.TeamHandler.ListTeams)
	teams.Post("/", c.TeamHandler.CreateTeam)
	teams.Get("/:id", c.TeamHandler.GetTeam)
	teams.Patch("/:id", c.TeamHandler.UpdateTeam)
}
