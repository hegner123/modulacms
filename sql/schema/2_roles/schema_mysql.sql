CREATE TABLE IF NOT EXISTS roles (
    role_id VARCHAR(26) PRIMARY KEY NOT NULL,
    label VARCHAR(255) NOT NULL,
    permissions JSON NULL,
    CONSTRAINT label
        UNIQUE (label)
);
