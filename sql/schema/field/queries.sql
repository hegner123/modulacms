-- name: GetField :one
SELECT * FROM fields 
WHERE field_id = ? LIMIT 1;

-- name: CountField :one
SELECT COUNT(*)
FROM fields ;

-- name: ListField :many
SELECT * FROM fields 
ORDER BY field_id;

-- name: CreateField :one
INSERT INTO fields  (
    route_id,
    parent_id,
    label,
    data,
    type,
    author,
    author_id,
    history,
    date_created,
    date_modified
    ) VALUES (
?,?,?,?,?,?,?,?,?,?
    ) RETURNING *;


-- name: UpdateField :exec
UPDATE fields 
set route_id = ?,
    parent_id = ?,
    label = ?,
    data = ?,
    type = ?,
    author = ?,
    author_id = ?,
    history =?,
    date_created = ?,
    date_modified = ?
    WHERE field_id = ?
    RETURNING *;

-- name: DeleteField :exec
DELETE FROM fields 
WHERE field_id = ?;

-- name: ListFieldByRouteId :many
SELECT field_id, route_id, parent_id, label, data, type
FROM fields 
WHERE route_id = ?;
