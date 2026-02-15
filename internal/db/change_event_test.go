// White-box tests for change_event.go: ChangeEvent mapper methods across all
// three database drivers, parameter type mappings, field name bridging
// (RequestId/Ip <-> RequestID/IP), and cross-database consistency.
//
// White-box access is needed because:
//   - The mapper methods are defined on the unexported-field Database,
//     MysqlDatabase, and PsqlDatabase receiver types that live in this package.
//   - Testing the RequestId/Ip -> RequestID/IP field name bridging requires
//     direct access to driver-specific sqlc types and the wrapper mappers.
package db

import (
	"testing"
	"time"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/types"
)

// --- SQLite Database.MapChangeEvent tests ---

func TestDatabase_MapChangeEvent_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	eventID := types.NewEventID()
	nodeID := types.NewNodeID()
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 3, 15, 10, 30, 0, 0, time.UTC))
	hlc := types.HLCNow()
	// Use nil-Data JSONData values to keep the struct comparable with !=.
	// JSONData with map/slice Data values are not comparable; we test the
	// .Valid flag pass-through instead.
	oldVals := types.NewJSONData(nil)
	newVals := types.NewJSONData(nil)
	meta := types.NewJSONData(nil)
	reqID := types.NewNullableString("req-sqlite-001")
	ip := types.NewNullableString("192.168.1.100")

	input := mdb.ChangeEvent{
		EventID:       eventID,
		HlcTimestamp:  hlc,
		WallTimestamp: ts,
		NodeID:        nodeID,
		TableName:     "users",
		RecordID:      "user-abc",
		Operation:     types.OpUpdate,
		Action:        types.ActionUpdate,
		UserID:        userID,
		OldValues:     oldVals,
		NewValues:     newVals,
		Metadata:      meta,
		RequestId:     reqID, // sqlc field name
		Ip:            ip,    // sqlc field name
		SyncedAt:      ts,
		ConsumedAt:    ts,
	}

	got := d.MapChangeEvent(input)

	if got.EventID != eventID {
		t.Errorf("EventID = %v, want %v", got.EventID, eventID)
	}
	if got.HlcTimestamp != hlc {
		t.Errorf("HlcTimestamp = %v, want %v", got.HlcTimestamp, hlc)
	}
	if got.WallTimestamp != ts {
		t.Errorf("WallTimestamp = %v, want %v", got.WallTimestamp, ts)
	}
	if got.NodeID != nodeID {
		t.Errorf("NodeID = %v, want %v", got.NodeID, nodeID)
	}
	if got.TableName != "users" {
		t.Errorf("TableName = %q, want %q", got.TableName, "users")
	}
	if got.RecordID != "user-abc" {
		t.Errorf("RecordID = %q, want %q", got.RecordID, "user-abc")
	}
	if got.Operation != types.OpUpdate {
		t.Errorf("Operation = %v, want %v", got.Operation, types.OpUpdate)
	}
	if got.Action != types.ActionUpdate {
		t.Errorf("Action = %v, want %v", got.Action, types.ActionUpdate)
	}
	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	// JSONData fields: verify Valid flag is preserved through mapping
	if got.OldValues.Valid != oldVals.Valid {
		t.Errorf("OldValues.Valid = %v, want %v", got.OldValues.Valid, oldVals.Valid)
	}
	if got.NewValues.Valid != newVals.Valid {
		t.Errorf("NewValues.Valid = %v, want %v", got.NewValues.Valid, newVals.Valid)
	}
	if got.Metadata.Valid != meta.Valid {
		t.Errorf("Metadata.Valid = %v, want %v", got.Metadata.Valid, meta.Valid)
	}
	// Critical field name mapping: sqlc RequestId -> wrapper RequestID
	if got.RequestID != reqID {
		t.Errorf("RequestID = %v, want %v", got.RequestID, reqID)
	}
	// Critical field name mapping: sqlc Ip -> wrapper IP
	if got.IP != ip {
		t.Errorf("IP = %v, want %v", got.IP, ip)
	}
	if got.SyncedAt != ts {
		t.Errorf("SyncedAt = %v, want %v", got.SyncedAt, ts)
	}
	if got.ConsumedAt != ts {
		t.Errorf("ConsumedAt = %v, want %v", got.ConsumedAt, ts)
	}
}

func TestDatabase_MapChangeEvent_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapChangeEvent(mdb.ChangeEvent{})

	if got.EventID != "" {
		t.Errorf("EventID = %v, want zero value", got.EventID)
	}
	if got.NodeID != "" {
		t.Errorf("NodeID = %v, want zero value", got.NodeID)
	}
	if got.TableName != "" {
		t.Errorf("TableName = %q, want empty string", got.TableName)
	}
	if got.RecordID != "" {
		t.Errorf("RecordID = %q, want empty string", got.RecordID)
	}
	if got.Operation != "" {
		t.Errorf("Operation = %v, want zero value", got.Operation)
	}
	if got.Action != "" {
		t.Errorf("Action = %v, want zero value", got.Action)
	}
	if got.UserID.Valid {
		t.Error("UserID.Valid = true, want false")
	}
	if got.RequestID.Valid {
		t.Error("RequestID.Valid = true, want false")
	}
	if got.IP.Valid {
		t.Error("IP.Valid = true, want false")
	}
}

// --- SQLite Database.MapRecordChangeEventParams tests ---

func TestDatabase_MapRecordChangeEventParams_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	eventID := types.NewEventID()
	nodeID := types.NewNodeID()
	hlc := types.HLCNow()
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	// Use nil-Data JSONData values to keep structs comparable
	oldVals := types.NewJSONData(nil)
	newVals := types.NewJSONData(nil)
	meta := types.NewJSONData(nil)
	reqID := types.NewNullableString("req-map-001")
	ip := types.NewNullableString("172.16.0.5")

	input := RecordChangeEventParams{
		EventID:      eventID,
		HlcTimestamp: hlc,
		NodeID:       nodeID,
		TableName:    "routes",
		RecordID:     "route-xyz",
		Operation:    types.OpInsert,
		Action:       types.ActionCreate,
		UserID:       userID,
		OldValues:    oldVals,
		NewValues:    newVals,
		Metadata:     meta,
		RequestID:    reqID,
		IP:           ip,
	}

	got := d.MapRecordChangeEventParams(input)

	if got.EventID != eventID {
		t.Errorf("EventID = %v, want %v", got.EventID, eventID)
	}
	if got.HlcTimestamp != hlc {
		t.Errorf("HlcTimestamp = %v, want %v", got.HlcTimestamp, hlc)
	}
	if got.NodeID != nodeID {
		t.Errorf("NodeID = %v, want %v", got.NodeID, nodeID)
	}
	if got.TableName != "routes" {
		t.Errorf("TableName = %q, want %q", got.TableName, "routes")
	}
	if got.RecordID != "route-xyz" {
		t.Errorf("RecordID = %q, want %q", got.RecordID, "route-xyz")
	}
	if got.Operation != types.OpInsert {
		t.Errorf("Operation = %v, want %v", got.Operation, types.OpInsert)
	}
	if got.Action != types.ActionCreate {
		t.Errorf("Action = %v, want %v", got.Action, types.ActionCreate)
	}
	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	// JSONData fields: verify Valid flag pass-through
	if got.OldValues.Valid != oldVals.Valid {
		t.Errorf("OldValues.Valid = %v, want %v", got.OldValues.Valid, oldVals.Valid)
	}
	if got.NewValues.Valid != newVals.Valid {
		t.Errorf("NewValues.Valid = %v, want %v", got.NewValues.Valid, newVals.Valid)
	}
	if got.Metadata.Valid != meta.Valid {
		t.Errorf("Metadata.Valid = %v, want %v", got.Metadata.Valid, meta.Valid)
	}
	// Critical: wrapper RequestID -> sqlc RequestId
	if got.RequestId != reqID {
		t.Errorf("RequestId = %v, want %v", got.RequestId, reqID)
	}
	// Critical: wrapper IP -> sqlc Ip
	if got.Ip != ip {
		t.Errorf("Ip = %v, want %v", got.Ip, ip)
	}
}

