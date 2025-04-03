-- name: DropTableTable :exec
DROP TABLE tables;

-- name: CreateTablesTable :exec
CREATE TABLE IF NOT EXISTS tables (
    id INTEGER
        PRIMARY KEY,
    label TEXT NOT NULL
        UNIQUE,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT
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
    label
) VALUES (
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
