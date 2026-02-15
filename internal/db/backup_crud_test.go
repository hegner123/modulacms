// Integration tests for the backup, backup_set, and backup_verification
// entity CRUD lifecycles. Uses testIntegrationDB (Tier 0: no FK dependencies).
//
// Backup methods are NON-audited (no ctx/ac parameters on mutations).
//
// NOTE: The backup_sets and backup_verifications tables are NOT created by
// CreateAllTables (only the backups table is). Tests for those entities
// create the tables via raw DDL.
package db

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

// backupSetsDDL is the DDL for the backup_sets table (not included in CreateAllTables).
const backupSetsDDL = `CREATE TABLE IF NOT EXISTS backup_sets (
    backup_set_id    TEXT PRIMARY KEY CHECK (length(backup_set_id) = 26),
    created_at       TEXT NOT NULL,
    hlc_timestamp    INTEGER NOT NULL,
    status           TEXT NOT NULL CHECK (status IN ('pending', 'complete', 'partial')),
    backup_ids       TEXT NOT NULL,
    node_count       INTEGER NOT NULL,
    completed_count  INTEGER DEFAULT 0,
    error_message    TEXT
);`

// backupVerificationsDDL is the DDL for the backup_verifications table.
const backupVerificationsDDL = `CREATE TABLE IF NOT EXISTS backup_verifications (
    verification_id  TEXT PRIMARY KEY CHECK (length(verification_id) = 26),
    backup_id        TEXT NOT NULL REFERENCES backups(backup_id),
    verified_at      TEXT NOT NULL,
    verified_by      TEXT,
    restore_tested   INTEGER DEFAULT 0,
    checksum_valid   INTEGER DEFAULT 0,
    record_count_match INTEGER DEFAULT 0,
    status           TEXT NOT NULL CHECK (status IN ('pending', 'verified', 'failed')),
    error_message    TEXT,
    duration_ms      INTEGER
);`

