package services

import (
	"errors"
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/models"
	"fiber-api-boilerplate/internal/repository"
	"fiber-api-boilerplate/internal/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// ErrActiveLeadsExist is returned when a caller tries to deactivate a user
// who still owns active leads without providing a reassignment target. The
// handler maps this to 422 with the active lead count as payload.
var ErrActiveLeadsExist = errors.New("user has active leads")

// ErrForbiddenTarget is returned when an ADMIN_SALES caller tries to
// mutate (update, reset password, deactivate, or reactivate) a user whose
// role is SU. SU-level accounts can only be managed by other SU users. The
// handler maps this to 403.
var ErrForbiddenTarget = errors.New("cannot manage this user")

// EmailTakenError is returned by CreateUser when the submitted email
// already belongs to another account. Carries the existing account's id,
// name, and active status so the caller can be offered reactivation
// instead of a duplicate-account error, when that account is inactive.
// ExistingUserID is uuid.Nil (zero value) in the rare case where the
// collision was only caught by the database's unique index at insert time
// (a race with a concurrent create) rather than the FindByEmail pre-check —
// there's no fetched record to describe in that fallback.
type EmailTakenError struct {
	ExistingUserID uuid.UUID
	ExistingName   string
	ExistingActive bool
}

func (e *EmailTakenError) Error() string { return "email already registered" }

// callerForbiddenFromTarget reports whether callerRole may not mutate a
// user whose current role is targetRole. The only restriction today: an
// ADMIN_SALES caller cannot act on an SU account. SU callers face no
// restriction on any target role.
func callerForbiddenFromTarget(callerRole, targetRole string) bool {
	return callerRole == "ADMIN_SALES" && targetRole == "SU"
}

// userRepositoryForUser is the subset of repository methods UserService needs.
// Defined consumer-side for testability (satisfied by the real repo or a mock).
type userRepositoryForUser interface {
	Create(user *models.User) error
	FindByID(id uuid.UUID) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	ListWithFilter(role, teamID string) ([]models.User, error)
	CountActiveByOwner(ownerID uuid.UUID) (int64, error)
	ReassignOwner(oldOwnerID, newOwnerID uuid.UUID) error
	CountActiveAdmins() (int64, error)
}

// teamRepositoryForUser is the subset of sales-team repository methods
// UserService needs, just enough to validate a submitted team_id exists and
// is active. Named distinctly from team_service.go's salesTeamRepository
// (same package, different - narrower - method set).
type teamRepositoryForUser interface {
	FindByID(id uuid.UUID) (*models.SalesTeam, error)
}

type UserService struct {
	db               *gorm.DB
	userRepo         userRepositoryForUser
	refreshTokenRepo refreshTokenRepository
	teamRepo         teamRepositoryForUser
}

func NewUserService(
	db *gorm.DB,
	userRepo userRepositoryForUser,
	refreshTokenRepo refreshTokenRepository,
	teamRepo teamRepositoryForUser,
) *UserService {
	return &UserService{
		db:               db,
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		teamRepo:         teamRepo,
	}
}

// run wraps a multi-step write in a single transaction. Reused by every
// method below that mutates a user and revokes refresh tokens together, and
// by later phases (follow-up's 3-step write, lead code generation).
func (s *UserService) run(fn func(tx *gorm.DB) error) error {
	return s.db.Transaction(fn)
}

func (s *UserService) GetProfile(userID uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	response := dto.ToUserResponse(user)
	return &response, nil
}

// GetUser is the admin-facing counterpart of GetProfile - fetching an
// arbitrary user by id rather than the caller's own profile.
func (s *UserService) GetUser(userID uuid.UUID) (*dto.UserResponse, error) {
	return s.GetProfile(userID)
}

func (s *UserService) UpdateProfile(userID uuid.UUID, input dto.UpdateProfileInput) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	user.Name = input.Name

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	response := dto.ToUserResponse(user)
	return &response, nil
}

// ListUsers returns users optionally narrowed by role and/or team_id query
// params. Empty string means "no filter" for that field.
//
// For a LEADER caller, teamID is forced to the caller's own team_id and
// whatever the client sent is discarded entirely - the client-supplied
// filter never gets a say, mirroring buildLeadScope in lead_service.go.
// ADMIN_SALES/SU are unrestricted: their role/teamID pass straight through
// as filters, unchanged from before.
func (s *UserService) ListUsers(callerID uuid.UUID, callerRole, role, teamID string) ([]dto.UserResponse, error) {
	if callerRole == "LEADER" {
		caller, err := s.userRepo.FindByID(callerID)
		if err != nil {
			return nil, errors.New("caller not found")
		}
		if caller.TeamID == nil {
			// Fail closed: a nil team_id must never fall through to an
			// unrestricted list. A DB CHECK constraint keeps this
			// unreachable today, but don't rely on that here either.
			return nil, errors.New("leader has no team")
		}
		teamID = caller.TeamID.String()
	}

	users, err := s.userRepo.ListWithFilter(role, teamID)
	if err != nil {
		return nil, err
	}

	return dto.ToUserResponseList(users), nil
}

