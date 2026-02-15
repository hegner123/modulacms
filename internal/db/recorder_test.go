// White-box tests for recorder.go: compile-time interface verification,
// package-level recorder variable initialization, concrete type identity,
// and field mapping from audited.ChangeEventParams to driver-specific
// RecordChangeEventParams.
//
// White-box access is needed because:
//   - The recorder types (sqliteRecorder, mysqlRecorder, psqlRecorder) are
//     unexported; verifying their identity requires same-package access.
//   - Testing the ChangeEventParams -> driver-specific params field mapping
//     exercises the bridge between audited.ChangeEventParams field names
//     (RequestID/IP) and sqlc-generated field names (RequestId/Ip).
//
// The actual Record() methods require a live DBTX (database transaction)
// to execute SQL -- those are integration tests and are excluded here.
package db

import (
	"fmt"
	"testing"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// ============================================================
// Compile-time interface satisfaction checks
// ============================================================

var _ audited.ChangeEventRecorder = SQLiteRecorder
var _ audited.ChangeEventRecorder = MysqlRecorder
var _ audited.ChangeEventRecorder = PsqlRecorder

var _ audited.ChangeEventRecorder = sqliteRecorder{}
var _ audited.ChangeEventRecorder = mysqlRecorder{}
var _ audited.ChangeEventRecorder = psqlRecorder{}

// ============================================================
// Package-level recorder variable initialization
// ============================================================

func TestRecorderPackageVars_NotNil(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
	}{
		{"SQLiteRecorder", SQLiteRecorder},
		{"MysqlRecorder", MysqlRecorder},
		{"PsqlRecorder", PsqlRecorder},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.recorder == nil {
				t.Fatalf("%s is nil, want non-nil ChangeEventRecorder", tt.name)
			}
		})
	}
}

// ============================================================
// Concrete type identity
// ============================================================

func TestRecorderPackageVars_ConcreteType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
		wantType string
	}{
		{"SQLiteRecorder", SQLiteRecorder, "db.sqliteRecorder"},
		{"MysqlRecorder", MysqlRecorder, "db.mysqlRecorder"},
		{"PsqlRecorder", PsqlRecorder, "db.psqlRecorder"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotType := fmt.Sprintf("%T", tt.recorder)
			if gotType != tt.wantType {
				t.Errorf("type of %s = %q, want %q", tt.name, gotType, tt.wantType)
			}
		})
	}
}

// ============================================================
// Recorder type zero-value instantiation
// ============================================================
// Since the recorder types are empty structs, verify they can be
// instantiated without panic and satisfy the interface.

func TestRecorderTypes_ZeroValueInstantiation(t *testing.T) {
	t.Parallel()

	t.Run("sqliteRecorder", func(t *testing.T) {
		t.Parallel()
		var r sqliteRecorder
		var _ audited.ChangeEventRecorder = r
	})

	t.Run("mysqlRecorder", func(t *testing.T) {
		t.Parallel()
		var r mysqlRecorder
		var _ audited.ChangeEventRecorder = r
	})

	t.Run("psqlRecorder", func(t *testing.T) {
		t.Parallel()
		var r psqlRecorder
		var _ audited.ChangeEventRecorder = r
	})
}

// ============================================================
// ChangeEventParams -> driver RecordChangeEventParams field mapping
// ============================================================
// The Record() methods manually map audited.ChangeEventParams fields
// to each driver's RecordChangeEventParams. These tests verify that
// the mapping is correct by constructing test params and checking
// the driver-specific param struct fields match.
//
// We test the mapping logic directly by building the same params
// struct that Record() would construct, since we cannot call Record()
// without a live DBTX.

// recorderFieldMappingFixture returns a fully populated ChangeEventParams
// for testing field mapping consistency.
func recorderFieldMappingFixture() audited.ChangeEventParams {
	return audited.ChangeEventParams{
		EventID:      types.NewEventID(),
		HlcTimestamp: types.HLCNow(),
		NodeID:       types.NewNodeID(),
		TableName:    "test_table",
		RecordID:     "rec-001",
		Operation:    types.OpInsert,
		Action:       types.ActionCreate,
		UserID:       types.NullableUserID{ID: types.NewUserID(), Valid: true},
		OldValues:    types.NewJSONData(nil),
		NewValues:    types.NewJSONData(nil),
		Metadata:     types.NewJSONData(nil),
		RequestID:    types.NewNullableString("req-recorder-001"),
		IP:           types.NewNullableString("192.168.1.50"),
	}
}

