//go:build integration

package services

import (
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/models"
	"fiber-api-boilerplate/internal/repository"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLeadCreate_ConcurrentCreates_NoDuplicateCodes drives the real
// LeadService.Create against a live database from multiple goroutines,
// proving NextCode + INSERT genuinely commit atomically per call - two
// concurrent creates must never produce the same code.
func TestLeadCreate_ConcurrentCreates_NoDuplicateCodes(t *testing.T) {
	db := setupServiceTestDB(t)
	leadRepo := repository.NewLeadRepository(db)
	userRepo := repository.NewUserRepository(db)
	referenceRepo := repository.NewReferenceRepository(db)
	svc := NewLeadService(db, leadRepo, userRepo, referenceRepo)

	ownerID := uuid.Must(uuid.NewV7())
	require.NoError(t, db.Create(&models.User{BaseModel: models.BaseModel{ID: ownerID}, Email: "owner@test.local", Password: "h", Name: "Owner", Role: "ADMIN_SALES", IsActive: true}).Error)

	const n = 10
	codes := make([]string, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			result, err := svc.Create(ownerID, "ADMIN_SALES", dto.CreateLeadInput{CompanyName: "Concurrent"})
			require.NoError(t, err)
			codes[i] = result.Code
		}(i)
	}
	wg.Wait()

	seen := make(map[string]bool, n)
	for _, code := range codes {
		assert.False(t, seen[code], "duplicate code generated: %s", code)
		seen[code] = true
	}
}
