CREATE TABLE IF NOT EXISTS permissions (
    permission_id VARCHAR(26) PRIMARY KEY NOT NULL,
    label VARCHAR(255) NOT NULL,
    system_protected BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT perm_label_unique UNIQUE (label)
);
