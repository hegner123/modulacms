-- name: GetField :one
SELECT * FROM field
WHERE id = ? LIMIT 1;

-- name: GetFieldId :one
SELECT id FROM field
WHERE id = ? LIMIT 1;

-- name: ListField :many
SELECT * FROM field
ORDER BY id;

-- name: ListFieldJoin :many
SELECT 
    f1.*,
    f2.*
FROM 
    field f1
LEFT JOIN 
    field f2
ON 
    f1.fieldid = f2.parentid
WHERE 
    f1.routeid = ?;


-- name: CreateField :one
INSERT INTO field (
    routeid,
    parentid,
    label,
    data,
    type,
    struct,
    author,
    authorid,
    datecreated,
    datemodified
    ) VALUES (
    ?,?,?, ?,?,?, ?,?,?,?
    ) RETURNING *;


-- name: UpdateField :exec
UPDATE field
set routeid = ?,
    parentid = ?,
    label = ?,
    data = ?,
    type = ?,
    struct = ?,
    author = ?,
    authorid = ?,
    datecreated = ?,
    datemodified = ?
    WHERE id = ?
    RETURNING *;

-- name: DeleteField :exec
DELETE FROM field
WHERE id = ?;
