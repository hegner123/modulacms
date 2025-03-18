-- name: CreateAdminContentFieldTable :exec
CREATE TABLE IF NOT EXISTS admin_content_fields (
    admin_content_field_id INTEGER PRIMARY KEY,
    admin_route_id       INTEGER NOT NULL
    REFERENCES admin_routes(admin_route_id)
    ON UPDATE CASCADE ON DELETE SET NULL,
    admin_content_data_id       INTEGER NOT NULL
        REFERENCES admin_content_data(admin_content_data_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    admin_field_id      INTEGER NOT NULL
        REFERENCES admin_fields(admin_field_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    admin_field_value         TEXT NOT NULL,
    date_created        TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified       TEXT DEFAULT CURRENT_TIMESTAMP,
    history             TEXT
);

-- name: GetAdminContentField :one
SELECT * FROM admin_content_fields
WHERE admin_content_field_id = ? LIMIT 1;

-- name: CountAdminContentField :one
SELECT COUNT(*)
FROM admin_content_fields;

-- name: ListAdminContentFields :many
SELECT * FROM admin_content_fields
ORDER BY admin_content_field_id;

-- name: CreateAdminContentField :one
INSERT INTO admin_content_fields (
    admin_content_field_id,
    admin_route_id,
    admin_content_data_id,
    admin_field_id,
    admin_field_value, 
    history,
    date_created, 
    date_modified
    ) VALUES (
  ?,?,?,?,?,?,?,?
    ) RETURNING *;


-- name: UpdateAdminContentField :exec
UPDATE admin_content_fields
set  admin_content_field_id=?,
    admin_route_id=?,
    admin_content_data_id=?,
    admin_field_id=?,
    admin_field_value=?, 
    history=?,
    date_created=?, 
    date_modified=?
    WHERE admin_content_field_id = ?
    RETURNING *;

-- name: DeleteAdminContentField :exec
DELETE FROM admin_content_fields
WHERE admin_content_field_id = ?;

-- name: ListAdminContentFieldsByRoute :many
SELECT * FROM admin_content_fields
WHERE admin_route_id = ?;
