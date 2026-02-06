-- name: DropBackupSetsTable :exec
DROP TABLE IF EXISTS backup_sets;

-- name: DropBackupVerificationsTable :exec
DROP TABLE IF EXISTS backup_verifications;

-- name: DropBackupsTable :exec
DROP TABLE IF EXISTS backups;

-- name: DropBackupTables :exec
DROP TABLE IF EXISTS backup_sets;
DROP TABLE IF EXISTS backup_verifications;
DROP TABLE IF EXISTS backups;

-- name: CreateBackupTables :exec
CREATE TABLE IF NOT EXISTS backups (
    backup_id       CHAR(26) PRIMARY KEY,
    node_id         CHAR(26) NOT NULL,
    backup_type     VARCHAR(20) NOT NULL CHECK (backup_type IN ('full', 'incremental', 'snapshot')),
    status          VARCHAR(20) NOT NULL CHECK (status IN ('started', 'completed', 'failed', 'verified')),
    started_at      TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at    TIMESTAMP WITH TIME ZONE,
    duration_ms     INTEGER,
    record_count    BIGINT,
    size_bytes      BIGINT,
    replication_lsn VARCHAR(64),
    hlc_timestamp   BIGINT,
    storage_path    TEXT NOT NULL,
    checksum        VARCHAR(64),
    triggered_by    VARCHAR(64),
    error_message   TEXT,
    metadata        JSONB
);

-- Backups CRUD

-- name: CreateBackup :one
INSERT INTO backups (
    backup_id, node_id, backup_type, status, started_at, storage_path,
    triggered_by, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetBackup :one
SELECT * FROM backups
WHERE backup_id = $1 LIMIT 1;

-- name: GetLatestBackup :one
SELECT * FROM backups
WHERE node_id = $1 AND status = 'completed'
ORDER BY started_at DESC
LIMIT 1;

-- name: GetLatestBackupByType :one
SELECT * FROM backups
WHERE node_id = $1 AND backup_type = $2 AND status = 'completed'
ORDER BY started_at DESC
LIMIT 1;

-- name: GetBackupsByNode :many
SELECT * FROM backups
WHERE node_id = $1
ORDER BY started_at DESC
LIMIT $2 OFFSET $3;

-- name: GetBackupsByStatus :many
SELECT * FROM backups
WHERE status = $1
ORDER BY started_at DESC
LIMIT $2 OFFSET $3;

-- name: GetBackupsByHLCRange :many
SELECT * FROM backups
WHERE hlc_timestamp >= $1 AND hlc_timestamp <= $2
ORDER BY hlc_timestamp ASC;

-- name: UpdateBackupStatus :exec
UPDATE backups
SET status = $1,
    completed_at = $2,
    duration_ms = $3,
    record_count = $4,
    size_bytes = $5,
    checksum = $6,
    error_message = $7
WHERE backup_id = $8;

-- name: UpdateBackupHLC :exec
UPDATE backups
SET hlc_timestamp = $1, replication_lsn = $2
WHERE backup_id = $3;

-- name: ListBackups :many
SELECT * FROM backups
ORDER BY started_at DESC
LIMIT $1 OFFSET $2;

-- name: CountBackups :one
SELECT COUNT(*) FROM backups;

-- name: CountBackupsByNode :one
SELECT COUNT(*) FROM backups
WHERE node_id = $1;

-- name: DeleteBackup :exec
DELETE FROM backups
WHERE backup_id = $1;

-- name: DeleteOldBackups :exec
DELETE FROM backups
WHERE started_at < $1 AND status IN ('completed', 'verified');

-- Backup Verifications CRUD

-- name: CreateVerification :one
INSERT INTO backup_verifications (
    verification_id, backup_id, verified_at, verified_by,
    restore_tested, checksum_valid, record_count_match,
    status, error_message, duration_ms
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: GetVerification :one
SELECT * FROM backup_verifications
WHERE verification_id = $1 LIMIT 1;

-- name: GetVerificationsByBackup :many
SELECT * FROM backup_verifications
WHERE backup_id = $1
ORDER BY verified_at DESC;

-- name: GetLatestVerification :one
SELECT * FROM backup_verifications
WHERE backup_id = $1
ORDER BY verified_at DESC
LIMIT 1;

-- name: ListVerifications :many
SELECT * FROM backup_verifications
ORDER BY verified_at DESC
LIMIT $1 OFFSET $2;

-- name: CountVerifications :one
SELECT COUNT(*) FROM backup_verifications;

-- name: DeleteVerification :exec
DELETE FROM backup_verifications
WHERE verification_id = $1;

-- Backup Sets CRUD

-- name: CreateBackupSet :one
INSERT INTO backup_sets (
    backup_set_id, created_at, hlc_timestamp, status,
    backup_ids, node_count, completed_count, error_message
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetBackupSet :one
SELECT * FROM backup_sets
WHERE backup_set_id = $1 LIMIT 1;

-- name: GetBackupSetByHLC :one
SELECT * FROM backup_sets
WHERE hlc_timestamp = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: UpdateBackupSetStatus :exec
UPDATE backup_sets
SET status = $1, completed_count = $2, error_message = $3
WHERE backup_set_id = $4;

-- name: IncrementBackupSetCompleted :exec
UPDATE backup_sets
SET completed_count = completed_count + 1
WHERE backup_set_id = $1;

-- name: ListBackupSets :many
SELECT * FROM backup_sets
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetPendingBackupSets :many
SELECT * FROM backup_sets
WHERE status = 'pending'
ORDER BY created_at ASC;

-- name: CountBackupSets :one
SELECT COUNT(*) FROM backup_sets;

-- name: DeleteBackupSet :exec
DELETE FROM backup_sets
WHERE backup_set_id = $1;
