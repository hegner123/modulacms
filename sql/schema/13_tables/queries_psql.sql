-- name: DropTableTable :exec
DROP TABLE tables;

-- name: CreateTablesTable :exec
CREATE TABLE IF NOT EXISTS tables (
    id TEXT PRIMARY KEY NOT NULL,
    label TEXT NOT NULL
        UNIQUE,
    author_id TEXT
        REFERENCES users
            ON UPDATE CASCADE ON DELETE SET NULL
);

-- name: CountTables :one
SELECT COUNT(*)
FROM tables;

-- name: GetTable :one
SELECT * FROM tables
WHERE id = $1
LIMIT 1;


-- name: GetTableId :one
SELECT id FROM tables
WHERE id = $1
LIMIT 1;

-- name: ListTable :many
SELECT * FROM tables 
ORDER BY label;

-- name: CreateTable :one
INSERT INTO tables (
    id,
    label
) VALUES (
    $1,
    $2
)
RETURNING *;

-- name: UpdateTable :exec
UPDATE tables
SET label = $1
WHERE id = $2;

-- name: DeleteTable :exec
DELETE FROM tables
WHERE id = $1;
