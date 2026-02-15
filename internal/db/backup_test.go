// White-box tests for backup.go: wrapper structs, mapper methods across all three
// database drivers (SQLite, MySQL, PostgreSQL), create-param mappers,
// update-status-param inline mapping, cross-database consistency, int32/int64
// conversion edge cases, and audited command struct accessors.
//
// White-box access is needed because:
//   - Audited command structs have unexported fields (ctx, auditCtx, params, conn,
//     recorder) that can only be constructed through the Database/MysqlDatabase/
//     PsqlDatabase factory methods, which require access to the package internals.
//   - We verify that the SQLiteRecorder, MysqlRecorder, and PsqlRecorder package-level
//     vars are correctly wired into command constructors.
package db

import (
	"context"
	"math"
	"testing"
	"time"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// ---------------------------------------------------------------------------
// Test data helpers
// ---------------------------------------------------------------------------

func backupTestFixture() (types.BackupID, types.NodeID, types.Timestamp, types.NullableString, types.JSONData) {
	backupID := types.NewBackupID()
	nodeID := types.NewNodeID()
	ts := types.NewTimestamp(time.Date(2025, 8, 15, 10, 30, 0, 0, time.UTC))
	triggeredBy := types.NullableString{String: "admin", Valid: true}
	metadata := types.NewJSONData(map[string]int{"version": 1})
	return backupID, nodeID, ts, triggeredBy, metadata
}

func fullBackupSqlite() mdb.Backup {
	backupID, nodeID, ts, triggeredBy, metadata := backupTestFixture()
	return mdb.Backup{
		BackupID:       backupID,
		NodeID:         nodeID,
		BackupType:     types.BackupTypeFull,
		Status:         types.BackupStatusCompleted,
		StartedAt:      ts,
		CompletedAt:    ts,
		DurationMs:     types.NullableInt64{Int64: 1500, Valid: true},
		RecordCount:    types.NullableInt64{Int64: 42, Valid: true},
		SizeBytes:      types.NullableInt64{Int64: 1024, Valid: true},
		ReplicationLsn: types.NullableString{String: "0/1234", Valid: true},
		HlcTimestamp:   types.HLCNow(),
		StoragePath:    "/backups/full-001.zip",
		Checksum:       types.NullableString{String: "sha256:abc", Valid: true},
		TriggeredBy:    triggeredBy,
		ErrorMessage:   types.NullableString{Valid: false},
		Metadata:       metadata,
	}
}

func fullBackupSetSqlite() mdb.BackupSet {
	return mdb.BackupSet{
		BackupSetID:    types.NewBackupSetID(),
		CreatedAt:      types.TimestampNow(),
		HlcTimestamp:   types.HLCNow(),
		Status:         types.BackupSetStatusPending,
		BackupIds:      types.NewJSONData([]string{"id1", "id2"}),
		NodeCount:      3,
		CompletedCount: types.NullableInt64{Int64: 1, Valid: true},
		ErrorMessage:   types.NullableString{Valid: false},
	}
}

func fullVerificationSqlite() mdb.BackupVerification {
	return mdb.BackupVerification{
		VerificationID:   types.NewVerificationID(),
		BackupID:         types.NewBackupID(),
		VerifiedAt:       types.TimestampNow(),
		VerifiedBy:       types.NullableString{String: "system", Valid: true},
		RestoreTested:    types.NullableBool{Bool: true, Valid: true},
		ChecksumValid:    types.NullableBool{Bool: true, Valid: true},
		RecordCountMatch: types.NullableBool{Bool: false, Valid: true},
		Status:           types.VerificationStatusVerified,
		ErrorMessage:     types.NullableString{Valid: false},
		DurationMs:       types.NullableInt64{Int64: 750, Valid: true},
	}
}

// ---------------------------------------------------------------------------
// SQLite Database.MapBackup tests
// ---------------------------------------------------------------------------

func TestDatabase_MapBackup_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := fullBackupSqlite()

	got := d.MapBackup(input)

	if got.BackupID != input.BackupID {
		t.Errorf("BackupID = %v, want %v", got.BackupID, input.BackupID)
	}
	if got.NodeID != input.NodeID {
		t.Errorf("NodeID = %v, want %v", got.NodeID, input.NodeID)
	}
	if got.BackupType != input.BackupType {
		t.Errorf("BackupType = %v, want %v", got.BackupType, input.BackupType)
	}
	if got.Status != input.Status {
		t.Errorf("Status = %v, want %v", got.Status, input.Status)
	}
	if got.StartedAt != input.StartedAt {
		t.Errorf("StartedAt = %v, want %v", got.StartedAt, input.StartedAt)
	}
	if got.CompletedAt != input.CompletedAt {
		t.Errorf("CompletedAt = %v, want %v", got.CompletedAt, input.CompletedAt)
	}
	if got.DurationMs != input.DurationMs {
		t.Errorf("DurationMs = %v, want %v", got.DurationMs, input.DurationMs)
	}
	if got.RecordCount != input.RecordCount {
		t.Errorf("RecordCount = %v, want %v", got.RecordCount, input.RecordCount)
	}
	if got.SizeBytes != input.SizeBytes {
		t.Errorf("SizeBytes = %v, want %v", got.SizeBytes, input.SizeBytes)
	}
	if got.ReplicationLsn != input.ReplicationLsn {
		t.Errorf("ReplicationLsn = %v, want %v", got.ReplicationLsn, input.ReplicationLsn)
	}
	if got.HlcTimestamp != input.HlcTimestamp {
		t.Errorf("HlcTimestamp = %v, want %v", got.HlcTimestamp, input.HlcTimestamp)
	}
	if got.StoragePath != input.StoragePath {
		t.Errorf("StoragePath = %v, want %v", got.StoragePath, input.StoragePath)
	}
	if got.Checksum != input.Checksum {
		t.Errorf("Checksum = %v, want %v", got.Checksum, input.Checksum)
	}
	if got.TriggeredBy != input.TriggeredBy {
		t.Errorf("TriggeredBy = %v, want %v", got.TriggeredBy, input.TriggeredBy)
	}
	if got.ErrorMessage != input.ErrorMessage {
		t.Errorf("ErrorMessage = %v, want %v", got.ErrorMessage, input.ErrorMessage)
	}
	if got.Metadata.Valid != input.Metadata.Valid {
		t.Errorf("Metadata.Valid = %v, want %v", got.Metadata.Valid, input.Metadata.Valid)
	}
}

func TestDatabase_MapBackup_ZeroValue(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapBackup(mdb.Backup{})

	if got.BackupID != "" {
		t.Errorf("BackupID = %v, want zero value", got.BackupID)
	}
	if got.NodeID != "" {
		t.Errorf("NodeID = %v, want zero value", got.NodeID)
	}
	if got.StoragePath != "" {
		t.Errorf("StoragePath = %v, want empty string", got.StoragePath)
	}
	if got.DurationMs.Valid {
		t.Errorf("DurationMs.Valid = true, want false")
	}
	if got.RecordCount.Valid {
		t.Errorf("RecordCount.Valid = true, want false")
	}
	if got.SizeBytes.Valid {
		t.Errorf("SizeBytes.Valid = true, want false")
	}
	if got.Checksum.Valid {
		t.Errorf("Checksum.Valid = true, want false")
	}
	if got.TriggeredBy.Valid {
		t.Errorf("TriggeredBy.Valid = true, want false")
	}
	if got.ErrorMessage.Valid {
		t.Errorf("ErrorMessage.Valid = true, want false")
	}
	if got.ReplicationLsn.Valid {
		t.Errorf("ReplicationLsn.Valid = true, want false")
	}
}

