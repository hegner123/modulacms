-- name: CreateTablesTable :exec
CREATE TABLE IF NOT EXISTS tables (
    id INT NOT NULL AUTO_INCREMENT,
    label VARCHAR(255) UNIQUE,
    author_id INT NOT NULL DEFAULT 1,
    PRIMARY KEY (id),
    CONSTRAINT fk_tables_author_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE
        ON DELETE NO ACTION
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- name: GetTable :one
SELECT * FROM `tables`
WHERE id = ?
LIMIT 1;

-- name: CountTables :one
SELECT COUNT(*)
FROM `tables`;

-- name: GetTableId :one
SELECT id FROM `tables`
WHERE id = ?
LIMIT 1;

-- name: ListTable :many
SELECT * FROM `tables`
ORDER BY label;

-- name: CreateTable :exec
INSERT INTO `tables` (
    label
) VALUES (
    ?
);
-- To retrieve the newly inserted row, you can run:
-- name: GetLastTable :one
 SELECT * FROM `tables` WHERE id = LAST_INSERT_ID();

-- name: UpdateTable :exec
UPDATE `tables`
SET label = ?
WHERE id = ?;
-- If needed, run a separate SELECT to fetch the updated row.

-- name: DeleteTable :exec
DELETE FROM `tables`
WHERE id = ?;

