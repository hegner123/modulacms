-- name: DropAdminRouteTable :exec
DROP TABLE admin_routes;

-- name: CreateAdminRouteTable :exec
CREATE TABLE admin_routes (
    admin_route_id SERIAL
        PRIMARY KEY,
    slug TEXT NOT NULL
        UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);

-- name: CreateAdminRouteSlugIndex :exec
CREATE INDEX IF NOT EXISTS idx_admin_routes_slug
ON admin_routes(slug);

-- name: CountAdminroute :one
SELECT COUNT(*)
FROM admin_routes;

-- name: GetAdminRouteBySlug :one
SELECT * FROM admin_routes
WHERE slug = $1 LIMIT 1;


-- name: GetAdminRoute :one
SELECT * FROM admin_routes
WHERE admin_route_id = $1 LIMIT 1;

-- name: GetAdminRouteIdBySlug :one
SELECT admin_route_id FROM admin_routes
WHERE slug = $1 LIMIT 1;

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
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
    ) RETURNING *;

-- name: UpdateAdminRoute :exec
UPDATE admin_routes
SET slug = $1,
    title = $2,
    status = $3, 
    author_id = $4,
    date_created = $5,
    date_modified = $6,
    history = $7
    WHERE slug = $8
    RETURNING *;

-- name: DeleteAdminRoute :exec
DELETE FROM admin_routes
WHERE admin_route_id = $1;
