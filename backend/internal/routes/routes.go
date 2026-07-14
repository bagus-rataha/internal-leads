package routes

import (
	"fiber-api-boilerplate/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

// SetupRoutes configures all application routes
func SetupRoutes(app *fiber.App, container interface{}, cfg *config.Config) {
	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"env":    cfg.AppEnv,
		})
	})

	// Swagger documentation (development only)
	if cfg.IsDevelopment() {
		app.Get("/swagger/*", swagger.HandlerDefault)
	}

	// API v1 routes
	SetupAPIRoutes(app, container, cfg)
}
