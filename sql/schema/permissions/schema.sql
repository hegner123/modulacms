CREATE TABLE IF NOT EXISTS permissions (
    permission_id INT PRIMARY KEY,
    table_id INT NOT NULL,
    mode INT NOT NULL,
    label TEXT NOT NULL
);
