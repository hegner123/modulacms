
-- name: GetTable :one
SELECT * FROM tables
WHERE ? = ? LIMIT 1;

-- name: GetTableId :one
SELECT id FROM tables
WHERE ? = ? LIMIT 1;

-- name: ListTable :many
SELECT * FROM tables 
ORDER BY label;

-- name: CreateTable :one
INSERT INTO tables (
    label
) VALUES (
  ?
)
RETURNING *;

-- name: UpdateTable :exec
UPDATE tables
set label = ?
WHERE ? = ?;

-- name: DeleteTable :exec
DELETE FROM tables
WHERE ? = ?;
