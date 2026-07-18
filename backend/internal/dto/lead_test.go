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

func TestToLeadResponse_CityName_NilCity_Nil(t *testing.T) {
	lead := &models.Lead{Status: "BARU"}
	assert.Nil(t, ToLeadResponse(lead).CityName)
}

func TestToLeadResponse_CityName_PopulatedCity(t *testing.T) {
	lead := &models.Lead{Status: "BARU", City: &models.City{Name: "Jakarta Selatan"}}
	if assert.NotNil(t, ToLeadResponse(lead).CityName) {
		assert.Equal(t, "Jakarta Selatan", *ToLeadResponse(lead).CityName)
	}
}
