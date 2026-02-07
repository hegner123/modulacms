-- name: DropChangeEventsTable :exec
DROP TABLE IF EXISTS change_events;

-- name: CreateChangeEventsTable :exec
CREATE TABLE IF NOT EXISTS change_events (
    event_id CHAR(26) PRIMARY KEY,
    hlc_timestamp BIGINT NOT NULL,
    wall_timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    node_id CHAR(26) NOT NULL,
    table_name VARCHAR(64) NOT NULL,
    record_id CHAR(26) NOT NULL,
    operation VARCHAR(20) NOT NULL CHECK (operation IN ('INSERT', 'UPDATE', 'DELETE')),
    action VARCHAR(20),
    user_id CHAR(26),
    old_values JSONB,
    new_values JSONB,
    metadata JSONB,
    request_id TEXT,
    ip TEXT,
    synced_at TIMESTAMP WITH TIME ZONE,
    consumed_at TIMESTAMP WITH TIME ZONE
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
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
RETURNING *;

-- name: GetChangeEvent :one
SELECT * FROM change_events
WHERE event_id = $1 LIMIT 1;

-- name: GetChangeEventsByRecord :many
SELECT * FROM change_events
WHERE table_name = $1 AND record_id = $2
ORDER BY hlc_timestamp DESC;

-- name: GetChangeEventsByRecordPaginated :many
SELECT * FROM change_events
WHERE table_name = $1 AND record_id = $2
ORDER BY hlc_timestamp DESC
LIMIT $3 OFFSET $4;

-- name: GetUnsyncedEvents :many
SELECT * FROM change_events
WHERE synced_at IS NULL
ORDER BY hlc_timestamp ASC
LIMIT $1;

-- name: GetUnsyncedEventsByNode :many
SELECT * FROM change_events
WHERE synced_at IS NULL AND node_id = $1
ORDER BY hlc_timestamp ASC
LIMIT $2;

-- name: MarkEventSynced :exec
UPDATE change_events
SET synced_at = CURRENT_TIMESTAMP
WHERE event_id = $1;

-- name: MarkEventsSyncedBatch :exec
UPDATE change_events
SET synced_at = CURRENT_TIMESTAMP
WHERE event_id = ANY($1::char(26)[]);

-- name: GetUnconsumedEvents :many
SELECT * FROM change_events
WHERE consumed_at IS NULL
ORDER BY hlc_timestamp ASC
LIMIT $1;

-- name: MarkEventConsumed :exec
UPDATE change_events
SET consumed_at = CURRENT_TIMESTAMP
WHERE event_id = $1;

-- name: MarkEventsConsumedBatch :exec
UPDATE change_events
SET consumed_at = CURRENT_TIMESTAMP
WHERE event_id = ANY($1::char(26)[]);

-- name: ListChangeEvents :many
SELECT * FROM change_events
ORDER BY hlc_timestamp DESC
LIMIT $1 OFFSET $2;

-- name: ListChangeEventsByUser :many
SELECT * FROM change_events
WHERE user_id = $1
ORDER BY hlc_timestamp DESC
LIMIT $2 OFFSET $3;

-- name: ListChangeEventsByAction :many
SELECT * FROM change_events
WHERE action = $1
ORDER BY hlc_timestamp DESC
LIMIT $2 OFFSET $3;

-- name: CountChangeEvents :one
SELECT COUNT(*) FROM change_events;

-- name: CountChangeEventsByRecord :one
SELECT COUNT(*) FROM change_events
WHERE table_name = $1 AND record_id = $2;

-- name: DeleteChangeEvent :exec
DELETE FROM change_events
WHERE event_id = $1;

-- name: DeleteChangeEventsOlderThan :exec
DELETE FROM change_events
WHERE wall_timestamp < $1
AND synced_at IS NOT NULL
AND consumed_at IS NOT NULL;