func TestDatabase_MapRecordChangeEventParams_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapRecordChangeEventParams(RecordChangeEventParams{})

	if got.EventID != "" {
		t.Errorf("EventID = %v, want zero value", got.EventID)
	}
	if got.TableName != "" {
		t.Errorf("TableName = %q, want empty string", got.TableName)
	}
	if got.RequestId.Valid {
		t.Error("RequestId.Valid = true, want false")
	}
	if got.Ip.Valid {
		t.Error("Ip.Valid = true, want false")
	}
}

func TestDatabase_MapRecordChangeEventParams_NullOptionalFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := RecordChangeEventParams{
		UserID:    types.NullableUserID{Valid: false},
		RequestID: types.NullableString{Valid: false},
		IP:        types.NullableString{Valid: false},
	}

	got := d.MapRecordChangeEventParams(input)

	if got.UserID.Valid {
		t.Error("UserID.Valid = true, want false")
	}
	if got.RequestId.Valid {
		t.Error("RequestId.Valid = true, want false")
	}
	if got.Ip.Valid {
		t.Error("Ip.Valid = true, want false")
	}
}

// --- MySQL MysqlDatabase.MapChangeEvent tests ---

func TestMysqlDatabase_MapChangeEvent_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	eventID := types.NewEventID()
	nodeID := types.NewNodeID()
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 1, 15, 8, 0, 0, 0, time.UTC))
	hlc := types.HLCNow()
	reqID := types.NewNullableString("req-mysql-001")
	ip := types.NewNullableString("10.10.10.10")

	input := mdbm.ChangeEvent{
		EventID:       eventID,
		HlcTimestamp:  hlc,
		WallTimestamp: ts,
		NodeID:        nodeID,
		TableName:     "media",
		RecordID:      "media-123",
		Operation:     types.OpDelete,
		Action:        types.ActionDelete,
		UserID:        userID,
		OldValues:     types.NewJSONData(nil),
		NewValues:     types.NewJSONData(nil),
		Metadata:      types.NewJSONData(nil),
		RequestId:     reqID,
		Ip:            ip,
		SyncedAt:      ts,
		ConsumedAt:    ts,
	}

	got := d.MapChangeEvent(input)

	if got.EventID != eventID {
		t.Errorf("EventID = %v, want %v", got.EventID, eventID)
	}
	if got.HlcTimestamp != hlc {
		t.Errorf("HlcTimestamp = %v, want %v", got.HlcTimestamp, hlc)
	}
	if got.WallTimestamp != ts {
		t.Errorf("WallTimestamp = %v, want %v", got.WallTimestamp, ts)
	}
	if got.NodeID != nodeID {
		t.Errorf("NodeID = %v, want %v", got.NodeID, nodeID)
	}
	if got.TableName != "media" {
		t.Errorf("TableName = %q, want %q", got.TableName, "media")
	}
	if got.RecordID != "media-123" {
		t.Errorf("RecordID = %q, want %q", got.RecordID, "media-123")
	}
	if got.Operation != types.OpDelete {
		t.Errorf("Operation = %v, want %v", got.Operation, types.OpDelete)
	}
	if got.Action != types.ActionDelete {
		t.Errorf("Action = %v, want %v", got.Action, types.ActionDelete)
	}
	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	// Field name mapping
	if got.RequestID != reqID {
		t.Errorf("RequestID = %v, want %v", got.RequestID, reqID)
	}
	if got.IP != ip {
		t.Errorf("IP = %v, want %v", got.IP, ip)
	}
	if got.SyncedAt != ts {
		t.Errorf("SyncedAt = %v, want %v", got.SyncedAt, ts)
	}
	if got.ConsumedAt != ts {
		t.Errorf("ConsumedAt = %v, want %v", got.ConsumedAt, ts)
	}
}

func TestMysqlDatabase_MapChangeEvent_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapChangeEvent(mdbm.ChangeEvent{})

	if got.EventID != "" {
		t.Errorf("EventID = %v, want zero value", got.EventID)
	}
	if got.NodeID != "" {
		t.Errorf("NodeID = %v, want zero value", got.NodeID)
	}
	if got.RequestID.Valid {
		t.Error("RequestID.Valid = true, want false")
	}
	if got.IP.Valid {
		t.Error("IP.Valid = true, want false")
	}
}

// --- MySQL MysqlDatabase.MapRecordChangeEventParams tests ---

func TestMysqlDatabase_MapRecordChangeEventParams_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	eventID := types.NewEventID()
	nodeID := types.NewNodeID()
	hlc := types.HLCNow()
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	reqID := types.NewNullableString("req-mysql-map")
	ip := types.NewNullableString("10.20.30.40")

	input := RecordChangeEventParams{
		EventID:      eventID,
		HlcTimestamp: hlc,
		NodeID:       nodeID,
		TableName:    "sessions",
		RecordID:     "sess-001",
		Operation:    types.OpUpdate,
		Action:       types.ActionUpdate,
		UserID:       userID,
		OldValues:    types.NewJSONData(nil),
		NewValues:    types.NewJSONData(nil),
		Metadata:     types.NewJSONData(nil),
		RequestID:    reqID,
		IP:           ip,
	}

	got := d.MapRecordChangeEventParams(input)

	if got.EventID != eventID {
		t.Errorf("EventID = %v, want %v", got.EventID, eventID)
	}
	if got.HlcTimestamp != hlc {
		t.Errorf("HlcTimestamp = %v, want %v", got.HlcTimestamp, hlc)
	}
	if got.NodeID != nodeID {
		t.Errorf("NodeID = %v, want %v", got.NodeID, nodeID)
	}
	if got.TableName != "sessions" {
		t.Errorf("TableName = %q, want %q", got.TableName, "sessions")
	}
	if got.RecordID != "sess-001" {
		t.Errorf("RecordID = %q, want %q", got.RecordID, "sess-001")
	}
	if got.Operation != types.OpUpdate {
		t.Errorf("Operation = %v, want %v", got.Operation, types.OpUpdate)
	}
	if got.Action != types.ActionUpdate {
		t.Errorf("Action = %v, want %v", got.Action, types.ActionUpdate)
	}
	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	// Critical field name mapping: wrapper RequestID -> sqlc RequestId
	if got.RequestId != reqID {
		t.Errorf("RequestId = %v, want %v", got.RequestId, reqID)
	}
	// Critical field name mapping: wrapper IP -> sqlc Ip
	if got.Ip != ip {
		t.Errorf("Ip = %v, want %v", got.Ip, ip)
	}
}

