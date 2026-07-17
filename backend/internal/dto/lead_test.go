package dto

import (
	"fiber-api-boilerplate/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

// is_stale computation lives in services/lead_service.go now (it needs
// repository.StaleLeadThresholdDays, and dto must not import repository) -
// see TestIsLeadStale_* in lead_service_test.go for that coverage.

func TestToLeadResponse_OwnerName_NilOwner_EmptyString(t *testing.T) {
	lead := &models.Lead{Status: "BARU"}
	assert.Equal(t, "", ToLeadResponse(lead).OwnerName)
}

func TestToLeadResponse_OwnerName_PopulatedOwner(t *testing.T) {
	lead := &models.Lead{Status: "BARU", Owner: &models.User{Name: "Budi Santoso"}}
	assert.Equal(t, "Budi Santoso", ToLeadResponse(lead).OwnerName)
}
