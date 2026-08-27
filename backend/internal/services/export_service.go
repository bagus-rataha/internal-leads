package services

import (
	"errors"
	"fmt"
	"time"

	"fiber-api-boilerplate/internal/models"
	"fiber-api-boilerplate/internal/repository"
	"fiber-api-boilerplate/internal/utils"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// MaxExportRows is the hard cap on how many leads a single export may
// return (ARCHITECTURE.md §10). A var, not a const, so tests can lower it
// instead of inserting 50,000 rows to exercise the cap.
var MaxExportRows int64 = 50000

// ErrExportInvalidPassword is returned when the caller's own re-entered
// password doesn't match their account - the same generic message
// /auth/login uses, so a wrong password never reveals more than that.
var ErrExportInvalidPassword = errors.New("invalid credentials")

// ErrExportTooManyRows carries the actual row count so the handler can
// report it back to the caller (ARCHITECTURE.md §10: "minta user
// mempersempit filter").
type ErrExportTooManyRows struct {
	Count int64
}

func (e *ErrExportTooManyRows) Error() string {
	return fmt.Sprintf("export row limit exceeded: %d rows", e.Count)
}

// ExportService has no repository of its own (ARCHITECTURE.md's explicit
// dashboard/export exception) - it builds the scoped+filtered query
// directly against *gorm.DB, reusing repository.ApplyLeadScope/
// ApplyLeadFilter so the scoping rule itself is never duplicated.
type ExportService struct {
	db       *gorm.DB
	userRepo userRepositoryForLead
}

func NewExportService(db *gorm.DB, userRepo userRepositoryForLead) *ExportService {
	return &ExportService{db: db, userRepo: userRepo}
}

// Export builds the two-sheet workbook for GET /leads/export: password
// re-auth first (cheap, fails fast), then the exact same scope+filter GET
// /leads uses (minus pagination), capped at MaxExportRows, plus every
// follow-up belonging to the matched leads in one bulk query.
func (s *ExportService) Export(callerID uuid.UUID, role, password string, filter repository.LeadFilter) (*excelize.File, error) {
	caller, err := s.userRepo.FindByID(callerID)
	if err != nil || !utils.VerifyPassword(password, caller.Password) {
		return nil, ErrExportInvalidPassword
	}

	scope, err := buildLeadScope(s.userRepo, callerID, role)
	if err != nil {
		return nil, err
	}

	tx := repository.ApplyLeadScope(s.db.Model(&models.Lead{}), scope)
	tx = repository.ApplyLeadFilter(tx, filter)

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, err
	}
	if total > MaxExportRows {
		return nil, &ErrExportTooManyRows{Count: total}
	}

	var leads []models.Lead
	err = tx.Preload("Owner.Team").Preload("Province").Preload("City").
		Preload("District").Preload("Village").Preload("ServiceType").
		Preload("LeadSource").Order("leads.code ASC").Find(&leads).Error
	if err != nil {
		return nil, err
	}

	leadIDs := make([]uuid.UUID, len(leads))
	for i, lead := range leads {
		leadIDs[i] = lead.ID
	}

	var followUps []models.FollowUp
	if len(leadIDs) > 0 {
		err = s.db.Where("lead_id IN ?", leadIDs).Preload("CreatedBy").
			Order("lead_id ASC, created_at ASC").Find(&followUps).Error
		if err != nil {
			return nil, err
		}
	}

	return buildExportWorkbook(leads, followUps), nil
}

const leadsSheetName = "Leads"
const followUpsSheetName = "Follow-up"

var leadsSheetHeader = []interface{}{
	"Kode", "Status", "Nama Perusahaan", "Bidang Usaha", "Website",
	"PIC", "Jabatan PIC", "Telepon Kantor", "Telepon Genggam", "Email",
	"Provinsi", "Kota/Kabupaten", "Kecamatan", "Kelurahan", "RT", "RW", "Alamat",
	"Jenis Layanan", "Kapasitas (Mbps)", "ISP Existing", "Harga", "Forecast MRR", "Layanan Lain",
	"Sumber Lead", "Sales", "Tim", "Jumlah Follow-up", "Follow-up Terakhir",
	"Alasan Lost", "Tanggal Dibuat",
}

