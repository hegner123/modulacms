package db

import (
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/types"
)

///////////////////////////////
// STRUCTS
//////////////////////////////

type Backup struct {
	BackupID       types.BackupID       `json:"backup_id"`
	NodeID         types.NodeID         `json:"node_id"`
	BackupType     types.BackupType     `json:"backup_type"`
	Status         types.BackupStatus   `json:"status"`
	StartedAt      types.Timestamp      `json:"started_at"`
	CompletedAt    types.Timestamp      `json:"completed_at"`
	DurationMs     types.NullableInt64  `json:"duration_ms"`
	RecordCount    types.NullableInt64  `json:"record_count"`
	SizeBytes      types.NullableInt64  `json:"size_bytes"`
	ReplicationLsn types.NullableString `json:"replication_lsn"`
	HlcTimestamp   types.HLC            `json:"hlc_timestamp"`
	StoragePath    string               `json:"storage_path"`
	Checksum       types.NullableString `json:"checksum"`
	TriggeredBy    types.NullableString `json:"triggered_by"`
	ErrorMessage   types.NullableString `json:"error_message"`
	Metadata       types.JSONData       `json:"metadata"`
}

type BackupSet struct {
	BackupSetID    types.BackupSetID     `json:"backup_set_id"`
	CreatedAt      types.Timestamp       `json:"created_at"`
	HlcTimestamp   types.HLC             `json:"hlc_timestamp"`
	Status         types.BackupSetStatus `json:"status"`
	BackupIds      types.JSONData        `json:"backup_ids"`
	NodeCount      int64                 `json:"node_count"`
	CompletedCount types.NullableInt64   `json:"completed_count"`
	ErrorMessage   types.NullableString  `json:"error_message"`
}

type BackupVerification struct {
	VerificationID   types.VerificationID     `json:"verification_id"`
	BackupID         types.BackupID           `json:"backup_id"`
	VerifiedAt       types.Timestamp          `json:"verified_at"`
	VerifiedBy       types.NullableString     `json:"verified_by"`
	RestoreTested    types.NullableBool       `json:"restore_tested"`
	ChecksumValid    types.NullableBool       `json:"checksum_valid"`
	RecordCountMatch types.NullableBool       `json:"record_count_match"`
	Status           types.VerificationStatus `json:"status"`
	ErrorMessage     types.NullableString     `json:"error_message"`
	DurationMs       types.NullableInt64      `json:"duration_ms"`
}

type CreateBackupParams struct {
	BackupID    types.BackupID       `json:"backup_id"`
	NodeID      types.NodeID         `json:"node_id"`
	BackupType  types.BackupType     `json:"backup_type"`
	Status      types.BackupStatus   `json:"status"`
	StartedAt   types.Timestamp      `json:"started_at"`
	StoragePath string               `json:"storage_path"`
	TriggeredBy types.NullableString `json:"triggered_by"`
	Metadata    types.JSONData       `json:"metadata"`
}

type CreateBackupSetParams struct {
	BackupSetID    types.BackupSetID     `json:"backup_set_id"`
	CreatedAt      types.Timestamp       `json:"created_at"`
	HlcTimestamp   types.HLC             `json:"hlc_timestamp"`
	Status         types.BackupSetStatus `json:"status"`
	BackupIds      types.JSONData        `json:"backup_ids"`
	NodeCount      int64                 `json:"node_count"`
	CompletedCount types.NullableInt64   `json:"completed_count"`
	ErrorMessage   types.NullableString  `json:"error_message"`
}

type CreateVerificationParams struct {
	VerificationID   types.VerificationID     `json:"verification_id"`
	BackupID         types.BackupID           `json:"backup_id"`
	VerifiedAt       types.Timestamp          `json:"verified_at"`
	VerifiedBy       types.NullableString     `json:"verified_by"`
	RestoreTested    types.NullableBool       `json:"restore_tested"`
	ChecksumValid    types.NullableBool       `json:"checksum_valid"`
	RecordCountMatch types.NullableBool       `json:"record_count_match"`
	Status           types.VerificationStatus `json:"status"`
	ErrorMessage     types.NullableString     `json:"error_message"`
	DurationMs       types.NullableInt64      `json:"duration_ms"`
}

