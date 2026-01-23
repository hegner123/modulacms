package db

import (
	"context"

	"github.com/hegner123/modulacms/internal/db/types"
)

// ChangeEventBuilder is a convenience builder for creating change events.
// The actual ChangeEvent struct is defined in change_event.go as the database wrapper.
type ChangeEventBuilder struct {
	EventID       types.EventID        `json:"event_id"`
	HLCTimestamp  types.HLC            `json:"hlc_timestamp"`
	WallTimestamp types.Timestamp      `json:"wall_timestamp"`
	NodeID        types.NodeID         `json:"node_id"`
	TableName     string               `json:"table_name"`
	RecordID      string               `json:"record_id"`
	Operation     types.Operation      `json:"operation"`
	Action        types.Action         `json:"action,omitempty"`
	UserID        types.NullableUserID `json:"user_id,omitempty"`
	OldValues     any                  `json:"old_values,omitempty"`
	NewValues     any                  `json:"new_values,omitempty"`
	Metadata      any                  `json:"metadata,omitempty"`
	SyncedAt      types.Timestamp      `json:"synced_at,omitempty"`
	ConsumedAt    types.Timestamp      `json:"consumed_at,omitempty"`
}

// EventLogger interface for change event operations
type EventLogger interface {
	// LogEvent records a change event
	LogEvent(ctx context.Context, event ChangeEvent) error

	// GetEventsByRecord retrieves events for a specific record
	GetEventsByRecord(ctx context.Context, tableName, recordID string) ([]ChangeEvent, error)

	// GetEventsSince retrieves events after an HLC timestamp (for replication)
	GetEventsSince(ctx context.Context, hlc types.HLC, limit int) ([]ChangeEvent, error)

	// GetUnsyncedEvents retrieves events not yet synced to other nodes
	GetUnsyncedEvents(ctx context.Context, limit int) ([]ChangeEvent, error)

	// GetUnconsumedEvents retrieves events not yet processed by webhooks
	GetUnconsumedEvents(ctx context.Context, limit int) ([]ChangeEvent, error)

	// MarkSynced marks events as synced
	MarkSynced(ctx context.Context, eventIDs []types.EventID) error

	// MarkConsumed marks events as consumed by webhooks
	MarkConsumed(ctx context.Context, eventIDs []types.EventID) error
}

// NewChangeEventBuilder creates a change event builder for the current node
func NewChangeEventBuilder(nodeID types.NodeID, tableName, recordID string, op types.Operation, action types.Action, userID types.UserID) ChangeEventBuilder {
	return ChangeEventBuilder{
		EventID:       types.NewEventID(),
		HLCTimestamp:  types.HLCNow(),
		WallTimestamp: types.TimestampNow(),
		NodeID:        nodeID,
		TableName:     tableName,
		RecordID:      recordID,
		Operation:     op,
		Action:        action,
		UserID:        types.NullableUserID{ID: userID, Valid: userID != ""},
	}
}

// WithChanges adds old/new values to the event
func (e ChangeEventBuilder) WithChanges(oldVal, newVal any) ChangeEventBuilder {
	e.OldValues = oldVal
	e.NewValues = newVal
	return e
}

// WithMetadata adds metadata to the event
func (e ChangeEventBuilder) WithMetadata(meta any) ChangeEventBuilder {
	e.Metadata = meta
	return e
}
