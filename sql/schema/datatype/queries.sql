-- name: GetDatatype :one
SELECT * FROM datatype
WHERE id = ? LIMIT 1;

-- name: GetDatatypeId :one
SELECT id FROM datatype
WHERE id = ? LIMIT 1;

-- name: ListDatatype :many
SELECT * FROM datatype
ORDER BY id;

-- name: ListDatatypeJoin :many
SELECT 
    f1.*,
    f2.*
FROM 
    datatype f1
LEFT JOIN 
    datatype f2
ON 
    f1.datatypeid = f2.parentid
WHERE 
    f1.routeid = ?;


-- name: CreateDatatype :one
INSERT INTO datatype (
    routeid,
    parentid,
    label,
    type,
    struct,
    author,
    authorid,
    datecreated,
    datemodified
    ) VALUES (
  ?,?, ?,?,?, ?,?,?,?
    ) RETURNING *;


-- name: UpdateDatatype :exec
UPDATE datatype
set routeid = ?,
    parentid = ?,
    label = ?,
    type = ?,
    struct = ?,
    author = ?,
    authorid = ?,
    datecreated = ?,
    datemodified = ?
    WHERE id = ?
    RETURNING *;

-- name: DeleteDatatype :exec
DELETE FROM datatype
WHERE id = ?;
