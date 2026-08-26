package services

import (
	"errors"
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/models"
	"fiber-api-boilerplate/internal/repository"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// ErrLeadNotFound is returned when a lead doesn't exist OR exists outside
// the caller's scope - the two cases are indistinguishable on purpose, so
// existence never leaks. The handler maps this to 404.
var ErrLeadNotFound = errors.New("lead not found")

// ErrOwnerNotTeamMember is returned when a LEADER submits an owner_id that
// isn't a member of their own team. The handler maps this to 400.
var ErrOwnerNotTeamMember = errors.New("owner is not a member of your team")

// ErrInvalidStatusTransition is returned by UpdateStatus for any transition
// not in the ARCHITECTURE.md §9 state machine, including BARU->FOLLOW_UP
// (which only happens through the follow-up endpoint). The handler maps
// this to 422.
var ErrInvalidStatusTransition = errors.New("invalid status transition")

// isLeadStale mirrors the SQL stale predicate in repository.ApplyLeadFilter
// exactly, so the query-param filter and this per-row flag can never
// disagree: status still active (BARU/FOLLOW_UP) and the last activity
// (last_follow_up_at, falling back to created_at) is older than
// repository.StaleLeadThresholdDays. now is passed in explicitly (rather than
// calling time.Now() internally) so tests can pin the clock. Lives here
// rather than in dto because dto must not import repository.
func isLeadStale(lead *models.Lead, now time.Time) bool {
	if lead.Status != "BARU" && lead.Status != "FOLLOW_UP" {
		return false
	}
	lastActivity := lead.CreatedAt
	if lead.LastFollowUpAt != nil {
		lastActivity = *lead.LastFollowUpAt
	}
	return lastActivity.Before(now.AddDate(0, 0, -repository.StaleLeadThresholdDays))
}

// ErrInvalidReference is returned when a create/update write fails a
// foreign-key constraint (invalid province_id/city_id/.../lead_source_id).
// Per ARCHITECTURE.md, these fields aren't existence-checked in the service
// - the DB's FK constraint is the only guard - so its rejection is
// translated here into a generic error rather than leaking constraint/table
// names to the client. The handler's default case already maps this to 400.
var ErrInvalidReference = errors.New("invalid reference id")

// leadRepositoryForLead is the subset of repository methods LeadService
// needs. Defined consumer-side for testability.
type leadRepositoryForLead interface {
	NextCode(month time.Time) (string, error)
	Create(lead *models.Lead) error
	List(scope repository.LeadScope, filter repository.LeadFilter) ([]models.Lead, int64, error)
	FindByCode(scope repository.LeadScope, code string) (*models.Lead, error)
	FindDetailByCode(scope repository.LeadScope, code string) (*models.Lead, error)
	Update(lead *models.Lead) error
}

// userRepositoryForLead is the subset of user-repository methods needed for
// LEADER team-scope lookup and owner-membership validation. Satisfied by the
// same *UserRepository already used elsewhere - no new repository type.
type userRepositoryForLead interface {
	FindByID(id uuid.UUID) (*models.User, error)
}

// cityRepositoryForLead is the subset needed to populate city_name on a
// lead that was just created in memory - List/FindByCode get City for free
// via Preload, but a freshly-inserted lead hasn't gone through one. Satisfied
// by the same *ReferenceRepository already used for /refs/cities - no new
// repository type.
type cityRepositoryForLead interface {
	FindCityByID(id int) (*models.City, error)
}

// buildLeadScope translates a caller's identity into a LeadScope. Shared by
// LeadService and FollowUpService so the role->scope mapping has exactly one
// definition. SALES scopes to self without touching the DB; LEADER looks up
// their own team_id (the one DB touch ARCHITECTURE.md §5 allows for auth);
// ADMIN_SALES/SU get an unrestricted scope.
func buildLeadScope(userRepo userRepositoryForLead, callerID uuid.UUID, role string) (repository.LeadScope, error) {
	switch role {
	case "SALES":
		return repository.LeadScope{OwnerID: &callerID}, nil
	case "LEADER":
		caller, err := userRepo.FindByID(callerID)
		if err != nil {
			return repository.LeadScope{}, errors.New("caller not found")
		}
		if caller.TeamID == nil {
			// Fail closed: a nil team_id must never produce an unrestricted
			// scope (both OwnerID and TeamID nil reads as ADMIN_SALES/SU to
			// ApplyLeadScope). A DB CHECK constraint keeps this unreachable
			// today, but the scope-building logic shouldn't rely on that.
			return repository.LeadScope{}, errors.New("leader has no team")
		}
		return repository.LeadScope{TeamID: caller.TeamID}, nil
	case "ADMIN_SALES", "SU":
		return repository.LeadScope{}, nil
	default:
		// Fail closed: an unrecognized role must never fall through to the
		// unrestricted ADMIN_SALES/SU scope. Should never happen in
		// production since role comes from a validated JWT claim.
		return repository.LeadScope{}, errors.New("unrecognized role")
	}
}

type LeadService struct {
	db       *gorm.DB
	leadRepo leadRepositoryForLead
	userRepo userRepositoryForLead
	cityRepo cityRepositoryForLead
}

func NewLeadService(db *gorm.DB, leadRepo leadRepositoryForLead, userRepo userRepositoryForLead, cityRepo cityRepositoryForLead) *LeadService {
	return &LeadService{db: db, leadRepo: leadRepo, userRepo: userRepo, cityRepo: cityRepo}
}

// run wraps a multi-step write in a single transaction, mirroring
// UserService.run - used here for Create, where code generation and the
// INSERT must commit together.
func (s *LeadService) run(fn func(tx *gorm.DB) error) error {
	return s.db.Transaction(fn)
}

// Create generates a code and inserts a lead, in one transaction so two
// concurrent creates can never produce colliding codes.
func (s *LeadService) Create(callerID uuid.UUID, role string, input dto.CreateLeadInput) (*dto.LeadResponse, error) {
	var result *dto.LeadResponse
	err := s.run(func(tx *gorm.DB) error {
		txLeadRepo := repository.NewLeadRepository(tx)
		response, err := s.createLead(txLeadRepo, s.userRepo, callerID, role, input)
		if err != nil {
			return err
		}
		result = response
		return nil
	})
	return result, err
}

// createLead holds the actual logic, taking leadRepo as a parameter rather
// than reading it off s - in production s.Create supplies a tx-scoped repo
// so code generation and the insert commit together; in tests this can be
// called directly with a mock, without needing a real *gorm.DB transaction.
func (s *LeadService) createLead(
	leadRepo leadRepositoryForLead,
	userRepo userRepositoryForLead,
	callerID uuid.UUID,
	role string,
	input dto.CreateLeadInput,
) (*dto.LeadResponse, error) {
	ownerID, err := resolveOwnerID(userRepo, callerID, role, input.OwnerID)
	if err != nil {
		return nil, err
	}

	code, err := leadRepo.NextCode(time.Now())
	if err != nil {
		return nil, err
	}

	lead := &models.Lead{
		Code:          code,
		OwnerID:       ownerID,
		CreatedByID:   callerID,
		CompanyName:   input.CompanyName,
		BusinessField: input.BusinessField,
		Website:       input.Website,
		ProvinceID:    input.ProvinceID,
		CityID:        input.CityID,
		DistrictID:    input.DistrictID,
		VillageID:     input.VillageID,
		ZipID:         input.ZipID,
		Rt:            input.Rt,
		Rw:            input.Rw,
		Street:        input.Street,
		PicName:       input.PicName,
		PicPosition:   input.PicPosition,
		OfficePhone:   input.OfficePhone,
		MobilePhone:   input.MobilePhone,
		Email:         input.Email,
		ServiceTypeID: input.ServiceTypeID,
		CapacityMbps:  input.CapacityMbps,
		ExistingIsp:   input.ExistingIsp,
		Price:         input.Price,
		OtherServices: input.OtherServices,
		LeadSourceID:  input.LeadSourceID,
		ForecastMrr:   input.ForecastMrr,
	}

	if err := leadRepo.Create(lead); err != nil {
		return nil, translateWriteError(err)
	}

	// Populate Owner for the response after Create (not before) - setting a
	// non-nil association before Create would make GORM try to upsert the
	// User row too. lead is in-memory only here, so this is exactly one
	// extra fetch, not a Preload/N+1 concern.
	owner, err := userRepo.FindByID(lead.OwnerID)
	if err != nil {
		return nil, err
	}
	lead.Owner = owner

	// Same reasoning as Owner above, but skipped entirely when CityID is nil
	// (unlike OwnerID, city_id is optional - a lead with no address filled in
	// legitimately has no city).
	if lead.CityID != nil {
		city, err := s.cityRepo.FindCityByID(*lead.CityID)
		if err != nil {
			return nil, err
		}
		lead.City = city
	}

	response := dto.ToLeadResponse(lead)
	response.IsStale = isLeadStale(lead, time.Now())
	return &response, nil
}

// translateWriteError maps a Postgres foreign-key-violation error (SQLSTATE
// 23503) - the only guard against an invalid province_id/city_id/.../
// lead_source_id, per ARCHITECTURE.md's no-existence-check rule - to the
// generic ErrInvalidReference, so the raw constraint/table name never
// reaches a client. Any other error passes through unchanged.
func translateWriteError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return ErrInvalidReference
	}
	return err
}