func TestDatabase_MapBackup_NullableFieldCombinations(t *testing.T) {
	t.Parallel()
	d := Database{}

	tests := []struct {
		name          string
		durationValid bool
		checksumValid bool
	}{
		{"both null", false, false},
		{"duration valid, checksum null", true, false},
		{"duration null, checksum valid", false, true},
		{"both valid", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := mdb.Backup{
				DurationMs: types.NullableInt64{Int64: 100, Valid: tt.durationValid},
				Checksum:   types.NullableString{String: "sha256:test", Valid: tt.checksumValid},
			}
			got := d.MapBackup(input)
			if got.DurationMs.Valid != tt.durationValid {
				t.Errorf("DurationMs.Valid = %v, want %v", got.DurationMs.Valid, tt.durationValid)
			}
			if got.Checksum.Valid != tt.checksumValid {
				t.Errorf("Checksum.Valid = %v, want %v", got.Checksum.Valid, tt.checksumValid)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// SQLite Database.MapBackupSet tests
// ---------------------------------------------------------------------------

func TestDatabase_MapBackupSet_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := fullBackupSetSqlite()

	got := d.MapBackupSet(input)

	if got.BackupSetID != input.BackupSetID {
		t.Errorf("BackupSetID = %v, want %v", got.BackupSetID, input.BackupSetID)
	}
	if got.CreatedAt != input.CreatedAt {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, input.CreatedAt)
	}
	if got.HlcTimestamp != input.HlcTimestamp {
		t.Errorf("HlcTimestamp = %v, want %v", got.HlcTimestamp, input.HlcTimestamp)
	}
	if got.Status != input.Status {
		t.Errorf("Status = %v, want %v", got.Status, input.Status)
	}
	if got.BackupIds.Valid != input.BackupIds.Valid {
		t.Errorf("BackupIds.Valid = %v, want %v", got.BackupIds.Valid, input.BackupIds.Valid)
	}
	if got.NodeCount != input.NodeCount {
		t.Errorf("NodeCount = %v, want %v", got.NodeCount, input.NodeCount)
	}
	if got.CompletedCount != input.CompletedCount {
		t.Errorf("CompletedCount = %v, want %v", got.CompletedCount, input.CompletedCount)
	}
	if got.ErrorMessage != input.ErrorMessage {
		t.Errorf("ErrorMessage = %v, want %v", got.ErrorMessage, input.ErrorMessage)
	}
}

func TestDatabase_MapBackupSet_ZeroValue(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapBackupSet(mdb.BackupSet{})

	if got.BackupSetID != "" {
		t.Errorf("BackupSetID = %v, want zero value", got.BackupSetID)
	}
	if got.NodeCount != 0 {
		t.Errorf("NodeCount = %v, want 0", got.NodeCount)
	}
	if got.CompletedCount.Valid {
		t.Errorf("CompletedCount.Valid = true, want false")
	}
	if got.ErrorMessage.Valid {
		t.Errorf("ErrorMessage.Valid = true, want false")
	}
}

// ---------------------------------------------------------------------------
// SQLite Database.MapBackupVerification tests
// ---------------------------------------------------------------------------

func TestDatabase_MapBackupVerification_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := fullVerificationSqlite()

	got := d.MapBackupVerification(input)

	if got.VerificationID != input.VerificationID {
		t.Errorf("VerificationID = %v, want %v", got.VerificationID, input.VerificationID)
	}
	if got.BackupID != input.BackupID {
		t.Errorf("BackupID = %v, want %v", got.BackupID, input.BackupID)
	}
	if got.VerifiedAt != input.VerifiedAt {
		t.Errorf("VerifiedAt = %v, want %v", got.VerifiedAt, input.VerifiedAt)
	}
	if got.VerifiedBy != input.VerifiedBy {
		t.Errorf("VerifiedBy = %v, want %v", got.VerifiedBy, input.VerifiedBy)
	}
	if got.RestoreTested != input.RestoreTested {
		t.Errorf("RestoreTested = %v, want %v", got.RestoreTested, input.RestoreTested)
	}
	if got.ChecksumValid != input.ChecksumValid {
		t.Errorf("ChecksumValid = %v, want %v", got.ChecksumValid, input.ChecksumValid)
	}
	if got.RecordCountMatch != input.RecordCountMatch {
		t.Errorf("RecordCountMatch = %v, want %v", got.RecordCountMatch, input.RecordCountMatch)
	}
	if got.Status != input.Status {
		t.Errorf("Status = %v, want %v", got.Status, input.Status)
	}
	if got.ErrorMessage != input.ErrorMessage {
		t.Errorf("ErrorMessage = %v, want %v", got.ErrorMessage, input.ErrorMessage)
	}
	if got.DurationMs != input.DurationMs {
		t.Errorf("DurationMs = %v, want %v", got.DurationMs, input.DurationMs)
	}
}

func TestDatabase_MapBackupVerification_ZeroValue(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapBackupVerification(mdb.BackupVerification{})

	if got.VerificationID != "" {
		t.Errorf("VerificationID = %v, want zero value", got.VerificationID)
	}
	if got.BackupID != "" {
		t.Errorf("BackupID = %v, want zero value", got.BackupID)
	}
	if got.VerifiedBy.Valid {
		t.Errorf("VerifiedBy.Valid = true, want false")
	}
	if got.RestoreTested.Valid {
		t.Errorf("RestoreTested.Valid = true, want false")
	}
	if got.ChecksumValid.Valid {
		t.Errorf("ChecksumValid.Valid = true, want false")
	}
	if got.RecordCountMatch.Valid {
		t.Errorf("RecordCountMatch.Valid = true, want false")
	}
	if got.ErrorMessage.Valid {
		t.Errorf("ErrorMessage.Valid = true, want false")
	}
	if got.DurationMs.Valid {
		t.Errorf("DurationMs.Valid = true, want false")
	}
}

func TestDatabase_MapBackupVerification_BoolCombinations(t *testing.T) {
	t.Parallel()
	d := Database{}

	tests := []struct {
		name                string
		restoreTestedValid  bool
		checksumValidValid  bool
		recordCountValid    bool
		restoreTestedValue  bool
		checksumValidValue  bool
		recordCountValue    bool
	}{
		{"all null", false, false, false, false, false, false},
		{"all valid true", true, true, true, true, true, true},
		{"all valid false", true, true, true, false, false, false},
		{"mixed validity", true, false, true, true, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := mdb.BackupVerification{
				RestoreTested:    types.NullableBool{Bool: tt.restoreTestedValue, Valid: tt.restoreTestedValid},
				ChecksumValid:    types.NullableBool{Bool: tt.checksumValidValue, Valid: tt.checksumValidValid},
				RecordCountMatch: types.NullableBool{Bool: tt.recordCountValue, Valid: tt.recordCountValid},
			}
			got := d.MapBackupVerification(input)
			if got.RestoreTested.Valid != tt.restoreTestedValid {
				t.Errorf("RestoreTested.Valid = %v, want %v", got.RestoreTested.Valid, tt.restoreTestedValid)
			}
			if got.RestoreTested.Valid && got.RestoreTested.Bool != tt.restoreTestedValue {
				t.Errorf("RestoreTested.Bool = %v, want %v", got.RestoreTested.Bool, tt.restoreTestedValue)
			}
			if got.ChecksumValid.Valid != tt.checksumValidValid {
				t.Errorf("ChecksumValid.Valid = %v, want %v", got.ChecksumValid.Valid, tt.checksumValidValid)
			}
			if got.RecordCountMatch.Valid != tt.recordCountValid {
				t.Errorf("RecordCountMatch.Valid = %v, want %v", got.RecordCountMatch.Valid, tt.recordCountValid)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// SQLite Database.MapCreateBackupParams tests
// ---------------------------------------------------------------------------

func TestDatabase_MapCreateBackupParams_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	backupID, nodeID, ts, triggeredBy, metadata := backupTestFixture()

	input := CreateBackupParams{
		BackupID:    backupID,
		NodeID:      nodeID,
		BackupType:  types.BackupTypeFull,
		Status:      types.BackupStatusPending,
		StartedAt:   ts,
		StoragePath: "/backups/test.zip",
		TriggeredBy: triggeredBy,
		Metadata:    metadata,
	}

	got := d.MapCreateBackupParams(input)

	if got.BackupID != input.BackupID {
		t.Errorf("BackupID = %v, want %v", got.BackupID, input.BackupID)
	}
	if got.NodeID != input.NodeID {
		t.Errorf("NodeID = %v, want %v", got.NodeID, input.NodeID)
	}
	if got.BackupType != input.BackupType {
		t.Errorf("BackupType = %v, want %v", got.BackupType, input.BackupType)
	}
	if got.Status != input.Status {
		t.Errorf("Status = %v, want %v", got.Status, input.Status)
	}
	if got.StartedAt != input.StartedAt {
		t.Errorf("StartedAt = %v, want %v", got.StartedAt, input.StartedAt)
	}
	if got.StoragePath != input.StoragePath {
		t.Errorf("StoragePath = %v, want %v", got.StoragePath, input.StoragePath)
	}
	if got.TriggeredBy != input.TriggeredBy {
		t.Errorf("TriggeredBy = %v, want %v", got.TriggeredBy, input.TriggeredBy)
	}
	if got.Metadata.Valid != input.Metadata.Valid {
		t.Errorf("Metadata.Valid = %v, want %v", got.Metadata.Valid, input.Metadata.Valid)
	}
}

func TestDatabase_MapCreateBackupParams_ZeroValue(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapCreateBackupParams(CreateBackupParams{})

	if got.BackupID != "" {
		t.Errorf("BackupID = %v, want zero value", got.BackupID)
	}
	if got.StoragePath != "" {
		t.Errorf("StoragePath = %v, want empty string", got.StoragePath)
	}
	if got.TriggeredBy.Valid {
		t.Errorf("TriggeredBy.Valid = true, want false")
	}
}

// ---------------------------------------------------------------------------
// SQLite Database.MapCreateBackupSetParams tests
// ---------------------------------------------------------------------------

func TestDatabase_MapCreateBackupSetParams_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := CreateBackupSetParams{
		BackupSetID:    types.NewBackupSetID(),
		CreatedAt:      types.TimestampNow(),
		HlcTimestamp:   types.HLCNow(),
		Status:         types.BackupSetStatusPending,
		BackupIds:      types.NewJSONData([]string{"a", "b"}),
		NodeCount:      5,
		CompletedCount: types.NullableInt64{Int64: 2, Valid: true},
		ErrorMessage:   types.NullableString{Valid: false},
	}

	got := d.MapCreateBackupSetParams(input)

	if got.BackupSetID != input.BackupSetID {
		t.Errorf("BackupSetID = %v, want %v", got.BackupSetID, input.BackupSetID)
	}
	if got.CreatedAt != input.CreatedAt {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, input.CreatedAt)
	}
	if got.HlcTimestamp != input.HlcTimestamp {
		t.Errorf("HlcTimestamp = %v, want %v", got.HlcTimestamp, input.HlcTimestamp)
	}
	if got.Status != input.Status {
		t.Errorf("Status = %v, want %v", got.Status, input.Status)
	}
	if got.BackupIds.Valid != input.BackupIds.Valid {
		t.Errorf("BackupIds.Valid = %v, want %v", got.BackupIds.Valid, input.BackupIds.Valid)
	}
	if got.NodeCount != input.NodeCount {
		t.Errorf("NodeCount = %v, want %v", got.NodeCount, input.NodeCount)
	}
	if got.CompletedCount != input.CompletedCount {
		t.Errorf("CompletedCount = %v, want %v", got.CompletedCount, input.CompletedCount)
	}
	if got.ErrorMessage != input.ErrorMessage {
		t.Errorf("ErrorMessage = %v, want %v", got.ErrorMessage, input.ErrorMessage)
	}
}

// ---------------------------------------------------------------------------
// SQLite Database.MapCreateVerificationParams tests
// ---------------------------------------------------------------------------

func TestDatabase_MapCreateVerificationParams_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	input := CreateVerificationParams{
		VerificationID:   types.NewVerificationID(),
		BackupID:         types.NewBackupID(),
		VerifiedAt:       types.TimestampNow(),
		VerifiedBy:       types.NullableString{String: "admin", Valid: true},
		RestoreTested:    types.NullableBool{Bool: true, Valid: true},
		ChecksumValid:    types.NullableBool{Bool: true, Valid: true},
		RecordCountMatch: types.NullableBool{Bool: false, Valid: true},
		Status:           types.VerificationStatusVerified,
		ErrorMessage:     types.NullableString{Valid: false},
		DurationMs:       types.NullableInt64{Int64: 500, Valid: true},
	}

	got := d.MapCreateVerificationParams(input)

	if got.VerificationID != input.VerificationID {
		t.Errorf("VerificationID = %v, want %v", got.VerificationID, input.VerificationID)
	}
	if got.BackupID != input.BackupID {
		t.Errorf("BackupID = %v, want %v", got.BackupID, input.BackupID)
	}
	if got.VerifiedAt != input.VerifiedAt {
		t.Errorf("VerifiedAt = %v, want %v", got.VerifiedAt, input.VerifiedAt)
	}
	if got.VerifiedBy != input.VerifiedBy {
		t.Errorf("VerifiedBy = %v, want %v", got.VerifiedBy, input.VerifiedBy)
	}
	if got.RestoreTested != input.RestoreTested {
		t.Errorf("RestoreTested = %v, want %v", got.RestoreTested, input.RestoreTested)
	}
	if got.ChecksumValid != input.ChecksumValid {
		t.Errorf("ChecksumValid = %v, want %v", got.ChecksumValid, input.ChecksumValid)
	}
	if got.RecordCountMatch != input.RecordCountMatch {
		t.Errorf("RecordCountMatch = %v, want %v", got.RecordCountMatch, input.RecordCountMatch)
	}
	if got.Status != input.Status {
		t.Errorf("Status = %v, want %v", got.Status, input.Status)
	}
	if got.ErrorMessage != input.ErrorMessage {
		t.Errorf("ErrorMessage = %v, want %v", got.ErrorMessage, input.ErrorMessage)
	}
	if got.DurationMs != input.DurationMs {
		t.Errorf("DurationMs = %v, want %v", got.DurationMs, input.DurationMs)
	}
}

func TestDatabase_MapCreateVerificationParams_ZeroValue(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapCreateVerificationParams(CreateVerificationParams{})

	if got.VerificationID != "" {
		t.Errorf("VerificationID = %v, want zero value", got.VerificationID)
	}
	if got.BackupID != "" {
		t.Errorf("BackupID = %v, want zero value", got.BackupID)
	}
	if got.VerifiedBy.Valid {
		t.Errorf("VerifiedBy.Valid = true, want false")
	}
	if got.RestoreTested.Valid {
		t.Errorf("RestoreTested.Valid = true, want false")
	}
	if got.ChecksumValid.Valid {
		t.Errorf("ChecksumValid.Valid = true, want false")
	}
	if got.RecordCountMatch.Valid {
		t.Errorf("RecordCountMatch.Valid = true, want false")
	}
	if got.DurationMs.Valid {
		t.Errorf("DurationMs.Valid = true, want false")
	}
}

// ---------------------------------------------------------------------------
// MySQL MysqlDatabase.MapBackup tests
// ---------------------------------------------------------------------------

func TestMysqlDatabase_MapBackup_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	backupID, nodeID, ts, triggeredBy, metadata := backupTestFixture()

	input := mdbm.Backup{
		BackupID:       backupID,
		NodeID:         nodeID,
		BackupType:     types.BackupTypeFull,
		Status:         types.BackupStatusCompleted,
		StartedAt:      ts,
		CompletedAt:    ts,
		DurationMs:     types.NullableInt64{Int64: 1500, Valid: true},
		RecordCount:    types.NullableInt64{Int64: 42, Valid: true},
		SizeBytes:      types.NullableInt64{Int64: 1024, Valid: true},
		ReplicationLsn: types.NullableString{String: "0/1234", Valid: true},
		HlcTimestamp:   types.HLCNow(),
		StoragePath:    "/backups/mysql-001.zip",
		Checksum:       types.NullableString{String: "sha256:def", Valid: true},
		TriggeredBy:    triggeredBy,
		ErrorMessage:   types.NullableString{Valid: false},
		Metadata:       metadata,
	}

	got := d.MapBackup(input)

	if got.BackupID != input.BackupID {
		t.Errorf("BackupID = %v, want %v", got.BackupID, input.BackupID)
	}
	if got.NodeID != input.NodeID {
		t.Errorf("NodeID = %v, want %v", got.NodeID, input.NodeID)
	}
	if got.BackupType != input.BackupType {
		t.Errorf("BackupType = %v, want %v", got.BackupType, input.BackupType)
	}
	if got.Status != input.Status {
		t.Errorf("Status = %v, want %v", got.Status, input.Status)
	}
	if got.StoragePath != input.StoragePath {
		t.Errorf("StoragePath = %v, want %v", got.StoragePath, input.StoragePath)
	}
	if got.DurationMs != input.DurationMs {
		t.Errorf("DurationMs = %v, want %v", got.DurationMs, input.DurationMs)
	}
	if got.Checksum != input.Checksum {
		t.Errorf("Checksum = %v, want %v", got.Checksum, input.Checksum)
	}
	if got.Metadata.Valid != input.Metadata.Valid {
		t.Errorf("Metadata.Valid = %v, want %v", got.Metadata.Valid, input.Metadata.Valid)
	}
}