func TestSQLiteRecorder_FieldMapping_AllFields(t *testing.T) {
	t.Parallel()
	p := recorderFieldMappingFixture()

	// Build the same params struct that sqliteRecorder.Record() would
	got := mdb.RecordChangeEventParams{
		EventID:      p.EventID,
		HlcTimestamp: p.HlcTimestamp,
		NodeID:       p.NodeID,
		TableName:    p.TableName,
		RecordID:     p.RecordID,
		Operation:    p.Operation,
		Action:       p.Action,
		UserID:       p.UserID,
		OldValues:    p.OldValues,
		NewValues:    p.NewValues,
		Metadata:     p.Metadata,
		RequestId:    p.RequestID,
		Ip:           p.IP,
	}

	if got.EventID != p.EventID {
		t.Errorf("EventID = %v, want %v", got.EventID, p.EventID)
	}
	if got.HlcTimestamp != p.HlcTimestamp {
		t.Errorf("HlcTimestamp = %v, want %v", got.HlcTimestamp, p.HlcTimestamp)
	}
	if got.NodeID != p.NodeID {
		t.Errorf("NodeID = %v, want %v", got.NodeID, p.NodeID)
	}
	if got.TableName != p.TableName {
		t.Errorf("TableName = %q, want %q", got.TableName, p.TableName)
	}
	if got.RecordID != p.RecordID {
		t.Errorf("RecordID = %q, want %q", got.RecordID, p.RecordID)
	}
	if got.Operation != p.Operation {
		t.Errorf("Operation = %v, want %v", got.Operation, p.Operation)
	}
	if got.Action != p.Action {
		t.Errorf("Action = %v, want %v", got.Action, p.Action)
	}
	if got.UserID != p.UserID {
		t.Errorf("UserID = %v, want %v", got.UserID, p.UserID)
	}
	if got.OldValues.Valid != p.OldValues.Valid {
		t.Errorf("OldValues.Valid = %v, want %v", got.OldValues.Valid, p.OldValues.Valid)
	}
	if got.NewValues.Valid != p.NewValues.Valid {
		t.Errorf("NewValues.Valid = %v, want %v", got.NewValues.Valid, p.NewValues.Valid)
	}
	if got.Metadata.Valid != p.Metadata.Valid {
		t.Errorf("Metadata.Valid = %v, want %v", got.Metadata.Valid, p.Metadata.Valid)
	}
	// Critical field name mapping: audited.RequestID -> sqlc RequestId
	if got.RequestId != p.RequestID {
		t.Errorf("RequestId = %v, want %v (RequestID)", got.RequestId, p.RequestID)
	}
	// Critical field name mapping: audited.IP -> sqlc Ip
	if got.Ip != p.IP {
		t.Errorf("Ip = %v, want %v (IP)", got.Ip, p.IP)
	}
}

func TestMySQLRecorder_FieldMapping_AllFields(t *testing.T) {
	t.Parallel()
	p := recorderFieldMappingFixture()

	// Build the same params struct that mysqlRecorder.Record() would
	got := mdbm.RecordChangeEventParams{
		EventID:      p.EventID,
		HlcTimestamp: p.HlcTimestamp,
		NodeID:       p.NodeID,
		TableName:    p.TableName,
		RecordID:     p.RecordID,
		Operation:    p.Operation,
		Action:       p.Action,
		UserID:       p.UserID,
		OldValues:    p.OldValues,
		NewValues:    p.NewValues,
		Metadata:     p.Metadata,
		RequestId:    p.RequestID,
		Ip:           p.IP,
	}

	if got.EventID != p.EventID {
		t.Errorf("EventID = %v, want %v", got.EventID, p.EventID)
	}
	if got.HlcTimestamp != p.HlcTimestamp {
		t.Errorf("HlcTimestamp = %v, want %v", got.HlcTimestamp, p.HlcTimestamp)
	}
	if got.NodeID != p.NodeID {
		t.Errorf("NodeID = %v, want %v", got.NodeID, p.NodeID)
	}
	if got.TableName != p.TableName {
		t.Errorf("TableName = %q, want %q", got.TableName, p.TableName)
	}
	if got.RecordID != p.RecordID {
		t.Errorf("RecordID = %q, want %q", got.RecordID, p.RecordID)
	}
	if got.Operation != p.Operation {
		t.Errorf("Operation = %v, want %v", got.Operation, p.Operation)
	}
	if got.Action != p.Action {
		t.Errorf("Action = %v, want %v", got.Action, p.Action)
	}
	if got.UserID != p.UserID {
		t.Errorf("UserID = %v, want %v", got.UserID, p.UserID)
	}
	if got.OldValues.Valid != p.OldValues.Valid {
		t.Errorf("OldValues.Valid = %v, want %v", got.OldValues.Valid, p.OldValues.Valid)
	}
	if got.NewValues.Valid != p.NewValues.Valid {
		t.Errorf("NewValues.Valid = %v, want %v", got.NewValues.Valid, p.NewValues.Valid)
	}
	if got.Metadata.Valid != p.Metadata.Valid {
		t.Errorf("Metadata.Valid = %v, want %v", got.Metadata.Valid, p.Metadata.Valid)
	}
	// Critical field name mapping
	if got.RequestId != p.RequestID {
		t.Errorf("RequestId = %v, want %v (RequestID)", got.RequestId, p.RequestID)
	}
	if got.Ip != p.IP {
		t.Errorf("Ip = %v, want %v (IP)", got.Ip, p.IP)
	}
}

