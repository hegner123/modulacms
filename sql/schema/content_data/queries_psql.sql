-- name: CreateContentDataTable :exec
CREATE TABLE IF NOT EXISTS content_data (
    content_data_id SERIAL PRIMARY KEY,
    route_id INTEGER,
    parent_id INTEGER,
    datatype_id INTEGER,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    history TEXT DEFAULT NULL,
    CONSTRAINT fk_routes FOREIGN KEY (route_id)
        REFERENCES routes(route_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_parent_id FOREIGN KEY (parent_id)
        REFERENCES content_data(content_data_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_datatypes FOREIGN KEY (datatype_id)
        REFERENCES datatypes(datatype_id)
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
    route_id,
    parent_id,
    datatype_id,
    history,
    date_created,
    date_modified
    ) VALUES (
$1,$2,$3,$4,$5,$6
    ) RETURNING *;


-- name: UpdateContentData :exec
UPDATE content_data
set route_id = $1,
    parent_id = $2,
    datatype_id =$3,
    history = $4,
    date_created = $5,
    date_modified = $6
    WHERE content_data_id = $7
    RETURNING *;

-- name: DeleteContentData :exec
DELETE FROM content_data
WHERE content_data_id = $1;

-- name: ListContentDataByRoute :many
SELECT * FROM content_data
WHERE route_id = $1;
