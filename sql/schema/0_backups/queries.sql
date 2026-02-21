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
    backup_id       TEXT PRIMARY KEY CHECK (length(backup_id) = 26),
    node_id         TEXT NOT NULL CHECK (length(node_id) = 26),
    backup_type     TEXT NOT NULL CHECK (backup_type IN ('full', 'incremental', 'differential')),
    status          TEXT NOT NULL CHECK (status IN ('pending', 'in_progress', 'completed', 'failed')),
    started_at      TEXT NOT NULL,
    completed_at    TEXT,
    duration_ms     INTEGER,
    record_count    INTEGER,
    size_bytes      INTEGER,
    replication_lsn TEXT,
    hlc_timestamp   INTEGER,
    storage_path    TEXT NOT NULL,
    checksum        TEXT,
    triggered_by    TEXT,
    error_message   TEXT,
    metadata        TEXT
);

-- Backups CRUD

-- name: CreateBackup :one
INSERT INTO backups (
    backup_id, node_id, backup_type, status, started_at, storage_path,
    triggered_by, metadata
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING *;

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
WHERE started_at < ? AND status IN ('completed', 'failed');

-- Backup Verifications CRUD

-- name: CreateVerification :one
INSERT INTO backup_verifications (
    verification_id, backup_id, verified_at, verified_by,
    restore_tested, checksum_valid, record_count_match,
    status, error_message, duration_ms
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING *;

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

-- name: CreateBackupSet :one
INSERT INTO backup_sets (
    backup_set_id, date_created, hlc_timestamp, status,
    backup_ids, node_count, completed_count, error_message
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING *;

-- name: GetBackupSet :one
SELECT * FROM backup_sets
WHERE backup_set_id = ? LIMIT 1;

-- name: GetBackupSetByHLC :one
SELECT * FROM backup_sets
WHERE hlc_timestamp = ?
ORDER BY date_created DESC
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
ORDER BY date_created DESC
LIMIT ? OFFSET ?;

-- name: GetPendingBackupSets :many
SELECT * FROM backup_sets
WHERE status = 'pending'
ORDER BY date_created ASC;

-- name: CountBackupSets :one
SELECT COUNT(*) FROM backup_sets;

-- name: DeleteBackupSet :exec
DELETE FROM backup_sets
WHERE backup_set_id = ?;
