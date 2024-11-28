-- name: GetField :one
SELECT * FROM field
WHERE id = ? LIMIT 1;

-- name: GetFieldId :one
SELECT id FROM field
WHERE id = ? LIMIT 1;

-- name: ListField :many
SELECT * FROM field
ORDER BY id;

-- name: CreateField :one
INSERT INTO field (
    routeid,
    adminrouteid,
    parentid,
    label,
    data,
    type,
    author,
    authorid,
    datecreated,
    datemodified
    ) VALUES (
?,?, ?,?, ?,?, ?,?, ?,?
    ) RETURNING *;


-- name: UpdateField :exec
UPDATE field
set routeid = ?,
    adminrouteid = ?,
    parentid = ?,
    label = ?,
    data = ?,
    type = ?,
    author = ?,
    authorid = ?,
    datecreated = ?,
    datemodified = ?
    WHERE id = ?
    RETURNING *;

-- name: DeleteField :exec
DELETE FROM field
WHERE id = ?;