// resolveOwnerID applies the role-dependent owner_id rule from
// ARCHITECTURE.md §9: SALES is always forced to self regardless of what was
// submitted; LEADER's submission must be a member of their own team, or
// defaults to self if omitted; ADMIN_SALES/SU may set anyone, defaulting to
// self if omitted.
func resolveOwnerID(userRepo userRepositoryForLead, callerID uuid.UUID, role string, submitted *uuid.UUID) (uuid.UUID, error) {
	if role == "SALES" {
		return callerID, nil
	}

	if submitted == nil {
		return callerID, nil
	}

	if role == "LEADER" {
		leader, err := userRepo.FindByID(callerID)
		if err != nil {
			return uuid.Nil, errors.New("caller not found")
		}
		target, err := userRepo.FindByID(*submitted)
		if err != nil {
			return uuid.Nil, errors.New("owner not found")
		}
		if leader.TeamID == nil || target.TeamID == nil || *leader.TeamID != *target.TeamID {
			return uuid.Nil, ErrOwnerNotTeamMember
		}
	}

	return *submitted, nil
}

// List returns a scoped, filtered, paginated page of leads.
func (s *LeadService) List(callerID uuid.UUID, role string, query dto.LeadListQuery) (*dto.PaginatedLeadResponse, error) {
	scope, err := buildLeadScope(s.userRepo, callerID, role)
	if err != nil {
		return nil, err
	}

	filter := repository.LeadFilter{
		Q:            query.Q,
		Status:       query.Status,
		SourceID:     query.SourceID,
		TeamID:       query.TeamID,
		OwnerID:      query.OwnerID,
		ProvinceID:   query.ProvinceID,
		CityID:       query.CityID,
		DateFrom:     query.DateFrom,
		DateTo:       query.DateTo,
		FollowUpFrom: query.FollowUpFrom,
		FollowUpTo:   query.FollowUpTo,
		Stale:        query.Stale,
		Sort:         query.Sort,
		Page:         query.Page,
		Limit:        query.Limit,
	}

	leads, total, err := s.leadRepo.List(scope, filter)
	if err != nil {
		return nil, err
	}

	items := dto.ToLeadResponseList(leads)
	now := time.Now()
	for i := range items {
		items[i].IsStale = isLeadStale(&leads[i], now)
	}

	return &dto.PaginatedLeadResponse{
		Items: items,
		Total: total,
		Page:  query.Page,
		Limit: query.Limit,
	}, nil
}

