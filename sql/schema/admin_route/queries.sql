-- name: CreateAdminRouteTable :exec
CREATE TABLE IF NOT EXISTS admin_routes (
    admin_route_id INTEGER
        PRIMARY KEY,
    slug TEXT NOT NULL
        UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    
    author TEXT DEFAULT "system" NOT NULL
        REFERENCES users (username)
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);
-- name: GetAdminRouteBySlug :one
SELECT * FROM admin_routes
WHERE slug = ? LIMIT 1;

-- name: CountAdminroute :one
SELECT COUNT(*)
FROM admin_routes;

-- name: GetAdminRouteById :one
SELECT * FROM admin_routes
WHERE admin_route_id = ? LIMIT 1;

-- name: GetAdminRouteId :one
SELECT admin_route_id FROM admin_routes
WHERE slug = ? LIMIT 1;

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
?,?,?,?,?,?,?,?
    ) RETURNING *;

-- name: UpdateAdminRoute :exec
UPDATE admin_routes
set slug = ?,
    title = ?,
    status = ?,
    author = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?,
    history = ?
    WHERE slug = ?
    RETURNING *;

-- name: DeleteAdminRoute :exec
DELETE FROM admin_routes
WHERE slug = ?;