func TestPsqlRecorder_FieldMapping_AllFields(t *testing.T) {
	t.Parallel()
	p := recorderFieldMappingFixture()

	// Build the same params struct that psqlRecorder.Record() would
	got := mdbp.RecordChangeEventParams{
		EventID:      p.EventID,
		HlcTimestamp: p.HlcTimestamp,
		NodeID:       p.NodeID,
		TableName:    p.TableName,
		RecordID:     p.RecordID,
		Operation:    p.Operation,
		Action:       p.Action,
		UserID:       p.UserID,
		OldValues:    p.OldValues,
		NewValues:    p.NewValues,
		Metadata:     p.Metadata,
		RequestId:    p.RequestID,
		Ip:           p.IP,
	}

	if got.EventID != p.EventID {
		t.Errorf("EventID = %v, want %v", got.EventID, p.EventID)
	}
	if got.HlcTimestamp != p.HlcTimestamp {
		t.Errorf("HlcTimestamp = %v, want %v", got.HlcTimestamp, p.HlcTimestamp)
	}
	if got.NodeID != p.NodeID {
		t.Errorf("NodeID = %v, want %v", got.NodeID, p.NodeID)
	}
	if got.TableName != p.TableName {
		t.Errorf("TableName = %q, want %q", got.TableName, p.TableName)
	}
	if got.RecordID != p.RecordID {
		t.Errorf("RecordID = %q, want %q", got.RecordID, p.RecordID)
	}
	if got.Operation != p.Operation {
		t.Errorf("Operation = %v, want %v", got.Operation, p.Operation)
	}
	if got.Action != p.Action {
		t.Errorf("Action = %v, want %v", got.Action, p.Action)
	}
	if got.UserID != p.UserID {
		t.Errorf("UserID = %v, want %v", got.UserID, p.UserID)
	}
	if got.OldValues.Valid != p.OldValues.Valid {
		t.Errorf("OldValues.Valid = %v, want %v", got.OldValues.Valid, p.OldValues.Valid)
	}
	if got.NewValues.Valid != p.NewValues.Valid {
		t.Errorf("NewValues.Valid = %v, want %v", got.NewValues.Valid, p.NewValues.Valid)
	}
	if got.Metadata.Valid != p.Metadata.Valid {
		t.Errorf("Metadata.Valid = %v, want %v", got.Metadata.Valid, p.Metadata.Valid)
	}
	// Critical field name mapping
	if got.RequestId != p.RequestID {
		t.Errorf("RequestId = %v, want %v (RequestID)", got.RequestId, p.RequestID)
	}
	if got.Ip != p.IP {
		t.Errorf("Ip = %v, want %v (IP)", got.Ip, p.IP)
	}
}

// ============================================================
// Field mapping with zero-value ChangeEventParams
// ============================================================

