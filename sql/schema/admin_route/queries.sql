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
template
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
    template = ?
    WHERE slug = ?
    RETURNING *;

-- name: DeleteAdminRoute :exec
DELETE FROM admin_routes
WHERE slug = ?;
