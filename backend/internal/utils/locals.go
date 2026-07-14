package utils

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// userIDKey is the Locals key used to carry the authenticated user's ID
// across middleware and handlers within a single request.
const userIDKey = "userID"

// SetUserID stores the authenticated user's ID into request Locals.
// Called by JWT middleware after successful token validation.
func SetUserID(c *fiber.Ctx, id uuid.UUID) {
	c.Locals(userIDKey, id)
}

// GetUserID retrieves the authenticated user's ID from request Locals.
// Returns (uuid.Nil, false) if the value is missing or has unexpected type.
func GetUserID(c *fiber.Ctx) (uuid.UUID, bool) {
	id, ok := c.Locals(userIDKey).(uuid.UUID)
	return id, ok
}
