
-- name: GetDatatype :one
SELECT * FROM datatype
WHERE datatype_id = ? LIMIT 1;

-- name: CountDatatype :one
SELECT COUNT(*)
FROM datatype;


-- name: ListDatatype :many
SELECT * FROM datatype
ORDER BY datatype_id;


-- name: CreateDatatype :one
INSERT INTO datatype (
    routeid,
    parentid,
    label,
    type,
    author,
    authorid,
    datecreated,
    datemodified
    ) VALUES (
  ?, ?,?, ?,?, ?,?,?
    ) RETURNING *;


-- name: UpdateDatatype :exec
UPDATE datatype
set routeid = ?,
    parentid = ?,
    label = ?,
    type = ?,
    author = ?,
    authorid = ?,
    datecreated = ?,
    datemodified = ?
    WHERE datatype_id = ?
    RETURNING *;

-- name: DeleteDatatype :exec
DELETE FROM datatype
WHERE datatype_id = ?;



-- name: ListDatatypeByRouteId :many
SELECT datatype_id, routeid, parentid, label, type
FROM datatype
WHERE routeid = ?;








