-- name: DropFieldTypesTable :exec
DROP TABLE field_types;

-- name: CreateFieldTypesTable :exec
CREATE TABLE IF NOT EXISTS field_types (
    field_type_id VARCHAR(26) PRIMARY KEY NOT NULL,
    type VARCHAR(255) NOT NULL,
    label VARCHAR(255) NOT NULL,
    CONSTRAINT field_types_type_unique UNIQUE (type)
);

-- name: GetFieldTypes :one
SELECT * FROM field_types
WHERE field_type_id = ? LIMIT 1;

-- name: GetFieldTypesByType :one
SELECT * FROM field_types
WHERE type = ? LIMIT 1;

-- name: CountFieldTypes :one
SELECT COUNT(*)
FROM field_types;

-- name: ListFieldTypes :many
SELECT * FROM field_types
ORDER BY label;

-- name: CreateFieldTypes :exec
INSERT INTO field_types(
    field_type_id,
    type,
    label
) VALUES (
    ?,
    ?,
    ?
);

-- name: UpdateFieldTypes :exec
UPDATE field_types
SET type=?,
    label=?
WHERE field_type_id = ?;

-- name: DeleteFieldTypes :exec
DELETE FROM field_types
WHERE field_type_id = ?;
