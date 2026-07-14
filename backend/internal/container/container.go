package container

import (
	"fiber-api-boilerplate/internal/config"
	"fiber-api-boilerplate/internal/handlers"
	"fiber-api-boilerplate/internal/repository"
	"fiber-api-boilerplate/internal/services"

	"gorm.io/gorm"
)

// Container holds all application dependencies
type Container struct {
	AuthHandler *handlers.AuthHandler
	UserHandler *handlers.UserHandler
	// Add more handlers here as you build features
}

// NewContainer creates and initializes all dependencies
func NewContainer(db *gorm.DB, cfg *config.Config) *Container {
	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, refreshTokenRepo, cfg)
	userService := services.NewUserService(userRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, cfg)
	userHandler := handlers.NewUserHandler(userService)

	return &Container{
		AuthHandler: authHandler,
		UserHandler: userHandler,
	}
}
