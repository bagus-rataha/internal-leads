package middleware

import (
	"fiber-api-boilerplate/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func SetupMiddleware(app *fiber.App, cfg *config.Config) {
	app.Use(recover.New(recover.Config{
		EnableStackTrace: cfg.IsDevelopment(),
	}))

	app.Use(SecurityHeaders(cfg))

	setupCORS(app, cfg)

	if cfg.IsProduction() {
		setupRateLimiter(app, cfg)
		app.Use(compress.New(compress.Config{
			Level: compress.LevelBestSpeed,
		}))
	}

	// Fiber's built-in logger
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${method} ${path} ${latency}\n",
		TimeFormat: "15:04:05",
		TimeZone:   "Local",
	}))
}

// ErrorHandler handles all errors globally
// Skips if response already written (status != 200)
func ErrorHandler(c *fiber.Ctx, err error) error {
	// If response already written, skip to prevent overwrite
	if c.Response().StatusCode() != fiber.StatusOK {
		return nil
	}

	// Default error code
	code := fiber.StatusInternalServerError

	// Check if it's a Fiber error with custom code
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	// Write error response (fallback for unhandled errors)
	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"message": err.Error(),
	})
}

func setupCORS(app *fiber.App, cfg *config.Config) {
	if cfg.IsProduction() {
		app.Use(cors.New(cors.Config{
			AllowOrigins:     cfg.AllowedOrigins,
			AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
			AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
			AllowCredentials: true,
		}))
	} else {
		app.Use(cors.New(cors.Config{
			AllowOrigins: "*",
			AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		}))
	}
}

func setupRateLimiter(app *fiber.App, cfg *config.Config) {
	app.Use(limiter.New(limiter.Config{
		Max:        cfg.RateLimitMax,
		Expiration: cfg.RateLimitWindow,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"message": "Too many requests, please try again later",
			})
		},
	}))
}
