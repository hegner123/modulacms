-- name: GetTable :one
SELECT * FROM tables
WHERE id = $1
LIMIT 1;

-- name: CountTables :one
SELECT COUNT(*)
FROM tables;

-- name: GetTableId :one
SELECT id FROM tables
WHERE id = $1
LIMIT 1;

-- name: ListTable :many
SELECT * FROM tables 
ORDER BY label;

-- name: CreateTable :one
INSERT INTO tables (
    label
) VALUES (
    $1
)
RETURNING *;

-- name: UpdateTable :exec
UPDATE tables
SET label = $1
WHERE id = $2;

-- name: DeleteTable :exec
DELETE FROM tables
WHERE id = $1;

