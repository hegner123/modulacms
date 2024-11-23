-- name: GetAdminRoute :one
SELECT * FROM adminroute
WHERE ? = ? LIMIT 1;

-- name: GetAdminRouteId :one
SELECT id FROM adminroute
WHERE ? = ? LIMIT 1;

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
    WHERE ? = ?
    RETURNING *;

-- name: DeleteAdminRoute :exec
DELETE FROM adminroute
WHERE ? = ?;
