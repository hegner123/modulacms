-- name: DropTableTable :exec
DROP TABLE tables;

-- name: CreateTablesTable :exec
CREATE TABLE IF NOT EXISTS tables (
    id TEXT
        PRIMARY KEY NOT NULL CHECK (length(id) = 26),
    label TEXT NOT NULL
        UNIQUE,
    author_id TEXT
        REFERENCES users
            ON DELETE SET NULL
);

-- name: CountTables :one
SELECT COUNT(*)
FROM tables;

-- name: GetTable :one
SELECT * FROM tables
WHERE id = ? LIMIT 1;

-- name: GetTableId :one
SELECT id FROM tables
WHERE id = ? LIMIT 1;

-- name: ListTable :many
SELECT * FROM tables 
ORDER BY label;

-- name: CreateTable :one
INSERT INTO tables (
    id,
    label
) VALUES (
    ?,
    ?
)
RETURNING *;

-- name: UpdateTable :exec
UPDATE tables
SET label = ?
WHERE id = ?;

-- name: DeleteTable :exec
DELETE FROM tables
WHERE id = ?;
