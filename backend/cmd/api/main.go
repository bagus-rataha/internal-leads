package main

import (
	"fiber-api-boilerplate/internal/app"
	"fiber-api-boilerplate/internal/config"
	"fiber-api-boilerplate/internal/container"
	"fiber-api-boilerplate/internal/routes"
	"log"

	_ "fiber-api-boilerplate/docs" // Swagger docs
)

// @title Fiber API Boilerplate
// @version 3.0
// @description Production-ready REST API with clean architecture
// @host localhost:8000
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Load configuration
	cfg := config.LoadConfig()
	log.Printf("Environment: %s", cfg.AppEnv)

	// Connect to database
	db := config.ConnectDB(cfg)

	// Initialize dependency injection container
	cnt := container.NewContainer(db, cfg)

	// Setup Fiber application
	fiberApp := app.NewFiberApp(cfg)

	// Setup routes
	routes.SetupRoutes(fiberApp, cnt, cfg)

	// Start server
	log.Printf("Server running on port %s", cfg.Port)
	if cfg.IsDevelopment() {
		log.Printf("Swagger UI: http://localhost:%s/swagger/", cfg.Port)
	}
	log.Fatal(fiberApp.Listen(":" + cfg.Port))
}