func TestDatabase_CRUD_Backup(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)

	nodeID := types.NodeID(d.Config.Node_ID)
	now := types.TimestampNow()

	// --- Count: starts at zero ---
	count, err := d.CountBackups()
	if err != nil {
		t.Fatalf("initial CountBackups: %v", err)
	}
	if *count != 0 {
		t.Fatalf("initial CountBackups = %d, want 0", *count)
	}

	// --- Create ---
	backupID := types.NewBackupID()
	created, err := d.CreateBackup(CreateBackupParams{
		BackupID:    backupID,
		NodeID:      nodeID,
		BackupType:  types.BackupTypeFull,
		Status:      types.BackupStatusCompleted,
		StartedAt:   now,
		StoragePath: "/backups/test-001.zip",
		TriggeredBy: types.NewNullableString("test-runner"),
		Metadata:    types.NewJSONData(nil),
	})
	if err != nil {
		t.Fatalf("CreateBackup: %v", err)
	}
	if created == nil {
		t.Fatal("CreateBackup returned nil")
	}
	if created.BackupID != backupID {
		t.Errorf("BackupID = %v, want %v", created.BackupID, backupID)
	}
	if created.NodeID != nodeID {
		t.Errorf("NodeID = %v, want %v", created.NodeID, nodeID)
	}
	if created.BackupType != types.BackupTypeFull {
		t.Errorf("BackupType = %v, want %v", created.BackupType, types.BackupTypeFull)
	}
	if created.Status != types.BackupStatusCompleted {
		t.Errorf("Status = %v, want %v", created.Status, types.BackupStatusCompleted)
	}
	if created.StoragePath != "/backups/test-001.zip" {
		t.Errorf("StoragePath = %q, want %q", created.StoragePath, "/backups/test-001.zip")
	}

	// --- Get ---
	got, err := d.GetBackup(backupID)
	if err != nil {
		t.Fatalf("GetBackup: %v", err)
	}
	if got == nil {
		t.Fatal("GetBackup returned nil")
	}
	if got.BackupID != backupID {
		t.Errorf("GetBackup BackupID = %v, want %v", got.BackupID, backupID)
	}
	if got.StoragePath != "/backups/test-001.zip" {
		t.Errorf("GetBackup StoragePath = %q, want %q", got.StoragePath, "/backups/test-001.zip")
	}

	// --- GetLatestBackup ---
	latest, err := d.GetLatestBackup(nodeID)
	if err != nil {
		t.Fatalf("GetLatestBackup: %v", err)
	}
	if latest == nil {
		t.Fatal("GetLatestBackup returned nil")
	}
	if latest.BackupID != backupID {
		t.Errorf("GetLatestBackup BackupID = %v, want %v", latest.BackupID, backupID)
	}

	// --- ListBackups ---
	list, err := d.ListBackups(ListBackupsParams{Limit: 100, Offset: 0})
	if err != nil {
		t.Fatalf("ListBackups: %v", err)
	}
	if list == nil {
		t.Fatal("ListBackups returned nil")
	}
	if len(*list) != 1 {
		t.Fatalf("ListBackups len = %d, want 1", len(*list))
	}

	// --- Count: now 1 ---
	count, err = d.CountBackups()
	if err != nil {
		t.Fatalf("CountBackups after create: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountBackups after create = %d, want 1", *count)
	}

	// --- UpdateBackupStatus ---
	completedAt := types.TimestampNow()
	err = d.UpdateBackupStatus(UpdateBackupStatusParams{
		Status:       types.BackupStatusFailed,
		CompletedAt:  completedAt,
		DurationMs:   types.NullableInt64{Int64: 5000, Valid: true},
		RecordCount:  types.NullableInt64{Int64: 150, Valid: true},
		SizeBytes:    types.NullableInt64{Int64: 1024000, Valid: true},
		Checksum:     types.NewNullableString("abc123"),
		ErrorMessage: types.NewNullableString("simulated failure"),
		BackupID:     backupID,
	})
	if err != nil {
		t.Fatalf("UpdateBackupStatus: %v", err)
	}

	// --- Get after update ---
	updated, err := d.GetBackup(backupID)
	if err != nil {
		t.Fatalf("GetBackup after update: %v", err)
	}
	if updated.Status != types.BackupStatusFailed {
		t.Errorf("updated Status = %v, want %v", updated.Status, types.BackupStatusFailed)
	}
	if !updated.DurationMs.Valid || updated.DurationMs.Int64 != 5000 {
		t.Errorf("updated DurationMs = %v, want valid 5000", updated.DurationMs)
	}
	if !updated.RecordCount.Valid || updated.RecordCount.Int64 != 150 {
		t.Errorf("updated RecordCount = %v, want valid 150", updated.RecordCount)
	}
	if !updated.SizeBytes.Valid || updated.SizeBytes.Int64 != 1024000 {
		t.Errorf("updated SizeBytes = %v, want valid 1024000", updated.SizeBytes)
	}
	if !updated.Checksum.Valid || updated.Checksum.String != "abc123" {
		t.Errorf("updated Checksum = %v, want valid 'abc123'", updated.Checksum)
	}

	// --- Delete ---
	err = d.DeleteBackup(backupID)
	if err != nil {
		t.Fatalf("DeleteBackup: %v", err)
	}

	// --- Get after delete: expect error ---
	_, err = d.GetBackup(backupID)
	if err == nil {
		t.Fatal("GetBackup after delete: expected error, got nil")
	}

	// --- Count: back to zero ---
	count, err = d.CountBackups()
	if err != nil {
		t.Fatalf("CountBackups after delete: %v", err)
	}
	if *count != 0 {
		t.Fatalf("CountBackups after delete = %d, want 0", *count)
	}
}

func TestDatabase_CRUD_BackupSet(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)

	// Create the backup_sets table (not included in CreateAllTables)
	if _, err := d.Connection.ExecContext(d.Context, backupSetsDDL); err != nil {
		t.Fatalf("create backup_sets table: %v", err)
	}

	now := types.TimestampNow()
	hlc := types.HLCNow()

	// --- Count: starts at zero ---
	count, err := d.CountBackupSets()
	if err != nil {
		t.Fatalf("initial CountBackupSets: %v", err)
	}
	if *count != 0 {
		t.Fatalf("initial CountBackupSets = %d, want 0", *count)
	}

	// --- Create ---
	bsID := types.NewBackupSetID()
	created, err := d.CreateBackupSet(CreateBackupSetParams{
		BackupSetID:    bsID,
		CreatedAt:      now,
		HlcTimestamp:   hlc,
		Status:         types.BackupSetStatusPending,
		BackupIds:      types.NewJSONData([]string{}),
		NodeCount:      3,
		CompletedCount: types.NullableInt64{},
		ErrorMessage:   types.NullableString{},
	})
	if err != nil {
		t.Fatalf("CreateBackupSet: %v", err)
	}
	if created == nil {
		t.Fatal("CreateBackupSet returned nil")
	}
	if created.BackupSetID != bsID {
		t.Errorf("BackupSetID = %v, want %v", created.BackupSetID, bsID)
	}
	if created.Status != types.BackupSetStatusPending {
		t.Errorf("Status = %v, want %v", created.Status, types.BackupSetStatusPending)
	}
	if created.NodeCount != 3 {
		t.Errorf("NodeCount = %d, want 3", created.NodeCount)
	}

	// --- Get ---
	got, err := d.GetBackupSet(bsID)
	if err != nil {
		t.Fatalf("GetBackupSet: %v", err)
	}
	if got == nil {
		t.Fatal("GetBackupSet returned nil")
	}
	if got.BackupSetID != bsID {
		t.Errorf("GetBackupSet BackupSetID = %v, want %v", got.BackupSetID, bsID)
	}
	if got.NodeCount != 3 {
		t.Errorf("GetBackupSet NodeCount = %d, want 3", got.NodeCount)
	}

	// --- GetPendingBackupSets ---
	pending, err := d.GetPendingBackupSets()
	if err != nil {
		t.Fatalf("GetPendingBackupSets: %v", err)
	}
	if pending == nil {
		t.Fatal("GetPendingBackupSets returned nil")
	}
	if len(*pending) != 1 {
		t.Fatalf("GetPendingBackupSets len = %d, want 1", len(*pending))
	}
	if (*pending)[0].BackupSetID != bsID {
		t.Errorf("GetPendingBackupSets[0].BackupSetID = %v, want %v", (*pending)[0].BackupSetID, bsID)
	}

	// --- Count: now 1 ---
	count, err = d.CountBackupSets()
	if err != nil {
		t.Fatalf("CountBackupSets after create: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountBackupSets after create = %d, want 1", *count)
	}
}

