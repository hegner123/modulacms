package audited

import "github.com/hegner123/modulacms/internal/db/types"

// AuditContext carries metadata for audit records.
// HookRunner is nil when hooks are disabled (CLI, TUI, bootstrap), backward
// compatible with all existing callers. When non-nil, the audited Create/Update/Delete
// functions invoke hooks at the appropriate lifecycle points.
type AuditContext struct {
	NodeID     types.NodeID
	UserID     types.UserID
	RequestID  string
	IP         string
	HookRunner HookRunner // nil = no hooks, backward compatible
}

// Ctx is a brief constructor for AuditContext.
func Ctx(nodeID types.NodeID, userID types.UserID, requestID, ip string) AuditContext {
	return AuditContext{
		NodeID:    nodeID,
		UserID:    userID,
		RequestID: requestID,
		IP:        ip,
	}
}
