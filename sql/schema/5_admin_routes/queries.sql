-- name: DropAdminRouteTable :exec
DROP TABLE admin_routes;
-- name: CreateAdminRouteTable :exec
CREATE TABLE admin_routes (
    admin_route_id INTEGER
        PRIMARY KEY,
    slug TEXT NOT NULL
        UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
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
    slug,
    title,
    status,
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

-- name: UpdateAdminRoute :exec
UPDATE admin_routes
SET slug = ?, 
    title = ?, 
    status = ?,  
    author_id = ?, 
    date_created = ?, 
    date_modified = ?, 
    history = ?
    WHERE slug = ?
    RETURNING *;

-- name: DeleteAdminRoute :exec
DELETE FROM admin_routes
WHERE admin_route_id = ?;
