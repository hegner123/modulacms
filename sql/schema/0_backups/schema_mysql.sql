CREATE TABLE IF NOT EXISTS backups (
    backup_id       CHAR(26) PRIMARY KEY,
    node_id         CHAR(26) NOT NULL,
    backup_type     VARCHAR(20) NOT NULL,
    status          VARCHAR(20) NOT NULL,
    started_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
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

CREATE INDEX idx_backups_node ON backups(node_id, started_at DESC);
CREATE INDEX idx_backups_status ON backups(status, started_at DESC);
CREATE INDEX idx_backups_hlc ON backups(hlc_timestamp);

CREATE TABLE IF NOT EXISTS backup_verifications (
    verification_id  CHAR(26) PRIMARY KEY,
    backup_id        CHAR(26) NOT NULL,
    verified_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    verified_by      VARCHAR(64),
    restore_tested   BOOLEAN DEFAULT FALSE,
    checksum_valid   BOOLEAN DEFAULT FALSE,
    record_count_match BOOLEAN DEFAULT FALSE,
    status           VARCHAR(20) NOT NULL,
    error_message    TEXT,
    duration_ms      INTEGER,
    CONSTRAINT chk_verification_status CHECK (status IN ('passed', 'failed')),
    FOREIGN KEY (backup_id) REFERENCES backups(backup_id)
);

CREATE INDEX idx_verifications_backup ON backup_verifications(backup_id, verified_at DESC);

CREATE TABLE IF NOT EXISTS backup_sets (
    backup_set_id    CHAR(26) PRIMARY KEY,
    created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    hlc_timestamp    BIGINT NOT NULL,
    status           VARCHAR(20) NOT NULL,
    backup_ids       JSON NOT NULL,
    node_count       INTEGER NOT NULL,
    completed_count  INTEGER DEFAULT 0,
    error_message    TEXT,
    CONSTRAINT chk_set_status CHECK (status IN ('pending', 'completed', 'failed'))
);

CREATE INDEX idx_backup_sets_time ON backup_sets(created_at DESC);
CREATE INDEX idx_backup_sets_hlc ON backup_sets(hlc_timestamp);
