-- name: CreateAdminRouteTable :exec
CREATE TABLE IF NOT EXISTS admin_routes (
    admin_route_id INT AUTO_INCREMENT PRIMARY KEY,
    slug VARCHAR(255) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    status INT NOT NULL,
    author VARCHAR(255) NOT NULL DEFAULT 'system',
    author_id INT NOT NULL DEFAULT 1,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    history TEXT,
    CONSTRAINT fk_admin_routes_users_username FOREIGN KEY (author)
        REFERENCES users(username)
        ON UPDATE CASCADE 
        -- ON DELETE SET DEFAULT is not supported in MySQL; consider using RESTRICT or SET NULL instead.
        ON DELETE RESTRICT,
    CONSTRAINT fk_admin_routes_users_user_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE 
        -- ON DELETE SET DEFAULT is not supported in MySQL; consider using RESTRICT instead.
        ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- name: GetAdminRouteBySlug :one
SELECT * FROM admin_routes
WHERE slug = ? LIMIT 1;

-- name: CountAdminroute :one
SELECT COUNT(*) FROM admin_routes;

-- name: GetAdminRouteById :one
SELECT * FROM admin_routes
WHERE admin_route_id = ? LIMIT 1;

-- name: GetAdminRouteId :one
SELECT admin_route_id FROM admin_routes
WHERE slug = ? LIMIT 1;

-- name: ListAdminRoute :many
SELECT * FROM admin_routes
ORDER BY slug;

-- name: CreateAdminRoute :exec
INSERT INTO admin_routes (
    author,
    author_id,
    slug,
    title,
    status,
    date_created,
    date_modified,
    history
) VALUES (
    ?,?,?,?,?,?,?,?
);
-- Note: MySQL does not support RETURNING *; execute a SELECT query afterward if needed.
-- name: GetLastAdminRoute :one
SELECT * FROM admin_routes WHERE admin_route_id = LAST_INSERT_ID();

-- name: UpdateAdminRoute :exec
UPDATE admin_routes
SET slug = ?,
    title = ?,
    status = ?,
    author = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?,
    history = ?
WHERE slug = ?;
-- Note: MySQL does not support RETURNING *; execute a SELECT query afterward if needed.

-- name: DeleteAdminRoute :exec
DELETE FROM admin_routes
WHERE slug = ?;