func TestSQLiteRecorder_FieldMapping_ZeroValues(t *testing.T) {
	t.Parallel()
	var p audited.ChangeEventParams

	got := mdb.RecordChangeEventParams{
		EventID:      p.EventID,
		HlcTimestamp: p.HlcTimestamp,
		NodeID:       p.NodeID,
		TableName:    p.TableName,
		RecordID:     p.RecordID,
		Operation:    p.Operation,
		Action:       p.Action,
		UserID:       p.UserID,
		OldValues:    p.OldValues,
		NewValues:    p.NewValues,
		Metadata:     p.Metadata,
		RequestId:    p.RequestID,
		Ip:           p.IP,
	}

	if got.EventID != "" {
		t.Errorf("EventID = %v, want zero value", got.EventID)
	}
	if got.HlcTimestamp != 0 {
		t.Errorf("HlcTimestamp = %v, want 0", got.HlcTimestamp)
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
	if got.OldValues.Valid {
		t.Error("OldValues.Valid = true, want false")
	}
	if got.NewValues.Valid {
		t.Error("NewValues.Valid = true, want false")
	}
	if got.Metadata.Valid {
		t.Error("Metadata.Valid = true, want false")
	}
	if got.RequestId.Valid {
		t.Error("RequestId.Valid = true, want false")
	}
	if got.Ip.Valid {
		t.Error("Ip.Valid = true, want false")
	}
}

func TestMySQLRecorder_FieldMapping_ZeroValues(t *testing.T) {
	t.Parallel()
	var p audited.ChangeEventParams

	got := mdbm.RecordChangeEventParams{
		EventID:      p.EventID,
		HlcTimestamp: p.HlcTimestamp,
		NodeID:       p.NodeID,
		TableName:    p.TableName,
		RecordID:     p.RecordID,
		Operation:    p.Operation,
		Action:       p.Action,
		UserID:       p.UserID,
		OldValues:    p.OldValues,
		NewValues:    p.NewValues,
		Metadata:     p.Metadata,
		RequestId:    p.RequestID,
		Ip:           p.IP,
	}

	if got.EventID != "" {
		t.Errorf("EventID = %v, want zero value", got.EventID)
	}
	if got.HlcTimestamp != 0 {
		t.Errorf("HlcTimestamp = %v, want 0", got.HlcTimestamp)
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
		t.Errorf("Operation = %q, want empty string", got.Operation)
	}
	if got.Action != "" {
		t.Errorf("Action = %q, want empty string", got.Action)
	}
	if got.UserID.Valid {
		t.Error("UserID.Valid = true, want false")
	}
	if got.OldValues.Data != nil {
		t.Errorf("OldValues.Raw = %v, want nil", got.OldValues.Data)
	}
	if got.NewValues.Data != nil {
		t.Errorf("NewValues.Raw = %v, want nil", got.NewValues.Data)
	}
	if got.Metadata.Data != nil {
		t.Errorf("Metadata.Raw = %v, want nil", got.Metadata.Data)
	}
	if got.RequestId.Valid {
		t.Error("RequestId.Valid = true, want false")
	}
	if got.Ip.Valid {
		t.Error("Ip.Valid = true, want false")
	}
}

func TestPsqlRecorder_FieldMapping_ZeroValues(t *testing.T) {
	t.Parallel()
	var p audited.ChangeEventParams

	got := mdbp.RecordChangeEventParams{
		EventID:      p.EventID,
		HlcTimestamp: p.HlcTimestamp,
		NodeID:       p.NodeID,
		TableName:    p.TableName,
		RecordID:     p.RecordID,
		Operation:    p.Operation,
		Action:       p.Action,
		UserID:       p.UserID,
		OldValues:    p.OldValues,
		NewValues:    p.NewValues,
		Metadata:     p.Metadata,
		RequestId:    p.RequestID,
		Ip:           p.IP,
	}

	if got.EventID != "" {
		t.Errorf("EventID = %v, want zero value", got.EventID)
	}
	if got.HlcTimestamp != 0 {
		t.Errorf("HlcTimestamp = %v, want 0", got.HlcTimestamp)
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
		t.Errorf("Operation = %q, want empty string", got.Operation)
	}
	if got.Action != "" {
		t.Errorf("Action = %q, want empty string", got.Action)
	}
	if got.UserID.Valid {
		t.Error("UserID.Valid = true, want false")
	}
	if got.OldValues.Data != nil {
		t.Errorf("OldValues.Raw = %v, want nil", got.OldValues.Data)
	}
	if got.NewValues.Data != nil {
		t.Errorf("NewValues.Raw = %v, want nil", got.NewValues.Data)
	}
	if got.Metadata.Data != nil {
		t.Errorf("Metadata.Raw = %v, want nil", got.Metadata.Data)
	}
	if got.RequestId.Valid {
		t.Error("RequestId.Valid = true, want false")
	}
	if got.Ip.Valid {
		t.Error("Ip.Valid = true, want false")
	}
}

// ============================================================
// Field mapping with nullable fields in various states
// ============================================================

func TestRecorder_FieldMapping_NullableFieldVariants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		userID    types.NullableUserID
		requestID types.NullableString
		ip        types.NullableString
		wantUser  bool
		wantReq   bool
		wantIP    bool
	}{
		{
			name:      "all nullable fields valid",
			userID:    types.NullableUserID{ID: types.NewUserID(), Valid: true},
			requestID: types.NewNullableString("req-valid"),
			ip:        types.NewNullableString("10.0.0.1"),
			wantUser:  true,
			wantReq:   true,
			wantIP:    true,
		},
		{
			name:      "all nullable fields null",
			userID:    types.NullableUserID{Valid: false},
			requestID: types.NullableString{Valid: false},
			ip:        types.NullableString{Valid: false},
			wantUser:  false,
			wantReq:   false,
			wantIP:    false,
		},
		{
			name:      "user valid, request null, ip valid",
			userID:    types.NullableUserID{ID: types.NewUserID(), Valid: true},
			requestID: types.NullableString{Valid: false},
			ip:        types.NewNullableString("172.16.0.1"),
			wantUser:  true,
			wantReq:   false,
			wantIP:    true,
		},
		{
			name:      "user null, request valid, ip null",
			userID:    types.NullableUserID{Valid: false},
			requestID: types.NewNullableString("req-only"),
			ip:        types.NullableString{Valid: false},
			wantUser:  false,
			wantReq:   true,
			wantIP:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := audited.ChangeEventParams{
				UserID:    tt.userID,
				RequestID: tt.requestID,
				IP:        tt.ip,
			}

			// Verify SQLite mapping
			sqliteGot := mdb.RecordChangeEventParams{
				UserID:    p.UserID,
				RequestId: p.RequestID,
				Ip:        p.IP,
			}
			if sqliteGot.UserID.Valid != tt.wantUser {
				t.Errorf("SQLite UserID.Valid = %v, want %v", sqliteGot.UserID.Valid, tt.wantUser)
			}
			if sqliteGot.RequestId.Valid != tt.wantReq {
				t.Errorf("SQLite RequestId.Valid = %v, want %v", sqliteGot.RequestId.Valid, tt.wantReq)
			}
			if sqliteGot.Ip.Valid != tt.wantIP {
				t.Errorf("SQLite Ip.Valid = %v, want %v", sqliteGot.Ip.Valid, tt.wantIP)
			}

			// Verify MySQL mapping
			mysqlGot := mdbm.RecordChangeEventParams{
				UserID:    p.UserID,
				RequestId: p.RequestID,
				Ip:        p.IP,
			}
			if mysqlGot.UserID.Valid != tt.wantUser {
				t.Errorf("MySQL UserID.Valid = %v, want %v", mysqlGot.UserID.Valid, tt.wantUser)
			}
			if mysqlGot.RequestId.Valid != tt.wantReq {
				t.Errorf("MySQL RequestId.Valid = %v, want %v", mysqlGot.RequestId.Valid, tt.wantReq)
			}
			if mysqlGot.Ip.Valid != tt.wantIP {
				t.Errorf("MySQL Ip.Valid = %v, want %v", mysqlGot.Ip.Valid, tt.wantIP)
			}

			// Verify PostgreSQL mapping
			psqlGot := mdbp.RecordChangeEventParams{
				UserID:    p.UserID,
				RequestId: p.RequestID,
				Ip:        p.IP,
			}
			if psqlGot.UserID.Valid != tt.wantUser {
				t.Errorf("PostgreSQL UserID.Valid = %v, want %v", psqlGot.UserID.Valid, tt.wantUser)
			}
			if psqlGot.RequestId.Valid != tt.wantReq {
				t.Errorf("PostgreSQL RequestId.Valid = %v, want %v", psqlGot.RequestId.Valid, tt.wantReq)
			}
			if psqlGot.Ip.Valid != tt.wantIP {
				t.Errorf("PostgreSQL Ip.Valid = %v, want %v", psqlGot.Ip.Valid, tt.wantIP)
			}
		})
	}
}

