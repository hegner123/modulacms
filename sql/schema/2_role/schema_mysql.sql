CREATE TABLE IF NOT EXISTS roles (
    role_id INT AUTO_INCREMENT
        PRIMARY KEY,
    label VARCHAR(255) NOT NULL,
    permissions LONGTEXT COLLATE utf8mb4_bin NULL
        CHECK (JSON_VALID(`permissions`)),
    CONSTRAINT label
        UNIQUE (label)
);
