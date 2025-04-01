-- name: DropAdminContentField :exec
DROP TABLE admin_content_fields;

-- name: CreateAdminContentFieldTable :exec
CREATE TABLE IF NOT EXISTS admin_content_fields (
    admin_content_field_id INTEGER
        PRIMARY KEY,
    admin_route_id INTEGER NOT NULL
        REFERENCES admin_routes
            ON DELETE SET NULL,
    admin_content_data_id INTEGER NOT NULL
        REFERENCES admin_content_data
            ON DELETE CASCADE,
    admin_field_id INTEGER NOT NULL
        REFERENCES admin_fields
            ON DELETE CASCADE,
    admin_field_value TEXT NOT NULL,
    author_id INTEGER NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);

-- name: CountAdminContentField :one
SELECT COUNT(*)
FROM admin_content_fields;

-- name: GetAdminContentField :one
SELECT * FROM admin_content_fields
WHERE admin_content_field_id = ? LIMIT 1;

-- name: ListAdminContentFields :many
SELECT * FROM admin_content_fields
ORDER BY admin_content_field_id;

-- name: ListAdminContentFieldsByRoute :many
SELECT * FROM admin_content_fields
WHERE admin_route_id = ?
ORDER BY admin_content_field_id;

-- name: CreateAdminContentField :one
INSERT INTO admin_content_fields (
    admin_content_field_id,    
    admin_route_id, 
    admin_content_data_id, 
    admin_field_id,
    admin_field_value,
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
    ?,
    ?,
    ?
) RETURNING *;

-- name: UpdateAdminContentField :exec
UPDATE admin_content_fields
SET  admin_content_field_id = ?,
    admin_route_id = ?,
    admin_content_data_id = ?,
    admin_field_id = ?,
    admin_field_value = ?,
    author_id = ?,
    history = ?,
    date_created = ?,
    date_modified = ?
WHERE admin_content_field_id = ?;

-- name: DeleteAdminContentField :exec
DELETE FROM admin_content_fields
WHERE admin_content_field_id = ?;
