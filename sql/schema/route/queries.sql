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
authorid,
slug,
title,
status,
datecreated,
datemodified, 
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
    authorid = ?,
    datecreated = ?,
    datemodified = ?
    WHERE slug = ?
    RETURNING *;

-- name: DeleteRoute :exec
DELETE FROM route
WHERE slug = ?;