func TestDatabase_CRUD_BackupVerification(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)

	// Create the backup_verifications table (not included in CreateAllTables)
	if _, err := d.Connection.ExecContext(d.Context, backupVerificationsDDL); err != nil {
		t.Fatalf("create backup_verifications table: %v", err)
	}

	now := types.TimestampNow()

	// First create a backup to reference (FK: backup_verifications.backup_id -> backups.backup_id)
	nodeID := types.NodeID(d.Config.Node_ID)
	backupID := types.NewBackupID()
	_, err := d.CreateBackup(CreateBackupParams{
		BackupID:    backupID,
		NodeID:      nodeID,
		BackupType:  types.BackupTypeFull,
		Status:      types.BackupStatusCompleted,
		StartedAt:   now,
		StoragePath: "/backups/verify-test.zip",
		TriggeredBy: types.NullableString{},
		Metadata:    types.NewJSONData(nil),
	})
	if err != nil {
		t.Fatalf("CreateBackup (setup): %v", err)
	}

	// --- Count: starts at zero ---
	count, err := d.CountVerifications()
	if err != nil {
		t.Fatalf("initial CountVerifications: %v", err)
	}
	if *count != 0 {
		t.Fatalf("initial CountVerifications = %d, want 0", *count)
	}

	// --- Create ---
	vID := types.NewVerificationID()
	created, err := d.CreateVerification(CreateVerificationParams{
		VerificationID:   vID,
		BackupID:         backupID,
		VerifiedAt:       now,
		VerifiedBy:       types.NewNullableString("test-verifier"),
		RestoreTested:    types.NullableBool{Bool: true, Valid: true},
		ChecksumValid:    types.NullableBool{Bool: true, Valid: true},
		RecordCountMatch: types.NullableBool{Bool: true, Valid: true},
		Status:           types.VerificationStatusVerified,
		ErrorMessage:     types.NullableString{},
		DurationMs:       types.NullableInt64{Int64: 1200, Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateVerification: %v", err)
	}
	if created == nil {
		t.Fatal("CreateVerification returned nil")
	}
	if created.VerificationID != vID {
		t.Errorf("VerificationID = %v, want %v", created.VerificationID, vID)
	}
	if created.BackupID != backupID {
		t.Errorf("BackupID = %v, want %v", created.BackupID, backupID)
	}
	if created.Status != types.VerificationStatusVerified {
		t.Errorf("Status = %v, want %v", created.Status, types.VerificationStatusVerified)
	}
	if !created.VerifiedBy.Valid || created.VerifiedBy.String != "test-verifier" {
		t.Errorf("VerifiedBy = %v, want valid 'test-verifier'", created.VerifiedBy)
	}
	if !created.RestoreTested.Valid || !created.RestoreTested.Bool {
		t.Errorf("RestoreTested = %v, want valid true", created.RestoreTested)
	}
	if !created.ChecksumValid.Valid || !created.ChecksumValid.Bool {
		t.Errorf("ChecksumValid = %v, want valid true", created.ChecksumValid)
	}
	if !created.RecordCountMatch.Valid || !created.RecordCountMatch.Bool {
		t.Errorf("RecordCountMatch = %v, want valid true", created.RecordCountMatch)
	}
	if !created.DurationMs.Valid || created.DurationMs.Int64 != 1200 {
		t.Errorf("DurationMs = %v, want valid 1200", created.DurationMs)
	}

	// --- GetVerification ---
	got, err := d.GetVerification(vID)
	if err != nil {
		t.Fatalf("GetVerification: %v", err)
	}
	if got == nil {
		t.Fatal("GetVerification returned nil")
	}
	if got.VerificationID != vID {
		t.Errorf("GetVerification VerificationID = %v, want %v", got.VerificationID, vID)
	}
	if got.BackupID != backupID {
		t.Errorf("GetVerification BackupID = %v, want %v", got.BackupID, backupID)
	}

	// --- GetLatestVerification ---
	latest, err := d.GetLatestVerification(backupID)
	if err != nil {
		t.Fatalf("GetLatestVerification: %v", err)
	}
	if latest == nil {
		t.Fatal("GetLatestVerification returned nil")
	}
	if latest.VerificationID != vID {
		t.Errorf("GetLatestVerification VerificationID = %v, want %v", latest.VerificationID, vID)
	}

	// --- Count: now 1 ---
	count, err = d.CountVerifications()
	if err != nil {
		t.Fatalf("CountVerifications after create: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountVerifications after create = %d, want 1", *count)
	}
}

// TestDatabase_CRUD_Backup_NullableFields verifies that creating a backup
// with all nullable fields left empty works correctly.
func TestDatabase_CRUD_Backup_NullableFields(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)

	nodeID := types.NodeID(d.Config.Node_ID)
	now := types.TimestampNow()
	backupID := types.NewBackupID()

	created, err := d.CreateBackup(CreateBackupParams{
		BackupID:    backupID,
		NodeID:      nodeID,
		BackupType:  types.BackupTypeFull,
		Status:      types.BackupStatusCompleted,
		StartedAt:   now,
		StoragePath: "/backups/nullable-test.zip",
		TriggeredBy: types.NullableString{},
		Metadata:    types.NewJSONData(nil),
	})
	if err != nil {
		t.Fatalf("CreateBackup: %v", err)
	}

	got, err := d.GetBackup(created.BackupID)
	if err != nil {
		t.Fatalf("GetBackup: %v", err)
	}

	// These should be null/invalid since we only set required fields on create
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
	if got.ErrorMessage.Valid {
		t.Errorf("ErrorMessage.Valid = true, want false")
	}
}

