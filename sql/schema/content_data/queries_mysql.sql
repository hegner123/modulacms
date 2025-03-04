-- name: CreateContentDataTable :exec
CREATE TABLE IF NOT EXISTS content_data (
    content_data_id INT AUTO_INCREMENT PRIMARY KEY,
    admin_dt_id INT DEFAULT NULL,
    history TEXT DEFAULT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_content_data_admin_datatypes FOREIGN KEY (admin_dt_id)
        REFERENCES admin_datatypes(admin_dt_id)
        ON UPDATE CASCADE ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
-- name: GetContentData :one
SELECT * FROM content_data
WHERE content_data_id = ? LIMIT 1;

-- name: CountContentData :one
SELECT COUNT(*)
FROM content_data;


-- name: ListContentData :many
SELECT * FROM content_data
ORDER BY content_data_id;


-- name: CreateContentData :exec
INSERT INTO content_data (
    admin_dt_id,
    history,
    date_created,
    date_modified
    ) VALUES (
?,?,?,?
    );
-- name: GetLastContentData :one
SELECT * FROM content_data WHERE content_data_id = LAST_INSERT_ID();


-- name: UpdateContentData :exec
UPDATE content_data
set admin_dt_id = ?,
    history = ?,
    date_created = ?,
    date_modified = ?
    WHERE content_data_id = ?;

-- name: DeleteContentData :exec
DELETE FROM content_data
WHERE content_data_id = ?;

-- name: ListFilteredContentData :many
SELECT * FROM content_data
WHERE content_data_id = ?;
