
-- name: CreateAdminContentDataTable :exec
CREATE TABLE IF NOT EXISTS admin_content_data (
    admin_content_data_id INTEGER PRIMARY KEY,
    admin_route_id      INTEGER NOT NULL
        REFERENCES admin_routes(admin_route_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    parent_id     INTEGER
        REFERENCES admin_content_data(admin_content_data_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    admin_datatype_id   INTEGER NOT NULL
        REFERENCES admin_datatypes(admin_datatype_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    date_created  TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT  DEFAULT NULL
);

-- name: GetAdminContentData :one
SELECT * FROM admin_content_data
WHERE admin_content_data_id = ? LIMIT 1;

-- name: CountAdminContentData :one
SELECT COUNT(*)
FROM admin_content_data;

-- name: ListAdminContentData :many
SELECT * FROM admin_content_data
ORDER BY admin_content_data_id;


-- name: CreateAdminContentData :one
INSERT INTO admin_content_data (
    admin_route_id,
    parent_id,
    admin_datatype_id,
    history,
    date_created,
    date_modified
    ) VALUES (
?,?,?,?,?,?
    ) RETURNING *;


-- name: UpdateAdminContentData :exec
UPDATE admin_content_data
set admin_route_id = ?,
    parent_id = ?,
    admin_datatype_id = ?,
    history = ?,
    date_created = ?,
    date_modified = ?
    WHERE admin_content_data_id = ?
    RETURNING *;

-- name: DeleteAdminContentData :exec
DELETE FROM admin_content_data
WHERE admin_content_data_id = ?;

-- name: ListAdminContentDataByRoute :many
SELECT * FROM admin_content_data
WHERE admin_route_id = ?;
