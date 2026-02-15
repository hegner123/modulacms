package types

import (
	"testing"
)

func TestNewChangeEvent(t *testing.T) {
	t.Parallel()
	nodeID := NewNodeID()
	userID := NewUserID()

	event := NewChangeEvent(nodeID, "content_data", "record-123", OpInsert, ActionCreate, userID)

	// EventID should be non-empty
	if event.EventID.IsZero() {
		t.Error("EventID is zero")
	}
	if err := event.EventID.Validate(); err != nil {
		t.Errorf("EventID.Validate() = %v", err)
	}

	// HLC should be non-zero
	if event.HLCTimestamp == 0 {
		t.Error("HLCTimestamp is 0")
	}

	// WallTimestamp should be valid
	if !event.WallTimestamp.Valid {
		t.Error("WallTimestamp.Valid = false")
	}

	// Fields should match inputs
	if event.NodeID != nodeID {
		t.Errorf("NodeID = %q, want %q", event.NodeID, nodeID)
	}
	if event.TableName != "content_data" {
		t.Errorf("TableName = %q", event.TableName)
	}
	if event.RecordID != "record-123" {
		t.Errorf("RecordID = %q", event.RecordID)
	}
	if event.Operation != OpInsert {
		t.Errorf("Operation = %q", event.Operation)
	}
	if event.Action != ActionCreate {
		t.Errorf("Action = %q", event.Action)
	}

	// UserID should be set and valid
	if !event.UserID.Valid {
		t.Error("UserID.Valid = false")
	}
	if event.UserID.ID != userID {
		t.Errorf("UserID.ID = %q, want %q", event.UserID.ID, userID)
	}
}

func TestNewChangeEvent_EmptyUserID(t *testing.T) {
	t.Parallel()
	nodeID := NewNodeID()
	event := NewChangeEvent(nodeID, "content_data", "record-123", OpUpdate, ActionUpdate, "")

	// Empty userID should result in Invalid nullable
	if event.UserID.Valid {
		t.Error("empty UserID should result in Valid=false")
	}
}

func TestChangeEvent_WithChanges(t *testing.T) {
	t.Parallel()
	nodeID := NewNodeID()
	event := NewChangeEvent(nodeID, "content_data", "record-123", OpUpdate, ActionUpdate, "")

	oldVal := map[string]string{"name": "old"}
	newVal := map[string]string{"name": "new"}

	updated := event.WithChanges(oldVal, newVal)

	// Original should be unchanged (value receiver)
	if event.OldValues != nil {
		t.Error("original event OldValues should be nil")
	}

	// Updated should have the values
	if updated.OldValues == nil {
		t.Error("updated OldValues is nil")
	}
	if updated.NewValues == nil {
		t.Error("updated NewValues is nil")
	}
}

func TestChangeEvent_WithMetadata(t *testing.T) {
	t.Parallel()
	nodeID := NewNodeID()
	event := NewChangeEvent(nodeID, "content_data", "record-123", OpDelete, ActionDelete, "")

	meta := map[string]string{"reason": "cleanup"}
	updated := event.WithMetadata(meta)

	// Original unchanged
	if event.Metadata != nil {
		t.Error("original Metadata should be nil")
	}
	// Updated has metadata
	if updated.Metadata == nil {
		t.Error("updated Metadata is nil")
	}
}

func TestChangeEvent_Chaining(t *testing.T) {
	t.Parallel()
	nodeID := NewNodeID()
	userID := NewUserID()

	event := NewChangeEvent(nodeID, "users", "user-1", OpUpdate, ActionUpdate, userID).
		WithChanges(map[string]string{"old": "a"}, map[string]string{"new": "b"}).
		WithMetadata(map[string]string{"ip": "127.0.0.1"})

	if event.OldValues == nil {
		t.Error("chained OldValues is nil")
	}
	if event.Metadata == nil {
		t.Error("chained Metadata is nil")
	}
	if event.Operation != OpUpdate {
		t.Errorf("chained Operation = %q", event.Operation)
	}
}
