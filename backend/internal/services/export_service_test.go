package services

import (
	"testing"
	"time"

	"fiber-api-boilerplate/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildExportWorkbook_LeadsSheet_HasNamesNotIDs(t *testing.T) {
	leadID := uuid.Must(uuid.NewV7())
	lastFollowUp := time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC)
	lead := models.Lead{
		BaseModel:      models.BaseModel{ID: leadID, CreatedAt: time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC)},
		Code:           "LD-2607-0001",
		Status:         "FOLLOW_UP",
		CompanyName:    "PT Contoh",
		FollowUpCount:  2,
		LastFollowUpAt: &lastFollowUp,
		Owner:          &models.User{Name: "Sales A", Team: &models.SalesTeam{Name: "Team A"}},
		Province:       &models.Province{Name: "DKI Jakarta"},
		City:           &models.City{Name: "Jakarta Selatan"},
	}

	f := buildExportWorkbook([]models.Lead{lead}, nil)

	rows, err := f.GetRows(leadsSheetName)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, "LD-2607-0001", rows[1][0])
	assert.Equal(t, "PT Contoh", rows[1][2])
	assert.Equal(t, "DKI Jakarta", rows[1][10])
	assert.Equal(t, "Jakarta Selatan", rows[1][11])
	assert.Equal(t, "Sales A", rows[1][24])
	assert.Equal(t, "Team A", rows[1][25])
	assert.Equal(t, "2", rows[1][26])
}

func TestBuildExportWorkbook_LeadsSheet_NilRelationsBecomeEmptyNotPanic(t *testing.T) {
	lead := models.Lead{
		BaseModel:   models.BaseModel{ID: uuid.Must(uuid.NewV7()), CreatedAt: time.Now()},
		Code:        "LD-2607-0002",
		Status:      "BARU",
		CompanyName: "PT Tanpa Relasi",
	}

	f := buildExportWorkbook([]models.Lead{lead}, nil)

	rows, err := f.GetRows(leadsSheetName)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, "", rows[1][10]) // province
	assert.Equal(t, "", rows[1][24]) // sales
	assert.Equal(t, "", rows[1][25]) // team
}

func TestBuildExportWorkbook_FollowUpSheet_UsesLeadCodeNotID(t *testing.T) {
	leadID := uuid.Must(uuid.NewV7())
	lead := models.Lead{BaseModel: models.BaseModel{ID: leadID}, Code: "LD-2607-0003"}
	followUp := models.FollowUp{
		LeadID:    leadID,
		Note:      "Follow up pertama",
		CreatedAt: time.Date(2026, 7, 10, 8, 30, 0, 0, time.UTC),
		CreatedBy: &models.User{Name: "Sales B"},
	}

	f := buildExportWorkbook([]models.Lead{lead}, []models.FollowUp{followUp})

	rows, err := f.GetRows(followUpsSheetName)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, "LD-2607-0003", rows[1][0])
	assert.Equal(t, "Sales B", rows[1][2])
	assert.Equal(t, "Follow up pertama", rows[1][3])
}

func TestBuildExportWorkbook_LeadsSheet_ForecastMrrColumn(t *testing.T) {
	forecast := 4500000.0
	withForecast := models.Lead{
		BaseModel:   models.BaseModel{ID: uuid.Must(uuid.NewV7()), CreatedAt: time.Now()},
		Code:        "LD-2607-0004",
		Status:      "FOLLOW_UP",
		CompanyName: "PT Forecast",
		ForecastMrr: &forecast,
	}
	withoutForecast := models.Lead{
		BaseModel:   models.BaseModel{ID: uuid.Must(uuid.NewV7()), CreatedAt: time.Now()},
		Code:        "LD-2607-0005",
		Status:      "BARU",
		CompanyName: "PT Tanpa Forecast",
	}

	f := buildExportWorkbook([]models.Lead{withForecast, withoutForecast}, nil)

	rows, err := f.GetRows(leadsSheetName)
	require.NoError(t, err)
	require.Len(t, rows, 3)
	assert.Equal(t, "Forecast MRR", rows[0][21])
	assert.Equal(t, "4500000", rows[1][21])
	assert.Equal(t, "", rows[2][21], "nil forecast_mrr becomes an empty cell, not 0 or an error")
}
