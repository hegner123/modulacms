package middleware

import (
	"net"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// AuditContextFromRequest builds an AuditContext from the HTTP request.
// Extracts the authenticated user, client IP, and request ID from the context.
func AuditContextFromRequest(r *http.Request, c config.Config) audited.AuditContext {
	var userID types.UserID
	if user := AuthenticatedUser(r.Context()); user != nil {
		userID = user.UserID
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}

	return audited.Ctx(
		types.NodeID(c.Node_ID),
		userID,
		RequestIDFromContext(r.Context()),
		ip,
	)
}

// AuditContextFromCLI builds an AuditContext for CLI/TUI operations.
func AuditContextFromCLI(c config.Config, userID types.UserID) audited.AuditContext {
	return audited.Ctx(
		types.NodeID(c.Node_ID),
		userID,
		"",
		"cli",
	)
}
