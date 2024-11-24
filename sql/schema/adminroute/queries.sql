-- name: GetAdminRouteBySlug :one
SELECT * FROM adminroute
WHERE slug = ? LIMIT 1;

-- name: GetAdminRouteById :one
SELECT * FROM adminroute
WHERE id = ? LIMIT 1;

-- name: GetAdminRouteId :one
SELECT id FROM adminroute
WHERE slug = ? LIMIT 1;

-- name: ListAdminRoute :many
SELECT * FROM adminroute
ORDER BY slug;

-- name: CreateAdminRoute :one
INSERT INTO adminroute (
author,
authorid,
slug,
title,
status,
datecreated,
datemodified, 
content, 
template
) VALUES (
?,?,?,?,?,?,?,?,?
) RETURNING *;


-- name: UpdateAdminRoute :exec
UPDATE adminroute
set slug = ?,
    title = ?,
    status = ?,
    content = ?, 
    template = ?,
    author = ?,
    authorid = ?,
    datecreated = ?,
    datemodified = ?
    WHERE slug = ?
    RETURNING *;

-- name: DeleteAdminRoute :exec
DELETE FROM adminroute
WHERE slug = ?;
