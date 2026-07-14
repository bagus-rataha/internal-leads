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
}

// setupAuthRoutes configures authentication routes
func setupAuthRoutes(api fiber.Router, c *container.Container, cfg *config.Config) {
	auth := api.Group("/auth")
	auth.Post("/register", c.AuthHandler.Register)
	auth.Post("/login", c.AuthHandler.Login)
	auth.Post("/refresh", c.AuthHandler.RefreshToken)
	auth.Post("/logout", c.AuthHandler.Logout)
	auth.Post("/logout-all", middleware.JWTProtected(cfg.JWTAccessSecret), c.AuthHandler.LogoutAll)
}

// setupUserRoutes configures user routes (protected)
func setupUserRoutes(api fiber.Router, c *container.Container, cfg *config.Config) {
	users := api.Group("/users")
	users.Use(middleware.JWTProtected(cfg.JWTAccessSecret))

	users.Get("/me", c.UserHandler.GetProfile)
	users.Put("/me", c.UserHandler.UpdateProfile)
	users.Get("/", c.UserHandler.ListUsers)
}
