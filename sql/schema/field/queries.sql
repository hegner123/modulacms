-- name: GetField :one
SELECT * FROM field
WHERE field_id = ? LIMIT 1;

-- name: CountField :one
SELECT COUNT(*)
FROM field;


-- name: ListField :many
SELECT * FROM field
ORDER BY field_id;

-- name: CreateField :one
INSERT INTO field (
    route_id,
    parent_id,
    label,
    data,
    type,
    author,
    author_id,
    date_created,
    date_modified
    ) VALUES (
?,?, ?,?, ?, ?,?, ?,?
    ) RETURNING *;


-- name: UpdateField :exec
UPDATE field
set route_id = ?,
    parent_id = ?,
    label = ?,
    data = ?,
    type = ?,
    author = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?
    WHERE field_id = ?
    RETURNING *;

-- name: DeleteField :exec
DELETE FROM field
WHERE field_id = ?;

-- name: ListFieldByRouteId :many
SELECT field_id, route_id, parent_id, label, data, type
FROM field
WHERE route_id = ?;
