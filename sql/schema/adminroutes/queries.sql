-- name: GetAdminRoute :one
SELECT * FROM adminroutes
WHERE slug = ? LIMIT 1;

-- name: ListAdminRoute :many
SELECT * FROM adminroutes
ORDER BY slug;

-- name: CreateAdminRoute :one
INSERT INTO adminroutes (
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
UPDATE adminroutes
set slug = ?,
    title = ?,
    status = ?,
    content = ?, 
    template = ?,
    author = ?,
    authorid = ?,
    datecreated = ?,
    datemodified = ?
    WHERE id = ?
    RETURNING *;

-- name: DeleteAdminRoute :exec
DELETE FROM adminroutes
WHERE id = ?;
