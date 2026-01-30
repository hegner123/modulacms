CREATE TABLE IF NOT EXISTS permissions (
    permission_id VARCHAR(26) PRIMARY KEY NOT NULL,
    table_id VARCHAR(26) NOT NULL,
    mode INT NOT NULL,
    label VARCHAR(255) NOT NULL
);
