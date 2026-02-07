package db

import (
	"context"
	"database/sql"
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/audited"
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

///////////////////////////////
// AUDITED COMMANDS
//////////////////////////////

// ===================================================================
// Backup — SQLite
// ===================================================================

// ----- SQLite CREATE -----

type NewBackupCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateBackupParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewBackupCmd) Context() context.Context              { return c.ctx }
func (c NewBackupCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewBackupCmd) Connection() *sql.DB                   { return c.conn }
func (c NewBackupCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewBackupCmd) TableName() string                     { return "backups" }
func (c NewBackupCmd) Params() any                           { return c.params }
func (c NewBackupCmd) GetID(r mdb.Backup) string             { return string(r.BackupID) }

func (c NewBackupCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.Backup, error) {
	id := c.params.BackupID
	if id.IsZero() {
		id = types.NewBackupID()
	}
	queries := mdb.New(tx)
	return queries.CreateBackup(ctx, mdb.CreateBackupParams{
		BackupID:    id,
		NodeID:      c.params.NodeID,
		BackupType:  c.params.BackupType,
		Status:      c.params.Status,
		StartedAt:   c.params.StartedAt,
		StoragePath: c.params.StoragePath,
		TriggeredBy: c.params.TriggeredBy,
		Metadata:    c.params.Metadata,
	})
}

func (d Database) NewBackupCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateBackupParams) NewBackupCmd {
	return NewBackupCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite DELETE -----

type DeleteBackupCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.BackupID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteBackupCmd) Context() context.Context              { return c.ctx }
func (c DeleteBackupCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteBackupCmd) Connection() *sql.DB                   { return c.conn }
func (c DeleteBackupCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteBackupCmd) TableName() string                     { return "backups" }
func (c DeleteBackupCmd) GetID() string                         { return string(c.id) }

func (c DeleteBackupCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Backup, error) {
	queries := mdb.New(tx)
	return queries.GetBackup(ctx, mdb.GetBackupParams{BackupID: c.id})
}

func (c DeleteBackupCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteBackup(ctx, mdb.DeleteBackupParams{BackupID: c.id})
}

func (d Database) DeleteBackupCmd(ctx context.Context, auditCtx audited.AuditContext, id types.BackupID) DeleteBackupCmd {
	return DeleteBackupCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

// ===================================================================
// Backup — MySQL
// ===================================================================

// ----- MySQL CREATE -----

type NewBackupCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateBackupParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewBackupCmdMysql) Context() context.Context              { return c.ctx }
func (c NewBackupCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewBackupCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c NewBackupCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewBackupCmdMysql) TableName() string                     { return "backups" }
func (c NewBackupCmdMysql) Params() any                           { return c.params }
func (c NewBackupCmdMysql) GetID(r mdbm.Backup) string            { return string(r.BackupID) }

func (c NewBackupCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.Backup, error) {
	id := c.params.BackupID
	if id.IsZero() {
		id = types.NewBackupID()
	}
	queries := mdbm.New(tx)
	params := mdbm.CreateBackupParams{
		BackupID:    id,
		NodeID:      c.params.NodeID,
		BackupType:  c.params.BackupType,
		Status:      c.params.Status,
		StartedAt:   c.params.StartedAt,
		StoragePath: c.params.StoragePath,
		TriggeredBy: c.params.TriggeredBy,
		Metadata:    c.params.Metadata,
	}
	if err := queries.CreateBackup(ctx, params); err != nil {
		return mdbm.Backup{}, err
	}
	return queries.GetBackup(ctx, mdbm.GetBackupParams{BackupID: params.BackupID})
}

func (d MysqlDatabase) NewBackupCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateBackupParams) NewBackupCmdMysql {
	return NewBackupCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL DELETE -----

type DeleteBackupCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.BackupID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteBackupCmdMysql) Context() context.Context              { return c.ctx }
func (c DeleteBackupCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteBackupCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c DeleteBackupCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteBackupCmdMysql) TableName() string                     { return "backups" }
func (c DeleteBackupCmdMysql) GetID() string                         { return string(c.id) }

func (c DeleteBackupCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Backup, error) {
	queries := mdbm.New(tx)
	return queries.GetBackup(ctx, mdbm.GetBackupParams{BackupID: c.id})
}

