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
    routeid,
    parentid,
    label,
    data,
    type,
    author,
    authorid,
    datecreated,
    datemodified
    ) VALUES (
?,?, ?,?, ?, ?,?, ?,?
    ) RETURNING *;


-- name: UpdateField :exec
UPDATE field
set routeid = ?,
    parentid = ?,
    label = ?,
    data = ?,
    type = ?,
    author = ?,
    authorid = ?,
    datecreated = ?,
    datemodified = ?
    WHERE field_id = ?
    RETURNING *;

-- name: DeleteField :exec
DELETE FROM field
WHERE field_id = ?;

-- name: ListFieldByRouteId :many
SELECT field_id, routeid, parentid, label, data, type
FROM field
WHERE routeid = ?;
