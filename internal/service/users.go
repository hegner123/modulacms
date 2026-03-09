package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/hegner123/modulacms/internal/auth"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// UserService manages user CRUD, password hashing, role assignment,
// uniqueness checks, and user reassign-delete orchestration.
type UserService struct {
	driver db.DbDriver
	mgr    *config.Manager
}

// NewUserService creates a UserService with the given dependencies.
func NewUserService(driver db.DbDriver, mgr *config.Manager) *UserService {
	return &UserService{driver: driver, mgr: mgr}
}

// CreateUserInput holds caller-provided fields for creating a user.
type CreateUserInput struct {
	Username string
	Name     string
	Email    types.Email
	Password string       // plaintext — service hashes; must be >= 8 chars
	Role     types.RoleID // zero value = default to viewer
	IsAdmin  bool         // whether caller is admin (for role assignment gating)
}

// UpdateUserInput holds caller-provided fields for updating a user.
type UpdateUserInput struct {
	UserID   types.UserID
	Username string
	Name     string
	Email    types.Email
	Password string       // plaintext; empty = keep existing hash
	Role     types.RoleID // zero value = keep existing role
	IsAdmin  bool         // whether caller is admin (for role change gating)
}

// ReassignDeleteInput holds parameters for the reassign-and-delete operation.
type ReassignDeleteInput struct {
	UserID     types.UserID
	ReassignTo types.UserID // zero = default to SystemUserID
}

// ReassignDeleteResult holds counts from a successful reassign-delete.
type ReassignDeleteResult struct {
	DeletedUserID              types.UserID
	ReassignedTo               types.UserID
	ContentDataReassigned      int64
	DatatypesReassigned        int64
	AdminContentDataReassigned int64
}

// --- Public Methods ---

