package config

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ConnectDB creates database connection and runs migrations
func ConnectDB(cfg *Config) *gorm.DB {
	// Build DSN (Data Source Name)
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort,
	)

	// Set log level based on environment
	logLevel := logger.Silent
	if cfg.IsDevelopment() {
		logLevel = logger.Info
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Get underlying SQL database
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get database instance:", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)                  // Idle connections
	sqlDB.SetMaxOpenConns(100)                 // Max open connections
	sqlDB.SetConnMaxLifetime(time.Hour)        // Connection lifetime
	sqlDB.SetConnMaxIdleTime(10 * time.Minute) // Idle timeout

	log.Println("Database connected successfully")
	log.Println("Schema is managed by golang-migrate. Run 'migrate -path migrations -database \"$DATABASE_URL\" up' to apply pending migrations.")
	return db
}
