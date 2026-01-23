-- name: DropAdminContentData :exec
DROP TABLE admin_content_data;

-- name: CreateAdminContentDataTable :exec
CREATE TABLE IF NOT EXISTS admin_content_data (
    admin_content_data_id INTEGER PRIMARY KEY,
    parent_id INTEGER,
    first_child_id INTEGER,
    next_sibling_id INTEGER,
    prev_sibling_id INTEGER,
    admin_route_id INTEGER NOT NULL,
    admin_datatype_id INTEGER NOT NULL,
    author_id INTEGER NOT NULL DEFAULT 1,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (parent_id) REFERENCES admin_content_data(admin_content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (first_child_id) REFERENCES admin_content_data(admin_content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (next_sibling_id) REFERENCES admin_content_data(admin_content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (prev_sibling_id) REFERENCES admin_content_data(admin_content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (admin_route_id) REFERENCES admin_routes(admin_route_id) ON DELETE RESTRICT,
    FOREIGN KEY (admin_datatype_id) REFERENCES admin_datatypes(admin_datatype_id) ON DELETE RESTRICT,
    FOREIGN KEY (author_id) REFERENCES users(user_id) ON DELETE SET DEFAULT
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
    parent_id,
    first_child_id,
    next_sibling_id,
    prev_sibling_id,
    admin_route_id,
    admin_datatype_id,
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
    ?,
    ?
) RETURNING *;

-- name: UpdateAdminContentData :exec
UPDATE admin_content_data
SET parent_id = ?,
    first_child_id = ?,
    next_sibling_id = ?,
    prev_sibling_id = ?,
    admin_route_id = ?,
    admin_datatype_id = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?
WHERE admin_content_data_id = ?;

-- name: DeleteAdminContentData :exec
DELETE FROM admin_content_data
WHERE admin_content_data_id = ?;
