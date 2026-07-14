package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	JWTAccessSecret  string
	JWTRefreshSecret string
	JWTAccessExpire  time.Duration
	JWTRefreshExpire time.Duration
	Port             string
	AppEnv           string
	AllowedOrigins   string
	RateLimitMax     int
	RateLimitWindow  time.Duration
}

// LoadConfig loads configuration from environment variables.
//
// Required settings (credentials, connection target, environment) have no
// default — a missing one aborts startup rather than letting the app run with
// placeholder values. Optional settings keep defaults that are genuinely valid
// universal values.
func LoadConfig() *Config {
	// Load .env file if exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Required settings — accumulate all missing keys so the operator sees
	// every problem at once instead of fixing them one by one.
	var missing []string
	dbHost := requireEnv("DB_HOST", &missing)
	dbPort := requireEnv("DB_PORT", &missing)
	dbUser := requireEnv("DB_USER", &missing)
	dbPassword := requireEnv("DB_PASSWORD", &missing)
	dbName := requireEnv("DB_NAME", &missing)
	jwtAccessSecret := requireEnv("JWT_ACCESS_SECRET", &missing)
	jwtRefreshSecret := requireEnv("JWT_REFRESH_SECRET", &missing)
	appEnv := requireEnv("APP_ENV", &missing)

	if len(missing) > 0 {
		log.Fatalf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	if appEnv != "development" && appEnv != "production" {
		log.Fatalf("APP_ENV must be 'development' or 'production', got %q", appEnv)
	}

	// Optional settings — defaults are valid universal values.
	accessExpire, err := time.ParseDuration(getEnv("JWT_ACCESS_EXPIRE", "24h"))
	if err != nil {
		log.Fatal("Invalid JWT_ACCESS_EXPIRE format")
	}

	refreshExpire, err := time.ParseDuration(getEnv("JWT_REFRESH_EXPIRE", "168h"))
	if err != nil {
		log.Fatal("Invalid JWT_REFRESH_EXPIRE format")
	}

	rateLimitMax, err := strconv.Atoi(getEnv("RATE_LIMIT_MAX", "100"))
	if err != nil {
		log.Fatal("Invalid RATE_LIMIT_MAX format")
	}

	rateLimitWindow, err := time.ParseDuration(getEnv("RATE_LIMIT_WINDOW", "1m"))
	if err != nil {
		log.Fatal("Invalid RATE_LIMIT_WINDOW format")
	}

	cfg := &Config{
		DBHost:           dbHost,
		DBPort:           dbPort,
		DBUser:           dbUser,
		DBPassword:       dbPassword,
		DBName:           dbName,
		JWTAccessSecret:  jwtAccessSecret,
		JWTRefreshSecret: jwtRefreshSecret,
		JWTAccessExpire:  accessExpire,
		JWTRefreshExpire: refreshExpire,
		Port:             getEnv("PORT", "8000"),
		AppEnv:           appEnv,
		AllowedOrigins:   getEnv("ALLOWED_ORIGINS", ""),
		RateLimitMax:     rateLimitMax,
		RateLimitWindow:  rateLimitWindow,
	}

	// ALLOWED_ORIGINS is only consumed in production (CORS). Require it there
	// so a misconfigured prod deploy fails loudly instead of allowing all.
	if cfg.IsProduction() && cfg.AllowedOrigins == "" {
		log.Fatal("ALLOWED_ORIGINS must be set when APP_ENV=production")
	}

	return cfg
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.AppEnv == "development"
}

// getEnv gets environment variable or returns default value.
// Use only for optional settings whose default is a genuinely valid value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// requireEnv returns the environment variable's value and, when it is unset,
// records the key into missing. Required settings have no safe default — the
// caller aborts startup once all missing keys are collected.
func requireEnv(key string, missing *[]string) string {
	value := os.Getenv(key)
	if value == "" {
		*missing = append(*missing, key)
	}
	return value
}
