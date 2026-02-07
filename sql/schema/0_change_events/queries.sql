-- name: DropChangeEventsTable :exec
DROP TABLE IF EXISTS change_events;

-- name: CreateChangeEventsTable :exec
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

-- name: RecordChangeEvent :one
INSERT INTO change_events (
    event_id,
    hlc_timestamp,
    node_id,
    table_name,
    record_id,
    operation,
    action,
    user_id,
    old_values,
    new_values,
    metadata,
    request_id,
    ip
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING *;

-- name: GetChangeEvent :one
SELECT * FROM change_events
WHERE event_id = ? LIMIT 1;

-- name: GetChangeEventsByRecord :many
SELECT * FROM change_events
WHERE table_name = ? AND record_id = ?
ORDER BY hlc_timestamp DESC;

-- name: GetChangeEventsByRecordPaginated :many
SELECT * FROM change_events
WHERE table_name = ? AND record_id = ?
ORDER BY hlc_timestamp DESC
LIMIT ? OFFSET ?;

-- name: GetUnsyncedEvents :many
SELECT * FROM change_events
WHERE synced_at IS NULL
ORDER BY hlc_timestamp ASC
LIMIT ?;

-- name: GetUnsyncedEventsByNode :many
SELECT * FROM change_events
WHERE synced_at IS NULL AND node_id = ?
ORDER BY hlc_timestamp ASC
LIMIT ?;

-- name: MarkEventSynced :exec
UPDATE change_events
SET synced_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE event_id = ?;

-- name: MarkEventsSyncedBatch :exec
UPDATE change_events
SET synced_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE event_id IN (sqlc.slice('event_ids'));

-- name: GetUnconsumedEvents :many
SELECT * FROM change_events
WHERE consumed_at IS NULL
ORDER BY hlc_timestamp ASC
LIMIT ?;

-- name: MarkEventConsumed :exec
UPDATE change_events
SET consumed_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE event_id = ?;

-- name: MarkEventsConsumedBatch :exec
UPDATE change_events
SET consumed_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE event_id IN (sqlc.slice('event_ids'));

-- name: ListChangeEvents :many
SELECT * FROM change_events
ORDER BY hlc_timestamp DESC
LIMIT ? OFFSET ?;

-- name: ListChangeEventsByUser :many
SELECT * FROM change_events
WHERE user_id = ?
ORDER BY hlc_timestamp DESC
LIMIT ? OFFSET ?;

-- name: ListChangeEventsByAction :many
SELECT * FROM change_events
WHERE action = ?
ORDER BY hlc_timestamp DESC
LIMIT ? OFFSET ?;

-- name: CountChangeEvents :one
SELECT COUNT(*) FROM change_events;

-- name: CountChangeEventsByRecord :one
SELECT COUNT(*) FROM change_events
WHERE table_name = ? AND record_id = ?;

-- name: DeleteChangeEvent :exec
DELETE FROM change_events
WHERE event_id = ?;

-- name: DeleteChangeEventsOlderThan :exec
DELETE FROM change_events
WHERE wall_timestamp < ?
AND synced_at IS NOT NULL
AND consumed_at IS NOT NULL;
