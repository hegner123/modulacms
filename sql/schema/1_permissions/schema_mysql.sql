CREATE TABLE IF NOT EXISTS permissions (
    permission_id INT AUTO_INCREMENT
        PRIMARY KEY,
    table_id INT NOT NULL,
    mode INT NOT NULL,
    label VARCHAR(255) NOT NULL
);
