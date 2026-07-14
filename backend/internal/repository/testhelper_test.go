//go:build integration

package repository

import (
	"os"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupTestDB connects to the dedicated test database, applies the existing
// golang-migrate migrations, and truncates tables so each test starts clean.
//
// The connection string is read from TEST_DATABASE_URL so no credentials live
// in source. Set it via your local .env (gitignored) or, in CI/CD, as a masked
// GitLab CI/CD variable. Expected URL form (used for both GORM and migrate):
//
//	postgres://user:password@host:port/dbname?sslmode=disable
//
// URL-encode special characters in the password (e.g. '@' -> '%40').
func setupTestDB(t *testing.T) *gorm.DB {
	// Convenience for local runs: load .env if present. Real environment
	// variables (e.g. from CI) always take precedence and are not overridden.
	_ = godotenv.Load("../../.env")

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Fatal("TEST_DATABASE_URL is not set; it is required to run repository integration tests")
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		t.Fatal("Failed to connect test DB:", err)
	}

	m, err := migrate.New("file://../../migrations", dbURL)
	if err != nil {
		t.Fatal("Failed to init migrate:", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatal("Failed to run migrations:", err)
	}

	truncateAllTables(t, db)

	return db
}

// truncateAllTables empties every table in the public schema so each test
// starts from a clean slate. Tables are discovered dynamically, so new models
// are cleaned automatically without updating this helper. The migrate
// bookkeeping table (schema_migrations) is excluded — truncating it would make
// golang-migrate re-run every migration.
func truncateAllTables(t *testing.T, db *gorm.DB) {
	var tables []string
	if err := db.Raw(`SELECT tablename FROM pg_tables
		WHERE schemaname = 'public' AND tablename <> 'schema_migrations'`).
		Scan(&tables).Error; err != nil {
		t.Fatal("Failed to list tables:", err)
	}

	if len(tables) == 0 {
		return
	}

	if err := db.Exec("TRUNCATE " + strings.Join(tables, ", ") + " RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatal("Failed to truncate tables:", err)
	}
}