func (c DeleteBackupCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteBackup(ctx, mdbm.DeleteBackupParams{BackupID: c.id})
}

func (d MysqlDatabase) DeleteBackupCmd(ctx context.Context, auditCtx audited.AuditContext, id types.BackupID) DeleteBackupCmdMysql {
	return DeleteBackupCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

// ===================================================================
// Backup — PostgreSQL
// ===================================================================

// ----- PostgreSQL CREATE -----

type NewBackupCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateBackupParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewBackupCmdPsql) Context() context.Context              { return c.ctx }
func (c NewBackupCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewBackupCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c NewBackupCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewBackupCmdPsql) TableName() string                     { return "backups" }
func (c NewBackupCmdPsql) Params() any                           { return c.params }
func (c NewBackupCmdPsql) GetID(r mdbp.Backup) string            { return string(r.BackupID) }

func (c NewBackupCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.Backup, error) {
	id := c.params.BackupID
	if id.IsZero() {
		id = types.NewBackupID()
	}
	queries := mdbp.New(tx)
	return queries.CreateBackup(ctx, mdbp.CreateBackupParams{
		BackupID:    id,
		NodeID:      c.params.NodeID,
		BackupType:  c.params.BackupType,
		Status:      c.params.Status,
		StartedAt:   c.params.StartedAt,
		StoragePath: c.params.StoragePath,
		TriggeredBy: c.params.TriggeredBy,
		Metadata:    c.params.Metadata,
	})
}

func (d PsqlDatabase) NewBackupCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateBackupParams) NewBackupCmdPsql {
	return NewBackupCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL DELETE -----

type DeleteBackupCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.BackupID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteBackupCmdPsql) Context() context.Context              { return c.ctx }
func (c DeleteBackupCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteBackupCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c DeleteBackupCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteBackupCmdPsql) TableName() string                     { return "backups" }
func (c DeleteBackupCmdPsql) GetID() string                         { return string(c.id) }

func (c DeleteBackupCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Backup, error) {
	queries := mdbp.New(tx)
	return queries.GetBackup(ctx, mdbp.GetBackupParams{BackupID: c.id})
}

func (c DeleteBackupCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteBackup(ctx, mdbp.DeleteBackupParams{BackupID: c.id})
}

func (d PsqlDatabase) DeleteBackupCmd(ctx context.Context, auditCtx audited.AuditContext, id types.BackupID) DeleteBackupCmdPsql {
	return DeleteBackupCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}

// ===================================================================
// BackupSet — SQLite
// ===================================================================

// ----- SQLite CREATE -----

type NewBackupSetCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateBackupSetParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewBackupSetCmd) Context() context.Context              { return c.ctx }
func (c NewBackupSetCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewBackupSetCmd) Connection() *sql.DB                   { return c.conn }
func (c NewBackupSetCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewBackupSetCmd) TableName() string                     { return "backup_sets" }
func (c NewBackupSetCmd) Params() any                           { return c.params }
func (c NewBackupSetCmd) GetID(r mdb.BackupSet) string          { return string(r.BackupSetID) }

func (c NewBackupSetCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.BackupSet, error) {
	id := c.params.BackupSetID
	if id.IsZero() {
		id = types.NewBackupSetID()
	}
	queries := mdb.New(tx)
	return queries.CreateBackupSet(ctx, mdb.CreateBackupSetParams{
		BackupSetID:    id,
		CreatedAt:      c.params.CreatedAt,
		HlcTimestamp:   c.params.HlcTimestamp,
		Status:         c.params.Status,
		BackupIds:      c.params.BackupIds,
		NodeCount:      c.params.NodeCount,
		CompletedCount: c.params.CompletedCount,
		ErrorMessage:   c.params.ErrorMessage,
	})
}

func (d Database) NewBackupSetCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateBackupSetParams) NewBackupSetCmd {
	return NewBackupSetCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite DELETE -----

type DeleteBackupSetCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.BackupSetID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteBackupSetCmd) Context() context.Context              { return c.ctx }
func (c DeleteBackupSetCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteBackupSetCmd) Connection() *sql.DB                   { return c.conn }
func (c DeleteBackupSetCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteBackupSetCmd) TableName() string                     { return "backup_sets" }
func (c DeleteBackupSetCmd) GetID() string                         { return string(c.id) }

