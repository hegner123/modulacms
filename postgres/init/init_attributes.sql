CREATE TABLE IF NOT EXISTS attributes (
    id INTEGER PRIMARY KEY,
    elementid INTEGER,
    key TEXT,
    value TEXT,
    FOREIGN KEY (elementid) REFERENCES elements(id)
);
