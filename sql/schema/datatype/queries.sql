
-- name: GetDatatype :one
SELECT * FROM datatypes
WHERE datatype_id = ? LIMIT 1;

-- name: CountDatatype :one
SELECT COUNT(*)
FROM datatypes;


-- name: ListDatatype :many
SELECT * FROM datatypes
ORDER BY datatype_id;


-- name: CreateDatatype :one
INSERT INTO datatypes (
    route_id,
    parent_id,
    label,
    type,
    author,
    author_id,
    history,
    date_created,
    date_modified
    ) VALUES (
  ?,?,?,?,?,?,?,?,?
    ) RETURNING *;


-- name: UpdateDatatype :exec
UPDATE datatypes
set route_id = ?,
    parent_id = ?,
    label = ?,
    type = ?,
    author = ?,
    author_id = ?,
    history = ?,
    date_created = ?,
    date_modified = ?
    WHERE datatype_id = ?
    RETURNING *;

-- name: DeleteDatatype :exec
DELETE FROM datatypes
WHERE datatype_id = ?;



-- name: ListDatatypeByRouteId :many
SELECT datatype_id, route_id, parent_id, label, type
FROM datatypes
WHERE route_id = ?;