func TestMysqlDatabase_MapRecordChangeEventParams_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapRecordChangeEventParams(RecordChangeEventParams{})

	if got.EventID != "" {
		t.Errorf("EventID = %v, want zero value", got.EventID)
	}
	if got.RequestId.Valid {
		t.Error("RequestId.Valid = true, want false")
	}
	if got.Ip.Valid {
		t.Error("Ip.Valid = true, want false")
	}
}

// --- PostgreSQL PsqlDatabase.MapChangeEvent tests ---

func TestPsqlDatabase_MapChangeEvent_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	eventID := types.NewEventID()
	nodeID := types.NewNodeID()
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 2, 20, 14, 0, 0, 0, time.UTC))
	hlc := types.HLCNow()
	reqID := types.NewNullableString("req-psql-001")
	ip := types.NewNullableString("172.16.0.1")

	input := mdbp.ChangeEvent{
		EventID:       eventID,
		HlcTimestamp:  hlc,
		WallTimestamp: ts,
		NodeID:        nodeID,
		TableName:     "datatypes",
		RecordID:      "dt-456",
		Operation:     types.OpInsert,
		Action:        types.ActionPublish,
		UserID:        userID,
		OldValues:     types.NewJSONData(nil),
		NewValues:     types.NewJSONData(nil),
		Metadata:      types.NewJSONData(nil),
		RequestId:     reqID,
		Ip:            ip,
		SyncedAt:      ts,
		ConsumedAt:    ts,
	}

	got := d.MapChangeEvent(input)

	if got.EventID != eventID {
		t.Errorf("EventID = %v, want %v", got.EventID, eventID)
	}
	if got.HlcTimestamp != hlc {
		t.Errorf("HlcTimestamp = %v, want %v", got.HlcTimestamp, hlc)
	}
	if got.WallTimestamp != ts {
		t.Errorf("WallTimestamp = %v, want %v", got.WallTimestamp, ts)
	}
	if got.NodeID != nodeID {
		t.Errorf("NodeID = %v, want %v", got.NodeID, nodeID)
	}
	if got.TableName != "datatypes" {
		t.Errorf("TableName = %q, want %q", got.TableName, "datatypes")
	}
	if got.RecordID != "dt-456" {
		t.Errorf("RecordID = %q, want %q", got.RecordID, "dt-456")
	}
	if got.Operation != types.OpInsert {
		t.Errorf("Operation = %v, want %v", got.Operation, types.OpInsert)
	}
	if got.Action != types.ActionPublish {
		t.Errorf("Action = %v, want %v", got.Action, types.ActionPublish)
	}
	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	// Field name mapping
	if got.RequestID != reqID {
		t.Errorf("RequestID = %v, want %v", got.RequestID, reqID)
	}
	if got.IP != ip {
		t.Errorf("IP = %v, want %v", got.IP, ip)
	}
	if got.SyncedAt != ts {
		t.Errorf("SyncedAt = %v, want %v", got.SyncedAt, ts)
	}
	if got.ConsumedAt != ts {
		t.Errorf("ConsumedAt = %v, want %v", got.ConsumedAt, ts)
	}
}

func TestPsqlDatabase_MapChangeEvent_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapChangeEvent(mdbp.ChangeEvent{})

	if got.EventID != "" {
		t.Errorf("EventID = %v, want zero value", got.EventID)
	}
	if got.NodeID != "" {
		t.Errorf("NodeID = %v, want zero value", got.NodeID)
	}
	if got.RequestID.Valid {
		t.Error("RequestID.Valid = true, want false")
	}
	if got.IP.Valid {
		t.Error("IP.Valid = true, want false")
	}
}

// --- PostgreSQL PsqlDatabase.MapRecordChangeEventParams tests ---

func TestPsqlDatabase_MapRecordChangeEventParams_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	eventID := types.NewEventID()
	nodeID := types.NewNodeID()
	hlc := types.HLCNow()
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	reqID := types.NewNullableString("req-psql-map")
	ip := types.NewNullableString("10.0.0.99")

	input := RecordChangeEventParams{
		EventID:      eventID,
		HlcTimestamp: hlc,
		NodeID:       nodeID,
		TableName:    "permissions",
		RecordID:     "perm-007",
		Operation:    types.OpDelete,
		Action:       types.ActionArchive,
		UserID:       userID,
		OldValues:    types.NewJSONData(nil),
		NewValues:    types.NewJSONData(nil),
		Metadata:     types.NewJSONData(nil),
		RequestID:    reqID,
		IP:           ip,
	}

	got := d.MapRecordChangeEventParams(input)

	if got.EventID != eventID {
		t.Errorf("EventID = %v, want %v", got.EventID, eventID)
	}
	if got.HlcTimestamp != hlc {
		t.Errorf("HlcTimestamp = %v, want %v", got.HlcTimestamp, hlc)
	}
	if got.NodeID != nodeID {
		t.Errorf("NodeID = %v, want %v", got.NodeID, nodeID)
	}
	if got.TableName != "permissions" {
		t.Errorf("TableName = %q, want %q", got.TableName, "permissions")
	}
	if got.RecordID != "perm-007" {
		t.Errorf("RecordID = %q, want %q", got.RecordID, "perm-007")
	}
	if got.Operation != types.OpDelete {
		t.Errorf("Operation = %v, want %v", got.Operation, types.OpDelete)
	}
	if got.Action != types.ActionArchive {
		t.Errorf("Action = %v, want %v", got.Action, types.ActionArchive)
	}
	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	// Critical field name mapping
	if got.RequestId != reqID {
		t.Errorf("RequestId = %v, want %v", got.RequestId, reqID)
	}
	if got.Ip != ip {
		t.Errorf("Ip = %v, want %v", got.Ip, ip)
	}
}

func TestPsqlDatabase_MapRecordChangeEventParams_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapRecordChangeEventParams(RecordChangeEventParams{})

	if got.EventID != "" {
		t.Errorf("EventID = %v, want zero value", got.EventID)
	}
	if got.RequestId.Valid {
		t.Error("RequestId.Valid = true, want false")
	}
	if got.Ip.Valid {
		t.Error("Ip.Valid = true, want false")
	}
}

