package middleware

import (
	"fiber-api-boilerplate/internal/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// JWTProtected middleware validates JWT token
func JWTProtected(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Missing authorization token")
		}

		// Check Bearer scheme
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid authorization format")
		}

		token := parts[1]

		// Validate token
		claims, err := utils.ValidateToken(token, secret)
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid or expired token")
		}

		// Set user info to context
		utils.SetUserID(c, claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}
