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

CREATE INDEX IF NOT EXISTS idx_backups_node ON backups(node_id, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_backups_status ON backups(status, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_backups_hlc ON backups(hlc_timestamp);

CREATE TABLE IF NOT EXISTS backup_verifications (
    verification_id  CHAR(26) PRIMARY KEY,
    backup_id        CHAR(26) NOT NULL REFERENCES backups(backup_id),
    verified_at      TIMESTAMP WITH TIME ZONE NOT NULL,
    verified_by      VARCHAR(64),
    restore_tested   BOOLEAN DEFAULT FALSE,
    checksum_valid   BOOLEAN DEFAULT FALSE,
    record_count_match BOOLEAN DEFAULT FALSE,
    status           VARCHAR(20) NOT NULL CHECK (status IN ('passed', 'failed')),
    error_message    TEXT,
    duration_ms      INTEGER
);

CREATE INDEX IF NOT EXISTS idx_verifications_backup ON backup_verifications(backup_id, verified_at DESC);

CREATE TABLE IF NOT EXISTS backup_sets (
    backup_set_id    CHAR(26) PRIMARY KEY,
    date_created       TIMESTAMP WITH TIME ZONE NOT NULL,
    hlc_timestamp    BIGINT NOT NULL,
    status           VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'completed', 'failed')),
    backup_ids       JSONB NOT NULL,
    node_count       INTEGER NOT NULL,
    completed_count  INTEGER DEFAULT 0,
    error_message    TEXT
);

CREATE INDEX IF NOT EXISTS idx_backup_sets_time ON backup_sets(date_created DESC);
CREATE INDEX IF NOT EXISTS idx_backup_sets_hlc ON backup_sets(hlc_timestamp);
