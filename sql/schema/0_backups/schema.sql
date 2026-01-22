CREATE TABLE IF NOT EXISTS backups (
    backup_id       TEXT PRIMARY KEY CHECK (length(backup_id) = 26),
    node_id         TEXT NOT NULL CHECK (length(node_id) = 26),
    backup_type     TEXT NOT NULL CHECK (backup_type IN ('full', 'incremental', 'snapshot')),
    status          TEXT NOT NULL CHECK (status IN ('started', 'completed', 'failed', 'verified')),
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

CREATE INDEX IF NOT EXISTS idx_backups_node ON backups(node_id, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_backups_status ON backups(status, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_backups_hlc ON backups(hlc_timestamp);

CREATE TABLE IF NOT EXISTS backup_verifications (
    verification_id  TEXT PRIMARY KEY CHECK (length(verification_id) = 26),
    backup_id        TEXT NOT NULL REFERENCES backups(backup_id),
    verified_at      TEXT NOT NULL,
    verified_by      TEXT,
    restore_tested   INTEGER DEFAULT 0,
    checksum_valid   INTEGER DEFAULT 0,
    record_count_match INTEGER DEFAULT 0,
    status           TEXT NOT NULL CHECK (status IN ('passed', 'failed')),
    error_message    TEXT,
    duration_ms      INTEGER
);

CREATE INDEX IF NOT EXISTS idx_verifications_backup ON backup_verifications(backup_id, verified_at DESC);

CREATE TABLE IF NOT EXISTS backup_sets (
    backup_set_id    TEXT PRIMARY KEY CHECK (length(backup_set_id) = 26),
    created_at       TEXT NOT NULL,
    hlc_timestamp    INTEGER NOT NULL,
    status           TEXT NOT NULL CHECK (status IN ('pending', 'completed', 'failed')),
    backup_ids       TEXT NOT NULL,
    node_count       INTEGER NOT NULL,
    completed_count  INTEGER DEFAULT 0,
    error_message    TEXT
);

CREATE INDEX IF NOT EXISTS idx_backup_sets_time ON backup_sets(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_backup_sets_hlc ON backup_sets(hlc_timestamp);
