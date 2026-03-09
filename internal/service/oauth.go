package service

import (
	"context"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// OAuthService manages OAuth provider connection CRUD.
type OAuthService struct {
	driver db.DbDriver
}

// NewOAuthService creates an OAuthService.
func NewOAuthService(driver db.DbDriver) *OAuthService {
	return &OAuthService{driver: driver}
}

// CreateUserOauth creates a new OAuth provider connection for a user.
func (s *OAuthService) CreateUserOauth(ctx context.Context, ac audited.AuditContext, params db.CreateUserOauthParams) (*db.UserOauth, error) {
	return s.driver.CreateUserOauth(ctx, ac, params)
}

// UpdateUserOauth updates an existing OAuth provider connection.
func (s *OAuthService) UpdateUserOauth(ctx context.Context, ac audited.AuditContext, params db.UpdateUserOauthParams) (*string, error) {
	return s.driver.UpdateUserOauth(ctx, ac, params)
}

// DeleteUserOauth removes an OAuth provider connection by ID.
func (s *OAuthService) DeleteUserOauth(ctx context.Context, ac audited.AuditContext, oauthID types.UserOauthID) error {
	return s.driver.DeleteUserOauth(ctx, ac, oauthID)
}
