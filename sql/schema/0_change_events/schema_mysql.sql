CREATE TABLE IF NOT EXISTS change_events (
    event_id CHAR(26) PRIMARY KEY,
    hlc_timestamp BIGINT NOT NULL,
    wall_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    node_id CHAR(26) NOT NULL,
    table_name VARCHAR(64) NOT NULL,
    record_id CHAR(26) NOT NULL,
    operation VARCHAR(20) NOT NULL,
    action VARCHAR(20),
    user_id CHAR(26),
    old_values JSON,
    new_values JSON,
    metadata JSON,
    request_id VARCHAR(255),
    ip VARCHAR(45),
    synced_at TIMESTAMP NULL,
    consumed_at TIMESTAMP NULL,
    CONSTRAINT chk_operation CHECK (operation IN ('INSERT', 'UPDATE', 'DELETE'))
);

CREATE INDEX idx_events_record ON change_events(table_name, record_id);
CREATE INDEX idx_events_hlc ON change_events(hlc_timestamp);
CREATE INDEX idx_events_node ON change_events(node_id);
CREATE INDEX idx_events_user ON change_events(user_id);
CREATE INDEX idx_events_unsynced ON change_events((synced_at IS NULL));
CREATE INDEX idx_events_unconsumed ON change_events((consumed_at IS NULL));
