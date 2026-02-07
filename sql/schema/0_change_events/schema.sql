CREATE TABLE IF NOT EXISTS change_events (
    event_id TEXT PRIMARY KEY CHECK (length(event_id) = 26),
    hlc_timestamp INTEGER NOT NULL,
    wall_timestamp TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    node_id TEXT NOT NULL CHECK (length(node_id) = 26),
    table_name TEXT NOT NULL,
    record_id TEXT NOT NULL CHECK (length(record_id) = 26),
    operation TEXT NOT NULL CHECK (operation IN ('INSERT', 'UPDATE', 'DELETE')),
    action TEXT,
    user_id TEXT CHECK (user_id IS NULL OR length(user_id) = 26),
    old_values TEXT,
    new_values TEXT,
    metadata TEXT,
    request_id TEXT,
    ip TEXT,
    synced_at TEXT,
    consumed_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_events_record ON change_events(table_name, record_id);
CREATE INDEX IF NOT EXISTS idx_events_hlc ON change_events(hlc_timestamp);
CREATE INDEX IF NOT EXISTS idx_events_node ON change_events(node_id);
CREATE INDEX IF NOT EXISTS idx_events_user ON change_events(user_id);
