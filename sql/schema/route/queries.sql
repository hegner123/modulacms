-- name: GetRoute :one
SELECT * FROM route
WHERE slug = ? LIMIT 1;

-- name: CountRoute :one
SELECT COUNT(*)
FROM route;

-- name: GetRouteId :one
SELECT route_id FROM route
WHERE slug = ? LIMIT 1;

-- name: ListRoute :many
SELECT * FROM route
ORDER BY slug;

-- name: CreateRoute :one
INSERT INTO route (
author,
author_id,
slug,
title,
status,
date_created,
date_modified, 
content
) VALUES (
?,?,?,?,?,?,?,?
) RETURNING *;

-- name: UpdateRoute :exec
UPDATE route
set slug = ?,
    title = ?,
    status = ?,
    content = ?, 
    author = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?
    WHERE slug = ?
    RETURNING *;

-- name: DeleteRoute :exec
DELETE FROM route
WHERE slug = ?;
