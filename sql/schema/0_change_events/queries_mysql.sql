-- name: DropChangeEventsTable :exec
DROP TABLE IF EXISTS change_events;

-- name: CreateChangeEventsTable :exec
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
    synced_at TIMESTAMP NULL,
    consumed_at TIMESTAMP NULL,
    CONSTRAINT chk_operation CHECK (operation IN ('INSERT', 'UPDATE', 'DELETE'))
);

-- name: RecordChangeEvent :exec
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
    metadata
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
);

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
SET synced_at = CURRENT_TIMESTAMP
WHERE event_id = ?;

-- name: MarkEventsSyncedBatch :exec
UPDATE change_events
SET synced_at = CURRENT_TIMESTAMP
WHERE event_id IN (sqlc.slice('event_ids'));

-- name: GetUnconsumedEvents :many
SELECT * FROM change_events
WHERE consumed_at IS NULL
ORDER BY hlc_timestamp ASC
LIMIT ?;

-- name: MarkEventConsumed :exec
UPDATE change_events
SET consumed_at = CURRENT_TIMESTAMP
WHERE event_id = ?;

-- name: MarkEventsConsumedBatch :exec
UPDATE change_events
SET consumed_at = CURRENT_TIMESTAMP
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
