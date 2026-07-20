package routes

import (
	"io"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newStaticTestApp registers a fake API route and health check the same way
// SetupRoutes does, then SetupStaticRoutes on top - the same order
// production uses, so these tests exercise the real "API wins, static is
// the trailing catch-all" behavior instead of just the middleware alone.
func newStaticTestApp(spaFS fstest.MapFS) *fiber.App {
	app := fiber.New()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	api := app.Group("/api/v1")
	api.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})

	SetupStaticRoutes(app, spaFS)
	return app
}

func TestSetupStaticRoutes_APIRouteNotShadowed(t *testing.T) {
	app := newStaticTestApp(fstest.MapFS{
		"index.html": {Data: []byte("<html>spa</html>")},
	})

	req := httptest.NewRequest(fiber.MethodGet, "/api/v1/ping", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "pong", string(body))
}

func TestSetupStaticRoutes_HealthNotShadowed(t *testing.T) {
	app := newStaticTestApp(fstest.MapFS{
		"index.html": {Data: []byte("<html>spa</html>")},
	})

	req := httptest.NewRequest(fiber.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestSetupStaticRoutes_UnknownPathFallsBackToIndex(t *testing.T) {
	app := newStaticTestApp(fstest.MapFS{
		"index.html": {Data: []byte("<html>spa-shell</html>")},
	})

	// A client-side route (e.g. the lead detail screen) has no matching
	// file on disk - the fallback must still return the SPA shell so React
	// Router can take over, not a 404.
	req := httptest.NewRequest(fiber.MethodGet, "/leads/LD-2607-0001", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "<html>spa-shell</html>", string(body))
}

func TestSetupStaticRoutes_KnownAssetServedDirectly(t *testing.T) {
	app := newStaticTestApp(fstest.MapFS{
		"index.html":    {Data: []byte("<html>spa</html>")},
		"assets/app.js": {Data: []byte("console.log('hi')")},
	})

	req := httptest.NewRequest(fiber.MethodGet, "/assets/app.js", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "console.log('hi')", string(body))
}

func TestSetupStaticRoutes_NoBuildEmbedded_UnknownPathReturns404(t *testing.T) {
	// Mirrors backend-only local development: dist/ has only .gitkeep, no
	// real frontend build, so there is no index.html to fall back to.
	app := newStaticTestApp(fstest.MapFS{
		".gitkeep": {Data: []byte{}},
	})

	req := httptest.NewRequest(fiber.MethodGet, "/leads", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}