func (c DeleteBackupSetCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.BackupSet, error) {
	queries := mdb.New(tx)
	return queries.GetBackupSet(ctx, mdb.GetBackupSetParams{BackupSetID: c.id})
}

func (c DeleteBackupSetCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteBackupSet(ctx, mdb.DeleteBackupSetParams{BackupSetID: c.id})
}

func (d Database) DeleteBackupSetCmd(ctx context.Context, auditCtx audited.AuditContext, id types.BackupSetID) DeleteBackupSetCmd {
	return DeleteBackupSetCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

// ===================================================================
// BackupSet — MySQL
// ===================================================================

// ----- MySQL CREATE -----

type NewBackupSetCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateBackupSetParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewBackupSetCmdMysql) Context() context.Context              { return c.ctx }
func (c NewBackupSetCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewBackupSetCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c NewBackupSetCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewBackupSetCmdMysql) TableName() string                     { return "backup_sets" }
func (c NewBackupSetCmdMysql) Params() any                           { return c.params }
func (c NewBackupSetCmdMysql) GetID(r mdbm.BackupSet) string         { return string(r.BackupSetID) }

func (c NewBackupSetCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.BackupSet, error) {
	id := c.params.BackupSetID
	if id.IsZero() {
		id = types.NewBackupSetID()
	}
	queries := mdbm.New(tx)
	params := mdbm.CreateBackupSetParams{
		BackupSetID:    id,
		CreatedAt:      c.params.CreatedAt,
		HlcTimestamp:   c.params.HlcTimestamp,
		Status:         c.params.Status,
		BackupIds:      c.params.BackupIds,
		NodeCount:      int32(c.params.NodeCount),
		CompletedCount: c.params.CompletedCount,
		ErrorMessage:   c.params.ErrorMessage,
	}
	if err := queries.CreateBackupSet(ctx, params); err != nil {
		return mdbm.BackupSet{}, err
	}
	return queries.GetBackupSet(ctx, mdbm.GetBackupSetParams{BackupSetID: params.BackupSetID})
}

func (d MysqlDatabase) NewBackupSetCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateBackupSetParams) NewBackupSetCmdMysql {
	return NewBackupSetCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL DELETE -----

type DeleteBackupSetCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.BackupSetID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteBackupSetCmdMysql) Context() context.Context              { return c.ctx }
func (c DeleteBackupSetCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteBackupSetCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c DeleteBackupSetCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteBackupSetCmdMysql) TableName() string                     { return "backup_sets" }
func (c DeleteBackupSetCmdMysql) GetID() string                         { return string(c.id) }

func (c DeleteBackupSetCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.BackupSet, error) {
	queries := mdbm.New(tx)
	return queries.GetBackupSet(ctx, mdbm.GetBackupSetParams{BackupSetID: c.id})
}

func (c DeleteBackupSetCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteBackupSet(ctx, mdbm.DeleteBackupSetParams{BackupSetID: c.id})
}

func (d MysqlDatabase) DeleteBackupSetCmd(ctx context.Context, auditCtx audited.AuditContext, id types.BackupSetID) DeleteBackupSetCmdMysql {
	return DeleteBackupSetCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

// ===================================================================
// BackupSet — PostgreSQL
// ===================================================================

// ----- PostgreSQL CREATE -----

type NewBackupSetCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateBackupSetParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewBackupSetCmdPsql) Context() context.Context              { return c.ctx }
func (c NewBackupSetCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewBackupSetCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c NewBackupSetCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewBackupSetCmdPsql) TableName() string                     { return "backup_sets" }
func (c NewBackupSetCmdPsql) Params() any                           { return c.params }
func (c NewBackupSetCmdPsql) GetID(r mdbp.BackupSet) string         { return string(r.BackupSetID) }

func (c NewBackupSetCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.BackupSet, error) {
	id := c.params.BackupSetID
	if id.IsZero() {
		id = types.NewBackupSetID()
	}
	queries := mdbp.New(tx)
	return queries.CreateBackupSet(ctx, mdbp.CreateBackupSetParams{
		BackupSetID:    id,
		CreatedAt:      c.params.CreatedAt,
		HlcTimestamp:   c.params.HlcTimestamp,
		Status:         c.params.Status,
		BackupIds:      c.params.BackupIds,
		NodeCount:      int32(c.params.NodeCount),
		CompletedCount: c.params.CompletedCount,
		ErrorMessage:   c.params.ErrorMessage,
	})
}

