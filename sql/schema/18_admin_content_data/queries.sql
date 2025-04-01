-- name: DropAdminContentData :exec
DROP TABLE admin_content_data;

-- name: CreateAdminContentDataTable :exec
CREATE TABLE admin_content_data (
    admin_content_data_id INTEGER
        PRIMARY KEY,
    parent_id INTEGER NOT NULL
        REFERENCES admin_content_data
            ON DELETE CASCADE,
    admin_route_id INTEGER NOT NULL
        REFERENCES admin_routes
            ON DELETE CASCADE,
    admin_datatype_id INTEGER NOT NULL
        REFERENCES admin_datatypes
            ON DELETE SET NULL,
    author_id INTEGER NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT DEFAULT NULL
);

-- name: CountAdminContentData :one
SELECT COUNT(*)
FROM admin_content_data;

-- name: GetAdminContentData :one
SELECT * FROM admin_content_data
WHERE admin_content_data_id = ? LIMIT 1;

-- name: ListAdminContentData :many
SELECT * FROM admin_content_data
ORDER BY admin_content_data_id;

-- name: ListAdminContentDataByRoute :many
SELECT * FROM admin_content_data
WHERE admin_route_id = ?
ORDER BY admin_content_data_id;

-- name: CreateAdminContentData :one
INSERT INTO admin_content_data (
    admin_route_id,
    parent_id,
    admin_datatype_id,
    author_id,
    date_created,
    date_modified,
    history
) VALUES ( 
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
) RETURNING *;

-- name: UpdateAdminContentData :exec
UPDATE admin_content_data
SET admin_route_id = ?,
    parent_id = ?,
    admin_datatype_id = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?,
    history = ?
WHERE admin_content_data_id = ?;

-- name: DeleteAdminContentData :exec
DELETE FROM admin_content_data
WHERE admin_content_data_id = ?;
