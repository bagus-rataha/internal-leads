package routes

import (
	"fiber-api-boilerplate/internal/config"
	"fiber-api-boilerplate/internal/container"
	"fiber-api-boilerplate/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

// SetupAPIRoutes configures all API v1 routes
func SetupAPIRoutes(app *fiber.App, cnt interface{}, cfg *config.Config) {
	// Type assertion for container
	c, ok := cnt.(*container.Container)
	if !ok {
		return
	}

	// API v1 group
	api := app.Group("/api/v1")

	// Setup module routes
	setupAuthRoutes(api, c, cfg)
	setupUserRoutes(api, c, cfg)
	setupTeamRoutes(api, c, cfg)
	setupLeadSourceRoutes(api, c, cfg)
	setupServiceTypeRoutes(api, c, cfg)
	setupReferenceRoutes(api, c, cfg)
	setupLeadRoutes(api, c, cfg)
	setupDashboardRoutes(api, c, cfg)
}

// setupAuthRoutes configures authentication routes
func setupAuthRoutes(api fiber.Router, c *container.Container, cfg *config.Config) {
	auth := api.Group("/auth")
	auth.Post("/login", c.AuthHandler.Login)
	auth.Post("/refresh", c.AuthHandler.RefreshToken)
	auth.Post("/logout", c.AuthHandler.Logout)
	auth.Post("/logout-all", middleware.JWTProtected(cfg.JWTAccessSecret), c.AuthHandler.LogoutAll)
}

// setupUserRoutes configures user routes. /me is open to every authenticated
// role. Listing is open to LEADER as well as ADMIN_SALES/SU (the service
// scopes a LEADER's results to their own team); the remaining user-management
// endpoints below stay ADMIN_SALES/SU-only.
func setupUserRoutes(api fiber.Router, c *container.Container, cfg *config.Config) {
	users := api.Group("/users")
	users.Use(middleware.JWTProtected(cfg.JWTAccessSecret))

	users.Get("/me", c.UserHandler.GetProfile)
	users.Put("/me", c.UserHandler.UpdateProfile)

	users.Get("/", middleware.RequireRole("LEADER", "ADMIN_SALES", "SU"), c.UserHandler.ListUsers)

	admin := users.Group("", middleware.RequireRole("ADMIN_SALES", "SU"))
	admin.Post("/", c.UserHandler.CreateUser)
	admin.Get("/:id", c.UserHandler.GetUser)
	admin.Patch("/:id", c.UserHandler.UpdateUser)
	admin.Post("/:id/reset-password", c.UserHandler.ResetPassword)
	admin.Post("/:id/deactivate", c.UserHandler.DeactivateUser)
}

// setupTeamRoutes configures sales team routes (admin-only)
func setupTeamRoutes(api fiber.Router, c *container.Container, cfg *config.Config) {
	teams := api.Group("/teams")
	teams.Use(middleware.JWTProtected(cfg.JWTAccessSecret))
	teams.Use(middleware.RequireRole("ADMIN_SALES", "SU"))

	teams.Get("/", c.TeamHandler.ListTeams)
	teams.Post("/", c.TeamHandler.CreateTeam)
	teams.Get("/:id", c.TeamHandler.GetTeam)
	teams.Patch("/:id", c.TeamHandler.UpdateTeam)
}

// setupLeadSourceRoutes configures lead source management routes
// (admin-only). Distinct from the read-only /refs/lead-sources group, which
// stays open to every role and filtered to active-only for the create-lead
// dropdown.
func setupLeadSourceRoutes(api fiber.Router, c *container.Container, cfg *config.Config) {
	sources := api.Group("/lead-sources")
	sources.Use(middleware.JWTProtected(cfg.JWTAccessSecret))
	sources.Use(middleware.RequireRole("ADMIN_SALES", "SU"))

	sources.Get("/", c.LeadSourceHandler.ListLeadSourcesAdmin)
	sources.Post("/", c.LeadSourceHandler.CreateLeadSource)
	sources.Patch("/:id", c.LeadSourceHandler.UpdateLeadSource)
}

// setupServiceTypeRoutes configures service type management routes
// (admin-only). Distinct from the read-only /refs/service-types group.
func setupServiceTypeRoutes(api fiber.Router, c *container.Container, cfg *config.Config) {
	types := api.Group("/service-types")
	types.Use(middleware.JWTProtected(cfg.JWTAccessSecret))
	types.Use(middleware.RequireRole("ADMIN_SALES", "SU"))

	types.Get("/", c.ServiceTypeHandler.ListServiceTypesAdmin)
	types.Post("/", c.ServiceTypeHandler.CreateServiceType)
	types.Patch("/:id", c.ServiceTypeHandler.UpdateServiceType)
}

// setupReferenceRoutes configures read-only reference routes (lead sources,
// service types, wilayah). Open to every authenticated role — this is
// dropdown data, not lead data, so it carries no ownership scoping.
func setupReferenceRoutes(api fiber.Router, c *container.Container, cfg *config.Config) {
	refs := api.Group("/refs")
	refs.Use(middleware.JWTProtected(cfg.JWTAccessSecret))

	refs.Get("/lead-sources", c.ReferenceHandler.ListLeadSources)
	refs.Get("/service-types", c.ReferenceHandler.ListServiceTypes)
	refs.Get("/provinces", c.ReferenceHandler.ListProvinces)
	refs.Get("/cities", c.ReferenceHandler.ListCities)
	refs.Get("/districts", c.ReferenceHandler.ListDistricts)
	refs.Get("/villages", c.ReferenceHandler.ListVillages)
}

// setupLeadRoutes configures lead and follow-up routes. No RequireRole -
// access is scope-based (every authenticated role may call these; the
// service decides what each caller can see), per ARCHITECTURE.md §9.
// /export is registered before /:code so a literal "export" path segment
// is never captured by the :code param.
func setupLeadRoutes(api fiber.Router, c *container.Container, cfg *config.Config) {
	leads := api.Group("/leads")
	leads.Use(middleware.JWTProtected(cfg.JWTAccessSecret))

	leads.Post("/", c.LeadHandler.CreateLead)
	leads.Get("/", c.LeadHandler.ListLeads)
	leads.Get("/export", c.ExportHandler.ExportLeads)
	leads.Get("/:code", c.LeadHandler.GetLead)
	leads.Patch("/:code", c.LeadHandler.UpdateLead)
	leads.Patch("/:code/status", c.LeadHandler.UpdateLeadStatus)
	leads.Post("/:code/followups", c.FollowUpHandler.CreateFollowUp)
	leads.Get("/:code/followups", c.FollowUpHandler.ListFollowUps)
}

// setupDashboardRoutes configures dashboard routes. No blanket RequireRole -
// every endpoint here is scope-based like /leads, except sales-activity,
// which adds its own role gate below per ARCHITECTURE.md §9.
func setupDashboardRoutes(api fiber.Router, c *container.Container, cfg *config.Config) {
	dash := api.Group("/dashboard")
	dash.Use(middleware.JWTProtected(cfg.JWTAccessSecret))

	dash.Get("/summary", c.DashboardHandler.Summary)
	dash.Get("/activity", c.DashboardHandler.Activity)
	dash.Get("/stale-leads", c.DashboardHandler.StaleLeads)
	dash.Get("/sales-activity", middleware.RequireRole("LEADER", "ADMIN_SALES", "SU"), c.DashboardHandler.SalesActivity)
	dash.Get("/segments", c.DashboardHandler.Segments)
}
