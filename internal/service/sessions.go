package service

import (
	"context"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// SessionService manages session CRUD operations.
type SessionService struct {
	driver db.DbDriver
}

// NewSessionService creates a SessionService.
func NewSessionService(driver db.DbDriver) *SessionService {
	return &SessionService{driver: driver}
}

// CreateSession creates a new session record.
func (s *SessionService) CreateSession(ctx context.Context, ac audited.AuditContext, params db.CreateSessionParams) (*db.Sessions, error) {
	session, err := s.driver.CreateSession(ctx, ac, params)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return session, nil
}

// GetSession retrieves a session by ID.
func (s *SessionService) GetSession(ctx context.Context, sessionID types.SessionID) (*db.Sessions, error) {
	session, err := s.driver.GetSession(sessionID)
	if err != nil {
		return nil, &NotFoundError{Resource: "session", ID: string(sessionID)}
	}
	return session, nil
}

// ListSessions returns all sessions.
func (s *SessionService) ListSessions(ctx context.Context) (*[]db.Sessions, error) {
	return s.driver.ListSessions()
}

// UpdateSession updates a session and returns the refreshed record.
func (s *SessionService) UpdateSession(ctx context.Context, ac audited.AuditContext, params db.UpdateSessionParams) (*db.Sessions, error) {
	_, err := s.driver.UpdateSession(ctx, ac, params)
	if err != nil {
		return nil, fmt.Errorf("update session: %w", err)
	}

	session, err := s.driver.GetSession(params.SessionID)
	if err != nil {
		return nil, fmt.Errorf("fetch updated session: %w", err)
	}
	return session, nil
}

// DeleteSession removes a session by ID.
func (s *SessionService) DeleteSession(ctx context.Context, ac audited.AuditContext, sessionID types.SessionID) error {
	if err := s.driver.DeleteSession(ctx, ac, sessionID); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