// --- Cross-database MapChangeEvent consistency ---
// All three database drivers use identical field types for ChangeEvent
// (no int32/int64 conversions). This verifies they all produce identical
// wrapper ChangeEvent structs from equivalent input.

func TestCrossDatabaseMapChangeEvent_Consistency(t *testing.T) {
	t.Parallel()
	eventID := types.NewEventID()
	nodeID := types.NewNodeID()
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	ts := types.NewTimestamp(time.Date(2025, 4, 10, 9, 15, 0, 0, time.UTC))
	hlc := types.HLCNow()
	// Use JSONData without map values to allow field-level comparison
	// of the comparable fields. JSONData contains `any`, so ChangeEvent
	// is not directly comparable with ==.
	reqID := types.NewNullableString("req-cross")
	ip := types.NewNullableString("10.0.0.1")

	sqliteInput := mdb.ChangeEvent{
		EventID: eventID, HlcTimestamp: hlc, WallTimestamp: ts,
		NodeID: nodeID, TableName: "cross_table", RecordID: "cross-001",
		Operation: types.OpInsert, Action: types.ActionCreate,
		UserID: userID, RequestId: reqID, Ip: ip,
		SyncedAt: ts, ConsumedAt: ts,
	}
	mysqlInput := mdbm.ChangeEvent{
		EventID: eventID, HlcTimestamp: hlc, WallTimestamp: ts,
		NodeID: nodeID, TableName: "cross_table", RecordID: "cross-001",
		Operation: types.OpInsert, Action: types.ActionCreate,
		UserID: userID, RequestId: reqID, Ip: ip,
		SyncedAt: ts, ConsumedAt: ts,
	}
	psqlInput := mdbp.ChangeEvent{
		EventID: eventID, HlcTimestamp: hlc, WallTimestamp: ts,
		NodeID: nodeID, TableName: "cross_table", RecordID: "cross-001",
		Operation: types.OpInsert, Action: types.ActionCreate,
		UserID: userID, RequestId: reqID, Ip: ip,
		SyncedAt: ts, ConsumedAt: ts,
	}

	sqliteResult := Database{}.MapChangeEvent(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapChangeEvent(mysqlInput)
	psqlResult := PsqlDatabase{}.MapChangeEvent(psqlInput)

	// Compare all comparable fields individually (JSONData contains `any`
	// which makes the struct non-comparable with ==)
	compareChangeEvents := func(t *testing.T, label string, a, b ChangeEvent) {
		t.Helper()
		if a.EventID != b.EventID {
			t.Errorf("%s EventID: %v vs %v", label, a.EventID, b.EventID)
		}
		if a.HlcTimestamp != b.HlcTimestamp {
			t.Errorf("%s HlcTimestamp: %v vs %v", label, a.HlcTimestamp, b.HlcTimestamp)
		}
		if a.WallTimestamp != b.WallTimestamp {
			t.Errorf("%s WallTimestamp: %v vs %v", label, a.WallTimestamp, b.WallTimestamp)
		}
		if a.NodeID != b.NodeID {
			t.Errorf("%s NodeID: %v vs %v", label, a.NodeID, b.NodeID)
		}
		if a.TableName != b.TableName {
			t.Errorf("%s TableName: %q vs %q", label, a.TableName, b.TableName)
		}
		if a.RecordID != b.RecordID {
			t.Errorf("%s RecordID: %q vs %q", label, a.RecordID, b.RecordID)
		}
		if a.Operation != b.Operation {
			t.Errorf("%s Operation: %v vs %v", label, a.Operation, b.Operation)
		}
		if a.Action != b.Action {
			t.Errorf("%s Action: %v vs %v", label, a.Action, b.Action)
		}
		if a.UserID != b.UserID {
			t.Errorf("%s UserID: %v vs %v", label, a.UserID, b.UserID)
		}
		if a.RequestID != b.RequestID {
			t.Errorf("%s RequestID: %v vs %v", label, a.RequestID, b.RequestID)
		}
		if a.IP != b.IP {
			t.Errorf("%s IP: %v vs %v", label, a.IP, b.IP)
		}
		if a.SyncedAt != b.SyncedAt {
			t.Errorf("%s SyncedAt: %v vs %v", label, a.SyncedAt, b.SyncedAt)
		}
		if a.ConsumedAt != b.ConsumedAt {
			t.Errorf("%s ConsumedAt: %v vs %v", label, a.ConsumedAt, b.ConsumedAt)
		}
	}

	compareChangeEvents(t, "SQLite vs MySQL", sqliteResult, mysqlResult)
	compareChangeEvents(t, "SQLite vs PostgreSQL", sqliteResult, psqlResult)
}

// --- Cross-database MapRecordChangeEventParams field name mapping ---
// This is the critical mapping to verify: the wrapper uses RequestID/IP
// while sqlc generates RequestId/Ip. All three param mappers must correctly
// bridge this naming difference.

func TestCrossDatabaseRequestIDFieldMapping(t *testing.T) {
	t.Parallel()
	reqID := types.NewNullableString("cross-req-id-001")
	ip := types.NewNullableString("192.168.0.1")

	input := RecordChangeEventParams{
		RequestID: reqID,
		IP:        ip,
	}

	t.Run("SQLite", func(t *testing.T) {
		t.Parallel()
		got := Database{}.MapRecordChangeEventParams(input)
		if got.RequestId != reqID {
			t.Errorf("RequestId = %v, want %v", got.RequestId, reqID)
		}
		if got.Ip != ip {
			t.Errorf("Ip = %v, want %v", got.Ip, ip)
		}
	})

	t.Run("MySQL", func(t *testing.T) {
		t.Parallel()
		got := MysqlDatabase{}.MapRecordChangeEventParams(input)
		if got.RequestId != reqID {
			t.Errorf("RequestId = %v, want %v", got.RequestId, reqID)
		}
		if got.Ip != ip {
			t.Errorf("Ip = %v, want %v", got.Ip, ip)
		}
	})

	t.Run("PostgreSQL", func(t *testing.T) {
		t.Parallel()
		got := PsqlDatabase{}.MapRecordChangeEventParams(input)
		if got.RequestId != reqID {
			t.Errorf("RequestId = %v, want %v", got.RequestId, reqID)
		}
		if got.Ip != ip {
			t.Errorf("Ip = %v, want %v", got.Ip, ip)
		}
	})
}

func TestCrossDatabaseRequestIDReverseMapping(t *testing.T) {
	t.Parallel()
	// Verify the reverse direction: sqlc RequestId/Ip -> wrapper RequestID/IP
	reqID := types.NewNullableString("reverse-req-001")
	ip := types.NewNullableString("10.0.0.42")

	t.Run("SQLite", func(t *testing.T) {
		t.Parallel()
		got := Database{}.MapChangeEvent(mdb.ChangeEvent{RequestId: reqID, Ip: ip})
		if got.RequestID != reqID {
			t.Errorf("RequestID = %v, want %v", got.RequestID, reqID)
		}
		if got.IP != ip {
			t.Errorf("IP = %v, want %v", got.IP, ip)
		}
	})

	t.Run("MySQL", func(t *testing.T) {
		t.Parallel()
		got := MysqlDatabase{}.MapChangeEvent(mdbm.ChangeEvent{RequestId: reqID, Ip: ip})
		if got.RequestID != reqID {
			t.Errorf("RequestID = %v, want %v", got.RequestID, reqID)
		}
		if got.IP != ip {
			t.Errorf("IP = %v, want %v", got.IP, ip)
		}
	})

	t.Run("PostgreSQL", func(t *testing.T) {
		t.Parallel()
		got := PsqlDatabase{}.MapChangeEvent(mdbp.ChangeEvent{RequestId: reqID, Ip: ip})
		if got.RequestID != reqID {
			t.Errorf("RequestID = %v, want %v", got.RequestID, reqID)
		}
		if got.IP != ip {
			t.Errorf("IP = %v, want %v", got.IP, ip)
		}
	})
}

// --- MapChangeEvent with all Operation/Action variants ---
// Ensures the mapper faithfully passes through every enum value without
// accidental transformation.

func TestDatabase_MapChangeEvent_AllOperations(t *testing.T) {
	t.Parallel()
	d := Database{}

	operations := []types.Operation{
		types.OpInsert,
		types.OpUpdate,
		types.OpDelete,
	}

	for _, op := range operations {
		t.Run(string(op), func(t *testing.T) {
			t.Parallel()
			input := mdb.ChangeEvent{Operation: op}
			got := d.MapChangeEvent(input)
			if got.Operation != op {
				t.Errorf("Operation = %v, want %v", got.Operation, op)
			}
		})
	}
}

func TestDatabase_MapChangeEvent_AllActions(t *testing.T) {
	t.Parallel()
	d := Database{}

	actions := []types.Action{
		types.ActionCreate,
		types.ActionUpdate,
		types.ActionDelete,
		types.ActionPublish,
		types.ActionArchive,
	}

	for _, action := range actions {
		t.Run(string(action), func(t *testing.T) {
			t.Parallel()
			input := mdb.ChangeEvent{Action: action}
			got := d.MapChangeEvent(input)
			if got.Action != action {
				t.Errorf("Action = %v, want %v", got.Action, action)
			}
		})
	}
}

// --- MapChangeEvent with null vs valid nullable fields ---
// Tests the boundary between Valid=true and Valid=false for all nullable fields.

func TestDatabase_MapChangeEvent_NullableFieldVariants(t *testing.T) {
	t.Parallel()
	d := Database{}

	tests := []struct {
		name      string
		input     mdb.ChangeEvent
		wantUser  bool
		wantReq   bool
		wantIP    bool
		wantReqS  string
		wantIPS   string
	}{
		{
			name: "all nullable fields valid",
			input: mdb.ChangeEvent{
				UserID:    types.NullableUserID{ID: types.NewUserID(), Valid: true},
				RequestId: types.NewNullableString("req-valid"),
				Ip:        types.NewNullableString("1.2.3.4"),
			},
			wantUser: true,
			wantReq:  true,
			wantIP:   true,
			wantReqS: "req-valid",
			wantIPS:  "1.2.3.4",
		},
		{
			name: "all nullable fields null",
			input: mdb.ChangeEvent{
				UserID:    types.NullableUserID{Valid: false},
				RequestId: types.NullableString{Valid: false},
				Ip:        types.NullableString{Valid: false},
			},
			wantUser: false,
			wantReq:  false,
			wantIP:   false,
		},
		{
			name: "mixed - user valid, request null, ip valid",
			input: mdb.ChangeEvent{
				UserID:    types.NullableUserID{ID: types.NewUserID(), Valid: true},
				RequestId: types.NullableString{Valid: false},
				Ip:        types.NewNullableString("5.6.7.8"),
			},
			wantUser: true,
			wantReq:  false,
			wantIP:   true,
			wantIPS:  "5.6.7.8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := d.MapChangeEvent(tt.input)

			if got.UserID.Valid != tt.wantUser {
				t.Errorf("UserID.Valid = %v, want %v", got.UserID.Valid, tt.wantUser)
			}
			if got.RequestID.Valid != tt.wantReq {
				t.Errorf("RequestID.Valid = %v, want %v", got.RequestID.Valid, tt.wantReq)
			}
			if got.IP.Valid != tt.wantIP {
				t.Errorf("IP.Valid = %v, want %v", got.IP.Valid, tt.wantIP)
			}
			if tt.wantReq && got.RequestID.String != tt.wantReqS {
				t.Errorf("RequestID.String = %q, want %q", got.RequestID.String, tt.wantReqS)
			}
			if tt.wantIP && got.IP.String != tt.wantIPS {
				t.Errorf("IP.String = %q, want %q", got.IP.String, tt.wantIPS)
			}
		})
	}
}

