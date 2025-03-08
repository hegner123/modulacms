-- name: CreateContentFieldTable :exec
CREATE TABLE IF NOT EXISTS content_fields (
    content_field_id SERIAL PRIMARY KEY,
    route_id INTEGER,
    content_data_id INTEGER NOT NULL,
    field_id INTEGER NOT NULL,
    field_value TEXT NOT NULL,
    history TEXT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_route_id FOREIGN KEY (route_id)
        REFERENCES routes(route_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
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

-- name: ListContentFields :many
SELECT * FROM content_fields
ORDER BY content_field_id;

-- name: CreateContentField :one
INSERT INTO content_fields (
    content_field_id,
    route_id,
    content_data_id,
    field_id,
    field_value, 
    history,
    date_created, 
    date_modified
    ) VALUES (
  $1,$2,$3,$4,$5,$6,$7,$8
    ) RETURNING *;


-- name: UpdateContentField :exec
UPDATE content_fields
set  content_field_id=$1,
    route_id=$2,
    content_data_id=$3,
    field_id=$4,
    field_value=$5, 
    history=$6,
    date_created=$7, 
    date_modified=$8
    WHERE content_field_id = $9
    RETURNING *;

-- name: DeleteContentField :exec
DELETE FROM content_fields
WHERE content_field_id = $1;

-- name: ListContentFieldsByRoute :many
SELECT * FROM content_fields
WHERE route_id = $1;
