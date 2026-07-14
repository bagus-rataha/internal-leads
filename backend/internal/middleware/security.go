package middleware

import (
	"fiber-api-boilerplate/internal/config"

	"github.com/gofiber/fiber/v2"
)

// SecurityHeaders sets security-related HTTP headers.
// HSTS is only emitted in production where the server is expected to serve HTTPS.
func SecurityHeaders(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")

		if cfg.IsProduction() {
			c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		return c.Next()
	}
}