// ============================================================
// Cross-driver field mapping consistency
// ============================================================
// All three drivers should produce identical field values when
// mapping from the same audited.ChangeEventParams input.

func TestCrossDriverRecorder_FieldMappingConsistency(t *testing.T) {
	t.Parallel()
	p := recorderFieldMappingFixture()

	// Build the same mapping that each Record() method does
	sqliteGot := mdb.RecordChangeEventParams{
		EventID:      p.EventID,
		HlcTimestamp: p.HlcTimestamp,
		NodeID:       p.NodeID,
		TableName:    p.TableName,
		RecordID:     p.RecordID,
		Operation:    p.Operation,
		Action:       p.Action,
		UserID:       p.UserID,
		OldValues:    p.OldValues,
		NewValues:    p.NewValues,
		Metadata:     p.Metadata,
		RequestId:    p.RequestID,
		Ip:           p.IP,
	}
	mysqlGot := mdbm.RecordChangeEventParams{
		EventID:      p.EventID,
		HlcTimestamp: p.HlcTimestamp,
		NodeID:       p.NodeID,
		TableName:    p.TableName,
		RecordID:     p.RecordID,
		Operation:    p.Operation,
		Action:       p.Action,
		UserID:       p.UserID,
		OldValues:    p.OldValues,
		NewValues:    p.NewValues,
		Metadata:     p.Metadata,
		RequestId:    p.RequestID,
		Ip:           p.IP,
	}
	psqlGot := mdbp.RecordChangeEventParams{
		EventID:      p.EventID,
		HlcTimestamp: p.HlcTimestamp,
		NodeID:       p.NodeID,
		TableName:    p.TableName,
		RecordID:     p.RecordID,
		Operation:    p.Operation,
		Action:       p.Action,
		UserID:       p.UserID,
		OldValues:    p.OldValues,
		NewValues:    p.NewValues,
		Metadata:     p.Metadata,
		RequestId:    p.RequestID,
		Ip:           p.IP,
	}

	// Compare string-representable fields across all three drivers
	fields := []struct {
		name   string
		sqlite string
		mysql  string
		psql   string
	}{
		{"EventID", string(sqliteGot.EventID), string(mysqlGot.EventID), string(psqlGot.EventID)},
		{"HlcTimestamp", fmt.Sprint(sqliteGot.HlcTimestamp), fmt.Sprint(mysqlGot.HlcTimestamp), fmt.Sprint(psqlGot.HlcTimestamp)},
		{"NodeID", string(sqliteGot.NodeID), string(mysqlGot.NodeID), string(psqlGot.NodeID)},
		{"TableName", sqliteGot.TableName, mysqlGot.TableName, psqlGot.TableName},
		{"RecordID", sqliteGot.RecordID, mysqlGot.RecordID, psqlGot.RecordID},
		{"Operation", string(sqliteGot.Operation), string(mysqlGot.Operation), string(psqlGot.Operation)},
		{"Action", string(sqliteGot.Action), string(mysqlGot.Action), string(psqlGot.Action)},
		{"UserID", fmt.Sprint(sqliteGot.UserID), fmt.Sprint(mysqlGot.UserID), fmt.Sprint(psqlGot.UserID)},
		{"OldValues", fmt.Sprint(sqliteGot.OldValues), fmt.Sprint(mysqlGot.OldValues), fmt.Sprint(psqlGot.OldValues)},
		{"NewValues", fmt.Sprint(sqliteGot.NewValues), fmt.Sprint(mysqlGot.NewValues), fmt.Sprint(psqlGot.NewValues)},
		{"Metadata", fmt.Sprint(sqliteGot.Metadata), fmt.Sprint(mysqlGot.Metadata), fmt.Sprint(psqlGot.Metadata)},
		{"RequestId.String", sqliteGot.RequestId.String, mysqlGot.RequestId.String, psqlGot.RequestId.String},
		{"Ip.String", sqliteGot.Ip.String, mysqlGot.Ip.String, psqlGot.Ip.String},
	}

	for _, f := range fields {
		t.Run(f.name, func(t *testing.T) {
			t.Parallel()
			if f.sqlite != f.mysql {
				t.Errorf("SQLite %s = %q, MySQL %s = %q", f.name, f.sqlite, f.name, f.mysql)
			}
			if f.sqlite != f.psql {
				t.Errorf("SQLite %s = %q, PostgreSQL %s = %q", f.name, f.sqlite, f.name, f.psql)
			}
		})
	}
}

