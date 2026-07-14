package utils

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

type testInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// newTestApp builds a Fiber app whose ErrorHandler leaves an already-written
// non-200 response untouched. ParseAndValidate writes its own 400 body and
// returns fiber.ErrBadRequest, so without this the default handler would
// overwrite the detailed validation response.
func newTestApp() *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if c.Response().StatusCode() != fiber.StatusOK {
				return nil
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		},
	})
}

func TestParseAndValidate_Success(t *testing.T) {
	app := newTestApp()
	app.Post("/test", func(c *fiber.Ctx) error {
		var input testInput
		if err := ParseAndValidate(c, &input); err != nil {
			return err
		}
		return c.SendStatus(200)
	})

	body := `{"email":"test@test.com","password":"123456"}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestParseAndValidate_MissingField(t *testing.T) {
	app := newTestApp()
	app.Post("/test", func(c *fiber.Ctx) error {
		var input testInput
		if err := ParseAndValidate(c, &input); err != nil {
			return err
		}
		return c.SendStatus(200)
	})

	body := `{}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestParseAndValidate_InvalidEmail(t *testing.T) {
	app := newTestApp()
	app.Post("/test", func(c *fiber.Ctx) error {
		var input testInput
		if err := ParseAndValidate(c, &input); err != nil {
			return err
		}
		return c.SendStatus(200)
	})

	body := `{"email":"notanemail","password":"123456"}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestParseAndValidate_InvalidJSON(t *testing.T) {
	app := newTestApp()
	app.Post("/test", func(c *fiber.Ctx) error {
		var input testInput
		if err := ParseAndValidate(c, &input); err != nil {
			return err
		}
		return c.SendStatus(200)
	})

	body := `{invalid json`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}
