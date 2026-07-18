//go:build integration

package repository

import (
	"fiber-api-boilerplate/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFollowUpRepository_Create_ListByLeadID(t *testing.T) {
	db := setupTestDB(t)
	leadRepo := NewLeadRepository(db)
	followUpRepo := NewFollowUpRepository(db)

	userID := seedTestUser(t, db)
	lead := &models.Lead{Code: "LD-2607-2001", OwnerID: userID, CreatedByID: userID, CompanyName: "A"}
	require.NoError(t, leadRepo.Create(lead))

	f1 := &models.FollowUp{LeadID: lead.ID, Note: "first contact", CreatedByID: userID}
	f2 := &models.FollowUp{LeadID: lead.ID, Note: "second contact", CreatedByID: userID}
	require.NoError(t, followUpRepo.Create(f1))
	require.NoError(t, followUpRepo.Create(f2))

	results, err := followUpRepo.ListByLeadID(lead.ID)
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestFollowUpRepository_ListByLeadID_OnlyThatLead(t *testing.T) {
	db := setupTestDB(t)
	leadRepo := NewLeadRepository(db)
	followUpRepo := NewFollowUpRepository(db)

	userID := seedTestUser(t, db)
	leadA := &models.Lead{Code: "LD-2607-2101", OwnerID: userID, CreatedByID: userID, CompanyName: "A"}
	leadB := &models.Lead{Code: "LD-2607-2102", OwnerID: userID, CreatedByID: userID, CompanyName: "B"}
	require.NoError(t, leadRepo.Create(leadA))
	require.NoError(t, leadRepo.Create(leadB))
	require.NoError(t, followUpRepo.Create(&models.FollowUp{LeadID: leadA.ID, Note: "for A", CreatedByID: userID}))
	require.NoError(t, followUpRepo.Create(&models.FollowUp{LeadID: leadB.ID, Note: "for B", CreatedByID: userID}))

	results, err := followUpRepo.ListByLeadID(leadA.ID)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "for A", results[0].Note)
}

func TestFollowUpRepository_ListByLeadID_PreloadsCreatedBy(t *testing.T) {
	db := setupTestDB(t)
	leadRepo := NewLeadRepository(db)
	followUpRepo := NewFollowUpRepository(db)
	userID := seedTestUser(t, db)
	lead := &models.Lead{Code: "LD-2607-1301", OwnerID: userID, CreatedByID: userID, CompanyName: "Acme"}
	require.NoError(t, leadRepo.Create(lead))
	require.NoError(t, followUpRepo.Create(&models.FollowUp{LeadID: lead.ID, Note: "first contact", CreatedByID: userID}))

	results, err := followUpRepo.ListByLeadID(lead.ID)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.NotNil(t, results[0].CreatedBy, "CreatedBy must be preloaded")
	assert.Equal(t, "Test User", results[0].CreatedBy.Name)
}
