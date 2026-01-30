CREATE TABLE IF NOT EXISTS roles (
    role_id VARCHAR(26) PRIMARY KEY NOT NULL,
    label VARCHAR(255) NOT NULL,
    permissions LONGTEXT COLLATE utf8mb4_bin NULL
        CHECK (JSON_VALID(`permissions`)),
    CONSTRAINT label
        UNIQUE (label)
);
