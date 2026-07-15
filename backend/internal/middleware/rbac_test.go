package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// newRoleApp builds a minimal Fiber app that stubs c.Locals("role") to the
// given value (standing in for JWTProtected, which normally sets it) and
// then runs RequireRole in front of a dummy 200 handler.
func newRoleApp(setRole string, allowed ...string) *fiber.App {
	app := fiber.New()
	app.Get("/protected", func(c *fiber.Ctx) error {
		c.Locals("role", setRole)
		return c.Next()
	}, RequireRole(allowed...), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	return app
}

func TestRequireRole_MatchingRole(t *testing.T) {
	app := newRoleApp("ADMIN_SALES", "ADMIN_SALES", "SU")

	req := httptest.NewRequest("GET", "/protected", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestRequireRole_NonMatchingRole(t *testing.T) {
	app := newRoleApp("SALES", "ADMIN_SALES", "SU")

	req := httptest.NewRequest("GET", "/protected", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestRequireRole_MultipleAllowedRoles_OneMatches(t *testing.T) {
	app := newRoleApp("LEADER", "ADMIN_SALES", "SU", "LEADER")

	req := httptest.NewRequest("GET", "/protected", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}
