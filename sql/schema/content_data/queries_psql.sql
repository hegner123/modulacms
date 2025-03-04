-- name: CreateContentDataTable :exec
CREATE TABLE IF NOT EXISTS content_data (
    content_data_id SERIAL PRIMARY KEY,
    admin_dt_id INTEGER,
    history TEXT DEFAULT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_admin_datatypes FOREIGN KEY (admin_dt_id)
        REFERENCES admin_datatypes(admin_dt_id)
        ON UPDATE CASCADE ON DELETE SET NULL
);
-- name: GetContentData :one
SELECT * FROM content_data
WHERE content_data_id = $1 LIMIT 1;

-- name: CountContentData :one
SELECT COUNT(*)
FROM content_data;


-- name: ListContentData :many
SELECT * FROM content_data
ORDER BY content_data_id;


-- name: CreateContentData :one
INSERT INTO content_data (
    admin_dt_id,
    history,
    date_created,
    date_modified
    ) VALUES (
$1,$2,$3,$4
    ) RETURNING *;


-- name: UpdateContentData :exec
UPDATE content_data
set admin_dt_id = $1,
    history = $2,
    date_created = $3,
    date_modified = $4
    WHERE content_data_id = $5
    RETURNING *;

-- name: DeleteContentData :exec
DELETE FROM content_data
WHERE content_data_id = $1;

-- name: ListFilteredContentData :many
SELECT * FROM content_data
WHERE content_data_id = $1;