// --- Parameter struct usability tests ---
// These verify the declared param types can be instantiated with all fields
// and that zero-value construction does not panic.

func TestChangeEvent_ZeroValue(t *testing.T) {
	t.Parallel()
	var ce ChangeEvent

	if ce.EventID != "" {
		t.Errorf("EventID = %v, want zero value", ce.EventID)
	}
	if ce.NodeID != "" {
		t.Errorf("NodeID = %v, want zero value", ce.NodeID)
	}
	if ce.TableName != "" {
		t.Errorf("TableName = %q, want empty string", ce.TableName)
	}
	if ce.RecordID != "" {
		t.Errorf("RecordID = %q, want empty string", ce.RecordID)
	}
	if ce.Operation != "" {
		t.Errorf("Operation = %v, want zero value", ce.Operation)
	}
	if ce.Action != "" {
		t.Errorf("Action = %v, want zero value", ce.Action)
	}
	if ce.UserID.Valid {
		t.Error("UserID.Valid = true, want false")
	}
	if ce.OldValues.Valid {
		t.Error("OldValues.Valid = true, want false")
	}
	if ce.NewValues.Valid {
		t.Error("NewValues.Valid = true, want false")
	}
	if ce.Metadata.Valid {
		t.Error("Metadata.Valid = true, want false")
	}
	if ce.RequestID.Valid {
		t.Error("RequestID.Valid = true, want false")
	}
	if ce.IP.Valid {
		t.Error("IP.Valid = true, want false")
	}
}

func TestRecordChangeEventParams_ZeroValue(t *testing.T) {
	t.Parallel()
	var p RecordChangeEventParams

	if p.EventID != "" {
		t.Errorf("EventID = %v, want zero value", p.EventID)
	}
	if p.NodeID != "" {
		t.Errorf("NodeID = %v, want zero value", p.NodeID)
	}
	if p.TableName != "" {
		t.Errorf("TableName = %q, want empty string", p.TableName)
	}
	if p.RecordID != "" {
		t.Errorf("RecordID = %q, want empty string", p.RecordID)
	}
	if p.Operation != "" {
		t.Errorf("Operation = %v, want zero value", p.Operation)
	}
	if p.Action != "" {
		t.Errorf("Action = %v, want zero value", p.Action)
	}
	if p.UserID.Valid {
		t.Error("UserID.Valid = true, want false")
	}
	if p.RequestID.Valid {
		t.Error("RequestID.Valid = true, want false")
	}
	if p.IP.Valid {
		t.Error("IP.Valid = true, want false")
	}
}

