package audited

import "github.com/hegner123/modulacms/internal/db/types"

// AuditContext carries metadata for audit records.
type AuditContext struct {
	NodeID    types.NodeID
	UserID    types.UserID
	RequestID string
	IP        string
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