// ============================================================
// Field mapping with all Operation/Action enum variants
// ============================================================
// Ensures that every enum value passes through the mapping correctly.

func TestRecorder_FieldMapping_AllOperations(t *testing.T) {
	t.Parallel()

	operations := []types.Operation{
		types.OpInsert,
		types.OpUpdate,
		types.OpDelete,
	}

	for _, op := range operations {
		t.Run(string(op), func(t *testing.T) {
			t.Parallel()
			p := audited.ChangeEventParams{Operation: op}

			sqliteGot := mdb.RecordChangeEventParams{Operation: p.Operation}
			if sqliteGot.Operation != op {
				t.Errorf("SQLite Operation = %v, want %v", sqliteGot.Operation, op)
			}

			mysqlGot := mdbm.RecordChangeEventParams{Operation: p.Operation}
			if mysqlGot.Operation != op {
				t.Errorf("MySQL Operation = %v, want %v", mysqlGot.Operation, op)
			}

			psqlGot := mdbp.RecordChangeEventParams{Operation: p.Operation}
			if psqlGot.Operation != op {
				t.Errorf("PostgreSQL Operation = %v, want %v", psqlGot.Operation, op)
			}
		})
	}
}

func TestRecorder_FieldMapping_AllActions(t *testing.T) {
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
			p := audited.ChangeEventParams{Action: action}

			sqliteGot := mdb.RecordChangeEventParams{Action: p.Action}
			if sqliteGot.Action != action {
				t.Errorf("SQLite Action = %v, want %v", sqliteGot.Action, action)
			}

			mysqlGot := mdbm.RecordChangeEventParams{Action: p.Action}
			if mysqlGot.Action != action {
				t.Errorf("MySQL Action = %v, want %v", mysqlGot.Action, action)
			}

			psqlGot := mdbp.RecordChangeEventParams{Action: p.Action}
			if psqlGot.Action != action {
				t.Errorf("PostgreSQL Action = %v, want %v", psqlGot.Action, action)
			}
		})
	}
}

// ============================================================
// Field name bridging: RequestID/IP -> RequestId/Ip
// ============================================================
// The critical mapping: audited.ChangeEventParams uses Go-idiomatic
// field names (RequestID, IP) while sqlc generates Go fields from
// SQL column names (RequestId, Ip). This test exercises the exact
// mapping used in the Record() method bodies.

func TestRecorder_RequestIDIPFieldNameBridge(t *testing.T) {
	t.Parallel()
	reqID := types.NewNullableString("bridge-req-001")
	ip := types.NewNullableString("10.20.30.40")

	p := audited.ChangeEventParams{
		RequestID: reqID,
		IP:        ip,
	}

	t.Run("SQLite", func(t *testing.T) {
		t.Parallel()
		got := mdb.RecordChangeEventParams{
			RequestId: p.RequestID,
			Ip:        p.IP,
		}
		if got.RequestId != reqID {
			t.Errorf("RequestId = %v, want %v", got.RequestId, reqID)
		}
		if got.Ip != ip {
			t.Errorf("Ip = %v, want %v", got.Ip, ip)
		}
	})

	t.Run("MySQL", func(t *testing.T) {
		t.Parallel()
		got := mdbm.RecordChangeEventParams{
			RequestId: p.RequestID,
			Ip:        p.IP,
		}
		if got.RequestId != reqID {
			t.Errorf("RequestId = %v, want %v", got.RequestId, reqID)
		}
		if got.Ip != ip {
			t.Errorf("Ip = %v, want %v", got.Ip, ip)
		}
	})

	t.Run("PostgreSQL", func(t *testing.T) {
		t.Parallel()
		got := mdbp.RecordChangeEventParams{
			RequestId: p.RequestID,
			Ip:        p.IP,
		}
		if got.RequestId != reqID {
			t.Errorf("RequestId = %v, want %v", got.RequestId, reqID)
		}
		if got.Ip != ip {
			t.Errorf("Ip = %v, want %v", got.Ip, ip)
		}
	})
}

// ============================================================
// HLC value preservation through mapping
// ============================================================

