
-- name: GetRoute :one
SELECT * FROM route
WHERE ? = ? LIMIT 1;

-- name: GetRouteId :one
SELECT id FROM route
WHERE ? = ? LIMIT 1;

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
content, 
template
) VALUES (
?,?,?,?,?,?,?,?,?
) RETURNING *;


-- name: UpdateRoute :exec
UPDATE route
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

-- name: DeleteRoute :exec
DELETE FROM route
WHERE ? = ?;
