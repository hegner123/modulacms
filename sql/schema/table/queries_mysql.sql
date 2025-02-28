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
 SELECT * FROM `tables` WHERE id = LAST_INSERT_ID();

-- name: UpdateTable :exec
UPDATE `tables`
SET label = ?
WHERE id = ?;
-- If needed, run a separate SELECT to fetch the updated row.

-- name: DeleteTable :exec
DELETE FROM `tables`
WHERE id = ?;

