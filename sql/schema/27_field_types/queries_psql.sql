-- name: DropFieldTypeTable :exec
DROP TABLE field_types;

-- name: CreateFieldTypeTable :exec
CREATE TABLE IF NOT EXISTS field_types (
    field_type_id TEXT PRIMARY KEY NOT NULL CHECK (length(field_type_id) = 26),
    type TEXT NOT NULL UNIQUE,
    label TEXT NOT NULL
);

-- name: GetFieldType :one
SELECT * FROM field_types
WHERE field_type_id = $1 LIMIT 1;

-- name: GetFieldTypesByType :one
SELECT * FROM field_types
WHERE type = $1 LIMIT 1;

-- name: CountFieldType :one
SELECT COUNT(*)
FROM field_types;

-- name: ListFieldType :many
SELECT * FROM field_types
ORDER BY label;

-- name: CreateFieldType :one
INSERT INTO field_types(
    field_type_id,
    type,
    label
) VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: UpdateFieldType :exec
UPDATE field_types
SET type=$1,
    label=$2
WHERE field_type_id = $3;

-- name: DeleteFieldType :exec
DELETE FROM field_types
WHERE field_type_id = $1;
