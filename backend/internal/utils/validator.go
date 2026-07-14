package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

// ValidateStruct validates struct using validator tags
func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

// ParseAndValidate parses JSON body and validates the resulting struct.
// On failure it writes a detailed response and returns fiber.ErrBadRequest
// so the handler can `return err` without writing the body itself.
func ParseAndValidate(c *fiber.Ctx, input interface{}) error {
	if err := c.BodyParser(input); err != nil {
		return writeBodyParseError(c, err)
	}

	if err := validate.Struct(input); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(Response{
			Success: false,
			Message: "Validation failed",
			Errors:  formatValidationErrors(err),
		})
		return fiber.ErrBadRequest
	}

	return nil
}

// ParseAndValidateParams parses URL params (e.g. /users/:id) into a struct
// and validates it. Mirrors ParseAndValidate's contract for handlers.
func ParseAndValidateParams(c *fiber.Ctx, input interface{}) error {
	if err := c.ParamsParser(input); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(Response{
			Success: false,
			Message: "Invalid URL parameters",
		})
		return fiber.ErrBadRequest
	}

	if err := validate.Struct(input); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(Response{
			Success: false,
			Message: "Validation failed",
			Errors:  formatValidationErrors(err),
		})
		return fiber.ErrBadRequest
	}

	return nil
}

// writeBodyParseError translates BodyParser errors into a friendly response.
func writeBodyParseError(c *fiber.Ctx, err error) error {
	if jsonErr, ok := err.(*json.SyntaxError); ok {
		c.Status(fiber.StatusBadRequest).JSON(Response{
			Success: false,
			Message: fmt.Sprintf("JSON syntax error at position %d", jsonErr.Offset),
		})
		return fiber.ErrBadRequest
	}

	if jsonErr, ok := err.(*json.UnmarshalTypeError); ok {
		c.Status(fiber.StatusBadRequest).JSON(Response{
			Success: false,
			Message: fmt.Sprintf("Invalid type for field '%s': expected %s, got %s",
				jsonErr.Field, jsonErr.Type, jsonErr.Value),
		})
		return fiber.ErrBadRequest
	}

	c.Status(fiber.StatusBadRequest).JSON(Response{
		Success: false,
		Message: "Invalid request body",
	})
	return fiber.ErrBadRequest
}

// formatValidationErrors converts validator.ValidationErrors into the
// API's ValidationError slice with human-readable messages per tag.
func formatValidationErrors(err error) []ValidationError {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return nil
	}

	out := make([]ValidationError, 0, len(validationErrors))
	for _, fe := range validationErrors {
		field := strings.ToLower(fe.Field())
		var message string

		switch fe.Tag() {
		case "required":
			message = fmt.Sprintf("%s is required", field)
		case "email":
			message = "Invalid email format"
		case "min":
			message = fmt.Sprintf("%s must be at least %s characters", field, fe.Param())
		case "max":
			message = fmt.Sprintf("%s must not exceed %s characters", field, fe.Param())
		case "oneof":
			message = fmt.Sprintf("%s must be one of: %s", field, fe.Param())
		case "gtfield":
			message = fmt.Sprintf("%s must be greater than %s", field, fe.Param())
		case "gt":
			message = fmt.Sprintf("%s must be greater than %s", field, fe.Param())
		case "gte":
			message = fmt.Sprintf("%s must be greater than or equal to %s", field, fe.Param())
		case "uuid":
			message = fmt.Sprintf("%s must be a valid UUID", field)
		default:
			message = fmt.Sprintf("Validation failed on '%s'", fe.Tag())
		}

		out = append(out, ValidationError{Field: field, Message: message})
	}
	return out
}
