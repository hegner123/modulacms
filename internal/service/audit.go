package service

import (
	"context"
	"fmt"

	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
)

// AuditCtx builds an AuditContext from an HTTP request context.
// Reads user ID from the authenticated session, request ID and client IP
// from middleware context values, and node ID from the config manager.
//
// Use this for requests originating from admin panel or API handlers.
func (r *Registry) AuditCtx(ctx context.Context) (audited.AuditContext, error) {
	cfg, err := r.mgr.Config()
	if err != nil {
		return audited.AuditContext{}, fmt.Errorf("audit context: config unavailable: %w", err)
	}

	user := middleware.AuthenticatedUser(ctx)
	var userID types.UserID
	if user != nil {
		userID = user.UserID
	}

	return audited.AuditContext{
		NodeID:    types.NodeID(cfg.Node_ID),
		UserID:    userID,
		RequestID: middleware.RequestIDFromContext(ctx),
		IP:        middleware.ClientIPFromContext(ctx),
	}, nil
}

// SystemAuditCtx builds an AuditContext for non-HTTP callers such as
// the TUI, scheduler, or CLI commands where no HTTP request context exists.
// The reason parameter serves as the request ID (e.g. "scheduled-publish",
// "tui-edit", "cli-import"). IP is set to "system".
func (r *Registry) SystemAuditCtx(userID types.UserID, reason string) (audited.AuditContext, error) {
	cfg, err := r.mgr.Config()
	if err != nil {
		return audited.AuditContext{}, fmt.Errorf("audit context: config unavailable: %w", err)
	}

	return audited.AuditContext{
		NodeID:    types.NodeID(cfg.Node_ID),
		UserID:    userID,
		RequestID: reason,
		IP:        "system",
	}, nil
}
