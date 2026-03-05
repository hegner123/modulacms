-- name: DropAdminContentData :exec
DROP TABLE admin_content_data;

-- name: CreateAdminContentDataTable :exec
CREATE TABLE IF NOT EXISTS admin_content_data (
    admin_content_data_id VARCHAR(26) PRIMARY KEY NOT NULL,
    parent_id VARCHAR(26) NULL,
    first_child_id VARCHAR(26) NULL,
    next_sibling_id VARCHAR(26) NULL,
    prev_sibling_id VARCHAR(26) NULL,
    admin_route_id VARCHAR(26) NOT NULL,
    admin_datatype_id VARCHAR(26) NOT NULL,
    author_id VARCHAR(26) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,
    published_at TIMESTAMP NULL,
    published_by VARCHAR(26) NULL,
    publish_at TIMESTAMP NULL,
    revision INT NOT NULL DEFAULT 0,

    CONSTRAINT fk_admin_content_data_published_by
        FOREIGN KEY (published_by) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_admin_content_data_parent_id
        FOREIGN KEY (parent_id) REFERENCES admin_content_data (admin_content_data_id)
             ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_data_first_child_id
        FOREIGN KEY (first_child_id) REFERENCES admin_content_data (admin_content_data_id)
             ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_data_next_sibling_id
        FOREIGN KEY (next_sibling_id) REFERENCES admin_content_data (admin_content_data_id)
             ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_data_prev_sibling_id
        FOREIGN KEY (prev_sibling_id) REFERENCES admin_content_data (admin_content_data_id)
             ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_data_admin_datatypes
        FOREIGN KEY (admin_datatype_id) REFERENCES admin_datatypes (admin_datatype_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_data_admin_route_id
        FOREIGN KEY (admin_route_id) REFERENCES admin_routes (admin_route_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_data_author_users_user_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE
);


-- name: CountAdminContentData :one
SELECT COUNT(*)
FROM admin_content_data;

-- name: GetAdminContentData :one
SELECT * FROM admin_content_data
WHERE admin_content_data_id = ? LIMIT 1;

-- name: ListAdminContentData :many
SELECT * FROM admin_content_data
ORDER BY admin_content_data_id;

-- name: ListAdminContentDataByRoute :many
SELECT * FROM admin_content_data
WHERE admin_route_id = ?
ORDER BY admin_content_data_id;

-- name: CreateAdminContentData :exec
INSERT INTO admin_content_data (
    admin_content_data_id,
    parent_id,
    first_child_id,
    next_sibling_id,
    prev_sibling_id,
    admin_route_id,
    admin_datatype_id,
    author_id,
    status,
    date_created,
    date_modified
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
);

-- name: UpdateAdminContentData :exec
UPDATE admin_content_data
SET parent_id = ?,
    first_child_id = ?,
    next_sibling_id = ?,
    prev_sibling_id = ?,
    admin_route_id = ?,
    admin_datatype_id = ?,
    author_id = ?,
    status = ?,
    date_created = ?,
    date_modified = ?
WHERE admin_content_data_id = ?;

-- name: DeleteAdminContentData :exec
DELETE FROM admin_content_data
WHERE admin_content_data_id = ?;

-- name: ListAdminContentDataPaginated :many
SELECT * FROM admin_content_data
ORDER BY admin_content_data_id
LIMIT ? OFFSET ?;

-- name: ListAdminContentDataByRoutePaginated :many
SELECT * FROM admin_content_data
WHERE admin_route_id = ?
ORDER BY admin_content_data_id
LIMIT ? OFFSET ?;

-- name: ListAdminContentDataTopLevelPaginated :many
SELECT acd.*, u.name AS author_name, COALESCE(ar.slug, '') AS route_slug, COALESCE(ar.title, '') AS route_title, COALESCE(adt.label, '') AS datatype_label FROM admin_content_data acd
LEFT JOIN admin_datatypes adt ON acd.admin_datatype_id = adt.admin_datatype_id
LEFT JOIN users u ON acd.author_id = u.user_id
LEFT JOIN admin_routes ar ON acd.admin_route_id = ar.admin_route_id
WHERE acd.admin_route_id IS NOT NULL OR adt.type = '_root'
ORDER BY acd.admin_content_data_id
LIMIT ? OFFSET ?;

-- name: CountAdminContentDataTopLevel :one
SELECT COUNT(*) FROM admin_content_data acd
LEFT JOIN admin_datatypes adt ON acd.admin_datatype_id = adt.admin_datatype_id
WHERE acd.admin_route_id IS NOT NULL OR adt.type = '_root';

-- name: UpdateAdminContentDataPublishMeta :exec
UPDATE admin_content_data
SET status = ?,
    published_at = ?,
    published_by = ?,
    revision = revision + 1,
    date_modified = ?
WHERE admin_content_data_id = ?;

-- name: UpdateAdminContentDataWithRevision :exec
UPDATE admin_content_data
SET admin_route_id = ?,
    parent_id = ?,
    first_child_id = ?,
    next_sibling_id = ?,
    prev_sibling_id = ?,
    admin_datatype_id = ?,
    author_id = ?,
    status = ?,
    date_created = ?,
    date_modified = ?
WHERE admin_content_data_id = ? AND revision = ?;

-- name: UpdateAdminContentDataSchedule :exec
UPDATE admin_content_data
SET publish_at = ?,
    date_modified = ?
WHERE admin_content_data_id = ?;

-- name: ClearAdminContentDataSchedule :exec
UPDATE admin_content_data
SET publish_at = NULL,
    date_modified = ?
WHERE admin_content_data_id = ?;

-- name: ListAdminContentDataDueForPublish :many
SELECT * FROM admin_content_data
WHERE publish_at IS NOT NULL AND publish_at <= ? AND status = 'draft';

-- name: GetAdminContentDataDescendants :many
WITH RECURSIVE tree AS (
    SELECT cd1.admin_content_data_id AS cid FROM admin_content_data cd1 WHERE cd1.admin_content_data_id = ?
    UNION ALL
    SELECT cd2.admin_content_data_id FROM admin_content_data cd2
    INNER JOIN tree t ON cd2.parent_id = t.cid
)
SELECT cd.* FROM admin_content_data cd
INNER JOIN tree t ON cd.admin_content_data_id = t.cid;

-- name: ReassignAdminContentDataAuthor :exec
UPDATE admin_content_data SET author_id = ? WHERE author_id = ?;

-- name: CountAdminContentDataByAuthor :one
SELECT COUNT(*) FROM admin_content_data WHERE author_id = ?;
