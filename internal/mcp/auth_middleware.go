package mcp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/middleware"
)

const errAuthRequired = "authentication required"
const errForbidden = "forbidden: requires permission '%s'"

// PassthroughHTTPContextFunc returns an HTTPContextFunc that is a no-op.
// The DefaultMiddlewareChain already populates all auth context values
// before this runs.
func PassthroughHTTPContextFunc() server.HTTPContextFunc {
	return func(ctx context.Context, r *http.Request) context.Context {
		return ctx
	}
}

// PermissionMiddleware returns a ToolHandlerMiddleware that checks the
// authenticated user's permissions before allowing a tool call.
// It reads auth state from context using the middleware package's
// exported extractors (not MCP-specific keys).
func PermissionMiddleware() server.ToolHandlerMiddleware {
	return func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
		return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			toolName := request.Params.Name

			// Look up required permission for this tool.
			permission, requiresAuth := toolPermissions[toolName]

			// If not in the map, check if it is a public tool.
			if !requiresAuth {
				if publicTools[toolName] {
					return next(ctx, request)
				}
				// Tool not in permission map and not public. This covers
				// connection tools (registered only with cm != nil) and any
				// future tools that don't need permission checks.
				return next(ctx, request)
			}

			// Authentication gate: must run before admin bypass check.
			user := middleware.AuthenticatedUser(ctx)
			if user == nil {
				return mcp.NewToolResultError(errAuthRequired), nil
			}

			// Admin bypass: admins skip permission checks.
			if middleware.ContextIsAdmin(ctx) {
				return next(ctx, request)
			}

			// Check permission set.
			ps := middleware.ContextPermissions(ctx)
			if ps == nil {
				return mcp.NewToolResultError(errAuthRequired), nil
			}

			if !ps.Has(permission) {
				return mcp.NewToolResultError(fmt.Sprintf(errForbidden, permission)), nil
			}

			return next(ctx, request)
		}
	}
}

// injectAuditContextMiddleware returns a ToolHandlerMiddleware that stores
// the provided AuditContext under mcpAuditKey{} in the tool call's context.
// Used by ServeDirect (stdio mode) where no HTTP middleware chain is present.
// AuditContextFromMCP checks for this key as a fallback when
// middleware.AuthenticatedUser(ctx) returns nil.
func injectAuditContextMiddleware(ac audited.AuditContext) server.ToolHandlerMiddleware {
	return func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
		return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			ctx = context.WithValue(ctx, mcpAuditKey{}, ac)
			return next(ctx, request)
		}
	}
}
