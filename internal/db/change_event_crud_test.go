// Integration tests for the change_event entity CRUD lifecycle.
// Uses testIntegrationDB (Tier 0: no FK dependencies).
//
// ChangeEvent methods are NON-audited (no ctx/ac parameters on mutations).
package db

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

func TestDatabase_CRUD_ChangeEvent(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)

	nodeID := types.NodeID(d.Config.Node_ID)

	// --- Count: starts at zero ---
	count, err := d.CountChangeEvents()
	if err != nil {
		t.Fatalf("initial CountChangeEvents: %v", err)
	}
	if *count != 0 {
		t.Fatalf("initial CountChangeEvents = %d, want 0", *count)
	}

	// --- RecordChangeEvent (Create) ---
	eventID := types.NewEventID()
	hlc := types.HLCNow()
	// record_id has CHECK(length(record_id) = 26), use a ULID
	recordID := types.NewULID().String()

	created, err := d.RecordChangeEvent(RecordChangeEventParams{
		EventID:      eventID,
		HlcTimestamp: hlc,
		NodeID:       nodeID,
		TableName:    "test_table",
		RecordID:     recordID,
		Operation:    types.OpInsert,
		Action:       types.ActionCreate,
		UserID:       types.NullableUserID{},
		OldValues:    types.NewJSONData(nil),
		NewValues:    types.NewJSONData(nil),
		Metadata:     types.NewJSONData(nil),
		RequestID:    types.NullableString{},
		IP:           types.NullableString{},
	})
	if err != nil {
		t.Fatalf("RecordChangeEvent: %v", err)
	}
	if created == nil {
		t.Fatal("RecordChangeEvent returned nil")
	}
	if created.EventID != eventID {
		t.Errorf("EventID = %v, want %v", created.EventID, eventID)
	}
	if created.NodeID != nodeID {
		t.Errorf("NodeID = %v, want %v", created.NodeID, nodeID)
	}
	if created.TableName != "test_table" {
		t.Errorf("TableName = %q, want %q", created.TableName, "test_table")
	}
	if created.RecordID != recordID {
		t.Errorf("RecordID = %q, want %q", created.RecordID, recordID)
	}
	if created.Operation != types.OpInsert {
		t.Errorf("Operation = %v, want %v", created.Operation, types.OpInsert)
	}
	if created.Action != types.ActionCreate {
		t.Errorf("Action = %v, want %v", created.Action, types.ActionCreate)
	}

	// --- GetChangeEvent ---
	got, err := d.GetChangeEvent(eventID)
	if err != nil {
		t.Fatalf("GetChangeEvent: %v", err)
	}
	if got == nil {
		t.Fatal("GetChangeEvent returned nil")
	}
	if got.EventID != eventID {
		t.Errorf("GetChangeEvent EventID = %v, want %v", got.EventID, eventID)
	}
	if got.TableName != "test_table" {
		t.Errorf("GetChangeEvent TableName = %q, want %q", got.TableName, "test_table")
	}

	// --- GetChangeEventsByRecord ---
	byRecord, err := d.GetChangeEventsByRecord("test_table", recordID)
	if err != nil {
		t.Fatalf("GetChangeEventsByRecord: %v", err)
	}
	if byRecord == nil {
		t.Fatal("GetChangeEventsByRecord returned nil")
	}
	if len(*byRecord) != 1 {
		t.Fatalf("GetChangeEventsByRecord len = %d, want 1", len(*byRecord))
	}
	if (*byRecord)[0].EventID != eventID {
		t.Errorf("GetChangeEventsByRecord[0].EventID = %v, want %v", (*byRecord)[0].EventID, eventID)
	}

	// --- ListChangeEvents ---
	list, err := d.ListChangeEvents(ListChangeEventsParams{Limit: 100, Offset: 0})
	if err != nil {
		t.Fatalf("ListChangeEvents: %v", err)
	}
	if list == nil {
		t.Fatal("ListChangeEvents returned nil")
	}
	if len(*list) != 1 {
		t.Fatalf("ListChangeEvents len = %d, want 1", len(*list))
	}

	// --- Count: now 1 ---
	count, err = d.CountChangeEvents()
	if err != nil {
		t.Fatalf("CountChangeEvents after record: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountChangeEvents after record = %d, want 1", *count)
	}

	// --- DeleteChangeEvent ---
	err = d.DeleteChangeEvent(eventID)
	if err != nil {
		t.Fatalf("DeleteChangeEvent: %v", err)
	}

	// --- Get after delete: expect error ---
	_, err = d.GetChangeEvent(eventID)
	if err == nil {
		t.Fatal("GetChangeEvent after delete: expected error, got nil")
	}

	// --- Count: back to zero ---
	count, err = d.CountChangeEvents()
	if err != nil {
		t.Fatalf("CountChangeEvents after delete: %v", err)
	}
	if *count != 0 {
		t.Fatalf("CountChangeEvents after delete = %d, want 0", *count)
	}
}

