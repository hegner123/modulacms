-- name: GetContentField :one
SELECT * FROM content_fields
WHERE content_field_id = ? LIMIT 1;

-- name: CountContentField :one
SELECT COUNT(*)
FROM content_fields;

-- name: ListContentField :many
SELECT * FROM content_fields
ORDER BY content_fields_id;

-- name: CreateContentField :one
INSERT INTO content_fields (
    content_field_id,
    content_data_id,
    admin_field_id,
    field_value, 
    history,
    date_created, 
    date_modified
    ) VALUES (
  ?,?,?,?,?,?,?
    ) RETURNING *;


-- name: UpdateContentField :exec
UPDATE content_fields
set  content_field_id=?,
    content_data_id=?,
    admin_field_id=?,
    field_value=?, 
    history=?,
    date_created=?, 
    date_modified=?
    WHERE content_field_id = ?
    RETURNING *;

-- name: DeleteContentField :exec
DELETE FROM content_fields
WHERE content_field_id = ?;

-- name: ListContentFields :many
SELECT * FROM content_fields
WHERE content_data_id = ?;
