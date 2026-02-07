package audited

import "github.com/hegner123/modulacms/internal/db/types"

// ChangeEventParams contains all fields needed to record a change event.
// Used by ChangeEventRecorder implementations.
type ChangeEventParams struct {
	EventID      types.EventID
	HlcTimestamp types.HLC
	NodeID       types.NodeID
	TableName    string
	RecordID     string
	Operation    types.Operation
	Action       types.Action
	UserID       types.NullableUserID
	OldValues    types.JSONData
	NewValues    types.JSONData
	Metadata     types.JSONData
	RequestID    types.NullableString
	IP           types.NullableString
}