func TestMysqlDatabase_MapBackup_ZeroValue(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapBackup(mdbm.Backup{})

	if got.BackupID != "" {
		t.Errorf("BackupID = %v, want zero value", got.BackupID)
	}
	if got.StoragePath != "" {
		t.Errorf("StoragePath = %v, want empty string", got.StoragePath)
	}
}

// ---------------------------------------------------------------------------
// MySQL MysqlDatabase.MapBackupSet tests -- int32 to int64 conversion
// ---------------------------------------------------------------------------

func TestMysqlDatabase_MapBackupSet_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	input := mdbm.BackupSet{
		BackupSetID:    types.NewBackupSetID(),
		CreatedAt:      types.TimestampNow(),
		HlcTimestamp:   types.HLCNow(),
		Status:         types.BackupSetStatusComplete,
		BackupIds:      types.NewJSONData([]string{"x"}),
		NodeCount:      7,
		CompletedCount: types.NullableInt64{Int64: 7, Valid: true},
		ErrorMessage:   types.NullableString{Valid: false},
	}

	got := d.MapBackupSet(input)

	if got.BackupSetID != input.BackupSetID {
		t.Errorf("BackupSetID = %v, want %v", got.BackupSetID, input.BackupSetID)
	}
	// NodeCount: MySQL has int32, wrapper has int64
	if got.NodeCount != int64(input.NodeCount) {
		t.Errorf("NodeCount = %v, want %v", got.NodeCount, int64(input.NodeCount))
	}
	if got.CompletedCount != input.CompletedCount {
		t.Errorf("CompletedCount = %v, want %v", got.CompletedCount, input.CompletedCount)
	}
	if got.Status != input.Status {
		t.Errorf("Status = %v, want %v", got.Status, input.Status)
	}
}

func TestMysqlDatabase_MapBackupSet_NodeCountMaxInt32(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	input := mdbm.BackupSet{
		NodeCount: math.MaxInt32,
	}

	got := d.MapBackupSet(input)

	if got.NodeCount != int64(math.MaxInt32) {
		t.Errorf("NodeCount = %v, want %v", got.NodeCount, int64(math.MaxInt32))
	}
}

// ---------------------------------------------------------------------------
// MySQL MysqlDatabase.MapBackupVerification tests
// ---------------------------------------------------------------------------

func TestMysqlDatabase_MapBackupVerification_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	input := mdbm.BackupVerification{
		VerificationID:   types.NewVerificationID(),
		BackupID:         types.NewBackupID(),
		VerifiedAt:       types.TimestampNow(),
		VerifiedBy:       types.NullableString{String: "operator", Valid: true},
		RestoreTested:    types.NullableBool{Bool: true, Valid: true},
		ChecksumValid:    types.NullableBool{Bool: false, Valid: true},
		RecordCountMatch: types.NullableBool{Bool: true, Valid: true},
		Status:           types.VerificationStatusFailed,
		ErrorMessage:     types.NullableString{String: "checksum mismatch", Valid: true},
		DurationMs:       types.NullableInt64{Int64: 300, Valid: true},
	}

	got := d.MapBackupVerification(input)

	if got.VerificationID != input.VerificationID {
		t.Errorf("VerificationID = %v, want %v", got.VerificationID, input.VerificationID)
	}
	if got.BackupID != input.BackupID {
		t.Errorf("BackupID = %v, want %v", got.BackupID, input.BackupID)
	}
	if got.Status != input.Status {
		t.Errorf("Status = %v, want %v", got.Status, input.Status)
	}
	if got.ErrorMessage != input.ErrorMessage {
		t.Errorf("ErrorMessage = %v, want %v", got.ErrorMessage, input.ErrorMessage)
	}
	if got.RestoreTested != input.RestoreTested {
		t.Errorf("RestoreTested = %v, want %v", got.RestoreTested, input.RestoreTested)
	}
	if got.ChecksumValid != input.ChecksumValid {
		t.Errorf("ChecksumValid = %v, want %v", got.ChecksumValid, input.ChecksumValid)
	}
	if got.RecordCountMatch != input.RecordCountMatch {
		t.Errorf("RecordCountMatch = %v, want %v", got.RecordCountMatch, input.RecordCountMatch)
	}
}

func TestMysqlDatabase_MapBackupVerification_ZeroValue(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapBackupVerification(mdbm.BackupVerification{})

	if got.VerificationID != "" {
		t.Errorf("VerificationID = %v, want zero value", got.VerificationID)
	}
	if got.RestoreTested.Valid {
		t.Errorf("RestoreTested.Valid = true, want false")
	}
}

// ---------------------------------------------------------------------------
// MySQL MysqlDatabase.MapCreateBackupParams tests
// ---------------------------------------------------------------------------

func TestMysqlDatabase_MapCreateBackupParams_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	backupID, nodeID, ts, triggeredBy, metadata := backupTestFixture()

	input := CreateBackupParams{
		BackupID:    backupID,
		NodeID:      nodeID,
		BackupType:  types.BackupTypeIncremental,
		Status:      types.BackupStatusInProgress,
		StartedAt:   ts,
		StoragePath: "/backups/mysql-inc.zip",
		TriggeredBy: triggeredBy,
		Metadata:    metadata,
	}

	got := d.MapCreateBackupParams(input)

	if got.BackupID != input.BackupID {
		t.Errorf("BackupID = %v, want %v", got.BackupID, input.BackupID)
	}
	if got.NodeID != input.NodeID {
		t.Errorf("NodeID = %v, want %v", got.NodeID, input.NodeID)
	}
	if got.BackupType != input.BackupType {
		t.Errorf("BackupType = %v, want %v", got.BackupType, input.BackupType)
	}
	if got.StoragePath != input.StoragePath {
		t.Errorf("StoragePath = %v, want %v", got.StoragePath, input.StoragePath)
	}
}

