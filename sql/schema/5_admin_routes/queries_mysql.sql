-- name: DropAdminRouteTable :exec
DROP TABLE admin_routes;

-- name: CreateAdminRouteTable :exec
CREATE TABLE IF NOT EXISTS admin_routes (
    admin_route_id VARCHAR(26) PRIMARY KEY NOT NULL,
    slug VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    status INT NOT NULL,
    author_id VARCHAR(26) NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT slug
        UNIQUE (slug),
    CONSTRAINT fk_admin_routes_users_user_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE
);

-- name: CreateAdminRouteSlugIndex :exec
CREATE INDEX idx_admin_routes_slug
ON admin_routes(slug);

-- name: CountAdminRoute :one
SELECT COUNT(*) FROM admin_routes;

-- name: GetAdminRouteBySlug :one
SELECT * FROM admin_routes
WHERE slug = ? LIMIT 1;

-- name: GetAdminRouteById :one
SELECT * FROM admin_routes
WHERE admin_route_id = ? LIMIT 1;

-- name: GetAdminRouteIdBySlug :one
SELECT admin_route_id FROM admin_routes
WHERE slug = ? LIMIT 1;

-- name: ListAdminRoute :many
SELECT * FROM admin_routes
ORDER BY slug;

-- name: CreateAdminRoute :exec
INSERT INTO admin_routes (
    admin_route_id,
    slug,
    title,
    status,
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
    ?
);

-- name: UpdateAdminRoute :exec
UPDATE admin_routes
SET slug = ?,
    title = ?,
    status = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?
WHERE slug = ?;

-- name: DeleteAdminRoute :exec
DELETE FROM admin_routes
WHERE admin_route_id = ?;

-- name: ListAdminRoutePaginated :many
SELECT * FROM admin_routes
ORDER BY slug
LIMIT ? OFFSET ?;