// roleRequiresTeam reports whether role must carry a non-nil team_id per
// the DB's chk_users_team_by_role constraint.
func roleRequiresTeam(role string) bool {
	return role == "SALES" || role == "LEADER"
}

// roleForbidsTeam reports whether role must carry a nil team_id per the
// DB's chk_users_team_by_role constraint.
func roleForbidsTeam(role string) bool {
	return role == "ADMIN_SALES" || role == "SU"
}

// roleIsAdmin reports whether role can manage other users (reach the admin
// endpoints). Used to protect the last active administrator from deactivation.
func roleIsAdmin(role string) bool {
	return role == "ADMIN_SALES" || role == "SU"
}

// validateRoleTeam rejects role/team_id combinations that would violate
// chk_users_team_by_role, so the error message is clear instead of a raw DB
// constraint violation.
func validateRoleTeam(role string, teamID *uuid.UUID) error {
	if roleRequiresTeam(role) && teamID == nil {
		return errors.New("role SALES/LEADER requires a team_id")
	}
	if roleForbidsTeam(role) && teamID != nil {
		return errors.New("role ADMIN_SALES/SU must not have a team_id")
	}
	return nil
}

// uuidPtrEqual compares two nullable UUIDs by value.
func uuidPtrEqual(a, b *uuid.UUID) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

// CreateUser is the only way a new account is created - self-registration
// was removed. Validates role/team_id consistency and that a submitted team
// exists and is active before hashing the password and inserting.
func (s *UserService) CreateUser(input dto.CreateUserInput) (*dto.UserResponse, error) {
	if err := validateRoleTeam(input.Role, input.TeamID); err != nil {
		return nil, err
	}

	if existing, err := s.userRepo.FindByEmail(input.Email); err == nil {
		return nil, &EmailTakenError{
			ExistingUserID: existing.ID,
			ExistingName:   existing.Name,
			ExistingActive: existing.IsActive,
		}
	}

	if input.TeamID != nil {
		team, err := s.teamRepo.FindByID(*input.TeamID)
		if err != nil {
			return nil, errors.New("team not found")
		}
		if !team.IsActive {
			return nil, errors.New("team is not active")
		}
	}

	hashed, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: hashed,
		Role:     input.Role,
		TeamID:   input.TeamID,
		IsActive: true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, translateCreateUserError(err)
	}

	response := dto.ToUserResponse(user)
	return &response, nil
}

// translateCreateUserError maps a Postgres unique-violation on the email
// index (SQLSTATE 23505) to an EmailTakenError with no existing-account
// details — the backstop for the rare race where two creates for the same
// email land concurrently and the FindByEmail pre-check above missed it.
// Any other error passes through unchanged.
func translateCreateUserError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return &EmailTakenError{}
	}
	return err
}

// UpdateUser applies the fields the admin sent, transactionally. If the
// effective role or team_id changes, every refresh token belonging to the
// user is revoked in the same transaction - role/team_id are JWT-adjacent
// facts and a stale token must not outlive the change.
func (s *UserService) UpdateUser(callerRole string, userID uuid.UUID, input dto.UpdateUserInput) (*dto.UserResponse, error) {
	var result *dto.UserResponse
	err := s.run(func(tx *gorm.DB) error {
		txUserRepo := repository.NewUserRepository(tx)
		txRefreshRepo := repository.NewRefreshTokenRepository(tx)

		response, err := s.updateUser(txUserRepo, txRefreshRepo, callerRole, userID, input)
		if err != nil {
			return err
		}
		result = response
		return nil
	})
	return result, err
}

// updateUser holds the actual logic, taking the repos as parameters rather
// than reading them off s. In production s.UpdateUser supplies tx-scoped
// repos so the write and the revoke commit together; in tests this can be
// called directly with mocks, without needing a real *gorm.DB transaction.
func (s *UserService) updateUser(
	userRepo userRepositoryForUser,
	refreshRepo refreshTokenRepository,
	callerRole string,
	userID uuid.UUID,
	input dto.UpdateUserInput,
) (*dto.UserResponse, error) {
	user, err := userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if callerForbiddenFromTarget(callerRole, user.Role) {
		return nil, ErrForbiddenTarget
	}

	if input.IsActive != nil && !*input.IsActive {
		return nil, errors.New("use the deactivate endpoint to deactivate a user")
	}

	newRole := user.Role
	if input.Role != nil {
		newRole = *input.Role
	}

	newTeamID := user.TeamID
	if input.TeamID != nil {
		newTeamID = input.TeamID
	} else if roleForbidsTeam(newRole) {
		// The DTO can't distinguish "team_id omitted" from "explicitly
		// cleared". Demoting to an admin-ish role without sending team_id
		// clears any existing team, since chk_users_team_by_role forbids
		// ADMIN_SALES/SU from carrying one.
		newTeamID = nil
	}

	if err := validateRoleTeam(newRole, newTeamID); err != nil {
		return nil, err
	}

	if input.TeamID != nil {
		team, err := s.teamRepo.FindByID(*newTeamID)
		if err != nil {
			return nil, errors.New("team not found")
		}
		if !team.IsActive {
			return nil, errors.New("team is not active")
		}
	}

	roleOrTeamChanged := newRole != user.Role || !uuidPtrEqual(newTeamID, user.TeamID)

	if input.Name != nil {
		user.Name = *input.Name
	}
	user.Role = newRole
	user.TeamID = newTeamID
	if input.IsActive != nil {
		user.IsActive = true // false already rejected above
	}

	if err := userRepo.Update(user); err != nil {
		return nil, err
	}

	if roleOrTeamChanged {
		if err := refreshRepo.DeleteAllByUserID(userID); err != nil {
			return nil, err
		}
	}

	response := dto.ToUserResponse(user)
	return &response, nil
}