// TestDatabase_CRUD_ChangeEvent_MarkSynced tests MarkEventSynced and
// verifies the event no longer appears in GetUnsyncedEvents.
func TestDatabase_CRUD_ChangeEvent_MarkSynced(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)

	nodeID := types.NodeID(d.Config.Node_ID)
	eventID := types.NewEventID()
	recordID := types.NewULID().String()

	_, err := d.RecordChangeEvent(RecordChangeEventParams{
		EventID:      eventID,
		HlcTimestamp: types.HLCNow(),
		NodeID:       nodeID,
		TableName:    "sync_table",
		RecordID:     recordID,
		Operation:    types.OpInsert,
		Action:       types.ActionCreate,
		UserID:       types.NullableUserID{},
		OldValues:    types.NewJSONData(nil),
		NewValues:    types.NewJSONData(nil),
		Metadata:     types.NewJSONData(nil),
		RequestID:    types.NullableString{},
		IP:           types.NullableString{},
	})
	if err != nil {
		t.Fatalf("RecordChangeEvent: %v", err)
	}

	// Should appear in unsynced events
	unsynced, err := d.GetUnsyncedEvents(100)
	if err != nil {
		t.Fatalf("GetUnsyncedEvents: %v", err)
	}
	if unsynced == nil || len(*unsynced) == 0 {
		t.Fatal("GetUnsyncedEvents returned empty; expected the recorded event")
	}

	// Mark synced
	err = d.MarkEventSynced(eventID)
	if err != nil {
		t.Fatalf("MarkEventSynced: %v", err)
	}

	// Should no longer appear in unsynced events
	unsynced, err = d.GetUnsyncedEvents(100)
	if err != nil {
		t.Fatalf("GetUnsyncedEvents after mark: %v", err)
	}
	for _, e := range *unsynced {
		if e.EventID == eventID {
			t.Fatalf("event %v still appears in unsynced events after MarkEventSynced", eventID)
		}
	}
}

// TestDatabase_CRUD_ChangeEvent_MarkConsumed tests MarkEventConsumed and
// verifies the event no longer appears in GetUnconsumedEvents.
func TestDatabase_CRUD_ChangeEvent_MarkConsumed(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)

	nodeID := types.NodeID(d.Config.Node_ID)
	eventID := types.NewEventID()
	recordID := types.NewULID().String()

	_, err := d.RecordChangeEvent(RecordChangeEventParams{
		EventID:      eventID,
		HlcTimestamp: types.HLCNow(),
		NodeID:       nodeID,
		TableName:    "consume_table",
		RecordID:     recordID,
		Operation:    types.OpUpdate,
		Action:       types.ActionUpdate,
		UserID:       types.NullableUserID{},
		OldValues:    types.NewJSONData(nil),
		NewValues:    types.NewJSONData(nil),
		Metadata:     types.NewJSONData(nil),
		RequestID:    types.NullableString{},
		IP:           types.NullableString{},
	})
	if err != nil {
		t.Fatalf("RecordChangeEvent: %v", err)
	}

	// Should appear in unconsumed events
	unconsumed, err := d.GetUnconsumedEvents(100)
	if err != nil {
		t.Fatalf("GetUnconsumedEvents: %v", err)
	}
	if unconsumed == nil || len(*unconsumed) == 0 {
		t.Fatal("GetUnconsumedEvents returned empty; expected the recorded event")
	}

	// Mark consumed
	err = d.MarkEventConsumed(eventID)
	if err != nil {
		t.Fatalf("MarkEventConsumed: %v", err)
	}

	// Should no longer appear in unconsumed events
	unconsumed, err = d.GetUnconsumedEvents(100)
	if err != nil {
		t.Fatalf("GetUnconsumedEvents after mark: %v", err)
	}
	for _, e := range *unconsumed {
		if e.EventID == eventID {
			t.Fatalf("event %v still appears in unconsumed events after MarkEventConsumed", eventID)
		}
	}
}

// TestDatabase_CRUD_ChangeEvent_WithOptionalFields verifies that nullable
// fields (UserID, RequestID, IP) are correctly stored and retrieved.
func TestDatabase_CRUD_ChangeEvent_WithOptionalFields(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)

	nodeID := types.NodeID(d.Config.Node_ID)
	eventID := types.NewEventID()
	recordID := types.NewULID().String()
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	reqID := types.NewNullableString("test-req-001")
	ip := types.NewNullableString("192.168.1.100")

	_, err := d.RecordChangeEvent(RecordChangeEventParams{
		EventID:      eventID,
		HlcTimestamp: types.HLCNow(),
		NodeID:       nodeID,
		TableName:    "optional_table",
		RecordID:     recordID,
		Operation:    types.OpDelete,
		Action:       types.ActionDelete,
		UserID:       userID,
		OldValues:    types.NewJSONData(nil),
		NewValues:    types.NewJSONData(nil),
		Metadata:     types.NewJSONData(nil),
		RequestID:    reqID,
		IP:           ip,
	})
	if err != nil {
		t.Fatalf("RecordChangeEvent: %v", err)
	}

	got, err := d.GetChangeEvent(eventID)
	if err != nil {
		t.Fatalf("GetChangeEvent: %v", err)
	}

	if !got.UserID.Valid {
		t.Error("UserID.Valid = false, want true")
	}
	if got.UserID.ID != userID.ID {
		t.Errorf("UserID.ID = %v, want %v", got.UserID.ID, userID.ID)
	}
	if !got.RequestID.Valid {
		t.Error("RequestID.Valid = false, want true")
	}
	if got.RequestID.String != "test-req-001" {
		t.Errorf("RequestID.String = %q, want %q", got.RequestID.String, "test-req-001")
	}
	if !got.IP.Valid {
		t.Error("IP.Valid = false, want true")
	}
	if got.IP.String != "192.168.1.100" {
		t.Errorf("IP.String = %q, want %q", got.IP.String, "192.168.1.100")
	}
}