func (d PsqlDatabase) NewBackupSetCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateBackupSetParams) NewBackupSetCmdPsql {
	return NewBackupSetCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL DELETE -----

type DeleteBackupSetCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.BackupSetID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteBackupSetCmdPsql) Context() context.Context              { return c.ctx }
func (c DeleteBackupSetCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteBackupSetCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c DeleteBackupSetCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteBackupSetCmdPsql) TableName() string                     { return "backup_sets" }
func (c DeleteBackupSetCmdPsql) GetID() string                         { return string(c.id) }

func (c DeleteBackupSetCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.BackupSet, error) {
	queries := mdbp.New(tx)
	return queries.GetBackupSet(ctx, mdbp.GetBackupSetParams{BackupSetID: c.id})
}

func (c DeleteBackupSetCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteBackupSet(ctx, mdbp.DeleteBackupSetParams{BackupSetID: c.id})
}

func (d PsqlDatabase) DeleteBackupSetCmd(ctx context.Context, auditCtx audited.AuditContext, id types.BackupSetID) DeleteBackupSetCmdPsql {
	return DeleteBackupSetCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}

// ===================================================================
// BackupVerification — SQLite
// ===================================================================

// ----- SQLite CREATE -----

type NewVerificationCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateVerificationParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewVerificationCmd) Context() context.Context              { return c.ctx }
func (c NewVerificationCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewVerificationCmd) Connection() *sql.DB                   { return c.conn }
func (c NewVerificationCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewVerificationCmd) TableName() string                     { return "backup_verifications" }
func (c NewVerificationCmd) Params() any                           { return c.params }
func (c NewVerificationCmd) GetID(r mdb.BackupVerification) string { return string(r.VerificationID) }

func (c NewVerificationCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.BackupVerification, error) {
	id := c.params.VerificationID
	if id.IsZero() {
		id = types.NewVerificationID()
	}
	queries := mdb.New(tx)
	return queries.CreateVerification(ctx, mdb.CreateVerificationParams{
		VerificationID:   id,
		BackupID:         c.params.BackupID,
		VerifiedAt:       c.params.VerifiedAt,
		VerifiedBy:       c.params.VerifiedBy,
		RestoreTested:    c.params.RestoreTested,
		ChecksumValid:    c.params.ChecksumValid,
		RecordCountMatch: c.params.RecordCountMatch,
		Status:           c.params.Status,
		ErrorMessage:     c.params.ErrorMessage,
		DurationMs:       c.params.DurationMs,
	})
}

func (d Database) NewVerificationCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateVerificationParams) NewVerificationCmd {
	return NewVerificationCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite DELETE -----

type DeleteVerificationCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.VerificationID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteVerificationCmd) Context() context.Context              { return c.ctx }
func (c DeleteVerificationCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteVerificationCmd) Connection() *sql.DB                   { return c.conn }
func (c DeleteVerificationCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteVerificationCmd) TableName() string                     { return "backup_verifications" }
func (c DeleteVerificationCmd) GetID() string                         { return string(c.id) }

func (c DeleteVerificationCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.BackupVerification, error) {
	queries := mdb.New(tx)
	return queries.GetVerification(ctx, mdb.GetVerificationParams{VerificationID: c.id})
}

func (c DeleteVerificationCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteVerification(ctx, mdb.DeleteVerificationParams{VerificationID: c.id})
}

func (d Database) DeleteVerificationCmd(ctx context.Context, auditCtx audited.AuditContext, id types.VerificationID) DeleteVerificationCmd {
	return DeleteVerificationCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

// ===================================================================
// BackupVerification — MySQL
// ===================================================================

// ----- MySQL CREATE -----

type NewVerificationCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateVerificationParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewVerificationCmdMysql) Context() context.Context              { return c.ctx }
func (c NewVerificationCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewVerificationCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c NewVerificationCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewVerificationCmdMysql) TableName() string                     { return "backup_verifications" }
func (c NewVerificationCmdMysql) Params() any                           { return c.params }
func (c NewVerificationCmdMysql) GetID(r mdbm.BackupVerification) string {
	return string(r.VerificationID)
}

