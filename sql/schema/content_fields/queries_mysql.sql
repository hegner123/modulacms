-- name: CreateContentFieldTable :exec
CREATE TABLE IF NOT EXISTS content_fields (
    content_field_id INT AUTO_INCREMENT PRIMARY KEY,
    route_id INT,
    content_data_id INT NOT NULL,
    field_id INT NOT NULL,
    field_value TEXT NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    history TEXT,
    CONSTRAINT fk_content_field_content_data FOREIGN KEY (content_data_id)
        REFERENCES content_data(content_data_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_content_field_route_id FOREIGN KEY (route_id)
        REFERENCES routes(route_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_content_field_fields FOREIGN KEY (field_id)
        REFERENCES fields(field_id)
        ON UPDATE CASCADE ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
-- name: GetContentField :one
SELECT * FROM content_fields
WHERE content_field_id = ? LIMIT 1;

-- name: CountContentField :one
SELECT COUNT(*)
FROM content_fields;

-- name: ListContentFields :many
SELECT * FROM content_fields
ORDER BY content_fields_id;

-- name: CreateContentField :exec
INSERT INTO content_fields (
    content_field_id,
    route_id,
    content_data_id,
    field_id,
    field_value, 
    history,
    date_created, 
    date_modified
    ) VALUES (
  ?,?,?,?,?,?,?,?
    );
-- name: GetLastContentField :one
SELECT * FROM content_fields WHERE content_fields_id = LAST_INSERT_ID();

-- name: UpdateContentField :exec
UPDATE content_fields
set  content_field_id=?,
    route_id=?,
    content_data_id=?,
    field_id=?,
    field_value=?, 
    history=?,
    date_created=?, 
    date_modified=?
    WHERE content_field_id = ?;

-- name: DeleteContentField :exec
DELETE FROM content_fields
WHERE content_field_id = ?;

-- name: ListContentFieldsByRoute :many
SELECT * FROM content_fields
WHERE route_id = ?;
