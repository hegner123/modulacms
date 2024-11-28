
-- name: GetDatatype :one
SELECT * FROM datatype
WHERE id = ? LIMIT 1;

-- name: GetDatatypeId :one
SELECT id FROM datatype
WHERE id = ? LIMIT 1;

-- name: ListDatatype :many
SELECT * FROM datatype
ORDER BY id;


-- name: CreateDatatype :one
INSERT INTO datatype (
    routeid,
    adminrouteid,
    parentid,
    label,
    type,
    author,
    authorid,
    datecreated,
    datemodified
    ) VALUES (
  ?,?, ?,?, ?,?, ?,?,?
    ) RETURNING *;


-- name: UpdateDatatype :exec
UPDATE datatype
set routeid = ?,
    adminrouteid = ?,
    parentid = ?,
    label = ?,
    type = ?,
    author = ?,
    authorid = ?,
    datecreated = ?,
    datemodified = ?
    WHERE id = ?
    RETURNING *;

-- name: DeleteDatatype :exec
DELETE FROM datatype
WHERE id = ?;


-- name: RecursiveJoinByRoute :many
WITH RECURSIVE datatype_hierarchy AS (
    -- Anchor member: Select datatypes with the given routeid
    SELECT
        dt.id,
        dt.parentid,
        dt.routeid,
        dt.adminrouteid,
        dt.label,
        dt.type,
        dt.author,
        dt.authorid,
        dt.datecreated,
        dt.datemodified
    FROM
        datatype dt
    WHERE
        dt.routeid = ?

    UNION ALL

    -- Recursive member: Select child datatypes
    SELECT
        child_dt.id,
        child_dt.parentid,
        child_dt.routeid,
        child_dt.adminrouteid,
        child_dt.label,
        child_dt.type,
        child_dt.author,
        child_dt.authorid,
        child_dt.datecreated,
        child_dt.datemodified
    FROM
        datatype child_dt
    INNER JOIN
        datatype_hierarchy dh ON child_dt.parentid = dh.id
)

SELECT
    dh.id AS datatype_id,
    dh.label AS datatype_label,
    dh.type AS datatype_type,
    dh.author AS datatype_author,
    dh.datecreated AS datatype_datecreated,
    dh.datemodified AS datatype_datemodified,
    f.id AS field_id,
    f.label AS field_label,
    f.data AS field_data,
    f.type AS field_type,
    f.author AS field_author,
    f.datecreated AS field_datecreated,
    f.datemodified AS field_datemodified
FROM
    datatype_hierarchy dh
LEFT JOIN
    field f ON f.parentid = dh.id