// CreateUser validates input, checks uniqueness, hashes password,
// resolves default role, and creates the user.
func (s *UserService) CreateUser(ctx context.Context, ac audited.AuditContext, input CreateUserInput) (*db.Users, error) {
	if err := validateCreateUserInput(input); err != nil {
		return nil, err
	}

	if err := s.checkEmailUniqueness(ctx, input.Email, ""); err != nil {
		return nil, err
	}
	if err := s.checkUsernameUniqueness(ctx, input.Username, ""); err != nil {
		return nil, err
	}

	// Role gating: non-admins cannot assign roles
	if !input.Role.IsZero() && !input.IsAdmin {
		return nil, &ForbiddenError{Message: "only administrators can assign roles"}
	}

	roleID, err := s.resolveDefaultRole(input.Role)
	if err != nil {
		return nil, fmt.Errorf("resolve default role: %w", err)
	}

	hash, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	now := types.NewTimestamp(time.Now().UTC())
	created, err := s.driver.CreateUser(ctx, ac, db.CreateUserParams{
		Username:     input.Username,
		Name:         input.Name,
		Email:        input.Email,
		Hash:         hash,
		Role:         string(roleID),
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return created, nil
}

// UpdateUser validates input, checks uniqueness, optionally re-hashes password,
// enforces role assignment gating, and updates the user.
func (s *UserService) UpdateUser(ctx context.Context, ac audited.AuditContext, input UpdateUserInput) (*db.Users, error) {
	existing, err := s.driver.GetUser(input.UserID)
	if err != nil {
		return nil, &NotFoundError{Resource: "user", ID: string(input.UserID)}
	}

	if err := validateUpdateUserInput(input); err != nil {
		return nil, err
	}

	// System user role protection
	if input.UserID == types.SystemUserID && !input.Role.IsZero() && string(input.Role) != existing.Role {
		return nil, &ForbiddenError{Message: "cannot change the system user's role"}
	}

	// Role change gating: non-admins cannot change roles
	if !input.Role.IsZero() && string(input.Role) != existing.Role && !input.IsAdmin {
		return nil, &ForbiddenError{Message: "only administrators can assign roles"}
	}

	// Password handling
	hash := existing.Hash
	if input.Password != "" {
		if len(input.Password) < 8 {
			return nil, NewValidationError("password", "password must be at least 8 characters")
		}
		h, hashErr := auth.HashPassword(input.Password)
		if hashErr != nil {
			return nil, fmt.Errorf("hash password: %w", hashErr)
		}
		hash = h
	}

	// Email uniqueness (only if changed)
	if string(input.Email) != string(existing.Email) {
		if err := s.checkEmailUniqueness(ctx, input.Email, string(input.UserID)); err != nil {
			return nil, err
		}
	}

	// Username uniqueness (only if changed)
	if input.Username != existing.Username {
		if err := s.checkUsernameUniqueness(ctx, input.Username, string(input.UserID)); err != nil {
			return nil, err
		}
	}

	// Resolve role: zero value = keep existing
	role := existing.Role
	if !input.Role.IsZero() {
		role = string(input.Role)
	}

	_, err = s.driver.UpdateUser(ctx, ac, db.UpdateUserParams{
		Username:     input.Username,
		Name:         input.Name,
		Email:        input.Email,
		Hash:         hash,
		Role:         role,
		DateCreated:  existing.DateCreated,
		DateModified: types.NewTimestamp(time.Now().UTC()),
		UserID:       input.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	updated, err := s.driver.GetUser(input.UserID)
	if err != nil {
		return nil, fmt.Errorf("fetch updated user: %w", err)
	}
	return updated, nil
}

// DeleteUser checks system user protection and deletes the user.
func (s *UserService) DeleteUser(ctx context.Context, ac audited.AuditContext, userID types.UserID) error {
	if userID == types.SystemUserID {
		return &ForbiddenError{Message: "cannot delete the system user"}
	}

	if _, err := s.driver.GetUser(userID); err != nil {
		return &NotFoundError{Resource: "user", ID: string(userID)}
	}

	if err := s.driver.DeleteUser(ctx, ac, userID); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

// GetUser retrieves a user by ID with NotFoundError mapping.
func (s *UserService) GetUser(ctx context.Context, userID types.UserID) (*db.Users, error) {
	user, err := s.driver.GetUser(userID)
	if err != nil {
		return nil, &NotFoundError{Resource: "user", ID: string(userID)}
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email with NotFoundError mapping.
func (s *UserService) GetUserByEmail(ctx context.Context, email types.Email) (*db.Users, error) {
	user, err := s.driver.GetUserByEmail(email)
	if err != nil {
		return nil, &NotFoundError{Resource: "user", ID: string(email)}
	}
	return user, nil
}

// ListUsers returns all users.
func (s *UserService) ListUsers(ctx context.Context) (*[]db.Users, error) {
	return s.driver.ListUsers()
}

// ListUsersWithRoleLabel returns all users with their role labels joined.
func (s *UserService) ListUsersWithRoleLabel(ctx context.Context) (*[]db.UserWithRoleLabelRow, error) {
	return s.driver.ListUsersWithRoleLabel()
}

// GetUserFull retrieves a composed user view with all related entities.
func (s *UserService) GetUserFull(ctx context.Context, userID types.UserID) (*db.UserFullView, error) {
	view, err := db.AssembleUserFullView(s.driver, userID)
	if err != nil {
		return nil, &NotFoundError{Resource: "user", ID: string(userID)}
	}
	return view, nil
}

// ReassignDelete atomically reassigns all authored content from a user to
// another user, then deletes the original user with an audit trail.
func (s *UserService) ReassignDelete(ctx context.Context, ac audited.AuditContext, input ReassignDeleteInput) (*ReassignDeleteResult, error) {
	if input.UserID.IsZero() {
		return nil, NewValidationError("user_id", "user_id is required")
	}
	if err := input.UserID.Validate(); err != nil {
		return nil, NewValidationError("user_id", fmt.Sprintf("invalid user_id: %v", err))
	}
	if input.UserID == types.SystemUserID {
		return nil, &ForbiddenError{Message: "cannot delete the system user"}
	}

	// Resolve reassign target (default to system user)
	reassignTo := input.ReassignTo
	if reassignTo == "" || reassignTo.IsZero() {
		reassignTo = types.SystemUserID
	}
	if err := reassignTo.Validate(); err != nil {
		return nil, NewValidationError("reassign_to", fmt.Sprintf("invalid reassign_to: %v", err))
	}
	if input.UserID == reassignTo {
		return nil, NewValidationError("reassign_to", "cannot reassign to the same user being deleted")
	}

	// Verify both users exist
	if _, err := s.driver.GetUser(input.UserID); err != nil {
		return nil, &NotFoundError{Resource: "user", ID: string(input.UserID)}
	}
	if _, err := s.driver.GetUser(reassignTo); err != nil {
		return nil, &NotFoundError{Resource: "reassign target user", ID: string(reassignTo)}
	}

	// Obtain connection and dialect for query builder
	conn, _, err := s.driver.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("get connection: %w", err)
	}

	cfg, err := s.mgr.Config()
	if err != nil {
		return nil, fmt.Errorf("get config: %w", err)
	}
	dialect := db.DialectFromString(string(cfg.Db_Driver))

	// Phase 1: counts + reassigns inside transaction
	result, err := db.WithTransactionResult[ReassignDeleteResult](ctx, conn, func(tx *sql.Tx) (ReassignDeleteResult, error) {
		var r ReassignDeleteResult
		r.DeletedUserID = input.UserID
		r.ReassignedTo = reassignTo

		authorWhere := map[string]any{"author_id": string(input.UserID)}
		authorSet := map[string]any{"author_id": string(reassignTo)}

		// Count content_data
		count, cErr := db.QCount(ctx, tx, dialect, "content_data", authorWhere)
		if cErr != nil {
			return r, fmt.Errorf("count content_data: %w", cErr)
		}
		r.ContentDataReassigned = count

		// Count datatypes
		count, cErr = db.QCount(ctx, tx, dialect, "datatypes", authorWhere)
		if cErr != nil {
			return r, fmt.Errorf("count datatypes: %w", cErr)
		}
		r.DatatypesReassigned = count

		// Count admin_content_data
		count, cErr = db.QCount(ctx, tx, dialect, "admin_content_data", authorWhere)
		if cErr != nil {
			return r, fmt.Errorf("count admin_content_data: %w", cErr)
		}
		r.AdminContentDataReassigned = count

		// Reassign content_data
		if r.ContentDataReassigned > 0 {
			if _, uErr := db.QUpdate(ctx, tx, dialect, db.UpdateParams{
				Table: "content_data",
				Set:   authorSet,
				Where: authorWhere,
			}); uErr != nil {
				return r, fmt.Errorf("reassign content_data: %w", uErr)
			}
		}

		// Reassign datatypes
		if r.DatatypesReassigned > 0 {
			if _, uErr := db.QUpdate(ctx, tx, dialect, db.UpdateParams{
				Table: "datatypes",
				Set:   authorSet,
				Where: authorWhere,
			}); uErr != nil {
				return r, fmt.Errorf("reassign datatypes: %w", uErr)
			}
		}

		// Reassign admin_content_data
		if r.AdminContentDataReassigned > 0 {
			if _, uErr := db.QUpdate(ctx, tx, dialect, db.UpdateParams{
				Table: "admin_content_data",
				Set:   authorSet,
				Where: authorWhere,
			}); uErr != nil {
				return r, fmt.Errorf("reassign admin_content_data: %w", uErr)
			}
		}

		return r, nil
	})
	if err != nil {
		return nil, fmt.Errorf("reassign transaction: %w", err)
	}

	// Phase 2: audited delete after tx commits
	if err := s.driver.DeleteUser(ctx, ac, input.UserID); err != nil {
		return nil, fmt.Errorf("delete user after reassign: %w", err)
	}

	return &result, nil
}

// --- Private Helpers ---

// validateCreateUserInput checks required fields for user creation.
func validateCreateUserInput(input CreateUserInput) *ValidationError {
	ve := &ValidationError{}
	if input.Username == "" {
		ve.Add("username", "username is required")
	}
	if input.Name == "" {
		ve.Add("name", "name is required")
	}
	if string(input.Email) == "" {
		ve.Add("email", "email is required")
	}
	if input.Password == "" {
		ve.Add("password", "password is required")
	} else if len(input.Password) < 8 {
		ve.Add("password", "password must be at least 8 characters")
	}
	if ve.HasErrors() {
		return ve
	}
	return nil
}

// validateUpdateUserInput checks required fields for user update.
func validateUpdateUserInput(input UpdateUserInput) *ValidationError {
	ve := &ValidationError{}
	if input.Username == "" {
		ve.Add("username", "username is required")
	}
	if input.Name == "" {
		ve.Add("name", "name is required")
	}
	if string(input.Email) == "" {
		ve.Add("email", "email is required")
	}
	if ve.HasErrors() {
		return ve
	}
	return nil
}

// checkEmailUniqueness checks if the email is already taken.
// excludeUserID is used for updates to exclude the current user (empty for create).
func (s *UserService) checkEmailUniqueness(ctx context.Context, email types.Email, excludeUserID string) error {
	conn, _, err := s.driver.GetConnection()
	if err != nil {
		return fmt.Errorf("get connection for email check: %w", err)
	}
	cfg, err := s.mgr.Config()
	if err != nil {
		return fmt.Errorf("get config for email check: %w", err)
	}
	dialect := db.DialectFromString(string(cfg.Db_Driver))

	where := map[string]any{"email": string(email)}
	if excludeUserID != "" {
		where["user_id"] = db.Neq(excludeUserID)
	}

	exists, err := db.QExists(ctx, conn, dialect, "users", where)
	if err != nil {
		return fmt.Errorf("check email uniqueness: %w", err)
	}
	if exists {
		return &ConflictError{Resource: "user", ID: string(email), Detail: "email already exists"}
	}
	return nil
}

// checkUsernameUniqueness checks if the username is already taken.
// excludeUserID is used for updates to exclude the current user (empty for create).
func (s *UserService) checkUsernameUniqueness(ctx context.Context, username string, excludeUserID string) error {
	conn, _, err := s.driver.GetConnection()
	if err != nil {
		return fmt.Errorf("get connection for username check: %w", err)
	}
	cfg, err := s.mgr.Config()
	if err != nil {
		return fmt.Errorf("get config for username check: %w", err)
	}
	dialect := db.DialectFromString(string(cfg.Db_Driver))

	where := map[string]any{"username": username}
	if excludeUserID != "" {
		where["user_id"] = db.Neq(excludeUserID)
	}

	exists, err := db.QExists(ctx, conn, dialect, "users", where)
	if err != nil {
		return fmt.Errorf("check username uniqueness: %w", err)
	}
	if exists {
		return &ConflictError{Resource: "user", ID: username, Detail: "username already exists"}
	}
	return nil
}

// resolveDefaultRole returns the viewer role ID if roleInput is zero,
// otherwise returns the input unchanged.
func (s *UserService) resolveDefaultRole(roleInput types.RoleID) (types.RoleID, error) {
	if !roleInput.IsZero() {
		return roleInput, nil
	}
	viewerRole, err := s.driver.GetRoleByLabel("viewer")
	if err != nil {
		return "", fmt.Errorf("get viewer role for default assignment: %w", err)
	}
	return viewerRole.RoleID, nil
}
