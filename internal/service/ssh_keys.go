package service

import (
	"context"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
)

// SSHKeyService manages SSH key CRUD with fingerprint extraction and ownership checks.
type SSHKeyService struct {
	driver db.DbDriver
}

// NewSSHKeyService creates an SSHKeyService.
func NewSSHKeyService(driver db.DbDriver) *SSHKeyService {
	return &SSHKeyService{driver: driver}
}

// AddSSHKeyInput contains parameters for adding a new SSH key.
type AddSSHKeyInput struct {
	UserID    types.UserID
	PublicKey string
	Label     string
}

// AddKey parses the public key, checks for duplicates by fingerprint, and
// creates a new SSH key record. Returns ConflictError if the fingerprint is
// already registered.
func (s *SSHKeyService) AddKey(ctx context.Context, ac audited.AuditContext, input AddSSHKeyInput) (*db.UserSshKeys, error) {
	keyType, fingerprint, err := middleware.ParseSSHPublicKey(input.PublicKey)
	if err != nil {
		return nil, NewValidationError("public_key", fmt.Sprintf("invalid SSH public key: %v", err))
	}

	existing, _ := s.driver.GetUserSshKeyByFingerprint(fingerprint)
	if existing != nil {
		return nil, &ConflictError{Resource: "ssh_key", Detail: "SSH key already registered"}
	}

	created, err := s.driver.CreateUserSshKey(ctx, ac, db.CreateUserSshKeyParams{
		UserID:      types.NullableUserID{ID: input.UserID, Valid: true},
		PublicKey:   input.PublicKey,
		KeyType:     keyType,
		Fingerprint: fingerprint,
		Label:       input.Label,
		DateCreated: types.TimestampNow(),
	})
	if err != nil {
		return nil, fmt.Errorf("create ssh key: %w", err)
	}

	return created, nil
}

// ListKeys returns all SSH keys for the given user.
func (s *SSHKeyService) ListKeys(ctx context.Context, userID types.NullableUserID) (*[]db.UserSshKeys, error) {
	keys, err := s.driver.ListUserSshKeys(userID)
	if err != nil {
		return nil, fmt.Errorf("list ssh keys: %w", err)
	}
	return keys, nil
}

// GetKeyByFingerprint retrieves a single SSH key by its fingerprint.
// Returns NotFoundError if no key matches.
func (s *SSHKeyService) GetKeyByFingerprint(ctx context.Context, fingerprint string) (*db.UserSshKeys, error) {
	key, err := s.driver.GetUserSshKeyByFingerprint(fingerprint)
	if err != nil {
		return nil, &NotFoundError{Resource: "ssh_key", ID: fingerprint}
	}
	return key, nil
}

// DeleteKey verifies ownership and deletes an SSH key. Returns ForbiddenError
// if the key belongs to a different user, NotFoundError if the key does not exist.
func (s *SSHKeyService) DeleteKey(ctx context.Context, ac audited.AuditContext, userID types.UserID, keyID string) error {
	key, err := s.driver.GetUserSshKey(keyID)
	if err != nil {
		return &NotFoundError{Resource: "ssh_key", ID: keyID}
	}

	if !key.UserID.Valid || key.UserID.ID != userID {
		return &ForbiddenError{Message: "cannot delete another user's SSH key"}
	}

	if err := s.driver.DeleteUserSshKey(ctx, ac, keyID); err != nil {
		return fmt.Errorf("delete ssh key: %w", err)
	}

	return nil
}
