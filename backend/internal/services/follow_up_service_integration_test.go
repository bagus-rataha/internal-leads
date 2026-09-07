//go:build integration

package services

import (
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/models"
	"fiber-api-boilerplate/internal/repository"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFollowUpCreate_Transaction_AllFourEffectsCommitTogether drives the
// real FollowUpService.Create against a live database, proving the insert
// and all three lead-side updates actually commit as one unit.
func TestFollowUpCreate_Transaction_AllFourEffectsCommitTogether(t *testing.T) {
	db := setupServiceTestDB(t)
	leadRepo := repository.NewLeadRepository(db)
	followUpRepo := repository.NewFollowUpRepository(db)
	userRepo := repository.NewUserRepository(db)

	ownerID := uuid.Must(uuid.NewV7())
	require.NoError(t, db.Create(&models.User{BaseModel: models.BaseModel{ID: ownerID}, Email: "o2@test.local", Password: "h", Name: "Owner", Role: "ADMIN_SALES", IsActive: true}).Error)
	lead := &models.Lead{Code: "LD-2607-4001", OwnerID: ownerID, CreatedByID: ownerID, CompanyName: "A", Status: "BARU"}
	require.NoError(t, leadRepo.Create(lead))

	svc := NewFollowUpService(db, followUpRepo, leadRepo, userRepo)
	result, err := svc.Create(ownerID, "ADMIN_SALES", lead.Code, dto.CreateFollowUpInput{Note: "first contact"})
	require.NoError(t, err)
	assert.Equal(t, "first contact", result.Note)

	found, err := leadRepo.FindByCode(repository.LeadScope{}, lead.Code)
	require.NoError(t, err)
	assert.Equal(t, "FOLLOW_UP", found.Status)
	assert.Equal(t, 1, found.FollowUpCount)
	assert.NotNil(t, found.LastFollowUpAt)

	followUps, err := followUpRepo.ListByLeadID(lead.ID)
	require.NoError(t, err)
	assert.Len(t, followUps, 1)
}

// TestFollowUpCreate_SecondFollowUp_DoesNotChangeAlreadyAdvancedStatus
// proves MaybeTransitionToFollowUp's WHERE-guarded update doesn't disturb a
// lead that's already past FOLLOW_UP, while follow_up_count keeps counting.
func TestFollowUpCreate_SecondFollowUp_DoesNotChangeAlreadyAdvancedStatus(t *testing.T) {
	db := setupServiceTestDB(t)
	leadRepo := repository.NewLeadRepository(db)
	followUpRepo := repository.NewFollowUpRepository(db)
	userRepo := repository.NewUserRepository(db)

	ownerID := uuid.Must(uuid.NewV7())
	require.NoError(t, db.Create(&models.User{BaseModel: models.BaseModel{ID: ownerID}, Email: "o3@test.local", Password: "h", Name: "Owner", Role: "ADMIN_SALES", IsActive: true}).Error)
	lead := &models.Lead{Code: "LD-2607-4002", OwnerID: ownerID, CreatedByID: ownerID, CompanyName: "A", Status: "SURVEY"}
	require.NoError(t, leadRepo.Create(lead))

	svc := NewFollowUpService(db, followUpRepo, leadRepo, userRepo)
	_, err := svc.Create(ownerID, "ADMIN_SALES", lead.Code, dto.CreateFollowUpInput{Note: "note after survey"})
	require.NoError(t, err)

	found, err := leadRepo.FindByCode(repository.LeadScope{}, lead.Code)
	require.NoError(t, err)
	assert.Equal(t, "SURVEY", found.Status, "already-advanced status must not be pulled back")
	assert.Equal(t, 1, found.FollowUpCount, "count still increments regardless of status")
}