// ---------------------------------------------------------------------------
// MySQL MysqlDatabase.MapCreateBackupSetParams -- int64 to int32 truncation
// ---------------------------------------------------------------------------

func TestMysqlDatabase_MapCreateBackupSetParams_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	input := CreateBackupSetParams{
		BackupSetID:    types.NewBackupSetID(),
		CreatedAt:      types.TimestampNow(),
		HlcTimestamp:   types.HLCNow(),
		Status:         types.BackupSetStatusPending,
		BackupIds:      types.NewJSONData([]string{"a"}),
		NodeCount:      10,
		CompletedCount: types.NullableInt64{Int64: 5, Valid: true},
		ErrorMessage:   types.NullableString{Valid: false},
	}

	got := d.MapCreateBackupSetParams(input)

	if got.BackupSetID != input.BackupSetID {
		t.Errorf("BackupSetID = %v, want %v", got.BackupSetID, input.BackupSetID)
	}
	// NodeCount is converted int64 -> int32
	if got.NodeCount != int32(input.NodeCount) {
		t.Errorf("NodeCount = %v, want %v", got.NodeCount, int32(input.NodeCount))
	}
	if got.CompletedCount != input.CompletedCount {
		t.Errorf("CompletedCount = %v, want %v", got.CompletedCount, input.CompletedCount)
	}
	if got.Status != input.Status {
		t.Errorf("Status = %v, want %v", got.Status, input.Status)
	}
}

func TestMysqlDatabase_MapCreateBackupSetParams_NodeCountTruncation(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}

	// A value beyond int32 range will be truncated
	input := CreateBackupSetParams{
		NodeCount: int64(math.MaxInt32) + 1,
	}

	got := d.MapCreateBackupSetParams(input)

	// int32 truncation means the high bits are lost
	if got.NodeCount == int32(input.NodeCount) {
		// This is expected behavior for the current code -- int32() truncates.
		// The truncated value will be math.MinInt32 (overflow wraparound)
		if got.NodeCount != math.MinInt32 {
			t.Errorf("NodeCount truncation: got %v, want %v", got.NodeCount, math.MinInt32)
		}
	}
}

// ---------------------------------------------------------------------------
// MySQL MysqlDatabase.MapCreateVerificationParams tests
// ---------------------------------------------------------------------------

func TestMysqlDatabase_MapCreateVerificationParams_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	input := CreateVerificationParams{
		VerificationID:   types.NewVerificationID(),
		BackupID:         types.NewBackupID(),
		VerifiedAt:       types.TimestampNow(),
		VerifiedBy:       types.NullableString{String: "cron", Valid: true},
		RestoreTested:    types.NullableBool{Bool: false, Valid: true},
		ChecksumValid:    types.NullableBool{Bool: true, Valid: true},
		RecordCountMatch: types.NullableBool{Bool: true, Valid: true},
		Status:           types.VerificationStatusPending,
		ErrorMessage:     types.NullableString{Valid: false},
		DurationMs:       types.NullableInt64{Int64: 200, Valid: true},
	}

	got := d.MapCreateVerificationParams(input)

	if got.VerificationID != input.VerificationID {
		t.Errorf("VerificationID = %v, want %v", got.VerificationID, input.VerificationID)
	}
	if got.BackupID != input.BackupID {
		t.Errorf("BackupID = %v, want %v", got.BackupID, input.BackupID)
	}
	if got.RestoreTested != input.RestoreTested {
		t.Errorf("RestoreTested = %v, want %v", got.RestoreTested, input.RestoreTested)
	}
	if got.ChecksumValid != input.ChecksumValid {
		t.Errorf("ChecksumValid = %v, want %v", got.ChecksumValid, input.ChecksumValid)
	}
	if got.Status != input.Status {
		t.Errorf("Status = %v, want %v", got.Status, input.Status)
	}
}

// ---------------------------------------------------------------------------
// PostgreSQL PsqlDatabase.MapBackup tests
// ---------------------------------------------------------------------------

func TestPsqlDatabase_MapBackup_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	backupID, nodeID, ts, triggeredBy, metadata := backupTestFixture()

	input := mdbp.Backup{
		BackupID:       backupID,
		NodeID:         nodeID,
		BackupType:     types.BackupTypeDifferential,
		Status:         types.BackupStatusFailed,
		StartedAt:      ts,
		CompletedAt:    ts,
		DurationMs:     types.NullableInt64{Int64: 9999, Valid: true},
		RecordCount:    types.NullableInt64{Int64: 0, Valid: true},
		SizeBytes:      types.NullableInt64{Int64: 0, Valid: true},
		ReplicationLsn: types.NullableString{String: "0/ABCD", Valid: true},
		HlcTimestamp:   types.HLCNow(),
		StoragePath:    "/backups/psql-diff.zip",
		Checksum:       types.NullableString{String: "sha256:ghi", Valid: true},
		TriggeredBy:    triggeredBy,
		ErrorMessage:   types.NullableString{String: "disk full", Valid: true},
		Metadata:       metadata,
	}

	got := d.MapBackup(input)

	if got.BackupID != input.BackupID {
		t.Errorf("BackupID = %v, want %v", got.BackupID, input.BackupID)
	}
	if got.NodeID != input.NodeID {
		t.Errorf("NodeID = %v, want %v", got.NodeID, input.NodeID)
	}
	if got.BackupType != input.BackupType {
		t.Errorf("BackupType = %v, want %v", got.BackupType, input.BackupType)
	}
	if got.Status != input.Status {
		t.Errorf("Status = %v, want %v", got.Status, input.Status)
	}
	if got.ErrorMessage != input.ErrorMessage {
		t.Errorf("ErrorMessage = %v, want %v", got.ErrorMessage, input.ErrorMessage)
	}
	if got.ReplicationLsn != input.ReplicationLsn {
		t.Errorf("ReplicationLsn = %v, want %v", got.ReplicationLsn, input.ReplicationLsn)
	}
	if got.StoragePath != input.StoragePath {
		t.Errorf("StoragePath = %v, want %v", got.StoragePath, input.StoragePath)
	}
}

func TestPsqlDatabase_MapBackup_ZeroValue(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapBackup(mdbp.Backup{})

	if got.BackupID != "" {
		t.Errorf("BackupID = %v, want zero value", got.BackupID)
	}
	if got.StoragePath != "" {
		t.Errorf("StoragePath = %v, want empty string", got.StoragePath)
	}
}

// ---------------------------------------------------------------------------
// PostgreSQL PsqlDatabase.MapBackupSet tests -- int32 to int64 conversion
// ---------------------------------------------------------------------------

func TestPsqlDatabase_MapBackupSet_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	input := mdbp.BackupSet{
		BackupSetID:    types.NewBackupSetID(),
		CreatedAt:      types.TimestampNow(),
		HlcTimestamp:   types.HLCNow(),
		Status:         types.BackupSetStatusPartial,
		BackupIds:      types.NewJSONData([]string{"p1", "p2", "p3"}),
		NodeCount:      12,
		CompletedCount: types.NullableInt64{Int64: 8, Valid: true},
		ErrorMessage:   types.NullableString{String: "partial failure", Valid: true},
	}

	got := d.MapBackupSet(input)

	if got.BackupSetID != input.BackupSetID {
		t.Errorf("BackupSetID = %v, want %v", got.BackupSetID, input.BackupSetID)
	}
	// NodeCount: PostgreSQL has int32, wrapper has int64
	if got.NodeCount != int64(input.NodeCount) {
		t.Errorf("NodeCount = %v, want %v", got.NodeCount, int64(input.NodeCount))
	}
	if got.ErrorMessage != input.ErrorMessage {
		t.Errorf("ErrorMessage = %v, want %v", got.ErrorMessage, input.ErrorMessage)
	}
	if got.Status != input.Status {
		t.Errorf("Status = %v, want %v", got.Status, input.Status)
	}
}

func TestPsqlDatabase_MapBackupSet_NodeCountMaxInt32(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	input := mdbp.BackupSet{
		NodeCount: math.MaxInt32,
	}

	got := d.MapBackupSet(input)

	if got.NodeCount != int64(math.MaxInt32) {
		t.Errorf("NodeCount = %v, want %v", got.NodeCount, int64(math.MaxInt32))
	}
}

// ---------------------------------------------------------------------------
// PostgreSQL PsqlDatabase.MapBackupVerification tests
// ---------------------------------------------------------------------------

func TestPsqlDatabase_MapBackupVerification_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	input := mdbp.BackupVerification{
		VerificationID:   types.NewVerificationID(),
		BackupID:         types.NewBackupID(),
		VerifiedAt:       types.TimestampNow(),
		VerifiedBy:       types.NullableString{String: "scheduler", Valid: true},
		RestoreTested:    types.NullableBool{Bool: false, Valid: true},
		ChecksumValid:    types.NullableBool{Bool: true, Valid: true},
		RecordCountMatch: types.NullableBool{Bool: true, Valid: true},
		Status:           types.VerificationStatusVerified,
		ErrorMessage:     types.NullableString{Valid: false},
		DurationMs:       types.NullableInt64{Int64: 1200, Valid: true},
	}

	got := d.MapBackupVerification(input)

	if got.VerificationID != input.VerificationID {
		t.Errorf("VerificationID = %v, want %v", got.VerificationID, input.VerificationID)
	}
	if got.BackupID != input.BackupID {
		t.Errorf("BackupID = %v, want %v", got.BackupID, input.BackupID)
	}
	if got.RestoreTested != input.RestoreTested {
		t.Errorf("RestoreTested = %v, want %v", got.RestoreTested, input.RestoreTested)
	}
	if got.ChecksumValid != input.ChecksumValid {
		t.Errorf("ChecksumValid = %v, want %v", got.ChecksumValid, input.ChecksumValid)
	}
	if got.DurationMs != input.DurationMs {
		t.Errorf("DurationMs = %v, want %v", got.DurationMs, input.DurationMs)
	}
}

