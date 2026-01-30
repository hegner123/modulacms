-- name: DropTableTable :exec
DROP TABLE tables;

-- name: CreateTablesTable :exec
CREATE TABLE IF NOT EXISTS tables (
    id VARCHAR(26) PRIMARY KEY NOT NULL,
    label VARCHAR(255) NOT NULL,
    author_id VARCHAR(26),
    CONSTRAINT label
        UNIQUE (label),
    CONSTRAINT fk_tables_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE
);

-- name: CountTables :one
SELECT COUNT(*)
FROM tables;

-- name: GetTable :one
SELECT * FROM tables
WHERE id = ?
LIMIT 1;

-- name: GetTableId :one
SELECT id FROM tables
WHERE id = ?
LIMIT 1;

-- name: ListTable :many
SELECT * FROM tables
ORDER BY label;

-- name: CreateTable :exec
INSERT INTO tables (
    id,
    label
) VALUES (
    ?,
    ?
);
-- name: GetLastTable :one
 SELECT * FROM tables WHERE id = LAST_INSERT_ID();

-- name: UpdateTable :exec
UPDATE tables
SET label = ?
WHERE id = ?;
-- If needed, run a separate SELECT to fetch the updated row.

-- name: DeleteTable :exec
DELETE FROM tables
WHERE id = ?;