// FindByCode fetches a single lead, scoped.
func (s *LeadService) FindByCode(callerID uuid.UUID, role, code string) (*dto.LeadDetailResponse, error) {
	scope, err := buildLeadScope(s.userRepo, callerID, role)
	if err != nil {
		return nil, err
	}

	lead, err := s.leadRepo.FindDetailByCode(scope, code)
	if err != nil {
		return nil, ErrLeadNotFound
	}

	response := dto.ToLeadDetailResponse(lead)
	response.IsStale = isLeadStale(lead, time.Now())
	return &response, nil
}

// Update applies submitted fields to an existing lead, scoped. owner_id
// reassignment is only honored for ADMIN_SALES/SU; any other role
// submitting it is silently ignored, matching create's posture.
func (s *LeadService) Update(callerID uuid.UUID, role, code string, input dto.UpdateLeadInput) (*dto.LeadResponse, error) {
	scope, err := buildLeadScope(s.userRepo, callerID, role)
	if err != nil {
		return nil, err
	}

	lead, err := s.leadRepo.FindByCode(scope, code)
	if err != nil {
		return nil, ErrLeadNotFound
	}

	applyLeadUpdate(lead, input, role)

	if err := s.leadRepo.Update(lead); err != nil {
		return nil, translateWriteError(err)
	}

	response := dto.ToLeadResponse(lead)
	response.IsStale = isLeadStale(lead, time.Now())
	return &response, nil
}

