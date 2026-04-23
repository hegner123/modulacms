package mcp

import (
	"context"

	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
)

// mcpAuditKey is used only in stdio mode to inject a default AuditContext
// when no HTTP middleware chain is available. In HTTP mode, the middleware
// chain populates the context with the authenticated user, permissions,
// and admin status directly. The tool middleware reads those values using
// the existing middleware extractors:
//   - middleware.AuthenticatedUser(ctx)
//   - middleware.ContextPermissions(ctx)
//   - middleware.ContextIsAdmin(ctx)
//   - middleware.RequestIDFromContext(ctx)
//   - middleware.ClientIPFromContext(ctx)
//
// No MCP-specific duplicates of these extractors are created.
type mcpAuditKey struct{}

// AuditContextFromMCP builds an AuditContext from the authenticated user
// in the MCP context. Uses the real request ID and client IP from the
// middleware chain when available. Falls back to a "mcp-anonymous" context
// if no user is present (should not happen after permission middleware,
// but defensive).
func AuditContextFromMCP(ctx context.Context) audited.AuditContext {
	requestID := middleware.RequestIDFromContext(ctx)
	clientIP := middleware.ClientIPFromContext(ctx)
	if clientIP == "" {
		clientIP = "127.0.0.1"
	}

	user := middleware.AuthenticatedUser(ctx)
	if user != nil {
		return audited.Ctx(types.NewNodeID(), user.UserID, requestID, clientIP)
	}

	// Stdio-mode fallback: check for injected AuditContext.
	if ac, ok := ctx.Value(mcpAuditKey{}).(audited.AuditContext); ok {
		return ac
	}

	return audited.Ctx(types.NewNodeID(), types.UserID("mcp-anonymous"), requestID, clientIP)
}