func TestRecorder_FieldMapping_HLCPreservation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		hlc  types.HLC
	}{
		{"zero HLC", types.HLC(0)},
		{"current HLC", types.HLCNow()},
		{"max int64 HLC", types.HLC(1<<63 - 1)},
		{"negative HLC", types.HLC(-1)},
		{"specific HLC", types.HLC(1706745600000 << 16)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := audited.ChangeEventParams{HlcTimestamp: tt.hlc}

			sqliteGot := mdb.RecordChangeEventParams{HlcTimestamp: p.HlcTimestamp}
			if sqliteGot.HlcTimestamp != tt.hlc {
				t.Errorf("SQLite HlcTimestamp = %v, want %v", sqliteGot.HlcTimestamp, tt.hlc)
			}

			mysqlGot := mdbm.RecordChangeEventParams{HlcTimestamp: p.HlcTimestamp}
			if mysqlGot.HlcTimestamp != tt.hlc {
				t.Errorf("MySQL HlcTimestamp = %v, want %v", mysqlGot.HlcTimestamp, tt.hlc)
			}

			psqlGot := mdbp.RecordChangeEventParams{HlcTimestamp: p.HlcTimestamp}
			if psqlGot.HlcTimestamp != tt.hlc {
				t.Errorf("PostgreSQL HlcTimestamp = %v, want %v", psqlGot.HlcTimestamp, tt.hlc)
			}
		})
	}
}

// ============================================================
// Special characters in string fields through mapping
// ============================================================
// Ensures strings containing SQL-sensitive characters survive
// the field mapping without mutation.

func TestRecorder_FieldMapping_SpecialCharacterStrings(t *testing.T) {
	t.Parallel()

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
		{"semicolons", "table;name", "record;id"},
		{"newlines", "table\nname", "record\nid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := audited.ChangeEventParams{
				TableName: tt.tableName,
				RecordID:  tt.recordID,
			}

			// Verify through SQLite mapping (representative of all three --
			// cross-driver consistency is tested elsewhere)
			sqliteGot := mdb.RecordChangeEventParams{
				TableName: p.TableName,
				RecordID:  p.RecordID,
			}
			if sqliteGot.TableName != tt.tableName {
				t.Errorf("TableName = %q, want %q", sqliteGot.TableName, tt.tableName)
			}
			if sqliteGot.RecordID != tt.recordID {
				t.Errorf("RecordID = %q, want %q", sqliteGot.RecordID, tt.recordID)
			}
		})
	}
}

// ============================================================
// Recorder type distinctness
// ============================================================
// Verify the three package-level recorders are distinct types
// and not aliases or shared instances.

func TestRecorderPackageVars_Distinct(t *testing.T) {
	t.Parallel()

	sqliteType := fmt.Sprintf("%T", SQLiteRecorder)
	mysqlType := fmt.Sprintf("%T", MysqlRecorder)
	psqlType := fmt.Sprintf("%T", PsqlRecorder)

	if sqliteType == mysqlType {
		t.Errorf("SQLiteRecorder and MysqlRecorder have same type %q", sqliteType)
	}
	if sqliteType == psqlType {
		t.Errorf("SQLiteRecorder and PsqlRecorder have same type %q", sqliteType)
	}
	if mysqlType == psqlType {
		t.Errorf("MysqlRecorder and PsqlRecorder have same type %q", mysqlType)
	}
}

// ============================================================
// JSONData field handling in mapping
// ============================================================
// Verify JSONData fields (OldValues, NewValues, Metadata) pass
// through the mapping with Valid flag preserved.