// applyLeadUpdate mutates lead in place with every submitted field.
func applyLeadUpdate(lead *models.Lead, input dto.UpdateLeadInput, role string) {
	if input.CompanyName != nil {
		lead.CompanyName = *input.CompanyName
	}
	if input.BusinessField != nil {
		lead.BusinessField = input.BusinessField
	}
	if input.Website != nil {
		lead.Website = input.Website
	}
	if input.ProvinceID != nil {
		lead.ProvinceID = input.ProvinceID
	}
	if input.CityID != nil {
		lead.CityID = input.CityID
	}
	if input.DistrictID != nil {
		lead.DistrictID = input.DistrictID
	}
	if input.VillageID != nil {
		lead.VillageID = input.VillageID
	}
	if input.ZipID != nil {
		lead.ZipID = input.ZipID
	}
	if input.Rt != nil {
		lead.Rt = input.Rt
	}
	if input.Rw != nil {
		lead.Rw = input.Rw
	}
	if input.Street != nil {
		lead.Street = input.Street
	}
	if input.PicName != nil {
		lead.PicName = input.PicName
	}
	if input.PicPosition != nil {
		lead.PicPosition = input.PicPosition
	}
	if input.OfficePhone != nil {
		lead.OfficePhone = input.OfficePhone
	}
	if input.MobilePhone != nil {
		lead.MobilePhone = input.MobilePhone
	}
	if input.Email != nil {
		lead.Email = input.Email
	}
	if input.ServiceTypeID != nil {
		lead.ServiceTypeID = input.ServiceTypeID
	}
	if input.CapacityMbps != nil {
		lead.CapacityMbps = input.CapacityMbps
	}
	if input.ExistingIsp != nil {
		lead.ExistingIsp = input.ExistingIsp
	}
	if input.Price != nil {
		lead.Price = input.Price
	}
	if input.OtherServices != nil {
		lead.OtherServices = input.OtherServices
	}
	if input.LeadSourceID != nil {
		lead.LeadSourceID = input.LeadSourceID
	}
	if input.ForecastMrr != nil {
		lead.ForecastMrr = input.ForecastMrr
	}
	if input.OwnerID != nil && (role == "ADMIN_SALES" || role == "SU") {
		lead.OwnerID = *input.OwnerID
	}
}

// leadStatusTransitions is the ARCHITECTURE.md §9 state machine, minus
// BARU->FOLLOW_UP - that transition only happens through the follow-up
// endpoint (see FollowUpService.Create), never through this one.
var leadStatusTransitions = map[string][]string{
	"BARU":         {"LOST"},
	"FOLLOW_UP":    {"HANDOFF_ODOO", "LOST"},
	"HANDOFF_ODOO": {},
	"LOST":         {},
}

// UpdateStatus validates and applies a status transition, scoped.
func (s *LeadService) UpdateStatus(callerID uuid.UUID, role, code string, input dto.UpdateLeadStatusInput) (*dto.LeadResponse, error) {
	scope, err := buildLeadScope(s.userRepo, callerID, role)
	if err != nil {
		return nil, err
	}

	lead, err := s.leadRepo.FindByCode(scope, code)
	if err != nil {
		return nil, ErrLeadNotFound
	}

	allowed := leadStatusTransitions[lead.Status]
	valid := false
	for _, target := range allowed {
		if target == input.Status {
			valid = true
			break
		}
	}
	if !valid {
		return nil, ErrInvalidStatusTransition
	}

	if input.Status == "LOST" && (input.LostReason == nil || *input.LostReason == "") {
		return nil, errors.New("lost_reason is required when status is LOST")
	}

	lead.Status = input.Status
	if input.Status == "LOST" {
		lead.LostReason = input.LostReason
	}

	if err := s.leadRepo.Update(lead); err != nil {
		return nil, err
	}

	response := dto.ToLeadResponse(lead)
	response.IsStale = isLeadStale(lead, time.Now())
	return &response, nil
}
