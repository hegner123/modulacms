-- name: DropBackupTables :exec
DROP TABLE IF EXISTS backup_sets;
DROP TABLE IF EXISTS backup_verifications;
DROP TABLE IF EXISTS backups;

-- name: CreateBackupTables :exec
CREATE TABLE IF NOT EXISTS backups (
    backup_id       CHAR(26) PRIMARY KEY,
    node_id         CHAR(26) NOT NULL,
    backup_type     VARCHAR(20) NOT NULL,
    status          VARCHAR(20) NOT NULL,
    started_at      TIMESTAMP NOT NULL,
    completed_at    TIMESTAMP NULL,
    duration_ms     INTEGER,
    record_count    BIGINT,
    size_bytes      BIGINT,
    replication_lsn VARCHAR(64),
    hlc_timestamp   BIGINT,
    storage_path    TEXT NOT NULL,
    checksum        VARCHAR(64),
    triggered_by    VARCHAR(64),
    error_message   TEXT,
    metadata        JSON,
    CONSTRAINT chk_backup_type CHECK (backup_type IN ('full', 'incremental', 'snapshot')),
    CONSTRAINT chk_backup_status CHECK (status IN ('started', 'completed', 'failed', 'verified'))
);

-- Backups CRUD

-- name: CreateBackup :exec
INSERT INTO backups (
    backup_id, node_id, backup_type, status, started_at, storage_path,
    triggered_by, metadata
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: GetBackup :one
SELECT * FROM backups
WHERE backup_id = ? LIMIT 1;

-- name: GetLatestBackup :one
SELECT * FROM backups
WHERE node_id = ? AND status = 'completed'
ORDER BY started_at DESC
LIMIT 1;

-- name: GetLatestBackupByType :one
SELECT * FROM backups
WHERE node_id = ? AND backup_type = ? AND status = 'completed'
ORDER BY started_at DESC
LIMIT 1;

-- name: GetBackupsByNode :many
SELECT * FROM backups
WHERE node_id = ?
ORDER BY started_at DESC
LIMIT ? OFFSET ?;

-- name: GetBackupsByStatus :many
SELECT * FROM backups
WHERE status = ?
ORDER BY started_at DESC
LIMIT ? OFFSET ?;

-- name: GetBackupsByHLCRange :many
SELECT * FROM backups
WHERE hlc_timestamp >= ? AND hlc_timestamp <= ?
ORDER BY hlc_timestamp ASC;

-- name: UpdateBackupStatus :exec
UPDATE backups
SET status = ?,
    completed_at = ?,
    duration_ms = ?,
    record_count = ?,
    size_bytes = ?,
    checksum = ?,
    error_message = ?
WHERE backup_id = ?;

-- name: UpdateBackupHLC :exec
UPDATE backups
SET hlc_timestamp = ?, replication_lsn = ?
WHERE backup_id = ?;

-- name: ListBackups :many
SELECT * FROM backups
ORDER BY started_at DESC
LIMIT ? OFFSET ?;

-- name: CountBackups :one
SELECT COUNT(*) FROM backups;

-- name: CountBackupsByNode :one
SELECT COUNT(*) FROM backups
WHERE node_id = ?;

-- name: DeleteBackup :exec
DELETE FROM backups
WHERE backup_id = ?;

-- name: DeleteOldBackups :exec
DELETE FROM backups
WHERE started_at < ? AND status IN ('completed', 'verified');

-- Backup Verifications CRUD

-- name: CreateVerification :exec
INSERT INTO backup_verifications (
    verification_id, backup_id, verified_at, verified_by,
    restore_tested, checksum_valid, record_count_match,
    status, error_message, duration_ms
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: GetVerification :one
SELECT * FROM backup_verifications
WHERE verification_id = ? LIMIT 1;

-- name: GetVerificationsByBackup :many
SELECT * FROM backup_verifications
WHERE backup_id = ?
ORDER BY verified_at DESC;

-- name: GetLatestVerification :one
SELECT * FROM backup_verifications
WHERE backup_id = ?
ORDER BY verified_at DESC
LIMIT 1;

-- name: ListVerifications :many
SELECT * FROM backup_verifications
ORDER BY verified_at DESC
LIMIT ? OFFSET ?;

-- name: CountVerifications :one
SELECT COUNT(*) FROM backup_verifications;

-- name: DeleteVerification :exec
DELETE FROM backup_verifications
WHERE verification_id = ?;

-- Backup Sets CRUD

-- name: CreateBackupSet :exec
INSERT INTO backup_sets (
    backup_set_id, created_at, hlc_timestamp, status,
    backup_ids, node_count, completed_count, error_message
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: GetBackupSet :one
SELECT * FROM backup_sets
WHERE backup_set_id = ? LIMIT 1;

-- name: GetBackupSetByHLC :one
SELECT * FROM backup_sets
WHERE hlc_timestamp = ?
ORDER BY created_at DESC
LIMIT 1;

-- name: UpdateBackupSetStatus :exec
UPDATE backup_sets
SET status = ?, completed_count = ?, error_message = ?
WHERE backup_set_id = ?;

-- name: IncrementBackupSetCompleted :exec
UPDATE backup_sets
SET completed_count = completed_count + 1
WHERE backup_set_id = ?;

-- name: ListBackupSets :many
SELECT * FROM backup_sets
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: GetPendingBackupSets :many
SELECT * FROM backup_sets
WHERE status = 'pending'
ORDER BY created_at ASC;

-- name: CountBackupSets :one
SELECT COUNT(*) FROM backup_sets;

-- name: DeleteBackupSet :exec
DELETE FROM backup_sets
WHERE backup_set_id = ?;
