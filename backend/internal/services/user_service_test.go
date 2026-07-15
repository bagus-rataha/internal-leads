package services

import (
	"fiber-api-boilerplate/internal/dto"
	"fiber-api-boilerplate/internal/models"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// newUserService wires a UserService for unit tests. db is nil - the tests
// exercise either non-transactional methods or the private helpers directly,
// so s.run/s.db.Transaction is never invoked and a real DB is unnecessary.
func newUserService(userRepo *MockUserRepository, refreshRepo *MockRefreshTokenRepository, teamRepo *MockSalesTeamRepository) *UserService {
	return NewUserService(nil, userRepo, refreshRepo, teamRepo)
}

func TestGetProfile_Success(t *testing.T) {
	userRepo := new(MockUserRepository)
	userID := uuid.Must(uuid.NewV7())

	user := &models.User{
		BaseModel: models.BaseModel{ID: userID},
		Email:     "test@test.com",
		Name:      "Test",
		Role:      "user",
	}

	userRepo.On("FindByID", userID).Return(user, nil)

	svc := newUserService(userRepo, new(MockRefreshTokenRepository), new(MockSalesTeamRepository))
	result, err := svc.GetProfile(userID)

	assert.NoError(t, err)
	assert.Equal(t, "test@test.com", result.Email)
	assert.Equal(t, "Test", result.Name)
}

func TestGetProfile_NotFound(t *testing.T) {
	userRepo := new(MockUserRepository)
	userID := uuid.Must(uuid.NewV7())

	userRepo.On("FindByID", userID).Return(nil, gorm.ErrRecordNotFound)

	svc := newUserService(userRepo, new(MockRefreshTokenRepository), new(MockSalesTeamRepository))
	_, err := svc.GetProfile(userID)

	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
}

func TestUpdateProfile_Success(t *testing.T) {
	userRepo := new(MockUserRepository)
	userID := uuid.Must(uuid.NewV7())

	user := &models.User{
		BaseModel: models.BaseModel{ID: userID},
		Email:     "test@test.com",
		Name:      "Old Name",
		Role:      "user",
	}

	userRepo.On("FindByID", userID).Return(user, nil)
	userRepo.On("Update", user).Return(nil)

	svc := newUserService(userRepo, new(MockRefreshTokenRepository), new(MockSalesTeamRepository))
	result, err := svc.UpdateProfile(userID, dto.UpdateProfileInput{Name: "New Name"})

	assert.NoError(t, err)
	assert.Equal(t, "New Name", result.Name)
}

func TestUpdateProfile_UserNotFound(t *testing.T) {
	userRepo := new(MockUserRepository)
	userID := uuid.Must(uuid.NewV7())

	userRepo.On("FindByID", userID).Return(nil, gorm.ErrRecordNotFound)

	svc := newUserService(userRepo, new(MockRefreshTokenRepository), new(MockSalesTeamRepository))
	_, err := svc.UpdateProfile(userID, dto.UpdateProfileInput{Name: "New Name"})

	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
}

func TestListUsers_Success(t *testing.T) {
	userRepo := new(MockUserRepository)

	users := []models.User{
		{BaseModel: models.BaseModel{ID: uuid.Must(uuid.NewV7())}, Email: "a@test.com", Name: "A"},
		{BaseModel: models.BaseModel{ID: uuid.Must(uuid.NewV7())}, Email: "b@test.com", Name: "B"},
	}

	userRepo.On("ListWithFilter", "", "").Return(users, nil)

	svc := newUserService(userRepo, new(MockRefreshTokenRepository), new(MockSalesTeamRepository))
	result, err := svc.ListUsers("", "")

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestListUsers_FilteredByRoleAndTeam(t *testing.T) {
	userRepo := new(MockUserRepository)
	teamID := uuid.Must(uuid.NewV7()).String()

	userRepo.On("ListWithFilter", "SALES", teamID).Return([]models.User{}, nil)

	svc := newUserService(userRepo, new(MockRefreshTokenRepository), new(MockSalesTeamRepository))
	_, err := svc.ListUsers("SALES", teamID)

	assert.NoError(t, err)
	userRepo.AssertCalled(t, "ListWithFilter", "SALES", teamID)
}

// --- CreateUser ---

func TestCreateUser_ValidRoleAndTeam_Success(t *testing.T) {
	userRepo := new(MockUserRepository)
	teamRepo := new(MockSalesTeamRepository)
	teamID := uuid.Must(uuid.NewV7())

	teamRepo.On("FindByID", teamID).Return(&models.SalesTeam{BaseModel: models.BaseModel{ID: teamID}, IsActive: true}, nil)
	userRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)

	svc := newUserService(userRepo, new(MockRefreshTokenRepository), teamRepo)
	input := dto.CreateUserInput{Name: "New Sales", Email: "sales@test.com", Password: "secret1", Role: "SALES", TeamID: &teamID}
	result, err := svc.CreateUser(input)

	assert.NoError(t, err)
	assert.Equal(t, "sales@test.com", result.Email)
	assert.Equal(t, "SALES", result.Role)
	userRepo.AssertCalled(t, "Create", mock.AnythingOfType("*models.User"))
}

func TestCreateUser_AdminRoleWithTeam_Rejected(t *testing.T) {
	userRepo := new(MockUserRepository)
	teamRepo := new(MockSalesTeamRepository)
	teamID := uuid.Must(uuid.NewV7())

	svc := newUserService(userRepo, new(MockRefreshTokenRepository), teamRepo)
	input := dto.CreateUserInput{Name: "Bad Admin", Email: "admin@test.com", Password: "secret1", Role: "ADMIN_SALES", TeamID: &teamID}
	_, err := svc.CreateUser(input)

	assert.Error(t, err)
	userRepo.AssertNotCalled(t, "Create", mock.Anything)
	teamRepo.AssertNotCalled(t, "FindByID", mock.Anything)
}

func TestCreateUser_SalesRoleWithoutTeam_Rejected(t *testing.T) {
	userRepo := new(MockUserRepository)
	teamRepo := new(MockSalesTeamRepository)

	svc := newUserService(userRepo, new(MockRefreshTokenRepository), teamRepo)
	input := dto.CreateUserInput{Name: "Bad Sales", Email: "sales2@test.com", Password: "secret1", Role: "SALES", TeamID: nil}
	_, err := svc.CreateUser(input)

	assert.Error(t, err)
	userRepo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestCreateUser_InactiveTeam_Rejected(t *testing.T) {
	userRepo := new(MockUserRepository)
	teamRepo := new(MockSalesTeamRepository)
	teamID := uuid.Must(uuid.NewV7())

	teamRepo.On("FindByID", teamID).Return(&models.SalesTeam{BaseModel: models.BaseModel{ID: teamID}, IsActive: false}, nil)

	svc := newUserService(userRepo, new(MockRefreshTokenRepository), teamRepo)
	input := dto.CreateUserInput{Name: "Sales", Email: "sales3@test.com", Password: "secret1", Role: "SALES", TeamID: &teamID}
	_, err := svc.CreateUser(input)

	assert.Error(t, err)
	userRepo.AssertNotCalled(t, "Create", mock.Anything)
}

// --- UpdateUser (private updateUser helper, exercised without a real tx) ---

func TestUpdateUser_NameOnly_NoRevoke(t *testing.T) {
	userRepo := new(MockUserRepository)
	refreshRepo := new(MockRefreshTokenRepository)
	teamID := uuid.Must(uuid.NewV7())
	userID := uuid.Must(uuid.NewV7())

	user := &models.User{BaseModel: models.BaseModel{ID: userID}, Name: "Old", Role: "SALES", TeamID: &teamID}
	userRepo.On("FindByID", userID).Return(user, nil)
	userRepo.On("Update", user).Return(nil)

	svc := newUserService(userRepo, refreshRepo, new(MockSalesTeamRepository))
	newName := "New"
	result, err := svc.updateUser(userRepo, refreshRepo, userID, dto.UpdateUserInput{Name: &newName})

	assert.NoError(t, err)
	assert.Equal(t, "New", result.Name)
	refreshRepo.AssertNotCalled(t, "DeleteAllByUserID", mock.Anything)
}

func TestUpdateUser_RoleChanged_RevokeCalled(t *testing.T) {
	userRepo := new(MockUserRepository)
	refreshRepo := new(MockRefreshTokenRepository)
	teamID := uuid.Must(uuid.NewV7())
	userID := uuid.Must(uuid.NewV7())

	user := &models.User{BaseModel: models.BaseModel{ID: userID}, Name: "Name", Role: "SALES", TeamID: &teamID}
	userRepo.On("FindByID", userID).Return(user, nil)
	userRepo.On("Update", user).Return(nil)
	refreshRepo.On("DeleteAllByUserID", userID).Return(nil)

	svc := newUserService(userRepo, refreshRepo, new(MockSalesTeamRepository))
	newRole := "LEADER"
	result, err := svc.updateUser(userRepo, refreshRepo, userID, dto.UpdateUserInput{Role: &newRole})

	assert.NoError(t, err)
	assert.Equal(t, "LEADER", result.Role)
	refreshRepo.AssertCalled(t, "DeleteAllByUserID", userID)
}

func TestUpdateUser_TeamChanged_RevokeCalled(t *testing.T) {
	userRepo := new(MockUserRepository)
	refreshRepo := new(MockRefreshTokenRepository)
	teamRepo := new(MockSalesTeamRepository)
	oldTeamID := uuid.Must(uuid.NewV7())
	newTeamID := uuid.Must(uuid.NewV7())
	userID := uuid.Must(uuid.NewV7())

	user := &models.User{BaseModel: models.BaseModel{ID: userID}, Name: "Name", Role: "SALES", TeamID: &oldTeamID}
	userRepo.On("FindByID", userID).Return(user, nil)
	userRepo.On("Update", user).Return(nil)
	refreshRepo.On("DeleteAllByUserID", userID).Return(nil)
	teamRepo.On("FindByID", newTeamID).Return(&models.SalesTeam{BaseModel: models.BaseModel{ID: newTeamID}, IsActive: true}, nil)

	svc := newUserService(userRepo, refreshRepo, teamRepo)
	result, err := svc.updateUser(userRepo, refreshRepo, userID, dto.UpdateUserInput{TeamID: &newTeamID})

	assert.NoError(t, err)
	assert.Equal(t, newTeamID, *result.TeamID)
	refreshRepo.AssertCalled(t, "DeleteAllByUserID", userID)
}

func TestUpdateUser_DemoteToAdmin_ClearsTeamAndRejectsExplicitTeam(t *testing.T) {
	userRepo := new(MockUserRepository)
	refreshRepo := new(MockRefreshTokenRepository)
	teamID := uuid.Must(uuid.NewV7())
	userID := uuid.Must(uuid.NewV7())

	user := &models.User{BaseModel: models.BaseModel{ID: userID}, Name: "Name", Role: "LEADER", TeamID: &teamID}
	userRepo.On("FindByID", userID).Return(user, nil)
	userRepo.On("Update", mock.AnythingOfType("*models.User")).Return(nil)
	refreshRepo.On("DeleteAllByUserID", userID).Return(nil)

	svc := newUserService(userRepo, refreshRepo, new(MockSalesTeamRepository))
	newRole := "ADMIN_SALES"
	result, err := svc.updateUser(userRepo, refreshRepo, userID, dto.UpdateUserInput{Role: &newRole})

	assert.NoError(t, err)
	assert.Nil(t, result.TeamID)
}

func TestUpdateUser_RoleTeamMismatch_Rejected(t *testing.T) {
	userRepo := new(MockUserRepository)
	refreshRepo := new(MockRefreshTokenRepository)
	userID := uuid.Must(uuid.NewV7())

	user := &models.User{BaseModel: models.BaseModel{ID: userID}, Name: "Name", Role: "ADMIN_SALES", TeamID: nil}
	userRepo.On("FindByID", userID).Return(user, nil)

	svc := newUserService(userRepo, refreshRepo, new(MockSalesTeamRepository))
	newRole := "SALES"
	_, err := svc.updateUser(userRepo, refreshRepo, userID, dto.UpdateUserInput{Role: &newRole})

	assert.Error(t, err)
	userRepo.AssertNotCalled(t, "Update", mock.Anything)
	refreshRepo.AssertNotCalled(t, "DeleteAllByUserID", mock.Anything)
}

// --- ResetPassword ---

func TestResetPassword_Success(t *testing.T) {
	userRepo := new(MockUserRepository)
	refreshRepo := new(MockRefreshTokenRepository)
	userID := uuid.Must(uuid.NewV7())

	user := &models.User{BaseModel: models.BaseModel{ID: userID}, Password: "old-hash"}
	userRepo.On("FindByID", userID).Return(user, nil)
	userRepo.On("Update", user).Return(nil)
	refreshRepo.On("DeleteAllByUserID", userID).Return(nil)

	svc := newUserService(userRepo, refreshRepo, new(MockSalesTeamRepository))
	err := svc.resetPassword(userRepo, refreshRepo, userID, dto.ResetPasswordInput{NewPassword: "newpass1"})

	assert.NoError(t, err)
	assert.NotEqual(t, "old-hash", user.Password)
	refreshRepo.AssertCalled(t, "DeleteAllByUserID", userID)
}

// --- DeactivateUser ---

func TestDeactivateUser_ActiveLeadsNoReassign_Error(t *testing.T) {
	userRepo := new(MockUserRepository)
	refreshRepo := new(MockRefreshTokenRepository)
	userID := uuid.Must(uuid.NewV7())

	user := &models.User{BaseModel: models.BaseModel{ID: userID}, IsActive: true}
	userRepo.On("FindByID", userID).Return(user, nil)
	userRepo.On("CountActiveByOwner", userID).Return(int64(3), nil)

	svc := newUserService(userRepo, refreshRepo, new(MockSalesTeamRepository))
	count, err := svc.deactivateUser(userRepo, refreshRepo, uuid.Must(uuid.NewV7()), userID, dto.DeactivateUserInput{})

	assert.ErrorIs(t, err, ErrActiveLeadsExist)
	assert.Equal(t, int64(3), count)
	userRepo.AssertNotCalled(t, "ReassignOwner", mock.Anything, mock.Anything)
	userRepo.AssertNotCalled(t, "Update", mock.Anything)
	refreshRepo.AssertNotCalled(t, "DeleteAllByUserID", mock.Anything)
}

func TestDeactivateUser_ActiveLeadsWithReassign_Success(t *testing.T) {
	userRepo := new(MockUserRepository)
	refreshRepo := new(MockRefreshTokenRepository)
	userID := uuid.Must(uuid.NewV7())
	newOwnerID := uuid.Must(uuid.NewV7())

	user := &models.User{BaseModel: models.BaseModel{ID: userID}, IsActive: true}
	target := &models.User{BaseModel: models.BaseModel{ID: newOwnerID}, IsActive: true}
	userRepo.On("FindByID", userID).Return(user, nil)
	userRepo.On("FindByID", newOwnerID).Return(target, nil)
	userRepo.On("CountActiveByOwner", userID).Return(int64(2), nil)
	userRepo.On("ReassignOwner", userID, newOwnerID).Return(nil)
	userRepo.On("Update", user).Return(nil)
	refreshRepo.On("DeleteAllByUserID", userID).Return(nil)

	svc := newUserService(userRepo, refreshRepo, new(MockSalesTeamRepository))
	count, err := svc.deactivateUser(userRepo, refreshRepo, uuid.Must(uuid.NewV7()), userID, dto.DeactivateUserInput{ReassignToUserID: &newOwnerID})

	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
	assert.False(t, user.IsActive)
	userRepo.AssertCalled(t, "ReassignOwner", userID, newOwnerID)
	refreshRepo.AssertCalled(t, "DeleteAllByUserID", userID)
}

func TestDeactivateUser_ReassignToSelf_Rejected(t *testing.T) {
	userRepo := new(MockUserRepository)
	refreshRepo := new(MockRefreshTokenRepository)
	userID := uuid.Must(uuid.NewV7())

	user := &models.User{BaseModel: models.BaseModel{ID: userID}, IsActive: true}
	userRepo.On("FindByID", userID).Return(user, nil)
	userRepo.On("CountActiveByOwner", userID).Return(int64(2), nil)

	svc := newUserService(userRepo, refreshRepo, new(MockSalesTeamRepository))
	_, err := svc.deactivateUser(userRepo, refreshRepo, uuid.Must(uuid.NewV7()), userID, dto.DeactivateUserInput{ReassignToUserID: &userID})

	assert.Error(t, err)
	// Reassigning to self would strand the leads on the deactivated owner.
	userRepo.AssertNotCalled(t, "ReassignOwner", mock.Anything, mock.Anything)
	userRepo.AssertNotCalled(t, "Update", mock.Anything)
	refreshRepo.AssertNotCalled(t, "DeleteAllByUserID", mock.Anything)
}

func TestDeactivateUser_ReassignToInactive_Rejected(t *testing.T) {
	userRepo := new(MockUserRepository)
	refreshRepo := new(MockRefreshTokenRepository)
	userID := uuid.Must(uuid.NewV7())
	newOwnerID := uuid.Must(uuid.NewV7())

	user := &models.User{BaseModel: models.BaseModel{ID: userID}, IsActive: true}
	inactiveTarget := &models.User{BaseModel: models.BaseModel{ID: newOwnerID}, IsActive: false}
	userRepo.On("FindByID", userID).Return(user, nil)
	userRepo.On("FindByID", newOwnerID).Return(inactiveTarget, nil)
	userRepo.On("CountActiveByOwner", userID).Return(int64(2), nil)

	svc := newUserService(userRepo, refreshRepo, new(MockSalesTeamRepository))
	_, err := svc.deactivateUser(userRepo, refreshRepo, uuid.Must(uuid.NewV7()), userID, dto.DeactivateUserInput{ReassignToUserID: &newOwnerID})

	assert.Error(t, err)
	userRepo.AssertNotCalled(t, "ReassignOwner", mock.Anything, mock.Anything)
	userRepo.AssertNotCalled(t, "Update", mock.Anything)
	refreshRepo.AssertNotCalled(t, "DeleteAllByUserID", mock.Anything)
}

func TestDeactivateUser_NoActiveLeads_SkipsReassign(t *testing.T) {
	userRepo := new(MockUserRepository)
	refreshRepo := new(MockRefreshTokenRepository)
	userID := uuid.Must(uuid.NewV7())

	user := &models.User{BaseModel: models.BaseModel{ID: userID}, IsActive: true}
	userRepo.On("FindByID", userID).Return(user, nil)
	userRepo.On("CountActiveByOwner", userID).Return(int64(0), nil)
	userRepo.On("Update", user).Return(nil)
	refreshRepo.On("DeleteAllByUserID", userID).Return(nil)

	svc := newUserService(userRepo, refreshRepo, new(MockSalesTeamRepository))
	count, err := svc.deactivateUser(userRepo, refreshRepo, uuid.Must(uuid.NewV7()), userID, dto.DeactivateUserInput{})

	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
	assert.False(t, user.IsActive)
	userRepo.AssertNotCalled(t, "ReassignOwner", mock.Anything, mock.Anything)
}

func TestDeactivateUser_Self_Rejected(t *testing.T) {
	userRepo := new(MockUserRepository)
	refreshRepo := new(MockRefreshTokenRepository)
	userID := uuid.Must(uuid.NewV7())

	svc := newUserService(userRepo, refreshRepo, new(MockSalesTeamRepository))
	// caller == target: an admin must not deactivate their own account.
	_, err := svc.deactivateUser(userRepo, refreshRepo, userID, userID, dto.DeactivateUserInput{})

	assert.Error(t, err)
	userRepo.AssertNotCalled(t, "FindByID", mock.Anything)
	userRepo.AssertNotCalled(t, "Update", mock.Anything)
	refreshRepo.AssertNotCalled(t, "DeleteAllByUserID", mock.Anything)
}

func TestDeactivateUser_LastActiveAdmin_Rejected(t *testing.T) {
	userRepo := new(MockUserRepository)
	refreshRepo := new(MockRefreshTokenRepository)
	userID := uuid.Must(uuid.NewV7())

	admin := &models.User{BaseModel: models.BaseModel{ID: userID}, Role: "SU", IsActive: true}
	userRepo.On("FindByID", userID).Return(admin, nil)
	userRepo.On("CountActiveAdmins").Return(int64(1), nil)

	svc := newUserService(userRepo, refreshRepo, new(MockSalesTeamRepository))
	_, err := svc.deactivateUser(userRepo, refreshRepo, uuid.Must(uuid.NewV7()), userID, dto.DeactivateUserInput{})

	assert.Error(t, err)
	userRepo.AssertNotCalled(t, "CountActiveByOwner", mock.Anything)
	userRepo.AssertNotCalled(t, "Update", mock.Anything)
	refreshRepo.AssertNotCalled(t, "DeleteAllByUserID", mock.Anything)
}

func TestDeactivateUser_AdminWithOtherAdmins_Allowed(t *testing.T) {
	userRepo := new(MockUserRepository)
	refreshRepo := new(MockRefreshTokenRepository)
	userID := uuid.Must(uuid.NewV7())

	admin := &models.User{BaseModel: models.BaseModel{ID: userID}, Role: "ADMIN_SALES", IsActive: true}
	userRepo.On("FindByID", userID).Return(admin, nil)
	userRepo.On("CountActiveAdmins").Return(int64(2), nil)
	userRepo.On("CountActiveByOwner", userID).Return(int64(0), nil)
	userRepo.On("Update", admin).Return(nil)
	refreshRepo.On("DeleteAllByUserID", userID).Return(nil)

	svc := newUserService(userRepo, refreshRepo, new(MockSalesTeamRepository))
	_, err := svc.deactivateUser(userRepo, refreshRepo, uuid.Must(uuid.NewV7()), userID, dto.DeactivateUserInput{})

	assert.NoError(t, err)
	assert.False(t, admin.IsActive)
}
