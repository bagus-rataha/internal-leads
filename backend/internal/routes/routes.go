package routes

import (
	"fiber-api-boilerplate/internal/config"
	"fiber-api-boilerplate/internal/static"
	"io/fs"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
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

	// Frontend build. Registered last: Fiber matches routes in registration
	// order, so every route above this line still wins over the catch-all
	// static/fallback handler below.
	SetupStaticRoutes(app, static.MustDistFS())
}

// SetupStaticRoutes serves the embedded frontend build and falls back to
// index.html for any path that isn't a real file, so client-side routing
// keeps working on a full page load or refresh. Must be the last route
// registered: it answers every GET/HEAD request that reaches it, matched or
// not, via the underlying filesystem middleware's NotFoundFile fallback.
func SetupStaticRoutes(app *fiber.App, spaFS fs.FS) {
	app.Use(filesystem.New(filesystem.Config{
		Root:         http.FS(spaFS),
		Index:        "index.html",
		NotFoundFile: "/index.html",
	}))
}