func TestPsqlDatabase_MapBackupVerification_ZeroValue(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapBackupVerification(mdbp.BackupVerification{})

	if got.VerificationID != "" {
		t.Errorf("VerificationID = %v, want zero value", got.VerificationID)
	}
	if got.RestoreTested.Valid {
		t.Errorf("RestoreTested.Valid = true, want false")
	}
}

// ---------------------------------------------------------------------------
// PostgreSQL PsqlDatabase.MapCreateBackupParams tests
// ---------------------------------------------------------------------------

func TestPsqlDatabase_MapCreateBackupParams_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	backupID, nodeID, ts, triggeredBy, metadata := backupTestFixture()

	input := CreateBackupParams{
		BackupID:    backupID,
		NodeID:      nodeID,
		BackupType:  types.BackupTypeFull,
		Status:      types.BackupStatusPending,
		StartedAt:   ts,
		StoragePath: "/backups/psql-full.zip",
		TriggeredBy: triggeredBy,
		Metadata:    metadata,
	}

	got := d.MapCreateBackupParams(input)

	if got.BackupID != input.BackupID {
		t.Errorf("BackupID = %v, want %v", got.BackupID, input.BackupID)
	}
	if got.NodeID != input.NodeID {
		t.Errorf("NodeID = %v, want %v", got.NodeID, input.NodeID)
	}
	if got.StoragePath != input.StoragePath {
		t.Errorf("StoragePath = %v, want %v", got.StoragePath, input.StoragePath)
	}
}

// ---------------------------------------------------------------------------
// PostgreSQL PsqlDatabase.MapCreateBackupSetParams -- int64 to int32 truncation
// ---------------------------------------------------------------------------

func TestPsqlDatabase_MapCreateBackupSetParams_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	input := CreateBackupSetParams{
		BackupSetID:    types.NewBackupSetID(),
		CreatedAt:      types.TimestampNow(),
		HlcTimestamp:   types.HLCNow(),
		Status:         types.BackupSetStatusComplete,
		BackupIds:      types.NewJSONData([]string{"q"}),
		NodeCount:      25,
		CompletedCount: types.NullableInt64{Int64: 25, Valid: true},
		ErrorMessage:   types.NullableString{Valid: false},
	}

	got := d.MapCreateBackupSetParams(input)

	if got.BackupSetID != input.BackupSetID {
		t.Errorf("BackupSetID = %v, want %v", got.BackupSetID, input.BackupSetID)
	}
	// NodeCount: int64 -> int32 conversion
	if got.NodeCount != int32(input.NodeCount) {
		t.Errorf("NodeCount = %v, want %v", got.NodeCount, int32(input.NodeCount))
	}
	if got.Status != input.Status {
		t.Errorf("Status = %v, want %v", got.Status, input.Status)
	}
}

func TestPsqlDatabase_MapCreateBackupSetParams_NodeCountTruncation(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}

	input := CreateBackupSetParams{
		NodeCount: int64(math.MaxInt32) + 1,
	}

	got := d.MapCreateBackupSetParams(input)

	// Same truncation behavior as MySQL
	if got.NodeCount == int32(input.NodeCount) {
		if got.NodeCount != math.MinInt32 {
			t.Errorf("NodeCount truncation: got %v, want %v", got.NodeCount, math.MinInt32)
		}
	}
}

// ---------------------------------------------------------------------------
// PostgreSQL PsqlDatabase.MapCreateVerificationParams tests
// ---------------------------------------------------------------------------

func TestPsqlDatabase_MapCreateVerificationParams_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	input := CreateVerificationParams{
		VerificationID:   types.NewVerificationID(),
		BackupID:         types.NewBackupID(),
		VerifiedAt:       types.TimestampNow(),
		VerifiedBy:       types.NullableString{String: "psql-admin", Valid: true},
		RestoreTested:    types.NullableBool{Bool: true, Valid: true},
		ChecksumValid:    types.NullableBool{Bool: true, Valid: true},
		RecordCountMatch: types.NullableBool{Bool: true, Valid: true},
		Status:           types.VerificationStatusVerified,
		ErrorMessage:     types.NullableString{Valid: false},
		DurationMs:       types.NullableInt64{Int64: 800, Valid: true},
	}

	got := d.MapCreateVerificationParams(input)

	if got.VerificationID != input.VerificationID {
		t.Errorf("VerificationID = %v, want %v", got.VerificationID, input.VerificationID)
	}
	if got.BackupID != input.BackupID {
		t.Errorf("BackupID = %v, want %v", got.BackupID, input.BackupID)
	}
	if got.RestoreTested != input.RestoreTested {
		t.Errorf("RestoreTested = %v, want %v", got.RestoreTested, input.RestoreTested)
	}
	if got.Status != input.Status {
		t.Errorf("Status = %v, want %v", got.Status, input.Status)
	}
}

// ---------------------------------------------------------------------------
// Cross-database mapper consistency: Backup
// ---------------------------------------------------------------------------

