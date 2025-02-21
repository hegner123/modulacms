-- name: GetRoute :one
SELECT * FROM routes
WHERE slug = ? LIMIT 1;

-- name: CountRoute :one
SELECT COUNT(*)
FROM routes;

-- name: GetRouteId :one
SELECT route_id FROM routes
WHERE slug = ? LIMIT 1;

-- name: ListRoute :many
SELECT * FROM routes
ORDER BY slug;

-- name: CreateRoute :one
INSERT INTO routes (
author,
author_id,
slug,
title,
status,
history,
date_created,
date_modified
) VALUES (
?,?,?,?,?,?,?,?
) RETURNING *;

-- name: UpdateRoute :exec
UPDATE routes
set slug = ?,
    title = ?,
    status = ?,
    history= ?,
    author = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?
    WHERE slug = ?
    RETURNING *;

-- name: DeleteRoute :exec
DELETE FROM routes
WHERE slug = ?;
