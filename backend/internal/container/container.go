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
	AuthHandler        *handlers.AuthHandler
	UserHandler        *handlers.UserHandler
	TeamHandler        *handlers.TeamHandler
	LeadSourceHandler  *handlers.LeadSourceHandler
	ServiceTypeHandler *handlers.ServiceTypeHandler
	ReferenceHandler   *handlers.ReferenceHandler
	LeadHandler        *handlers.LeadHandler
	FollowUpHandler    *handlers.FollowUpHandler
	// Add more handlers here as you build features
}

// NewContainer creates and initializes all dependencies
func NewContainer(db *gorm.DB, cfg *config.Config) *Container {
	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	salesTeamRepo := repository.NewSalesTeamRepository(db)
	leadSourceRepo := repository.NewLeadSourceRepository(db)
	serviceTypeRepo := repository.NewServiceTypeRepository(db)
	referenceRepo := repository.NewReferenceRepository(db)
	leadRepo := repository.NewLeadRepository(db)
	followUpRepo := repository.NewFollowUpRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, refreshTokenRepo, cfg)
	userService := services.NewUserService(db, userRepo, refreshTokenRepo, salesTeamRepo)
	teamService := services.NewTeamService(salesTeamRepo)
	leadSourceService := services.NewLeadSourceService(leadSourceRepo)
	serviceTypeService := services.NewServiceTypeService(serviceTypeRepo)
	referenceService := services.NewReferenceService(referenceRepo)
	leadService := services.NewLeadService(db, leadRepo, userRepo, referenceRepo)
	followUpService := services.NewFollowUpService(db, followUpRepo, leadRepo, userRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, cfg)
	userHandler := handlers.NewUserHandler(userService)
	teamHandler := handlers.NewTeamHandler(teamService)
	leadSourceHandler := handlers.NewLeadSourceHandler(leadSourceService)
	serviceTypeHandler := handlers.NewServiceTypeHandler(serviceTypeService)
	referenceHandler := handlers.NewReferenceHandler(referenceService)
	leadHandler := handlers.NewLeadHandler(leadService)
	followUpHandler := handlers.NewFollowUpHandler(followUpService)

	return &Container{
		AuthHandler:        authHandler,
		UserHandler:        userHandler,
		TeamHandler:        teamHandler,
		LeadSourceHandler:  leadSourceHandler,
		ServiceTypeHandler: serviceTypeHandler,
		ReferenceHandler:   referenceHandler,
		LeadHandler:        leadHandler,
		FollowUpHandler:    followUpHandler,
	}
}
