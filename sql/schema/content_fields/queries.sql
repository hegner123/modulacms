-- name: GetDatatype :one
SELECT * FROM content_fields
WHERE content_field_id = ? LIMIT 1;

-- name: CountDatatype :one
SELECT COUNT(*)
FROM content_fields;

-- name: ListDatatype :many
SELECT * FROM content_fields
ORDER BY content_fields_id;

-- name: CreateDatatype :one
INSERT INTO content_fields (
    content_field_id,
    content_data_id,
    admin_field_id,
    field_value, 
    date_created, 
    date_modified,
    ) VALUES (
  ?,?,?,?,?,?
    ) RETURNING *;


-- name: UpdateDatatype :exec
UPDATE content_fields
set  content_field_id=?,
    content_data_id=?,
    admin_field_id=?,
    field_value=?, 
    date_created=?, 
    date_modified=?,
    WHERE content_fields_id = ?
    RETURNING *;

-- name: DeleteDatatype :exec
DELETE FROM content_fields
WHERE content_field_id = ?;

-- name: ListDatatypes :many
SELECT * FROM content_fields
WHERE content_data_id = ?;