type UpdateBackupStatusParams struct {
	Status       types.BackupStatus   `json:"status"`
	CompletedAt  types.Timestamp      `json:"completed_at"`
	DurationMs   types.NullableInt64  `json:"duration_ms"`
	RecordCount  types.NullableInt64  `json:"record_count"`
	SizeBytes    types.NullableInt64  `json:"size_bytes"`
	Checksum     types.NullableString `json:"checksum"`
	ErrorMessage types.NullableString `json:"error_message"`
	BackupID     types.BackupID       `json:"backup_id"`
}

type ListBackupsParams struct {
	Limit  int64 `json:"limit"`
	Offset int64 `json:"offset"`
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

func (d Database) MapBackup(a mdb.Backup) Backup {
	return Backup{
		BackupID:       a.BackupID,
		NodeID:         a.NodeID,
		BackupType:     a.BackupType,
		Status:         a.Status,
		StartedAt:      a.StartedAt,
		CompletedAt:    a.CompletedAt,
		DurationMs:     a.DurationMs,
		RecordCount:    a.RecordCount,
		SizeBytes:      a.SizeBytes,
		ReplicationLsn: a.ReplicationLsn,
		HlcTimestamp:   a.HlcTimestamp,
		StoragePath:    a.StoragePath,
		Checksum:       a.Checksum,
		TriggeredBy:    a.TriggeredBy,
		ErrorMessage:   a.ErrorMessage,
		Metadata:       a.Metadata,
	}
}

func (d Database) MapBackupSet(a mdb.BackupSet) BackupSet {
	return BackupSet{
		BackupSetID:    a.BackupSetID,
		CreatedAt:      a.CreatedAt,
		HlcTimestamp:   a.HlcTimestamp,
		Status:         a.Status,
		BackupIds:      a.BackupIds,
		NodeCount:      a.NodeCount,
		CompletedCount: a.CompletedCount,
		ErrorMessage:   a.ErrorMessage,
	}
}

func (d Database) MapBackupVerification(a mdb.BackupVerification) BackupVerification {
	return BackupVerification{
		VerificationID:   a.VerificationID,
		BackupID:         a.BackupID,
		VerifiedAt:       a.VerifiedAt,
		VerifiedBy:       a.VerifiedBy,
		RestoreTested:    a.RestoreTested,
		ChecksumValid:    a.ChecksumValid,
		RecordCountMatch: a.RecordCountMatch,
		Status:           a.Status,
		ErrorMessage:     a.ErrorMessage,
		DurationMs:       a.DurationMs,
	}
}

func (d Database) MapCreateBackupParams(a CreateBackupParams) mdb.CreateBackupParams {
	return mdb.CreateBackupParams{
		BackupID:    a.BackupID,
		NodeID:      a.NodeID,
		BackupType:  a.BackupType,
		Status:      a.Status,
		StartedAt:   a.StartedAt,
		StoragePath: a.StoragePath,
		TriggeredBy: a.TriggeredBy,
		Metadata:    a.Metadata,
	}
}

func (d Database) MapCreateBackupSetParams(a CreateBackupSetParams) mdb.CreateBackupSetParams {
	return mdb.CreateBackupSetParams{
		BackupSetID:    a.BackupSetID,
		CreatedAt:      a.CreatedAt,
		HlcTimestamp:   a.HlcTimestamp,
		Status:         a.Status,
		BackupIds:      a.BackupIds,
		NodeCount:      a.NodeCount,
		CompletedCount: a.CompletedCount,
		ErrorMessage:   a.ErrorMessage,
	}
}

func (d Database) MapCreateVerificationParams(a CreateVerificationParams) mdb.CreateVerificationParams {
	return mdb.CreateVerificationParams{
		VerificationID:   a.VerificationID,
		BackupID:         a.BackupID,
		VerifiedAt:       a.VerifiedAt,
		VerifiedBy:       a.VerifiedBy,
		RestoreTested:    a.RestoreTested,
		ChecksumValid:    a.ChecksumValid,
		RecordCountMatch: a.RecordCountMatch,
		Status:           a.Status,
		ErrorMessage:     a.ErrorMessage,
		DurationMs:       a.DurationMs,
	}
}

// QUERIES - Backups

func (d Database) CreateBackupTables() error {
	queries := mdb.New(d.Connection)
	return queries.CreateBackupTables(d.Context)
}

func (d Database) DropBackupTables() error {
	queries := mdb.New(d.Connection)
	return queries.DropBackupTables(d.Context)
}

func (d Database) CreateBackup(params CreateBackupParams) (*Backup, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.CreateBackup(d.Context, d.MapCreateBackupParams(params))
	if err != nil {
		return nil, fmt.Errorf("failed to create backup: %v", err)
	}
	res := d.MapBackup(row)
	return &res, nil
}

func (d Database) GetBackup(id types.BackupID) (*Backup, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetBackup(d.Context, mdb.GetBackupParams{BackupID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapBackup(row)
	return &res, nil
}

func (d Database) GetLatestBackup(nodeID types.NodeID) (*Backup, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetLatestBackup(d.Context, mdb.GetLatestBackupParams{NodeID: nodeID})
	if err != nil {
		return nil, err
	}
	res := d.MapBackup(row)
	return &res, nil
}

func (d Database) ListBackups(params ListBackupsParams) (*[]Backup, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListBackups(d.Context, mdb.ListBackupsParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %v", err)
	}
	res := []Backup{}
	for _, v := range rows {
		res = append(res, d.MapBackup(v))
	}
	return &res, nil
}

func (d Database) UpdateBackupStatus(params UpdateBackupStatusParams) error {
	queries := mdb.New(d.Connection)
	return queries.UpdateBackupStatus(d.Context, mdb.UpdateBackupStatusParams{
		Status:       params.Status,
		CompletedAt:  params.CompletedAt,
		DurationMs:   params.DurationMs,
		RecordCount:  params.RecordCount,
		SizeBytes:    params.SizeBytes,
		Checksum:     params.Checksum,
		ErrorMessage: params.ErrorMessage,
		BackupID:     params.BackupID,
	})
}

func (d Database) CountBackups() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountBackups(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count backups: %v", err)
	}
	return &c, nil
}

func (d Database) DeleteBackup(id types.BackupID) error {
	queries := mdb.New(d.Connection)
	return queries.DeleteBackup(d.Context, mdb.DeleteBackupParams{BackupID: id})
}

// QUERIES - Backup Sets

func (d Database) CreateBackupSet(params CreateBackupSetParams) (*BackupSet, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.CreateBackupSet(d.Context, d.MapCreateBackupSetParams(params))
	if err != nil {
		return nil, fmt.Errorf("failed to create backup set: %v", err)
	}
	res := d.MapBackupSet(row)
	return &res, nil
}

func (d Database) GetBackupSet(id types.BackupSetID) (*BackupSet, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetBackupSet(d.Context, mdb.GetBackupSetParams{BackupSetID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapBackupSet(row)
	return &res, nil
}

func (d Database) GetPendingBackupSets() (*[]BackupSet, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.GetPendingBackupSets(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending backup sets: %v", err)
	}
	res := []BackupSet{}
	for _, v := range rows {
		res = append(res, d.MapBackupSet(v))
	}
	return &res, nil
}

func (d Database) CountBackupSets() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountBackupSets(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count backup sets: %v", err)
	}
	return &c, nil
}

// QUERIES - Verifications

func (d Database) CreateVerification(params CreateVerificationParams) (*BackupVerification, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.CreateVerification(d.Context, d.MapCreateVerificationParams(params))
	if err != nil {
		return nil, fmt.Errorf("failed to create verification: %v", err)
	}
	res := d.MapBackupVerification(row)
	return &res, nil
}

func (d Database) GetVerification(id types.VerificationID) (*BackupVerification, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetVerification(d.Context, mdb.GetVerificationParams{VerificationID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapBackupVerification(row)
	return &res, nil
}

func (d Database) GetLatestVerification(backupID types.BackupID) (*BackupVerification, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetLatestVerification(d.Context, mdb.GetLatestVerificationParams{BackupID: backupID})
	if err != nil {
		return nil, err
	}
	res := d.MapBackupVerification(row)
	return &res, nil
}

func (d Database) CountVerifications() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountVerifications(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count verifications: %v", err)
	}
	return &c, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

func (d MysqlDatabase) MapBackup(a mdbm.Backup) Backup {
	return Backup{
		BackupID:       a.BackupID,
		NodeID:         a.NodeID,
		BackupType:     a.BackupType,
		Status:         a.Status,
		StartedAt:      a.StartedAt,
		CompletedAt:    a.CompletedAt,
		DurationMs:     a.DurationMs,
		RecordCount:    a.RecordCount,
		SizeBytes:      a.SizeBytes,
		ReplicationLsn: a.ReplicationLsn,
		HlcTimestamp:   a.HlcTimestamp,
		StoragePath:    a.StoragePath,
		Checksum:       a.Checksum,
		TriggeredBy:    a.TriggeredBy,
		ErrorMessage:   a.ErrorMessage,
		Metadata:       a.Metadata,
	}
}

func (d MysqlDatabase) MapBackupSet(a mdbm.BackupSet) BackupSet {
	return BackupSet{
		BackupSetID:    a.BackupSetID,
		CreatedAt:      a.CreatedAt,
		HlcTimestamp:   a.HlcTimestamp,
		Status:         a.Status,
		BackupIds:      a.BackupIds,
		NodeCount:      int64(a.NodeCount),
		CompletedCount: a.CompletedCount,
		ErrorMessage:   a.ErrorMessage,
	}
}

func (d MysqlDatabase) MapBackupVerification(a mdbm.BackupVerification) BackupVerification {
	return BackupVerification{
		VerificationID:   a.VerificationID,
		BackupID:         a.BackupID,
		VerifiedAt:       a.VerifiedAt,
		VerifiedBy:       a.VerifiedBy,
		RestoreTested:    a.RestoreTested,
		ChecksumValid:    a.ChecksumValid,
		RecordCountMatch: a.RecordCountMatch,
		Status:           a.Status,
		ErrorMessage:     a.ErrorMessage,
		DurationMs:       a.DurationMs,
	}
}

func (d MysqlDatabase) MapCreateBackupParams(a CreateBackupParams) mdbm.CreateBackupParams {
	return mdbm.CreateBackupParams{
		BackupID:    a.BackupID,
		NodeID:      a.NodeID,
		BackupType:  a.BackupType,
		Status:      a.Status,
		StartedAt:   a.StartedAt,
		StoragePath: a.StoragePath,
		TriggeredBy: a.TriggeredBy,
		Metadata:    a.Metadata,
	}
}

func (d MysqlDatabase) MapCreateBackupSetParams(a CreateBackupSetParams) mdbm.CreateBackupSetParams {
	return mdbm.CreateBackupSetParams{
		BackupSetID:    a.BackupSetID,
		CreatedAt:      a.CreatedAt,
		HlcTimestamp:   a.HlcTimestamp,
		Status:         a.Status,
		BackupIds:      a.BackupIds,
		NodeCount:      int32(a.NodeCount),
		CompletedCount: a.CompletedCount,
		ErrorMessage:   a.ErrorMessage,
	}
}

func (d MysqlDatabase) MapCreateVerificationParams(a CreateVerificationParams) mdbm.CreateVerificationParams {
	return mdbm.CreateVerificationParams{
		VerificationID:   a.VerificationID,
		BackupID:         a.BackupID,
		VerifiedAt:       a.VerifiedAt,
		VerifiedBy:       a.VerifiedBy,
		RestoreTested:    a.RestoreTested,
		ChecksumValid:    a.ChecksumValid,
		RecordCountMatch: a.RecordCountMatch,
		Status:           a.Status,
		ErrorMessage:     a.ErrorMessage,
		DurationMs:       a.DurationMs,
	}
}

// QUERIES - Backups

func (d MysqlDatabase) CreateBackupTables() error {
	queries := mdbm.New(d.Connection)
	return queries.CreateBackupTables(d.Context)
}

func (d MysqlDatabase) DropBackupTables() error {
	queries := mdbm.New(d.Connection)
	return queries.DropBackupTables(d.Context)
}

func (d MysqlDatabase) CreateBackup(params CreateBackupParams) (*Backup, error) {
	queries := mdbm.New(d.Connection)
	err := queries.CreateBackup(d.Context, d.MapCreateBackupParams(params))
	if err != nil {
		return nil, fmt.Errorf("failed to create backup: %v", err)
	}
	row, err := queries.GetBackup(d.Context, mdbm.GetBackupParams{BackupID: params.BackupID})
	if err != nil {
		return nil, fmt.Errorf("failed to get created backup: %v", err)
	}
	res := d.MapBackup(row)
	return &res, nil
}

func (d MysqlDatabase) GetBackup(id types.BackupID) (*Backup, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetBackup(d.Context, mdbm.GetBackupParams{BackupID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapBackup(row)
	return &res, nil
}

func (d MysqlDatabase) GetLatestBackup(nodeID types.NodeID) (*Backup, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetLatestBackup(d.Context, mdbm.GetLatestBackupParams{NodeID: nodeID})
	if err != nil {
		return nil, err
	}
	res := d.MapBackup(row)
	return &res, nil
}

func (d MysqlDatabase) ListBackups(params ListBackupsParams) (*[]Backup, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListBackups(d.Context, mdbm.ListBackupsParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %v", err)
	}
	res := []Backup{}
	for _, v := range rows {
		res = append(res, d.MapBackup(v))
	}
	return &res, nil
}

func (d MysqlDatabase) UpdateBackupStatus(params UpdateBackupStatusParams) error {
	queries := mdbm.New(d.Connection)
	return queries.UpdateBackupStatus(d.Context, mdbm.UpdateBackupStatusParams{
		Status:       params.Status,
		CompletedAt:  params.CompletedAt,
		DurationMs:   params.DurationMs,
		RecordCount:  params.RecordCount,
		SizeBytes:    params.SizeBytes,
		Checksum:     params.Checksum,
		ErrorMessage: params.ErrorMessage,
		BackupID:     params.BackupID,
	})
}

func (d MysqlDatabase) CountBackups() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountBackups(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count backups: %v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) DeleteBackup(id types.BackupID) error {
	queries := mdbm.New(d.Connection)
	return queries.DeleteBackup(d.Context, mdbm.DeleteBackupParams{BackupID: id})
}

// QUERIES - Backup Sets

func (d MysqlDatabase) CreateBackupSet(params CreateBackupSetParams) (*BackupSet, error) {
	queries := mdbm.New(d.Connection)
	err := queries.CreateBackupSet(d.Context, d.MapCreateBackupSetParams(params))
	if err != nil {
		return nil, fmt.Errorf("failed to create backup set: %v", err)
	}
	row, err := queries.GetBackupSet(d.Context, mdbm.GetBackupSetParams{BackupSetID: params.BackupSetID})
	if err != nil {
		return nil, fmt.Errorf("failed to get created backup set: %v", err)
	}
	res := d.MapBackupSet(row)
	return &res, nil
}

func (d MysqlDatabase) GetBackupSet(id types.BackupSetID) (*BackupSet, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetBackupSet(d.Context, mdbm.GetBackupSetParams{BackupSetID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapBackupSet(row)
	return &res, nil
}

func (d MysqlDatabase) GetPendingBackupSets() (*[]BackupSet, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.GetPendingBackupSets(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending backup sets: %v", err)
	}
	res := []BackupSet{}
	for _, v := range rows {
		res = append(res, d.MapBackupSet(v))
	}
	return &res, nil
}

func (d MysqlDatabase) CountBackupSets() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountBackupSets(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count backup sets: %v", err)
	}
	return &c, nil
}

// QUERIES - Verifications

func (d MysqlDatabase) CreateVerification(params CreateVerificationParams) (*BackupVerification, error) {
	queries := mdbm.New(d.Connection)
	err := queries.CreateVerification(d.Context, d.MapCreateVerificationParams(params))
	if err != nil {
		return nil, fmt.Errorf("failed to create verification: %v", err)
	}
	row, err := queries.GetVerification(d.Context, mdbm.GetVerificationParams{VerificationID: params.VerificationID})
	if err != nil {
		return nil, fmt.Errorf("failed to get created verification: %v", err)
	}
	res := d.MapBackupVerification(row)
	return &res, nil
}

func (d MysqlDatabase) GetVerification(id types.VerificationID) (*BackupVerification, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetVerification(d.Context, mdbm.GetVerificationParams{VerificationID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapBackupVerification(row)
	return &res, nil
}

func (d MysqlDatabase) GetLatestVerification(backupID types.BackupID) (*BackupVerification, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetLatestVerification(d.Context, mdbm.GetLatestVerificationParams{BackupID: backupID})
	if err != nil {
		return nil, err
	}
	res := d.MapBackupVerification(row)
	return &res, nil
}

func (d MysqlDatabase) CountVerifications() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountVerifications(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count verifications: %v", err)
	}
	return &c, nil
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

func (d PsqlDatabase) MapBackup(a mdbp.Backup) Backup {
	return Backup{
		BackupID:       a.BackupID,
		NodeID:         a.NodeID,
		BackupType:     a.BackupType,
		Status:         a.Status,
		StartedAt:      a.StartedAt,
		CompletedAt:    a.CompletedAt,
		DurationMs:     a.DurationMs,
		RecordCount:    a.RecordCount,
		SizeBytes:      a.SizeBytes,
		ReplicationLsn: a.ReplicationLsn,
		HlcTimestamp:   a.HlcTimestamp,
		StoragePath:    a.StoragePath,
		Checksum:       a.Checksum,
		TriggeredBy:    a.TriggeredBy,
		ErrorMessage:   a.ErrorMessage,
		Metadata:       a.Metadata,
	}
}

func (d PsqlDatabase) MapBackupSet(a mdbp.BackupSet) BackupSet {
	return BackupSet{
		BackupSetID:    a.BackupSetID,
		CreatedAt:      a.CreatedAt,
		HlcTimestamp:   a.HlcTimestamp,
		Status:         a.Status,
		BackupIds:      a.BackupIds,
		NodeCount:      int64(a.NodeCount),
		CompletedCount: a.CompletedCount,
		ErrorMessage:   a.ErrorMessage,
	}
}

func (d PsqlDatabase) MapBackupVerification(a mdbp.BackupVerification) BackupVerification {
	return BackupVerification{
		VerificationID:   a.VerificationID,
		BackupID:         a.BackupID,
		VerifiedAt:       a.VerifiedAt,
		VerifiedBy:       a.VerifiedBy,
		RestoreTested:    a.RestoreTested,
		ChecksumValid:    a.ChecksumValid,
		RecordCountMatch: a.RecordCountMatch,
		Status:           a.Status,
		ErrorMessage:     a.ErrorMessage,
		DurationMs:       a.DurationMs,
	}
}

func (d PsqlDatabase) MapCreateBackupParams(a CreateBackupParams) mdbp.CreateBackupParams {
	return mdbp.CreateBackupParams{
		BackupID:    a.BackupID,
		NodeID:      a.NodeID,
		BackupType:  a.BackupType,
		Status:      a.Status,
		StartedAt:   a.StartedAt,
		StoragePath: a.StoragePath,
		TriggeredBy: a.TriggeredBy,
		Metadata:    a.Metadata,
	}
}

func (d PsqlDatabase) MapCreateBackupSetParams(a CreateBackupSetParams) mdbp.CreateBackupSetParams {
	return mdbp.CreateBackupSetParams{
		BackupSetID:    a.BackupSetID,
		CreatedAt:      a.CreatedAt,
		HlcTimestamp:   a.HlcTimestamp,
		Status:         a.Status,
		BackupIds:      a.BackupIds,
		NodeCount:      int32(a.NodeCount),
		CompletedCount: a.CompletedCount,
		ErrorMessage:   a.ErrorMessage,
	}
}

func (d PsqlDatabase) MapCreateVerificationParams(a CreateVerificationParams) mdbp.CreateVerificationParams {
	return mdbp.CreateVerificationParams{
		VerificationID:   a.VerificationID,
		BackupID:         a.BackupID,
		VerifiedAt:       a.VerifiedAt,
		VerifiedBy:       a.VerifiedBy,
		RestoreTested:    a.RestoreTested,
		ChecksumValid:    a.ChecksumValid,
		RecordCountMatch: a.RecordCountMatch,
		Status:           a.Status,
		ErrorMessage:     a.ErrorMessage,
		DurationMs:       a.DurationMs,
	}
}

// QUERIES - Backups

func (d PsqlDatabase) CreateBackupTables() error {
	queries := mdbp.New(d.Connection)
	return queries.CreateBackupTables(d.Context)
}

func (d PsqlDatabase) DropBackupTables() error {
	queries := mdbp.New(d.Connection)
	return queries.DropBackupTables(d.Context)
}

func (d PsqlDatabase) CreateBackup(params CreateBackupParams) (*Backup, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateBackup(d.Context, d.MapCreateBackupParams(params))
	if err != nil {
		return nil, fmt.Errorf("failed to create backup: %v", err)
	}
	res := d.MapBackup(row)
	return &res, nil
}

func (d PsqlDatabase) GetBackup(id types.BackupID) (*Backup, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetBackup(d.Context, mdbp.GetBackupParams{BackupID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapBackup(row)
	return &res, nil
}

func (d PsqlDatabase) GetLatestBackup(nodeID types.NodeID) (*Backup, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetLatestBackup(d.Context, mdbp.GetLatestBackupParams{NodeID: nodeID})
	if err != nil {
		return nil, err
	}
	res := d.MapBackup(row)
	return &res, nil
}

func (d PsqlDatabase) ListBackups(params ListBackupsParams) (*[]Backup, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListBackups(d.Context, mdbp.ListBackupsParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %v", err)
	}
	res := []Backup{}
	for _, v := range rows {
		res = append(res, d.MapBackup(v))
	}
	return &res, nil
}

func (d PsqlDatabase) UpdateBackupStatus(params UpdateBackupStatusParams) error {
	queries := mdbp.New(d.Connection)
	return queries.UpdateBackupStatus(d.Context, mdbp.UpdateBackupStatusParams{
		Status:       params.Status,
		CompletedAt:  params.CompletedAt,
		DurationMs:   params.DurationMs,
		RecordCount:  params.RecordCount,
		SizeBytes:    params.SizeBytes,
		Checksum:     params.Checksum,
		ErrorMessage: params.ErrorMessage,
		BackupID:     params.BackupID,
	})
}

func (d PsqlDatabase) CountBackups() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountBackups(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count backups: %v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) DeleteBackup(id types.BackupID) error {
	queries := mdbp.New(d.Connection)
	return queries.DeleteBackup(d.Context, mdbp.DeleteBackupParams{BackupID: id})
}

// QUERIES - Backup Sets

func (d PsqlDatabase) CreateBackupSet(params CreateBackupSetParams) (*BackupSet, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateBackupSet(d.Context, d.MapCreateBackupSetParams(params))
	if err != nil {
		return nil, fmt.Errorf("failed to create backup set: %v", err)
	}
	res := d.MapBackupSet(row)
	return &res, nil
}

func (d PsqlDatabase) GetBackupSet(id types.BackupSetID) (*BackupSet, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetBackupSet(d.Context, mdbp.GetBackupSetParams{BackupSetID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapBackupSet(row)
	return &res, nil
}

func (d PsqlDatabase) GetPendingBackupSets() (*[]BackupSet, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.GetPendingBackupSets(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending backup sets: %v", err)
	}
	res := []BackupSet{}
	for _, v := range rows {
		res = append(res, d.MapBackupSet(v))
	}
	return &res, nil
}

func (d PsqlDatabase) CountBackupSets() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountBackupSets(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count backup sets: %v", err)
	}
	return &c, nil
}

// QUERIES - Verifications

func (d PsqlDatabase) CreateVerification(params CreateVerificationParams) (*BackupVerification, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateVerification(d.Context, d.MapCreateVerificationParams(params))
	if err != nil {
		return nil, fmt.Errorf("failed to create verification: %v", err)
	}
	res := d.MapBackupVerification(row)
	return &res, nil
}

func (d PsqlDatabase) GetVerification(id types.VerificationID) (*BackupVerification, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetVerification(d.Context, mdbp.GetVerificationParams{VerificationID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapBackupVerification(row)
	return &res, nil
}

func (d PsqlDatabase) GetLatestVerification(backupID types.BackupID) (*BackupVerification, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetLatestVerification(d.Context, mdbp.GetLatestVerificationParams{BackupID: backupID})
	if err != nil {
		return nil, err
	}
	res := d.MapBackupVerification(row)
	return &res, nil
}

func (d PsqlDatabase) CountVerifications() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountVerifications(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count verifications: %v", err)
	}
	return &c, nil
}
