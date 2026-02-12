-- name: DropAdminContentField :exec
DROP TABLE admin_content_fields;

-- name: CreateAdminContentFieldTable :exec
CREATE TABLE IF NOT EXISTS admin_content_fields (
    admin_content_field_id TEXT NOT NULL CHECK (length(admin_content_field_id) = 26),
    admin_route_id TEXT,
    admin_content_data_id TEXT NOT NULL,
    admin_field_id TEXT NOT NULL,
    admin_field_value TEXT NOT NULL,
    author_id TEXT NOT NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (admin_content_field_id),
    FOREIGN KEY (admin_route_id) REFERENCES admin_routes(admin_route_id)
        ON DELETE SET NULL,
    FOREIGN KEY (admin_content_data_id) REFERENCES admin_content_data(admin_content_data_id)
        ON DELETE CASCADE,
    FOREIGN KEY (admin_field_id) REFERENCES admin_fields(admin_field_id)
        ON DELETE CASCADE,
    FOREIGN KEY (author_id) REFERENCES users(user_id)
        ON DELETE SET NULL
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
    date_modified
) VALUES (
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
SET admin_route_id = ?,
    admin_content_data_id = ?,
    admin_field_id = ?,
    admin_field_value = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?
WHERE admin_content_field_id = ?;

-- name: DeleteAdminContentField :exec
DELETE FROM admin_content_fields
WHERE admin_content_field_id = ?;

-- name: ListAdminContentFieldsPaginated :many
SELECT * FROM admin_content_fields
ORDER BY admin_content_field_id
LIMIT ? OFFSET ?;

-- name: ListAdminContentFieldsByRoutePaginated :many
SELECT * FROM admin_content_fields
WHERE admin_route_id = ?
ORDER BY admin_content_field_id
LIMIT ? OFFSET ?;
