-- name: CreateContentFieldTable :exec
CREATE TABLE IF NOT EXISTS content_fields (
    content_field_id SERIAL PRIMARY KEY,
    content_data_id INTEGER NOT NULL,
    admin_field_id INTEGER NOT NULL,
    field_value TEXT NOT NULL,
    history TEXT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_content_data FOREIGN KEY (content_data_id)
        REFERENCES content_data(content_data_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_fields FOREIGN KEY (admin_field_id)
        REFERENCES admin_fields(admin_field_id)
        ON UPDATE CASCADE ON DELETE CASCADE
);

-- name: GetContentField :one
SELECT * FROM content_fields
WHERE content_field_id = $1 LIMIT 1;

-- name: CountContentField :one
SELECT COUNT(*)
FROM content_fields;

-- name: ListContentField :many
SELECT * FROM content_fields
ORDER BY content_field_id;

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
  $1,$2,$3,$4,$5,$6,$7
    ) RETURNING *;


-- name: UpdateContentField :exec
UPDATE content_fields
set  content_field_id=$1,
    content_data_id=$2,
    admin_field_id=$3,
    field_value=$4, 
    history=$5,
    date_created=$6, 
    date_modified=$7
    WHERE content_field_id = $8
    RETURNING *;

-- name: DeleteContentField :exec
DELETE FROM content_fields
WHERE content_field_id = $1;

-- name: ListContentFields :many
SELECT * FROM content_fields
WHERE content_data_id = $1;
