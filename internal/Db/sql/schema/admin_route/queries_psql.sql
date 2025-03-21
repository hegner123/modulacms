-- name: CreateAdminRouteTable :exec
CREATE TABLE IF NOT EXISTS admin_routes (
    admin_route_id SERIAL PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    author TEXT NOT NULL DEFAULT 'system'
        REFERENCES users(username)
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    author_id INTEGER NOT NULL DEFAULT 1
        REFERENCES users(user_id)
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);
-- name: GetAdminRouteBySlug :one
SELECT * FROM admin_routes
WHERE slug = $1 LIMIT 1;

-- name: CountAdminroute :one
SELECT COUNT(*)
FROM admin_routes;

-- name: GetAdminRouteById :one
SELECT * FROM admin_routes
WHERE admin_route_id = $1 LIMIT 1;

-- name: GetAdminRouteId :one
SELECT admin_route_id FROM admin_routes
WHERE slug = $1 LIMIT 1;

-- name: ListAdminRoute :many
SELECT * FROM admin_routes
ORDER BY slug;

-- name: CreateAdminRoute :one
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
$1,$2,$3,$4,$5,$6,$7,$8
    ) RETURNING *;

-- name: UpdateAdminRoute :exec
UPDATE admin_routes
set slug = $1,
    title = $2,
    status = $3,
    author = $4,
    author_id = $5,
    date_created = $6,
    date_modified = $7,
    history = $8
    WHERE slug = $9
    RETURNING *;

-- name: DeleteAdminRoute :exec
DELETE FROM admin_routes
WHERE slug = $1;