// ResetPassword sets a new password and revokes every refresh token
// belonging to the user, in one transaction.
func (s *UserService) ResetPassword(callerRole string, userID uuid.UUID, input dto.ResetPasswordInput) error {
	return s.run(func(tx *gorm.DB) error {
		txUserRepo := repository.NewUserRepository(tx)
		txRefreshRepo := repository.NewRefreshTokenRepository(tx)
		return s.resetPassword(txUserRepo, txRefreshRepo, callerRole, userID, input)
	})
}

func (s *UserService) resetPassword(
	userRepo userRepositoryForUser,
	refreshRepo refreshTokenRepository,
	callerRole string,
	userID uuid.UUID,
	input dto.ResetPasswordInput,
) error {
	user, err := userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}
	if callerForbiddenFromTarget(callerRole, user.Role) {
		return ErrForbiddenTarget
	}

	hashed, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		return err
	}
	user.Password = hashed

	if err := userRepo.Update(user); err != nil {
		return err
	}

	return refreshRepo.DeleteAllByUserID(userID)
}

// DeactivateUser flips is_active off. If the user still owns active leads
// (status BARU/FOLLOW_UP) and no reassignment target was given, it returns
// ErrActiveLeadsExist with the count, without touching the DB beyond the
// count query - the caller decides whether to reassign or abort. With a
// reassignment target, reassign + deactivate + revoke happen in one
// transaction.
func (s *UserService) DeactivateUser(callerID uuid.UUID, callerRole string, userID uuid.UUID, input dto.DeactivateUserInput) (int64, error) {
	var activeCount int64
	err := s.run(func(tx *gorm.DB) error {
		txUserRepo := repository.NewUserRepository(tx)
		txRefreshRepo := repository.NewRefreshTokenRepository(tx)

		count, err := s.deactivateUser(txUserRepo, txRefreshRepo, callerID, callerRole, userID, input)
		activeCount = count
		return err
	})
	return activeCount, err
}

func (s *UserService) deactivateUser(
	userRepo userRepositoryForUser,
	refreshRepo refreshTokenRepository,
	callerID uuid.UUID,
	callerRole string,
	userID uuid.UUID,
	input dto.DeactivateUserInput,
) (int64, error) {
	// An admin deactivating themselves would revoke their own session and,
	// if they were the last admin, lock everyone out of user management.
	if userID == callerID {
		return 0, errors.New("cannot deactivate your own account")
	}

	user, err := userRepo.FindByID(userID)
	if err != nil {
		return 0, errors.New("user not found")
	}

	if callerForbiddenFromTarget(callerRole, user.Role) {
		return 0, ErrForbiddenTarget
	}

	// Never remove the last active administrator. Only admins can reach this
	// endpoint, so blocking self-deactivation already keeps at least the
	// caller active; this is the explicit backstop for that invariant.
	if roleIsAdmin(user.Role) && user.IsActive {
		admins, err := userRepo.CountActiveAdmins()
		if err != nil {
			return 0, err
		}
		if admins <= 1 {
			return 0, errors.New("cannot deactivate the last active administrator")
		}
	}

	count, err := userRepo.CountActiveByOwner(userID)
	if err != nil {
		return 0, err
	}

	if count > 0 {
		if input.ReassignToUserID == nil {
			return count, ErrActiveLeadsExist
		}
		// The reassignment target must be able to hold the leads. Reassigning
		// to the same user is a no-op that would strand active leads on the
		// now-inactive owner, and reassigning to another inactive user just
		// moves the problem — both defeat the 422 guard above. A non-existent
		// id would otherwise surface only as an opaque FK error.
		if *input.ReassignToUserID == userID {
			return count, errors.New("reassignment target must be a different user")
		}
		target, err := userRepo.FindByID(*input.ReassignToUserID)
		if err != nil {
			return count, errors.New("reassignment target not found")
		}
		if !target.IsActive {
			return count, errors.New("reassignment target is not active")
		}
		if err := userRepo.ReassignOwner(userID, *input.ReassignToUserID); err != nil {
			return count, err
		}
	}

	user.IsActive = false
	if err := userRepo.Update(user); err != nil {
		return count, err
	}

	if err := refreshRepo.DeleteAllByUserID(userID); err != nil {
		return count, err
	}

	return count, nil
}