func (c NewVerificationCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.BackupVerification, error) {
	id := c.params.VerificationID
	if id.IsZero() {
		id = types.NewVerificationID()
	}
	queries := mdbm.New(tx)
	params := mdbm.CreateVerificationParams{
		VerificationID:   id,
		BackupID:         c.params.BackupID,
		VerifiedAt:       c.params.VerifiedAt,
		VerifiedBy:       c.params.VerifiedBy,
		RestoreTested:    c.params.RestoreTested,
		ChecksumValid:    c.params.ChecksumValid,
		RecordCountMatch: c.params.RecordCountMatch,
		Status:           c.params.Status,
		ErrorMessage:     c.params.ErrorMessage,
		DurationMs:       c.params.DurationMs,
	}
	if err := queries.CreateVerification(ctx, params); err != nil {
		return mdbm.BackupVerification{}, err
	}
	return queries.GetVerification(ctx, mdbm.GetVerificationParams{VerificationID: params.VerificationID})
}

func (d MysqlDatabase) NewVerificationCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateVerificationParams) NewVerificationCmdMysql {
	return NewVerificationCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL DELETE -----

type DeleteVerificationCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.VerificationID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteVerificationCmdMysql) Context() context.Context              { return c.ctx }
func (c DeleteVerificationCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteVerificationCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c DeleteVerificationCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteVerificationCmdMysql) TableName() string                     { return "backup_verifications" }
func (c DeleteVerificationCmdMysql) GetID() string                         { return string(c.id) }

func (c DeleteVerificationCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.BackupVerification, error) {
	queries := mdbm.New(tx)
	return queries.GetVerification(ctx, mdbm.GetVerificationParams{VerificationID: c.id})
}

func (c DeleteVerificationCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteVerification(ctx, mdbm.DeleteVerificationParams{VerificationID: c.id})
}

func (d MysqlDatabase) DeleteVerificationCmd(ctx context.Context, auditCtx audited.AuditContext, id types.VerificationID) DeleteVerificationCmdMysql {
	return DeleteVerificationCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

// ===================================================================
// BackupVerification — PostgreSQL
// ===================================================================

// ----- PostgreSQL CREATE -----

type NewVerificationCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateVerificationParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewVerificationCmdPsql) Context() context.Context              { return c.ctx }
func (c NewVerificationCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewVerificationCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c NewVerificationCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewVerificationCmdPsql) TableName() string                     { return "backup_verifications" }
func (c NewVerificationCmdPsql) Params() any                           { return c.params }
func (c NewVerificationCmdPsql) GetID(r mdbp.BackupVerification) string {
	return string(r.VerificationID)
}

func (c NewVerificationCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.BackupVerification, error) {
	id := c.params.VerificationID
	if id.IsZero() {
		id = types.NewVerificationID()
	}
	queries := mdbp.New(tx)
	return queries.CreateVerification(ctx, mdbp.CreateVerificationParams{
		VerificationID:   id,
		BackupID:         c.params.BackupID,
		VerifiedAt:       c.params.VerifiedAt,
		VerifiedBy:       c.params.VerifiedBy,
		RestoreTested:    c.params.RestoreTested,
		ChecksumValid:    c.params.ChecksumValid,
		RecordCountMatch: c.params.RecordCountMatch,
		Status:           c.params.Status,
		ErrorMessage:     c.params.ErrorMessage,
		DurationMs:       c.params.DurationMs,
	})
}

func (d PsqlDatabase) NewVerificationCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateVerificationParams) NewVerificationCmdPsql {
	return NewVerificationCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL DELETE -----

type DeleteVerificationCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.VerificationID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteVerificationCmdPsql) Context() context.Context              { return c.ctx }
func (c DeleteVerificationCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteVerificationCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c DeleteVerificationCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteVerificationCmdPsql) TableName() string                     { return "backup_verifications" }
func (c DeleteVerificationCmdPsql) GetID() string                         { return string(c.id) }

func (c DeleteVerificationCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.BackupVerification, error) {
	queries := mdbp.New(tx)
	return queries.GetVerification(ctx, mdbp.GetVerificationParams{VerificationID: c.id})
}

func (c DeleteVerificationCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteVerification(ctx, mdbp.DeleteVerificationParams{VerificationID: c.id})
}

func (d PsqlDatabase) DeleteVerificationCmd(ctx context.Context, auditCtx audited.AuditContext, id types.VerificationID) DeleteVerificationCmdPsql {
	return DeleteVerificationCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}
