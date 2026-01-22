CREATE TABLE IF NOT EXISTS routes (
    route_id INTEGER
        PRIMARY KEY,
    slug TEXT NOT NULL
        UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL
    REFERENCES users
    ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_routes_author ON routes(author_id);

CREATE TRIGGER IF NOT EXISTS update_routes_modified
    AFTER UPDATE ON routes
    FOR EACH ROW
    BEGIN
        UPDATE routes SET date_modified = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
        WHERE route_id = NEW.route_id;
    END;
