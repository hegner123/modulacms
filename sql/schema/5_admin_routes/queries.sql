-- name: DropAdminRouteTable :exec
DROP TABLE admin_routes;
-- name: CreateAdminRouteTable :exec
CREATE TABLE IF NOT EXISTS admin_routes (
    admin_route_id TEXT PRIMARY KEY NOT NULL CHECK (length(admin_route_id) = 26),
    slug TEXT NOT NULL
        UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    author_id TEXT NOT NULL
        REFERENCES users
            ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

-- name: CreateAdminRouteSlugIndex :exec
CREATE INDEX IF NOT EXISTS idx_admin_routes_slug
ON admin_routes(slug);

-- name: CountAdminRoute :one
SELECT COUNT(*)
FROM admin_routes;

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

-- name: CreateAdminRoute :one
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
    ) RETURNING *;

-- name: UpdateAdminRoute :exec
UPDATE admin_routes
SET slug = ?, 
    title = ?, 
    status = ?,  
    author_id = ?, 
    date_created = ?, 
    date_modified = ?
    WHERE slug = ?
    RETURNING *;

-- name: DeleteAdminRoute :exec
DELETE FROM admin_routes
WHERE admin_route_id = ?;

-- name: ListAdminRoutePaginated :many
SELECT * FROM admin_routes
ORDER BY slug
LIMIT ? OFFSET ?;
