package middleware

import (
	"context"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// hookRunnerKey is the context key for the HookRunner injected by middleware.
// Uses a private type to prevent collisions with other packages.
type hookRunnerKeyType struct{}

var hookRunnerKey = hookRunnerKeyType{}

// WithHookRunner returns a new context with the given HookRunner stored.
// Used by the serve command to inject the plugin hook engine into the request
// context so that AuditContextFromRequest can pass it to audited operations.
func WithHookRunner(ctx context.Context, runner audited.HookRunner) context.Context {
	return context.WithValue(ctx, hookRunnerKey, runner)
}

// hookRunnerFromContext extracts the HookRunner from the context, or nil if absent.
func hookRunnerFromContext(ctx context.Context) audited.HookRunner {
	runner, _ := ctx.Value(hookRunnerKey).(audited.HookRunner)
	return runner
}

// AuditContextFromRequest builds an AuditContext from the HTTP request.
// Extracts the authenticated user, client IP, request ID, and HookRunner from
// the context. HookRunner is nil when not injected (CLI/TUI, or when plugins
// are disabled), which is backward compatible.
func AuditContextFromRequest(r *http.Request, c config.Config) audited.AuditContext {
	var userID types.UserID
	if user := AuthenticatedUser(r.Context()); user != nil {
		userID = user.UserID
	}

	ip := ClientIPFromContext(r.Context())

	actx := audited.Ctx(
		types.NodeID(c.Node_ID),
		userID,
		RequestIDFromContext(r.Context()),
		ip,
	)

	// Phase 3: Inject HookRunner if present in the request context.
	actx.HookRunner = hookRunnerFromContext(r.Context())

	return actx
}

// AuditContextFromCLI builds an AuditContext for CLI/TUI operations.
// CLI/TUI operations never have a HookRunner (hooks are HTTP-only).
func AuditContextFromCLI(c config.Config, userID types.UserID) audited.AuditContext {
	return audited.Ctx(
		types.NodeID(c.Node_ID),
		userID,
		"",
		"cli",
	)
}

// readHookRunnerKeyType is the context key for ReadHookRunner.
type readHookRunnerKeyType struct{}

var readHookRunnerKey = readHookRunnerKeyType{}

// ReadHookRunnerFromContext extracts the ReadHookRunner from the context, or nil
// if absent (plugins disabled). Callers must handle nil — skip hook calls and
// proceed with normal delivery.
func ReadHookRunnerFromContext(ctx context.Context) audited.ReadHookRunner {
	runner, _ := ctx.Value(readHookRunnerKey).(audited.ReadHookRunner)
	return runner
}

// ReadHookRunnerMiddleware returns a middleware that injects the given
// ReadHookRunner into every request's context. The delivery handler extracts
// it via ReadHookRunnerFromContext to dispatch before_read/after_read hooks.
//
// Returns a no-op middleware if runner is nil (plugins disabled).
func ReadHookRunnerMiddleware(runner audited.ReadHookRunner) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if runner == nil {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), readHookRunnerKey, runner)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// HookRunnerMiddleware returns a middleware that injects the given HookRunner into
// every request's context. This makes it available to AuditContextFromRequest,
// which passes it through to audited Create/Update/Delete for hook dispatch.
//
// Returns a no-op middleware if runner is nil (plugins disabled).
func HookRunnerMiddleware(runner audited.HookRunner) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if runner == nil {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := WithHookRunner(r.Context(), runner)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
