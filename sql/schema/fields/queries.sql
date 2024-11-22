-- name: Getfield :one
SELECT * FROM fields
WHERE id = ? LIMIT 1;

-- name: ListFields :many
SELECT * FROM fields
ORDER BY id;

-- name: CreateField :one
INSERT INTO fields (
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
UPDATE fields
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

-- name: DeleteFields :exec
DELETE FROM fields
WHERE id = ?;