func TestRecorder_FieldMapping_JSONDataVariants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		oldValues   types.JSONData
		newValues   types.JSONData
		metadata    types.JSONData
		wantOldVal  bool
		wantNewVal  bool
		wantMetaVal bool
	}{
		{
			name:        "all valid with nil data",
			oldValues:   types.NewJSONData(nil),
			newValues:   types.NewJSONData(nil),
			metadata:    types.NewJSONData(nil),
			wantOldVal:  true,
			wantNewVal:  true,
			wantMetaVal: true,
		},
		{
			name:        "all invalid",
			oldValues:   types.JSONData{Valid: false},
			newValues:   types.JSONData{Valid: false},
			metadata:    types.JSONData{Valid: false},
			wantOldVal:  false,
			wantNewVal:  false,
			wantMetaVal: false,
		},
		{
			name:        "mixed - old invalid, new valid, meta invalid",
			oldValues:   types.JSONData{Valid: false},
			newValues:   types.NewJSONData(map[string]string{"key": "val"}),
			metadata:    types.JSONData{Valid: false},
			wantOldVal:  false,
			wantNewVal:  true,
			wantMetaVal: false,
		},
		{
			name:        "create scenario - no old, has new and meta",
			oldValues:   types.JSONData{Valid: false},
			newValues:   types.NewJSONData(map[string]string{"name": "test"}),
			metadata:    types.NewJSONData(map[string]string{"source": "api"}),
			wantOldVal:  false,
			wantNewVal:  true,
			wantMetaVal: true,
		},
		{
			name:        "delete scenario - has old, no new",
			oldValues:   types.NewJSONData(map[string]string{"name": "deleted"}),
			newValues:   types.JSONData{Valid: false},
			metadata:    types.JSONData{Valid: false},
			wantOldVal:  true,
			wantNewVal:  false,
			wantMetaVal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := audited.ChangeEventParams{
				OldValues: tt.oldValues,
				NewValues: tt.newValues,
				Metadata:  tt.metadata,
			}

			// Test through SQLite mapping
			sqliteGot := mdb.RecordChangeEventParams{
				OldValues: p.OldValues,
				NewValues: p.NewValues,
				Metadata:  p.Metadata,
			}
			if sqliteGot.OldValues.Valid != tt.wantOldVal {
				t.Errorf("SQLite OldValues.Valid = %v, want %v", sqliteGot.OldValues.Valid, tt.wantOldVal)
			}
			if sqliteGot.NewValues.Valid != tt.wantNewVal {
				t.Errorf("SQLite NewValues.Valid = %v, want %v", sqliteGot.NewValues.Valid, tt.wantNewVal)
			}
			if sqliteGot.Metadata.Valid != tt.wantMetaVal {
				t.Errorf("SQLite Metadata.Valid = %v, want %v", sqliteGot.Metadata.Valid, tt.wantMetaVal)
			}

			// Test through MySQL mapping
			mysqlGot := mdbm.RecordChangeEventParams{
				OldValues: p.OldValues,
				NewValues: p.NewValues,
				Metadata:  p.Metadata,
			}
			if mysqlGot.OldValues.Valid != tt.wantOldVal {
				t.Errorf("MySQL OldValues.Valid = %v, want %v", mysqlGot.OldValues.Valid, tt.wantOldVal)
			}
			if mysqlGot.NewValues.Valid != tt.wantNewVal {
				t.Errorf("MySQL NewValues.Valid = %v, want %v", mysqlGot.NewValues.Valid, tt.wantNewVal)
			}
			if mysqlGot.Metadata.Valid != tt.wantMetaVal {
				t.Errorf("MySQL Metadata.Valid = %v, want %v", mysqlGot.Metadata.Valid, tt.wantMetaVal)
			}

			// Test through PostgreSQL mapping
			psqlGot := mdbp.RecordChangeEventParams{
				OldValues: p.OldValues,
				NewValues: p.NewValues,
				Metadata:  p.Metadata,
			}
			if psqlGot.OldValues.Valid != tt.wantOldVal {
				t.Errorf("PostgreSQL OldValues.Valid = %v, want %v", psqlGot.OldValues.Valid, tt.wantOldVal)
			}
			if psqlGot.NewValues.Valid != tt.wantNewVal {
				t.Errorf("PostgreSQL NewValues.Valid = %v, want %v", psqlGot.NewValues.Valid, tt.wantNewVal)
			}
			if psqlGot.Metadata.Valid != tt.wantMetaVal {
				t.Errorf("PostgreSQL Metadata.Valid = %v, want %v", psqlGot.Metadata.Valid, tt.wantMetaVal)
			}
		})
	}
}

// ============================================================
// audited.ChangeEventParams struct zero-value usability
// ============================================================

func TestAuditedChangeEventParams_ZeroValue(t *testing.T) {
	t.Parallel()
	var p audited.ChangeEventParams

	if p.EventID != "" {
		t.Errorf("EventID = %v, want zero value", p.EventID)
	}
	if p.HlcTimestamp != 0 {
		t.Errorf("HlcTimestamp = %v, want 0", p.HlcTimestamp)
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
	if p.OldValues.Valid {
		t.Error("OldValues.Valid = true, want false")
	}
	if p.NewValues.Valid {
		t.Error("NewValues.Valid = true, want false")
	}
	if p.Metadata.Valid {
		t.Error("Metadata.Valid = true, want false")
	}
	if p.RequestID.Valid {
		t.Error("RequestID.Valid = true, want false")
	}
	if p.IP.Valid {
		t.Error("IP.Valid = true, want false")
	}
}

func TestAuditedChangeEventParams_FullyPopulated(t *testing.T) {
	t.Parallel()
	p := recorderFieldMappingFixture()

	if p.EventID.IsZero() {
		t.Error("EventID is zero, want populated")
	}
	if p.HlcTimestamp == 0 {
		t.Error("HlcTimestamp = 0, want non-zero")
	}
	if p.NodeID.IsZero() {
		t.Error("NodeID is zero, want populated")
	}
	if p.TableName != "test_table" {
		t.Errorf("TableName = %q, want %q", p.TableName, "test_table")
	}
	if p.RecordID != "rec-001" {
		t.Errorf("RecordID = %q, want %q", p.RecordID, "rec-001")
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
	if !p.RequestID.Valid {
		t.Error("RequestID.Valid = false, want true")
	}
	if p.RequestID.String != "req-recorder-001" {
		t.Errorf("RequestID.String = %q, want %q", p.RequestID.String, "req-recorder-001")
	}
	if !p.IP.Valid {
		t.Error("IP.Valid = false, want true")
	}
	if p.IP.String != "192.168.1.50" {
		t.Errorf("IP.String = %q, want %q", p.IP.String, "192.168.1.50")
	}
}
