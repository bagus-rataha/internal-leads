package middleware

import (
	"fiber-api-boilerplate/internal/utils"

	"github.com/gofiber/fiber/v2"
)

// RequireRole restricts a route to the given roles. It must run after
// JWTProtected, which populates c.Locals("role") from the JWT claim.
func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals("role").(string)

		for _, r := range roles {
			if role == r {
				return c.Next()
			}
		}

		return utils.ErrorResponse(c, fiber.StatusForbidden, "insufficient role")
	}
}