func TestRecordChangeEventParams_FullyPopulated(t *testing.T) {
	t.Parallel()
	eventID := types.NewEventID()
	nodeID := types.NewNodeID()
	hlc := types.HLCNow()
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}

	p := RecordChangeEventParams{
		EventID:      eventID,
		HlcTimestamp: hlc,
		NodeID:       nodeID,
		TableName:    "tokens",
		RecordID:     "tok-abc",
		Operation:    types.OpInsert,
		Action:       types.ActionCreate,
		UserID:       userID,
		OldValues:    types.NewJSONData(nil),
		NewValues:    types.NewJSONData(map[string]string{"token": "***"}),
		Metadata:     types.NewJSONData(map[string]string{"client": "web"}),
		RequestID:    types.NewNullableString("req-full"),
		IP:           types.NewNullableString("127.0.0.1"),
	}

	if p.EventID != eventID {
		t.Errorf("EventID = %v, want %v", p.EventID, eventID)
	}
	if p.HlcTimestamp != hlc {
		t.Errorf("HlcTimestamp = %v, want %v", p.HlcTimestamp, hlc)
	}
	if p.NodeID != nodeID {
		t.Errorf("NodeID = %v, want %v", p.NodeID, nodeID)
	}
	if p.TableName != "tokens" {
		t.Errorf("TableName = %q, want %q", p.TableName, "tokens")
	}
	if p.RecordID != "tok-abc" {
		t.Errorf("RecordID = %q, want %q", p.RecordID, "tok-abc")
	}
	if p.Operation != types.OpInsert {
		t.Errorf("Operation = %v, want %v", p.Operation, types.OpInsert)
	}
	if p.Action != types.ActionCreate {
		t.Errorf("Action = %v, want %v", p.Action, types.ActionCreate)
	}
	if !p.UserID.Valid {
		t.Error("UserID.Valid = false, want true")
	}
	if p.OldValues.Data != nil {
		t.Errorf("OldValues.Raw = %v, want nil", p.OldValues.Data)
	}
	if p.NewValues.Data == nil {
		t.Error("NewValues.Raw = nil, want non-nil")
	}
	if p.Metadata.Data == nil {
		t.Error("Metadata.Raw = nil, want non-nil")
	}
	if !p.RequestID.Valid {
		t.Error("RequestID.Valid = false, want true")
	}
	if !p.IP.Valid {
		t.Error("IP.Valid = false, want true")
	}
}

func TestListChangeEventsParams_ZeroValue(t *testing.T) {
	t.Parallel()
	var p ListChangeEventsParams

	if p.Limit != 0 {
		t.Errorf("Limit = %d, want 0", p.Limit)
	}
	if p.Offset != 0 {
		t.Errorf("Offset = %d, want 0", p.Offset)
	}
}

func TestListChangeEventsParams_Populated(t *testing.T) {
	t.Parallel()
	p := ListChangeEventsParams{Limit: 50, Offset: 100}

	if p.Limit != 50 {
		t.Errorf("Limit = %d, want 50", p.Limit)
	}
	if p.Offset != 100 {
		t.Errorf("Offset = %d, want 100", p.Offset)
	}
}

func TestListChangeEventsByUserParams_ZeroValue(t *testing.T) {
	t.Parallel()
	var p ListChangeEventsByUserParams

	if p.UserID.Valid {
		t.Error("UserID.Valid = true, want false")
	}
	if p.Limit != 0 {
		t.Errorf("Limit = %d, want 0", p.Limit)
	}
	if p.Offset != 0 {
		t.Errorf("Offset = %d, want 0", p.Offset)
	}
}

func TestListChangeEventsByUserParams_Populated(t *testing.T) {
	t.Parallel()
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	p := ListChangeEventsByUserParams{
		UserID: userID,
		Limit:  25,
		Offset: 50,
	}

	if p.UserID != userID {
		t.Errorf("UserID = %v, want %v", p.UserID, userID)
	}
	if p.Limit != 25 {
		t.Errorf("Limit = %d, want 25", p.Limit)
	}
	if p.Offset != 50 {
		t.Errorf("Offset = %d, want 50", p.Offset)
	}
}

func TestListChangeEventsByActionParams_ZeroValue(t *testing.T) {
	t.Parallel()
	var p ListChangeEventsByActionParams

	if p.Action != "" {
		t.Errorf("Action = %v, want zero value", p.Action)
	}
	if p.Limit != 0 {
		t.Errorf("Limit = %d, want 0", p.Limit)
	}
	if p.Offset != 0 {
		t.Errorf("Offset = %d, want 0", p.Offset)
	}
}

func TestListChangeEventsByActionParams_AllActions(t *testing.T) {
	t.Parallel()
	actions := []types.Action{
		types.ActionCreate,
		types.ActionUpdate,
		types.ActionDelete,
		types.ActionPublish,
		types.ActionArchive,
	}

	for _, action := range actions {
		t.Run(string(action), func(t *testing.T) {
			t.Parallel()
			p := ListChangeEventsByActionParams{
				Action: action,
				Limit:  10,
				Offset: 0,
			}
			if p.Action != action {
				t.Errorf("Action = %v, want %v", p.Action, action)
			}
			if p.Limit != 10 {
				t.Errorf("Limit = %d, want 10", p.Limit)
			}
			if p.Offset != 0 {
				t.Errorf("Offset = %d, want 0", p.Offset)
			}
		})
	}
}

// --- Round-trip mapping tests ---
// Verify that mapping params -> sqlc -> result -> wrapper produces consistent
// field values for the name-bridged fields (RequestID/IP).