var followUpsSheetHeader = []interface{}{"Kode Lead", "Tanggal", "Penulis", "Catatan"}

// buildExportWorkbook is a pure function over already-fetched rows - no DB
// access - so it's covered by a plain unit test rather than the integration
// suite the rest of ExportService needs.
func buildExportWorkbook(leads []models.Lead, followUps []models.FollowUp) *excelize.File {
	leadCodes := make(map[uuid.UUID]string, len(leads))
	for _, lead := range leads {
		leadCodes[lead.ID] = lead.Code
	}

	f := excelize.NewFile()
	f.SetSheetName("Sheet1", leadsSheetName)
	f.NewSheet(followUpsSheetName)

	f.SetSheetRow(leadsSheetName, "A1", &leadsSheetHeader)
	for i := range leads {
		row := leadExportRow(&leads[i])
		f.SetSheetRow(leadsSheetName, fmt.Sprintf("A%d", i+2), &row)
	}

	f.SetSheetRow(followUpsSheetName, "A1", &followUpsSheetHeader)
	for i := range followUps {
		row := followUpExportRow(&followUps[i], leadCodes[followUps[i].LeadID])
		f.SetSheetRow(followUpsSheetName, fmt.Sprintf("A%d", i+2), &row)
	}

	return f
}

func leadExportRow(lead *models.Lead) []interface{} {
	ownerName, teamName := "", ""
	if lead.Owner != nil {
		ownerName = lead.Owner.Name
		if lead.Owner.Team != nil {
			teamName = lead.Owner.Team.Name
		}
	}
	return []interface{}{
		lead.Code,
		lead.Status,
		lead.CompanyName,
		strOrEmpty(lead.BusinessField),
		strOrEmpty(lead.Website),
		strOrEmpty(lead.PicName),
		strOrEmpty(lead.PicPosition),
		strOrEmpty(lead.OfficePhone),
		strOrEmpty(lead.MobilePhone),
		strOrEmpty(lead.Email),
		provinceName(lead.Province),
		cityName(lead.City),
		districtName(lead.District),
		villageName(lead.Village),
		strOrEmpty(lead.Rt),
		strOrEmpty(lead.Rw),
		strOrEmpty(lead.Street),
		serviceTypeName(lead.ServiceType),
		intOrEmpty(lead.CapacityMbps),
		strOrEmpty(lead.ExistingIsp),
		floatOrEmpty(lead.Price),
		floatOrEmpty(lead.ForecastMrr),
		strOrEmpty(lead.OtherServices),
		leadSourceName(lead.LeadSource),
		ownerName,
		teamName,
		lead.FollowUpCount,
		timeOrEmpty(lead.LastFollowUpAt),
		strOrEmpty(lead.LostReason),
		lead.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func followUpExportRow(fu *models.FollowUp, leadCode string) []interface{} {
	authorName := ""
	if fu.CreatedBy != nil {
		authorName = fu.CreatedBy.Name
	}
	return []interface{}{
		leadCode,
		fu.CreatedAt.Format("2006-01-02 15:04:05"),
		authorName,
		fu.Note,
	}
}

func strOrEmpty(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func intOrEmpty(v *int) interface{} {
	if v == nil {
		return ""
	}
	return *v
}

func floatOrEmpty(v *float64) interface{} {
	if v == nil {
		return ""
	}
	return *v
}

func timeOrEmpty(v *time.Time) string {
	if v == nil {
		return ""
	}
	return v.Format("2006-01-02 15:04:05")
}

func provinceName(p *models.Province) string {
	if p == nil {
		return ""
	}
	return p.Name
}

func cityName(c *models.City) string {
	if c == nil {
		return ""
	}
	return c.Name
}

func districtName(d *models.District) string {
	if d == nil {
		return ""
	}
	return d.Name
}

func villageName(v *models.Village) string {
	if v == nil {
		return ""
	}
	return v.Name
}

func serviceTypeName(s *models.ServiceType) string {
	if s == nil {
		return ""
	}
	return s.Name
}

func leadSourceName(l *models.LeadSource) string {
	if l == nil {
		return ""
	}
	return l.Name
}
