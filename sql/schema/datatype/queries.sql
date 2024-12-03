
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
    route_id,
    parent_id,
    label,
    type,
    author,
    author_id,
    date_created,
    date_modified
    ) VALUES (
  ?, ?,?, ?,?, ?,?,?
    ) RETURNING *;


-- name: UpdateDatatype :exec
UPDATE datatype
set route_id = ?,
    parent_id = ?,
    label = ?,
    type = ?,
    author = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?
    WHERE datatype_id = ?
    RETURNING *;

-- name: DeleteDatatype :exec
DELETE FROM datatype
WHERE datatype_id = ?;



-- name: ListDatatypeByRouteId :many
SELECT datatype_id, route_id, parent_id, label, type
FROM datatype
WHERE route_id = ?;