func TestDatabase_MapParamsAndResult_RoundTrip(t *testing.T) {
	t.Parallel()
	d := Database{}
	reqID := types.NewNullableString("round-trip-req")
	ip := types.NewNullableString("192.168.1.1")
	eventID := types.NewEventID()
	nodeID := types.NewNodeID()
	hlc := types.HLCNow()
	ts := types.NewTimestamp(time.Date(2025, 5, 1, 12, 0, 0, 0, time.UTC))

	// Map wrapper params to sqlc params
	inputParams := RecordChangeEventParams{
		EventID:      eventID,
		HlcTimestamp: hlc,
		NodeID:       nodeID,
		TableName:    "round_trip_table",
		RecordID:     "rt-001",
		Operation:    types.OpInsert,
		Action:       types.ActionCreate,
		UserID:       types.NullableUserID{Valid: false},
		OldValues:    types.JSONData{Valid: false},
		NewValues:    types.NewJSONData(map[string]string{"x": "y"}),
		Metadata:     types.JSONData{Valid: false},
		RequestID:    reqID,
		IP:           ip,
	}
	sqlcParams := d.MapRecordChangeEventParams(inputParams)

	// Simulate what a database would return by building a ChangeEvent from the sqlc params
	sqlcResult := mdb.ChangeEvent{
		EventID:       sqlcParams.EventID,
		HlcTimestamp:  sqlcParams.HlcTimestamp,
		WallTimestamp: ts,
		NodeID:        sqlcParams.NodeID,
		TableName:     sqlcParams.TableName,
		RecordID:      sqlcParams.RecordID,
		Operation:     sqlcParams.Operation,
		Action:        sqlcParams.Action,
		UserID:        sqlcParams.UserID,
		OldValues:     sqlcParams.OldValues,
		NewValues:     sqlcParams.NewValues,
		Metadata:      sqlcParams.Metadata,
		RequestId:     sqlcParams.RequestId,
		Ip:            sqlcParams.Ip,
		SyncedAt:      ts,
		ConsumedAt:    ts,
	}

	// Map sqlc result back to wrapper
	wrapperResult := d.MapChangeEvent(sqlcResult)

	// The RequestID and IP should survive the round-trip
	if wrapperResult.RequestID != reqID {
		t.Errorf("RequestID round-trip: got %v, want %v", wrapperResult.RequestID, reqID)
	}
	if wrapperResult.IP != ip {
		t.Errorf("IP round-trip: got %v, want %v", wrapperResult.IP, ip)
	}
	if wrapperResult.EventID != eventID {
		t.Errorf("EventID round-trip: got %v, want %v", wrapperResult.EventID, eventID)
	}
	if wrapperResult.NodeID != nodeID {
		t.Errorf("NodeID round-trip: got %v, want %v", wrapperResult.NodeID, nodeID)
	}
	if wrapperResult.TableName != "round_trip_table" {
		t.Errorf("TableName round-trip: got %q, want %q", wrapperResult.TableName, "round_trip_table")
	}
}

func TestMysqlDatabase_MapParamsAndResult_RoundTrip(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	reqID := types.NewNullableString("mysql-round-trip")
	ip := types.NewNullableString("10.20.30.40")
	eventID := types.NewEventID()
	nodeID := types.NewNodeID()
	hlc := types.HLCNow()
	ts := types.NewTimestamp(time.Date(2025, 5, 1, 12, 0, 0, 0, time.UTC))

	inputParams := RecordChangeEventParams{
		EventID:      eventID,
		HlcTimestamp: hlc,
		NodeID:       nodeID,
		TableName:    "mysql_table",
		RecordID:     "mysql-rt-001",
		Operation:    types.OpUpdate,
		Action:       types.ActionUpdate,
		RequestID:    reqID,
		IP:           ip,
	}
	sqlcParams := d.MapRecordChangeEventParams(inputParams)

	sqlcResult := mdbm.ChangeEvent{
		EventID:       sqlcParams.EventID,
		HlcTimestamp:  sqlcParams.HlcTimestamp,
		WallTimestamp: ts,
		NodeID:        sqlcParams.NodeID,
		TableName:     sqlcParams.TableName,
		RecordID:      sqlcParams.RecordID,
		Operation:     sqlcParams.Operation,
		Action:        sqlcParams.Action,
		UserID:        sqlcParams.UserID,
		OldValues:     sqlcParams.OldValues,
		NewValues:     sqlcParams.NewValues,
		Metadata:      sqlcParams.Metadata,
		RequestId:     sqlcParams.RequestId,
		Ip:            sqlcParams.Ip,
		SyncedAt:      ts,
		ConsumedAt:    ts,
	}

	wrapperResult := d.MapChangeEvent(sqlcResult)

	if wrapperResult.RequestID != reqID {
		t.Errorf("RequestID round-trip: got %v, want %v", wrapperResult.RequestID, reqID)
	}
	if wrapperResult.IP != ip {
		t.Errorf("IP round-trip: got %v, want %v", wrapperResult.IP, ip)
	}
}

func TestPsqlDatabase_MapParamsAndResult_RoundTrip(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	reqID := types.NewNullableString("psql-round-trip")
	ip := types.NewNullableString("172.16.0.99")
	eventID := types.NewEventID()
	nodeID := types.NewNodeID()
	hlc := types.HLCNow()
	ts := types.NewTimestamp(time.Date(2025, 5, 1, 12, 0, 0, 0, time.UTC))

	inputParams := RecordChangeEventParams{
		EventID:      eventID,
		HlcTimestamp: hlc,
		NodeID:       nodeID,
		TableName:    "psql_table",
		RecordID:     "psql-rt-001",
		Operation:    types.OpDelete,
		Action:       types.ActionDelete,
		RequestID:    reqID,
		IP:           ip,
	}
	sqlcParams := d.MapRecordChangeEventParams(inputParams)

	sqlcResult := mdbp.ChangeEvent{
		EventID:       sqlcParams.EventID,
		HlcTimestamp:  sqlcParams.HlcTimestamp,
		WallTimestamp: ts,
		NodeID:        sqlcParams.NodeID,
		TableName:     sqlcParams.TableName,
		RecordID:      sqlcParams.RecordID,
		Operation:     sqlcParams.Operation,
		Action:        sqlcParams.Action,
		UserID:        sqlcParams.UserID,
		OldValues:     sqlcParams.OldValues,
		NewValues:     sqlcParams.NewValues,
		Metadata:      sqlcParams.Metadata,
		RequestId:     sqlcParams.RequestId,
		Ip:            sqlcParams.Ip,
		SyncedAt:      ts,
		ConsumedAt:    ts,
	}

	wrapperResult := d.MapChangeEvent(sqlcResult)

	if wrapperResult.RequestID != reqID {
		t.Errorf("RequestID round-trip: got %v, want %v", wrapperResult.RequestID, reqID)
	}
	if wrapperResult.IP != ip {
		t.Errorf("IP round-trip: got %v, want %v", wrapperResult.IP, ip)
	}
}

// --- MapChangeEvent with NullableUserID edge cases ---

func TestDatabase_MapChangeEvent_NullUserID(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := mdb.ChangeEvent{
		UserID: types.NullableUserID{Valid: false},
	}
	got := d.MapChangeEvent(input)

	if got.UserID.Valid {
		t.Error("UserID.Valid = true, want false")
	}
}

func TestDatabase_MapChangeEvent_ValidUserID(t *testing.T) {
	t.Parallel()
	d := Database{}
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	input := mdb.ChangeEvent{UserID: userID}
	got := d.MapChangeEvent(input)

	if !got.UserID.Valid {
		t.Fatal("UserID.Valid = false, want true")
	}
	if got.UserID.ID != userID.ID {
		t.Errorf("UserID.ID = %v, want %v", got.UserID.ID, userID.ID)
	}
}

// --- Cross-database MapRecordChangeEventParams consistency ---
// All three drivers should produce identical sqlc params (same field values,
// same field types) when given the same wrapper input.

