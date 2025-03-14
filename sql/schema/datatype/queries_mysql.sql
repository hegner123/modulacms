-- name: CreateDatatypeTable :exec
CREATE TABLE IF NOT EXISTS datatypes (
    datatype_id INT AUTO_INCREMENT PRIMARY KEY,
    
    parent_id INT DEFAULT NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author VARCHAR(255) NOT NULL DEFAULT 'system',
    author_id INT NOT NULL DEFAULT 1,
    history TEXT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_dt_datatypes_parent FOREIGN KEY (parent_id)
        REFERENCES datatypes(datatype_id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_dt_users_author FOREIGN KEY (author)
        REFERENCES users(username)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_dt_users_author_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


-- name: GetDatatype :one
SELECT * FROM datatypes
WHERE datatype_id = ? LIMIT 1;

-- name: CountDatatype :one
SELECT COUNT(*)
FROM datatypes;

-- name: ListDatatype :many
SELECT * FROM datatypes
ORDER BY datatype_id;


-- name: CreateDatatype :exec
INSERT INTO datatypes (    
    parent_id,
    label,
    type,
    author,
    author_id,
    history,
    date_created,
    date_modified
    ) VALUES (
  ?,?,?,?,?,?,?,?
    );
-- name: GetLastDatatype :one
SELECT * FROM datatypes WHERE datatype_id = LAST_INSERT_ID();


-- name: UpdateDatatype :exec
UPDATE datatypes
set 
    parent_id = ?,
    label = ?,
    type = ?,
    author = ?,
    author_id = ?,
    history = ?,
    date_created = ?,
    date_modified = ?
    WHERE datatype_id = ?;

-- name: DeleteDatatype :exec
DELETE FROM datatypes
WHERE datatype_id = ?;



