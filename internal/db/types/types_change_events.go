package types

import (
	"context"
)

// ChangeEvent represents a row in the change_events table.
// It serves as:
// 1. Audit trail - Who changed what, when
// 2. Replication log - What to sync to other nodes
// 3. Webhook source - What events to fire
type ChangeEvent struct {
	EventID       EventID        `json:"event_id"`
	HLCTimestamp  HLC            `json:"hlc_timestamp"`
	WallTimestamp Timestamp      `json:"wall_timestamp"`
	NodeID        NodeID         `json:"node_id"`
	TableName     string         `json:"table_name"`
	RecordID      string         `json:"record_id"`
	Operation     Operation      `json:"operation"`
	Action        Action         `json:"action,omitempty"`
	UserID        NullableUserID `json:"user_id,omitempty"`
	OldValues     any            `json:"old_values,omitempty"`
	NewValues     any            `json:"new_values,omitempty"`
	Metadata      any            `json:"metadata,omitempty"`
	SyncedAt      Timestamp      `json:"synced_at,omitempty"`
	ConsumedAt    Timestamp      `json:"consumed_at,omitempty"`
}

// EventLogger interface for change event operations
type EventLogger interface {
	// LogEvent records a change event
	LogEvent(ctx context.Context, event ChangeEvent) error

	// GetEventsByRecord retrieves events for a specific record
	GetEventsByRecord(ctx context.Context, tableName, recordID string) ([]ChangeEvent, error)

	// GetEventsSince retrieves events after an HLC timestamp (for replication)
	GetEventsSince(ctx context.Context, hlc HLC, limit int) ([]ChangeEvent, error)

	// GetUnsyncedEvents retrieves events not yet synced to other nodes
	GetUnsyncedEvents(ctx context.Context, limit int) ([]ChangeEvent, error)

	// GetUnconsumedEvents retrieves events not yet processed by webhooks
	GetUnconsumedEvents(ctx context.Context, limit int) ([]ChangeEvent, error)

	// MarkSynced marks events as synced
	MarkSynced(ctx context.Context, eventIDs []EventID) error

	// MarkConsumed marks events as consumed by webhooks
	MarkConsumed(ctx context.Context, eventIDs []EventID) error
}

// NewChangeEvent creates a change event for the current node
func NewChangeEvent(nodeID NodeID, tableName, recordID string, op Operation, action Action, userID UserID) ChangeEvent {
	return ChangeEvent{
		EventID:       NewEventID(),
		HLCTimestamp:  HLCNow(),
		WallTimestamp: TimestampNow(),
		NodeID:        nodeID,
		TableName:     tableName,
		RecordID:      recordID,
		Operation:     op,
		Action:        action,
		UserID:        NullableUserID{ID: userID, Valid: userID != ""},
	}
}

// WithChanges adds old/new values to the event
func (e ChangeEvent) WithChanges(oldVal, newVal any) ChangeEvent {
	e.OldValues = oldVal
	e.NewValues = newVal
	return e
}

// WithMetadata adds metadata to the event
func (e ChangeEvent) WithMetadata(meta any) ChangeEvent {
	e.Metadata = meta
	return e
}