// TestDatabase_CRUD_BackupVerification_NullableFields verifies verification
// with null optional fields.
func TestDatabase_CRUD_BackupVerification_NullableFields(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)

	// Create the backup_verifications table
	if _, err := d.Connection.ExecContext(d.Context, backupVerificationsDDL); err != nil {
		t.Fatalf("create backup_verifications table: %v", err)
	}

	nodeID := types.NodeID(d.Config.Node_ID)
	now := types.TimestampNow()

	// Create a backup first
	backupID := types.NewBackupID()
	_, err := d.CreateBackup(CreateBackupParams{
		BackupID:    backupID,
		NodeID:      nodeID,
		BackupType:  types.BackupTypeFull,
		Status:      types.BackupStatusCompleted,
		StartedAt:   now,
		StoragePath: "/backups/null-verify.zip",
		TriggeredBy: types.NullableString{},
		Metadata:    types.NewJSONData(nil),
	})
	if err != nil {
		t.Fatalf("CreateBackup (setup): %v", err)
	}

	vID := types.NewVerificationID()
	created, err := d.CreateVerification(CreateVerificationParams{
		VerificationID:   vID,
		BackupID:         backupID,
		VerifiedAt:       now,
		VerifiedBy:       types.NullableString{},
		RestoreTested:    types.NullableBool{},
		ChecksumValid:    types.NullableBool{},
		RecordCountMatch: types.NullableBool{},
		Status:           types.VerificationStatusPending,
		ErrorMessage:     types.NullableString{},
		DurationMs:       types.NullableInt64{},
	})
	if err != nil {
		t.Fatalf("CreateVerification: %v", err)
	}

	got, err := d.GetVerification(created.VerificationID)
	if err != nil {
		t.Fatalf("GetVerification: %v", err)
	}

	if got.VerifiedBy.Valid {
		t.Errorf("VerifiedBy.Valid = true, want false")
	}
	if got.ErrorMessage.Valid {
		t.Errorf("ErrorMessage.Valid = true, want false")
	}
	if got.DurationMs.Valid {
		t.Errorf("DurationMs.Valid = true, want false")
	}
}