func TestCrossDatabaseMapRecordChangeEventParams_FieldConsistency(t *testing.T) {
	t.Parallel()
	eventID := types.NewEventID()
	nodeID := types.NewNodeID()
	hlc := types.HLCNow()
	userID := types.NullableUserID{ID: types.NewUserID(), Valid: true}
	oldVals := types.NewJSONData(map[string]string{"field": "old"})
	newVals := types.NewJSONData(map[string]string{"field": "new"})
	meta := types.NewJSONData(map[string]string{"key": "val"})
	reqID := types.NewNullableString("consistency-req")
	ip := types.NewNullableString("10.0.0.50")

	input := RecordChangeEventParams{
		EventID:      eventID,
		HlcTimestamp: hlc,
		NodeID:       nodeID,
		TableName:    "consistency_table",
		RecordID:     "con-001",
		Operation:    types.OpInsert,
		Action:       types.ActionCreate,
		UserID:       userID,
		OldValues:    oldVals,
		NewValues:    newVals,
		Metadata:     meta,
		RequestID:    reqID,
		IP:           ip,
	}

	sqliteGot := Database{}.MapRecordChangeEventParams(input)
	mysqlGot := MysqlDatabase{}.MapRecordChangeEventParams(input)
	psqlGot := PsqlDatabase{}.MapRecordChangeEventParams(input)

	// All three should have identical field values
	// (they all use the same types for RecordChangeEventParams fields)
	tests := []struct {
		name    string
		sqlite  string
		mysql   string
		psql    string
	}{
		{"EventID", string(sqliteGot.EventID), string(mysqlGot.EventID), string(psqlGot.EventID)},
		{"NodeID", string(sqliteGot.NodeID), string(mysqlGot.NodeID), string(psqlGot.NodeID)},
		{"TableName", sqliteGot.TableName, mysqlGot.TableName, psqlGot.TableName},
		{"RecordID", sqliteGot.RecordID, mysqlGot.RecordID, psqlGot.RecordID},
		{"Operation", string(sqliteGot.Operation), string(mysqlGot.Operation), string(psqlGot.Operation)},
		{"Action", string(sqliteGot.Action), string(mysqlGot.Action), string(psqlGot.Action)},
		{"RequestId", sqliteGot.RequestId.String, mysqlGot.RequestId.String, psqlGot.RequestId.String},
		{"Ip", sqliteGot.Ip.String, mysqlGot.Ip.String, psqlGot.Ip.String},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.sqlite != tt.mysql {
				t.Errorf("SQLite %s = %q, MySQL %s = %q", tt.name, tt.sqlite, tt.name, tt.mysql)
			}
			if tt.sqlite != tt.psql {
				t.Errorf("SQLite %s = %q, PostgreSQL %s = %q", tt.name, tt.sqlite, tt.name, tt.psql)
			}
		})
	}
}

// --- MapChangeEvent preserves HLC values correctly ---
// HLC is a uint64; ensure it survives mapping without truncation or mutation.

func TestDatabase_MapChangeEvent_HLCPreservation(t *testing.T) {
	t.Parallel()
	d := Database{}

	tests := []struct {
		name string
		hlc  types.HLC
	}{
		{"zero HLC", types.HLC(0)},
		{"current HLC", types.HLCNow()},
		{"max int64 HLC", types.HLC(1<<63 - 1)}, // max int64
		{"negative HLC", types.HLC(-1)},
		{"specific HLC", types.HLC(1706745600000 << 16)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := mdb.ChangeEvent{HlcTimestamp: tt.hlc}
			got := d.MapChangeEvent(input)
			if got.HlcTimestamp != tt.hlc {
				t.Errorf("HlcTimestamp = %v, want %v", got.HlcTimestamp, tt.hlc)
			}
		})
	}
}

// --- MapChangeEvent preserves timestamps correctly ---

func TestDatabase_MapChangeEvent_TimestampPreservation(t *testing.T) {
	t.Parallel()
	d := Database{}

	tests := []struct {
		name string
		ts   types.Timestamp
	}{
		{"epoch", types.NewTimestamp(time.Unix(0, 0).UTC())},
		{"recent", types.NewTimestamp(time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC))},
		{"now", types.TimestampNow()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := mdb.ChangeEvent{
				WallTimestamp: tt.ts,
				SyncedAt:      tt.ts,
				ConsumedAt:    tt.ts,
			}
			got := d.MapChangeEvent(input)
			if got.WallTimestamp != tt.ts {
				t.Errorf("WallTimestamp = %v, want %v", got.WallTimestamp, tt.ts)
			}
			if got.SyncedAt != tt.ts {
				t.Errorf("SyncedAt = %v, want %v", got.SyncedAt, tt.ts)
			}
			if got.ConsumedAt != tt.ts {
				t.Errorf("ConsumedAt = %v, want %v", got.ConsumedAt, tt.ts)
			}
		})
	}
}

// --- Mapper with special characters in string fields ---
// Ensures strings containing SQL-sensitive characters pass through without mutation.

func TestDatabase_MapChangeEvent_SpecialCharacterStrings(t *testing.T) {
	t.Parallel()
	d := Database{}

	tests := []struct {
		name      string
		tableName string
		recordID  string
	}{
		{"simple", "users", "user-001"},
		{"underscores", "content_data", "cd_abc_123"},
		{"dots", "schema.table", "ns.record.id"},
		{"quotes", `table"name`, `record'id`},
		{"unicode", "table_\u00e9\u00e0\u00fc", "record_\u4e16\u754c"},
		{"empty", "", ""},
		{"spaces", "table name", "record id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := mdb.ChangeEvent{
				TableName: tt.tableName,
				RecordID:  tt.recordID,
			}
			got := d.MapChangeEvent(input)
			if got.TableName != tt.tableName {
				t.Errorf("TableName = %q, want %q", got.TableName, tt.tableName)
			}
			if got.RecordID != tt.recordID {
				t.Errorf("RecordID = %q, want %q", got.RecordID, tt.recordID)
			}
		})
	}
}

func TestDatabase_MapRecordChangeEventParams_SpecialCharacterStrings(t *testing.T) {
	t.Parallel()
	d := Database{}

	tests := []struct {
		name      string
		tableName string
		recordID  string
	}{
		{"simple", "routes", "route-abc"},
		{"dots", "public.routes", "ns.route.id"},
		{"unicode", "\u00fc\u00e8\u00e9\u00e0", "\u4e16\u754c\u0414"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := RecordChangeEventParams{
				TableName: tt.tableName,
				RecordID:  tt.recordID,
			}
			got := d.MapRecordChangeEventParams(input)
			if got.TableName != tt.tableName {
				t.Errorf("TableName = %q, want %q", got.TableName, tt.tableName)
			}
			if got.RecordID != tt.recordID {
				t.Errorf("RecordID = %q, want %q", got.RecordID, tt.recordID)
			}
		})
	}
}