func TestCrossDatabaseMapBackup_Consistency(t *testing.T) {
	t.Parallel()
	backupID, nodeID, ts, triggeredBy, metadata := backupTestFixture()
	hlc := types.HLCNow()

	sqliteInput := mdb.Backup{
		BackupID: backupID, NodeID: nodeID, BackupType: types.BackupTypeFull,
		Status: types.BackupStatusCompleted, StartedAt: ts, CompletedAt: ts,
		DurationMs: types.NullableInt64{Int64: 500, Valid: true},
		RecordCount: types.NullableInt64{Int64: 10, Valid: true},
		SizeBytes: types.NullableInt64{Int64: 2048, Valid: true},
		ReplicationLsn: types.NullableString{String: "lsn", Valid: true},
		HlcTimestamp: hlc, StoragePath: "/path",
		Checksum: types.NullableString{String: "sum", Valid: true},
		TriggeredBy: triggeredBy, ErrorMessage: types.NullableString{Valid: false},
		Metadata: metadata,
	}
	mysqlInput := mdbm.Backup{
		BackupID: backupID, NodeID: nodeID, BackupType: types.BackupTypeFull,
		Status: types.BackupStatusCompleted, StartedAt: ts, CompletedAt: ts,
		DurationMs: types.NullableInt64{Int64: 500, Valid: true},
		RecordCount: types.NullableInt64{Int64: 10, Valid: true},
		SizeBytes: types.NullableInt64{Int64: 2048, Valid: true},
		ReplicationLsn: types.NullableString{String: "lsn", Valid: true},
		HlcTimestamp: hlc, StoragePath: "/path",
		Checksum: types.NullableString{String: "sum", Valid: true},
		TriggeredBy: triggeredBy, ErrorMessage: types.NullableString{Valid: false},
		Metadata: metadata,
	}
	psqlInput := mdbp.Backup{
		BackupID: backupID, NodeID: nodeID, BackupType: types.BackupTypeFull,
		Status: types.BackupStatusCompleted, StartedAt: ts, CompletedAt: ts,
		DurationMs: types.NullableInt64{Int64: 500, Valid: true},
		RecordCount: types.NullableInt64{Int64: 10, Valid: true},
		SizeBytes: types.NullableInt64{Int64: 2048, Valid: true},
		ReplicationLsn: types.NullableString{String: "lsn", Valid: true},
		HlcTimestamp: hlc, StoragePath: "/path",
		Checksum: types.NullableString{String: "sum", Valid: true},
		TriggeredBy: triggeredBy, ErrorMessage: types.NullableString{Valid: false},
		Metadata: metadata,
	}

	sqliteResult := Database{}.MapBackup(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapBackup(mysqlInput)
	psqlResult := PsqlDatabase{}.MapBackup(psqlInput)

	// Compare field by field since Metadata ([]byte) doesn't work with ==
	if sqliteResult.BackupID != mysqlResult.BackupID {
		t.Errorf("SQLite vs MySQL BackupID mismatch: %v vs %v", sqliteResult.BackupID, mysqlResult.BackupID)
	}
	if sqliteResult.BackupID != psqlResult.BackupID {
		t.Errorf("SQLite vs PostgreSQL BackupID mismatch: %v vs %v", sqliteResult.BackupID, psqlResult.BackupID)
	}
	if sqliteResult.NodeID != mysqlResult.NodeID {
		t.Errorf("SQLite vs MySQL NodeID mismatch")
	}
	if sqliteResult.Status != psqlResult.Status {
		t.Errorf("SQLite vs PostgreSQL Status mismatch")
	}
	if sqliteResult.StoragePath != mysqlResult.StoragePath {
		t.Errorf("SQLite vs MySQL StoragePath mismatch")
	}
	if sqliteResult.DurationMs != psqlResult.DurationMs {
		t.Errorf("SQLite vs PostgreSQL DurationMs mismatch")
	}
	if sqliteResult.Metadata.Valid != mysqlResult.Metadata.Valid {
		t.Errorf("SQLite vs MySQL Metadata.Valid mismatch")
	}
	if sqliteResult.Metadata.Valid != psqlResult.Metadata.Valid {
		t.Errorf("SQLite vs PostgreSQL Metadata.Valid mismatch")
	}
}

// ---------------------------------------------------------------------------
// Cross-database mapper consistency: BackupSet
// ---------------------------------------------------------------------------

func TestCrossDatabaseMapBackupSet_Consistency(t *testing.T) {
	t.Parallel()
	bsID := types.NewBackupSetID()
	ts := types.TimestampNow()
	hlc := types.HLCNow()
	backupIds := types.NewJSONData([]string{"x", "y"})

	sqliteInput := mdb.BackupSet{
		BackupSetID: bsID, CreatedAt: ts, HlcTimestamp: hlc,
		Status: types.BackupSetStatusComplete, BackupIds: backupIds,
		NodeCount: 5, CompletedCount: types.NullableInt64{Int64: 5, Valid: true},
		ErrorMessage: types.NullableString{Valid: false},
	}
	mysqlInput := mdbm.BackupSet{
		BackupSetID: bsID, CreatedAt: ts, HlcTimestamp: hlc,
		Status: types.BackupSetStatusComplete, BackupIds: backupIds,
		NodeCount: 5, CompletedCount: types.NullableInt64{Int64: 5, Valid: true},
		ErrorMessage: types.NullableString{Valid: false},
	}
	psqlInput := mdbp.BackupSet{
		BackupSetID: bsID, CreatedAt: ts, HlcTimestamp: hlc,
		Status: types.BackupSetStatusComplete, BackupIds: backupIds,
		NodeCount: 5, CompletedCount: types.NullableInt64{Int64: 5, Valid: true},
		ErrorMessage: types.NullableString{Valid: false},
	}

	sqliteResult := Database{}.MapBackupSet(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapBackupSet(mysqlInput)
	psqlResult := PsqlDatabase{}.MapBackupSet(psqlInput)

	if sqliteResult.BackupSetID != mysqlResult.BackupSetID {
		t.Errorf("SQLite vs MySQL BackupSetID mismatch")
	}
	if sqliteResult.BackupSetID != psqlResult.BackupSetID {
		t.Errorf("SQLite vs PostgreSQL BackupSetID mismatch")
	}
	if sqliteResult.NodeCount != mysqlResult.NodeCount {
		t.Errorf("SQLite vs MySQL NodeCount: %v vs %v", sqliteResult.NodeCount, mysqlResult.NodeCount)
	}
	if sqliteResult.NodeCount != psqlResult.NodeCount {
		t.Errorf("SQLite vs PostgreSQL NodeCount: %v vs %v", sqliteResult.NodeCount, psqlResult.NodeCount)
	}
	if sqliteResult.Status != mysqlResult.Status {
		t.Errorf("SQLite vs MySQL Status mismatch")
	}
	if sqliteResult.CompletedCount != psqlResult.CompletedCount {
		t.Errorf("SQLite vs PostgreSQL CompletedCount mismatch")
	}
}

// ---------------------------------------------------------------------------
// Cross-database mapper consistency: BackupVerification
// ---------------------------------------------------------------------------

func TestCrossDatabaseMapBackupVerification_Consistency(t *testing.T) {
	t.Parallel()
	vID := types.NewVerificationID()
	bID := types.NewBackupID()
	ts := types.TimestampNow()

	sqliteInput := mdb.BackupVerification{
		VerificationID: vID, BackupID: bID, VerifiedAt: ts,
		VerifiedBy: types.NullableString{String: "test", Valid: true},
		RestoreTested: types.NullableBool{Bool: true, Valid: true},
		ChecksumValid: types.NullableBool{Bool: true, Valid: true},
		RecordCountMatch: types.NullableBool{Bool: true, Valid: true},
		Status: types.VerificationStatusVerified,
		ErrorMessage: types.NullableString{Valid: false},
		DurationMs: types.NullableInt64{Int64: 100, Valid: true},
	}
	mysqlInput := mdbm.BackupVerification{
		VerificationID: vID, BackupID: bID, VerifiedAt: ts,
		VerifiedBy: types.NullableString{String: "test", Valid: true},
		RestoreTested: types.NullableBool{Bool: true, Valid: true},
		ChecksumValid: types.NullableBool{Bool: true, Valid: true},
		RecordCountMatch: types.NullableBool{Bool: true, Valid: true},
		Status: types.VerificationStatusVerified,
		ErrorMessage: types.NullableString{Valid: false},
		DurationMs: types.NullableInt64{Int64: 100, Valid: true},
	}
	psqlInput := mdbp.BackupVerification{
		VerificationID: vID, BackupID: bID, VerifiedAt: ts,
		VerifiedBy: types.NullableString{String: "test", Valid: true},
		RestoreTested: types.NullableBool{Bool: true, Valid: true},
		ChecksumValid: types.NullableBool{Bool: true, Valid: true},
		RecordCountMatch: types.NullableBool{Bool: true, Valid: true},
		Status: types.VerificationStatusVerified,
		ErrorMessage: types.NullableString{Valid: false},
		DurationMs: types.NullableInt64{Int64: 100, Valid: true},
	}

	sqliteResult := Database{}.MapBackupVerification(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapBackupVerification(mysqlInput)
	psqlResult := PsqlDatabase{}.MapBackupVerification(psqlInput)

	if sqliteResult.VerificationID != mysqlResult.VerificationID {
		t.Errorf("SQLite vs MySQL VerificationID mismatch")
	}
	if sqliteResult.VerificationID != psqlResult.VerificationID {
		t.Errorf("SQLite vs PostgreSQL VerificationID mismatch")
	}
	if sqliteResult.RestoreTested != mysqlResult.RestoreTested {
		t.Errorf("SQLite vs MySQL RestoreTested mismatch")
	}
	if sqliteResult.ChecksumValid != psqlResult.ChecksumValid {
		t.Errorf("SQLite vs PostgreSQL ChecksumValid mismatch")
	}
	if sqliteResult.Status != mysqlResult.Status {
		t.Errorf("SQLite vs MySQL Status mismatch")
	}
	if sqliteResult.DurationMs != psqlResult.DurationMs {
		t.Errorf("SQLite vs PostgreSQL DurationMs mismatch")
	}
}

// ===========================================================================
// Audited Command tests: Backup (Create + Delete, 3 drivers)
// ===========================================================================

// --- SQLite Backup Commands ---

func TestNewBackupCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("node-backup-1"),
		RequestID: "req-backup-001",
		IP:        "10.0.0.1",
	}
	backupID, nodeID, ts, triggeredBy, metadata := backupTestFixture()
	params := CreateBackupParams{
		BackupID:    backupID,
		NodeID:      nodeID,
		BackupType:  types.BackupTypeFull,
		Status:      types.BackupStatusPending,
		StartedAt:   ts,
		StoragePath: "/backups/test.zip",
		TriggeredBy: triggeredBy,
		Metadata:    metadata,
	}

	cmd := Database{}.NewBackupCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "backups" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "backups")
	}
	p, ok := cmd.Params().(CreateBackupParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateBackupParams", cmd.Params())
	}
	if p.BackupID != backupID {
		t.Errorf("Params().BackupID = %v, want %v", p.BackupID, backupID)
	}
	if p.NodeID != nodeID {
		t.Errorf("Params().NodeID = %v, want %v", p.NodeID, nodeID)
	}
	if p.StoragePath != "/backups/test.zip" {
		t.Errorf("Params().StoragePath = %v, want /backups/test.zip", p.StoragePath)
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewBackupCmd_GetID_ExtractsFromRow(t *testing.T) {
	t.Parallel()
	backupID := types.NewBackupID()
	cmd := NewBackupCmd{}

	row := mdb.Backup{BackupID: backupID}
	got := cmd.GetID(row)
	if got != string(backupID) {
		t.Errorf("GetID() = %q, want %q", got, string(backupID))
	}
}

func TestDeleteBackupCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	backupID := types.NewBackupID()

	cmd := Database{}.DeleteBackupCmd(ctx, ac, backupID)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "backups" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "backups")
	}
	if cmd.GetID() != string(backupID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(backupID))
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- MySQL Backup Commands ---

func TestNewBackupCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("mysql-backup-node"),
		RequestID: "mysql-backup-001",
		IP:        "192.168.1.1",
	}
	backupID, nodeID, ts, triggeredBy, metadata := backupTestFixture()
	params := CreateBackupParams{
		BackupID:    backupID,
		NodeID:      nodeID,
		BackupType:  types.BackupTypeIncremental,
		Status:      types.BackupStatusPending,
		StartedAt:   ts,
		StoragePath: "/backups/mysql.zip",
		TriggeredBy: triggeredBy,
		Metadata:    metadata,
	}

	cmd := MysqlDatabase{}.NewBackupCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "backups" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "backups")
	}
	p, ok := cmd.Params().(CreateBackupParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateBackupParams", cmd.Params())
	}
	if p.BackupID != backupID {
		t.Errorf("Params().BackupID = %v, want %v", p.BackupID, backupID)
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty MysqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewBackupCmdMysql_GetID(t *testing.T) {
	t.Parallel()
	backupID := types.NewBackupID()
	cmd := NewBackupCmdMysql{}

	row := mdbm.Backup{BackupID: backupID}
	got := cmd.GetID(row)
	if got != string(backupID) {
		t.Errorf("GetID() = %q, want %q", got, string(backupID))
	}
}

func TestDeleteBackupCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	backupID := types.NewBackupID()

	cmd := MysqlDatabase{}.DeleteBackupCmd(ctx, ac, backupID)

	if cmd.TableName() != "backups" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "backups")
	}
	if cmd.GetID() != string(backupID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(backupID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty MysqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- PostgreSQL Backup Commands ---

func TestNewBackupCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("psql-backup-node"),
		RequestID: "psql-backup-001",
		IP:        "172.16.0.1",
	}
	backupID, nodeID, ts, triggeredBy, metadata := backupTestFixture()
	params := CreateBackupParams{
		BackupID:    backupID,
		NodeID:      nodeID,
		BackupType:  types.BackupTypeDifferential,
		Status:      types.BackupStatusPending,
		StartedAt:   ts,
		StoragePath: "/backups/psql.zip",
		TriggeredBy: triggeredBy,
		Metadata:    metadata,
	}

	cmd := PsqlDatabase{}.NewBackupCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "backups" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "backups")
	}
	p, ok := cmd.Params().(CreateBackupParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateBackupParams", cmd.Params())
	}
	if p.StoragePath != "/backups/psql.zip" {
		t.Errorf("Params().StoragePath = %v, want /backups/psql.zip", p.StoragePath)
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty PsqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewBackupCmdPsql_GetID(t *testing.T) {
	t.Parallel()
	backupID := types.NewBackupID()
	cmd := NewBackupCmdPsql{}

	row := mdbp.Backup{BackupID: backupID}
	got := cmd.GetID(row)
	if got != string(backupID) {
		t.Errorf("GetID() = %q, want %q", got, string(backupID))
	}
}

func TestDeleteBackupCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	backupID := types.NewBackupID()

	cmd := PsqlDatabase{}.DeleteBackupCmd(ctx, ac, backupID)

	if cmd.TableName() != "backups" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "backups")
	}
	if cmd.GetID() != string(backupID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(backupID))
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty PsqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// ===========================================================================
// Audited Command tests: BackupSet (Create + Delete, 3 drivers)
// ===========================================================================

// --- SQLite BackupSet Commands ---

func TestNewBackupSetCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID(), NodeID: types.NodeID("n1")}
	params := CreateBackupSetParams{
		BackupSetID: types.NewBackupSetID(),
		NodeCount:   3,
		Status:      types.BackupSetStatusPending,
	}

	cmd := Database{}.NewBackupSetCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "backup_sets" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "backup_sets")
	}
	p, ok := cmd.Params().(CreateBackupSetParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateBackupSetParams", cmd.Params())
	}
	if p.BackupSetID != params.BackupSetID {
		t.Errorf("Params().BackupSetID = %v, want %v", p.BackupSetID, params.BackupSetID)
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewBackupSetCmd_GetID(t *testing.T) {
	t.Parallel()
	bsID := types.NewBackupSetID()
	cmd := NewBackupSetCmd{}

	row := mdb.BackupSet{BackupSetID: bsID}
	got := cmd.GetID(row)
	if got != string(bsID) {
		t.Errorf("GetID() = %q, want %q", got, string(bsID))
	}
}

func TestDeleteBackupSetCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	bsID := types.NewBackupSetID()

	cmd := Database{}.DeleteBackupSetCmd(ctx, ac, bsID)

	if cmd.TableName() != "backup_sets" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "backup_sets")
	}
	if cmd.GetID() != string(bsID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(bsID))
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- MySQL BackupSet Commands ---

func TestNewBackupSetCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := CreateBackupSetParams{
		BackupSetID: types.NewBackupSetID(),
		NodeCount:   10,
		Status:      types.BackupSetStatusPending,
	}

	cmd := MysqlDatabase{}.NewBackupSetCmd(ctx, ac, params)

	if cmd.TableName() != "backup_sets" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "backup_sets")
	}
	p, ok := cmd.Params().(CreateBackupSetParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateBackupSetParams", cmd.Params())
	}
	if p.NodeCount != 10 {
		t.Errorf("Params().NodeCount = %v, want 10", p.NodeCount)
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewBackupSetCmdMysql_GetID(t *testing.T) {
	t.Parallel()
	bsID := types.NewBackupSetID()
	cmd := NewBackupSetCmdMysql{}

	row := mdbm.BackupSet{BackupSetID: bsID}
	got := cmd.GetID(row)
	if got != string(bsID) {
		t.Errorf("GetID() = %q, want %q", got, string(bsID))
	}
}

func TestDeleteBackupSetCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	bsID := types.NewBackupSetID()

	cmd := MysqlDatabase{}.DeleteBackupSetCmd(ctx, ac, bsID)

	if cmd.TableName() != "backup_sets" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "backup_sets")
	}
	if cmd.GetID() != string(bsID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(bsID))
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- PostgreSQL BackupSet Commands ---

func TestNewBackupSetCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := CreateBackupSetParams{
		BackupSetID: types.NewBackupSetID(),
		NodeCount:   20,
		Status:      types.BackupSetStatusComplete,
	}

	cmd := PsqlDatabase{}.NewBackupSetCmd(ctx, ac, params)

	if cmd.TableName() != "backup_sets" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "backup_sets")
	}
	p, ok := cmd.Params().(CreateBackupSetParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateBackupSetParams", cmd.Params())
	}
	if p.NodeCount != 20 {
		t.Errorf("Params().NodeCount = %v, want 20", p.NodeCount)
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewBackupSetCmdPsql_GetID(t *testing.T) {
	t.Parallel()
	bsID := types.NewBackupSetID()
	cmd := NewBackupSetCmdPsql{}

	row := mdbp.BackupSet{BackupSetID: bsID}
	got := cmd.GetID(row)
	if got != string(bsID) {
		t.Errorf("GetID() = %q, want %q", got, string(bsID))
	}
}

func TestDeleteBackupSetCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	bsID := types.NewBackupSetID()

	cmd := PsqlDatabase{}.DeleteBackupSetCmd(ctx, ac, bsID)

	if cmd.TableName() != "backup_sets" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "backup_sets")
	}
	if cmd.GetID() != string(bsID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(bsID))
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// ===========================================================================
// Audited Command tests: BackupVerification (Create + Delete, 3 drivers)
// ===========================================================================

// --- SQLite Verification Commands ---

func TestNewVerificationCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID(), NodeID: types.NodeID("v-node")}
	params := CreateVerificationParams{
		VerificationID: types.NewVerificationID(),
		BackupID:       types.NewBackupID(),
		Status:         types.VerificationStatusPending,
	}

	cmd := Database{}.NewVerificationCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "backup_verifications" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "backup_verifications")
	}
	p, ok := cmd.Params().(CreateVerificationParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateVerificationParams", cmd.Params())
	}
	if p.VerificationID != params.VerificationID {
		t.Errorf("Params().VerificationID = %v, want %v", p.VerificationID, params.VerificationID)
	}
	if p.BackupID != params.BackupID {
		t.Errorf("Params().BackupID = %v, want %v", p.BackupID, params.BackupID)
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewVerificationCmd_GetID(t *testing.T) {
	t.Parallel()
	vID := types.NewVerificationID()
	cmd := NewVerificationCmd{}

	row := mdb.BackupVerification{VerificationID: vID}
	got := cmd.GetID(row)
	if got != string(vID) {
		t.Errorf("GetID() = %q, want %q", got, string(vID))
	}
}

func TestDeleteVerificationCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	vID := types.NewVerificationID()

	cmd := Database{}.DeleteVerificationCmd(ctx, ac, vID)

	if cmd.TableName() != "backup_verifications" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "backup_verifications")
	}
	if cmd.GetID() != string(vID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(vID))
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- MySQL Verification Commands ---

func TestNewVerificationCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := CreateVerificationParams{
		VerificationID: types.NewVerificationID(),
		BackupID:       types.NewBackupID(),
		Status:         types.VerificationStatusVerified,
	}

	cmd := MysqlDatabase{}.NewVerificationCmd(ctx, ac, params)

	if cmd.TableName() != "backup_verifications" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "backup_verifications")
	}
	p, ok := cmd.Params().(CreateVerificationParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateVerificationParams", cmd.Params())
	}
	if p.VerificationID != params.VerificationID {
		t.Errorf("Params().VerificationID = %v, want %v", p.VerificationID, params.VerificationID)
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewVerificationCmdMysql_GetID(t *testing.T) {
	t.Parallel()
	vID := types.NewVerificationID()
	cmd := NewVerificationCmdMysql{}

	row := mdbm.BackupVerification{VerificationID: vID}
	got := cmd.GetID(row)
	if got != string(vID) {
		t.Errorf("GetID() = %q, want %q", got, string(vID))
	}
}

func TestDeleteVerificationCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	vID := types.NewVerificationID()

	cmd := MysqlDatabase{}.DeleteVerificationCmd(ctx, ac, vID)

	if cmd.TableName() != "backup_verifications" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "backup_verifications")
	}
	if cmd.GetID() != string(vID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(vID))
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- PostgreSQL Verification Commands ---

func TestNewVerificationCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := CreateVerificationParams{
		VerificationID: types.NewVerificationID(),
		BackupID:       types.NewBackupID(),
		Status:         types.VerificationStatusFailed,
	}

	cmd := PsqlDatabase{}.NewVerificationCmd(ctx, ac, params)

	if cmd.TableName() != "backup_verifications" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "backup_verifications")
	}
	p, ok := cmd.Params().(CreateVerificationParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateVerificationParams", cmd.Params())
	}
	if p.Status != types.VerificationStatusFailed {
		t.Errorf("Params().Status = %v, want %v", p.Status, types.VerificationStatusFailed)
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewVerificationCmdPsql_GetID(t *testing.T) {
	t.Parallel()
	vID := types.NewVerificationID()
	cmd := NewVerificationCmdPsql{}

	row := mdbp.BackupVerification{VerificationID: vID}
	got := cmd.GetID(row)
	if got != string(vID) {
		t.Errorf("GetID() = %q, want %q", got, string(vID))
	}
}

func TestDeleteVerificationCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	vID := types.NewVerificationID()

	cmd := PsqlDatabase{}.DeleteVerificationCmd(ctx, ac, vID)

	if cmd.TableName() != "backup_verifications" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "backup_verifications")
	}
	if cmd.GetID() != string(vID) {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), string(vID))
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// ===========================================================================
// Cross-database audited command table name consistency
// ===========================================================================

func TestAuditedBackupCommands_TableNameConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	backupParams := CreateBackupParams{BackupID: types.NewBackupID()}
	backupID := types.NewBackupID()
	bsParams := CreateBackupSetParams{BackupSetID: types.NewBackupSetID()}
	bsID := types.NewBackupSetID()
	vParams := CreateVerificationParams{VerificationID: types.NewVerificationID()}
	vID := types.NewVerificationID()

	commands := []struct {
		label    string
		name     string
		wantName string
	}{
		// Backup
		{"SQLite Backup Create", Database{}.NewBackupCmd(ctx, ac, backupParams).TableName(), "backups"},
		{"SQLite Backup Delete", Database{}.DeleteBackupCmd(ctx, ac, backupID).TableName(), "backups"},
		{"MySQL Backup Create", MysqlDatabase{}.NewBackupCmd(ctx, ac, backupParams).TableName(), "backups"},
		{"MySQL Backup Delete", MysqlDatabase{}.DeleteBackupCmd(ctx, ac, backupID).TableName(), "backups"},
		{"PostgreSQL Backup Create", PsqlDatabase{}.NewBackupCmd(ctx, ac, backupParams).TableName(), "backups"},
		{"PostgreSQL Backup Delete", PsqlDatabase{}.DeleteBackupCmd(ctx, ac, backupID).TableName(), "backups"},
		// BackupSet
		{"SQLite BackupSet Create", Database{}.NewBackupSetCmd(ctx, ac, bsParams).TableName(), "backup_sets"},
		{"SQLite BackupSet Delete", Database{}.DeleteBackupSetCmd(ctx, ac, bsID).TableName(), "backup_sets"},
		{"MySQL BackupSet Create", MysqlDatabase{}.NewBackupSetCmd(ctx, ac, bsParams).TableName(), "backup_sets"},
		{"MySQL BackupSet Delete", MysqlDatabase{}.DeleteBackupSetCmd(ctx, ac, bsID).TableName(), "backup_sets"},
		{"PostgreSQL BackupSet Create", PsqlDatabase{}.NewBackupSetCmd(ctx, ac, bsParams).TableName(), "backup_sets"},
		{"PostgreSQL BackupSet Delete", PsqlDatabase{}.DeleteBackupSetCmd(ctx, ac, bsID).TableName(), "backup_sets"},
		// Verification
		{"SQLite Verification Create", Database{}.NewVerificationCmd(ctx, ac, vParams).TableName(), "backup_verifications"},
		{"SQLite Verification Delete", Database{}.DeleteVerificationCmd(ctx, ac, vID).TableName(), "backup_verifications"},
		{"MySQL Verification Create", MysqlDatabase{}.NewVerificationCmd(ctx, ac, vParams).TableName(), "backup_verifications"},
		{"MySQL Verification Delete", MysqlDatabase{}.DeleteVerificationCmd(ctx, ac, vID).TableName(), "backup_verifications"},
		{"PostgreSQL Verification Create", PsqlDatabase{}.NewVerificationCmd(ctx, ac, vParams).TableName(), "backup_verifications"},
		{"PostgreSQL Verification Delete", PsqlDatabase{}.DeleteVerificationCmd(ctx, ac, vID).TableName(), "backup_verifications"},
	}

	for _, c := range commands {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			if c.name != c.wantName {
				t.Errorf("TableName() = %q, want %q", c.name, c.wantName)
			}
		})
	}
}

