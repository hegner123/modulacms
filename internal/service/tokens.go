package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
)

// CreateTokenInput holds the caller-provided fields for token creation.
// The raw token value and its hash are generated server-side by CreateToken.
type CreateTokenInput struct {
	UserID    types.NullableUserID
	TokenType string
	Label     string
	Expiry    types.Timestamp
}

// CreateTokenResult contains the persisted token record and the raw token
// value. The raw value is only available at creation time and must be shown
// to the user immediately -- it is never stored or retrievable again.
type CreateTokenResult struct {
	Token    *db.Tokens
	RawToken string
}

// TokenService manages authentication token CRUD with server-side generation and hashing.
type TokenService struct {
	driver db.DbDriver
}

// NewTokenService creates a TokenService.
func NewTokenService(driver db.DbDriver) *TokenService {
	return &TokenService{driver: driver}
}

// GetToken retrieves a token by ID. Returns NotFoundError if not found.
func (s *TokenService) GetToken(ctx context.Context, tokenID string) (*db.Tokens, error) {
	token, err := s.driver.GetToken(tokenID)
	if err != nil {
		return nil, &NotFoundError{Resource: "token", ID: tokenID}
	}
	return token, nil
}

// ListTokens returns all tokens.
func (s *TokenService) ListTokens(ctx context.Context) (*[]db.Tokens, error) {
	return s.driver.ListTokens()
}

// CreateToken generates a cryptographically random token, hashes it for
// storage, and returns both the persisted record and the one-time raw value.
//
// The raw token format is "mcms_" followed by 64 hex characters (32 random
// bytes). Only the SHA-256 hash is stored in the database.
func (s *TokenService) CreateToken(ctx context.Context, ac audited.AuditContext, input CreateTokenInput) (*CreateTokenResult, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("generate token bytes: %w", err)
	}
	rawToken := "mcms_" + hex.EncodeToString(tokenBytes)
	hashedToken := utility.HashToken(rawToken)

	params := db.CreateTokenParams{
		UserID:    input.UserID,
		TokenType: input.TokenType,
		Token:     hashedToken,
		IssuedAt:  types.TimestampNow(),
		ExpiresAt: input.Expiry,
		Revoked:   false,
	}

	created, err := s.driver.CreateToken(ctx, ac, params)
	if err != nil {
		return nil, fmt.Errorf("create token: %w", err)
	}

	return &CreateTokenResult{
		Token:    created,
		RawToken: rawToken,
	}, nil
}

// UpdateToken updates a token and returns the refreshed record.
// The driver's UpdateToken returns *string (a message), so we fetch the
// updated record via GetToken to return the full object.
func (s *TokenService) UpdateToken(ctx context.Context, ac audited.AuditContext, params db.UpdateTokenParams) (*db.Tokens, error) {
	_, err := s.driver.UpdateToken(ctx, ac, params)
	if err != nil {
		return nil, fmt.Errorf("update token: %w", err)
	}

	token, err := s.driver.GetToken(params.ID)
	if err != nil {
		return nil, fmt.Errorf("fetch updated token: %w", err)
	}
	return token, nil
}

// DeleteToken removes a token by ID.
func (s *TokenService) DeleteToken(ctx context.Context, ac audited.AuditContext, tokenID string) error {
	if err := s.driver.DeleteToken(ctx, ac, tokenID); err != nil {
		return fmt.Errorf("delete token: %w", err)
	}
	return nil
}
