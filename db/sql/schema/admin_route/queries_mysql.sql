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
-- name: RecursiveDataTypeJoin :many
WITH RECURSIVE datatype_tree AS (
    SELECT
        dt.id,
        dt.label AS datatype_label,
        dt.type AS datatype_type,
        dt.parentid,
        0 AS level
    FROM datatype dt
    WHERE dt.parentid IS NULL AND dt.adminrouteid = ?

    UNION ALL

    SELECT
        dt.id,
        dt.label AS datatype_label,
        dt.type AS datatype_type,
        dt.parentid,
        datatype_tree.level + 1 AS level
    FROM datatype dt
    INNER JOIN datatype_tree ON dt.parentid = datatype_tree.id
    WHERE dt.adminrouteid = ?
)

SELECT
    substr('..........', 1, level * 3) || datatype_label AS hierarchy,
    datatype_type,
    f.id AS field_id,
    f.label AS field_label,
    f.type AS field_type,
    f.data AS field_data
FROM datatype_tree
LEFT JOIN field f ON f.parentid = datatype_tree.id
WHERE f.adminrouteid = ? OR f.adminrouteid IS NULL
ORDER BY hierarchy;