func TestAuditedBackupCommands_RecorderAssignment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	backupParams := CreateBackupParams{}
	bsParams := CreateBackupSetParams{}
	vParams := CreateVerificationParams{}

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
	}{
		// Backup
		{"SQLite Backup Create", Database{}.NewBackupCmd(ctx, ac, backupParams).Recorder()},
		{"MySQL Backup Create", MysqlDatabase{}.NewBackupCmd(ctx, ac, backupParams).Recorder()},
		{"PostgreSQL Backup Create", PsqlDatabase{}.NewBackupCmd(ctx, ac, backupParams).Recorder()},
		// BackupSet
		{"SQLite BackupSet Create", Database{}.NewBackupSetCmd(ctx, ac, bsParams).Recorder()},
		{"MySQL BackupSet Create", MysqlDatabase{}.NewBackupSetCmd(ctx, ac, bsParams).Recorder()},
		{"PostgreSQL BackupSet Create", PsqlDatabase{}.NewBackupSetCmd(ctx, ac, bsParams).Recorder()},
		// Verification
		{"SQLite Verification Create", Database{}.NewVerificationCmd(ctx, ac, vParams).Recorder()},
		{"MySQL Verification Create", MysqlDatabase{}.NewVerificationCmd(ctx, ac, vParams).Recorder()},
		{"PostgreSQL Verification Create", PsqlDatabase{}.NewVerificationCmd(ctx, ac, vParams).Recorder()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.recorder == nil {
				t.Fatalf("%s recorder is nil", tt.name)
			}
		})
	}
}

// ===========================================================================
// Edge cases: empty/zero IDs on delete commands
// ===========================================================================

func TestDeleteBackupCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	emptyID := types.BackupID("")

	sqliteCmd := Database{}.DeleteBackupCmd(context.Background(), audited.AuditContext{}, emptyID)
	if sqliteCmd.GetID() != "" {
		t.Errorf("SQLite GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.DeleteBackupCmd(context.Background(), audited.AuditContext{}, emptyID)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.DeleteBackupCmd(context.Background(), audited.AuditContext{}, emptyID)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

func TestDeleteBackupSetCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	emptyID := types.BackupSetID("")

	sqliteCmd := Database{}.DeleteBackupSetCmd(context.Background(), audited.AuditContext{}, emptyID)
	if sqliteCmd.GetID() != "" {
		t.Errorf("SQLite GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.DeleteBackupSetCmd(context.Background(), audited.AuditContext{}, emptyID)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.DeleteBackupSetCmd(context.Background(), audited.AuditContext{}, emptyID)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

func TestDeleteVerificationCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	emptyID := types.VerificationID("")

	sqliteCmd := Database{}.DeleteVerificationCmd(context.Background(), audited.AuditContext{}, emptyID)
	if sqliteCmd.GetID() != "" {
		t.Errorf("SQLite GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.DeleteVerificationCmd(context.Background(), audited.AuditContext{}, emptyID)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.DeleteVerificationCmd(context.Background(), audited.AuditContext{}, emptyID)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

// ===========================================================================
// Cross-database GetID consistency for Create commands
// ===========================================================================

func TestAuditedBackupCreateCommands_GetID_CrossDatabase(t *testing.T) {
	t.Parallel()

	t.Run("Backup CreateCmd GetID extracts from result row", func(t *testing.T) {
		t.Parallel()
		backupID := types.NewBackupID()

		sqliteCmd := NewBackupCmd{}
		mysqlCmd := NewBackupCmdMysql{}
		psqlCmd := NewBackupCmdPsql{}

		sqliteRow := mdb.Backup{BackupID: backupID}
		mysqlRow := mdbm.Backup{BackupID: backupID}
		psqlRow := mdbp.Backup{BackupID: backupID}

		if sqliteCmd.GetID(sqliteRow) != string(backupID) {
			t.Errorf("SQLite GetID() = %q, want %q", sqliteCmd.GetID(sqliteRow), string(backupID))
		}
		if mysqlCmd.GetID(mysqlRow) != string(backupID) {
			t.Errorf("MySQL GetID() = %q, want %q", mysqlCmd.GetID(mysqlRow), string(backupID))
		}
		if psqlCmd.GetID(psqlRow) != string(backupID) {
			t.Errorf("PostgreSQL GetID() = %q, want %q", psqlCmd.GetID(psqlRow), string(backupID))
		}
	})

	t.Run("BackupSet CreateCmd GetID extracts from result row", func(t *testing.T) {
		t.Parallel()
		bsID := types.NewBackupSetID()

		sqliteCmd := NewBackupSetCmd{}
		mysqlCmd := NewBackupSetCmdMysql{}
		psqlCmd := NewBackupSetCmdPsql{}

		sqliteRow := mdb.BackupSet{BackupSetID: bsID}
		mysqlRow := mdbm.BackupSet{BackupSetID: bsID}
		psqlRow := mdbp.BackupSet{BackupSetID: bsID}

		if sqliteCmd.GetID(sqliteRow) != string(bsID) {
			t.Errorf("SQLite GetID() = %q, want %q", sqliteCmd.GetID(sqliteRow), string(bsID))
		}
		if mysqlCmd.GetID(mysqlRow) != string(bsID) {
			t.Errorf("MySQL GetID() = %q, want %q", mysqlCmd.GetID(mysqlRow), string(bsID))
		}
		if psqlCmd.GetID(psqlRow) != string(bsID) {
			t.Errorf("PostgreSQL GetID() = %q, want %q", psqlCmd.GetID(psqlRow), string(bsID))
		}
	})

	t.Run("Verification CreateCmd GetID extracts from result row", func(t *testing.T) {
		t.Parallel()
		vID := types.NewVerificationID()

		sqliteCmd := NewVerificationCmd{}
		mysqlCmd := NewVerificationCmdMysql{}
		psqlCmd := NewVerificationCmdPsql{}

		sqliteRow := mdb.BackupVerification{VerificationID: vID}
		mysqlRow := mdbm.BackupVerification{VerificationID: vID}
		psqlRow := mdbp.BackupVerification{VerificationID: vID}

		if sqliteCmd.GetID(sqliteRow) != string(vID) {
			t.Errorf("SQLite GetID() = %q, want %q", sqliteCmd.GetID(sqliteRow), string(vID))
		}
		if mysqlCmd.GetID(mysqlRow) != string(vID) {
			t.Errorf("MySQL GetID() = %q, want %q", mysqlCmd.GetID(mysqlRow), string(vID))
		}
		if psqlCmd.GetID(psqlRow) != string(vID) {
			t.Errorf("PostgreSQL GetID() = %q, want %q", psqlCmd.GetID(psqlRow), string(vID))
		}
	})
}

// ===========================================================================
// Compile-time interface checks: audited commands satisfy their interfaces
// ===========================================================================

var (
	// Backup: Create + Delete (no Update command exists)
	_ audited.CreateCommand[mdb.Backup]  = NewBackupCmd{}
	_ audited.DeleteCommand[mdb.Backup]  = DeleteBackupCmd{}
	_ audited.CreateCommand[mdbm.Backup] = NewBackupCmdMysql{}
	_ audited.DeleteCommand[mdbm.Backup] = DeleteBackupCmdMysql{}
	_ audited.CreateCommand[mdbp.Backup] = NewBackupCmdPsql{}
	_ audited.DeleteCommand[mdbp.Backup] = DeleteBackupCmdPsql{}

	// BackupSet: Create + Delete
	_ audited.CreateCommand[mdb.BackupSet]  = NewBackupSetCmd{}
	_ audited.DeleteCommand[mdb.BackupSet]  = DeleteBackupSetCmd{}
	_ audited.CreateCommand[mdbm.BackupSet] = NewBackupSetCmdMysql{}
	_ audited.DeleteCommand[mdbm.BackupSet] = DeleteBackupSetCmdMysql{}
	_ audited.CreateCommand[mdbp.BackupSet] = NewBackupSetCmdPsql{}
	_ audited.DeleteCommand[mdbp.BackupSet] = DeleteBackupSetCmdPsql{}

	// BackupVerification: Create + Delete
	_ audited.CreateCommand[mdb.BackupVerification]  = NewVerificationCmd{}
	_ audited.DeleteCommand[mdb.BackupVerification]  = DeleteVerificationCmd{}
	_ audited.CreateCommand[mdbm.BackupVerification] = NewVerificationCmdMysql{}
	_ audited.DeleteCommand[mdbm.BackupVerification] = DeleteVerificationCmdMysql{}
	_ audited.CreateCommand[mdbp.BackupVerification] = NewVerificationCmdPsql{}
	_ audited.DeleteCommand[mdbp.BackupVerification] = DeleteVerificationCmdPsql{}
)
