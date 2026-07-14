package app

import (
	"fiber-api-boilerplate/internal/config"
	"fiber-api-boilerplate/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

// NewFiberApp creates and configures Fiber application
func NewFiberApp(cfg *config.Config) *fiber.App {
	// Create Fiber app with configuration
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
		Prefork:      cfg.IsProduction(),
		ServerHeader: "",
		AppName:      "Fiber API v3.0",
	})

	// Setup global middleware
	middleware.SetupMiddleware(app, cfg)

	return app
}
