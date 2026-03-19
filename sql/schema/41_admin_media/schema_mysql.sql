CREATE TABLE IF NOT EXISTS admin_media (
    admin_media_id VARCHAR(26) PRIMARY KEY NOT NULL,
    name TEXT NULL,
    display_name TEXT NULL,
    alt TEXT NULL,
    caption TEXT NULL,
    description TEXT NULL,
    class TEXT NULL,
    mimetype TEXT NULL,
    dimensions TEXT NULL,
    url VARCHAR(255) NULL,
    srcset TEXT NULL,
    focal_x FLOAT NULL,
    focal_y FLOAT NULL,
    author_id VARCHAR(26) NULL,
    folder_id VARCHAR(26) NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT admin_media_url
        UNIQUE (url),
    CONSTRAINT fk_admin_media_users_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_admin_media_admin_media_folders_folder_id
        FOREIGN KEY (folder_id) REFERENCES admin_media_folders (admin_folder_id)
            ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE INDEX idx_admin_media_author ON admin_media(author_id);
CREATE INDEX idx_admin_media_folder ON admin_media(folder_id);
