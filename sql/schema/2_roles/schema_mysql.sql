CREATE TABLE IF NOT EXISTS roles (
    role_id VARCHAR(26) PRIMARY KEY NOT NULL,
    label VARCHAR(255) NOT NULL,
    system_protected BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT label
        UNIQUE (label)
);
