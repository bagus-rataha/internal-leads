package utils

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// RefreshTokenCookieName is the cookie key used to transport refresh tokens
const RefreshTokenCookieName = "refreshToken"

// RefreshTokenCookiePath scopes the refresh token cookie to the auth endpoints
// so it is never sent to other parts of the API.
const RefreshTokenCookiePath = "/api/v1/auth"

// SetRefreshTokenCookie writes the refresh token as an HTTP-only cookie
func SetRefreshTokenCookie(c *fiber.Ctx, token string, maxAge time.Duration, secure bool) {
	c.Cookie(&fiber.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    token,
		Path:     RefreshTokenCookiePath,
		Expires:  time.Now().Add(maxAge),
		HTTPOnly: true,
		Secure:   secure,
		SameSite: "Strict",
	})
}

// ClearRefreshTokenCookie removes the refresh token cookie
func ClearRefreshTokenCookie(c *fiber.Ctx, secure bool) {
	c.Cookie(&fiber.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    "",
		Path:     RefreshTokenCookiePath,
		Expires:  time.Now().Add(-time.Hour),
		MaxAge:   -1,
		HTTPOnly: true,
		Secure:   secure,
		SameSite: "Strict",
	})
}
