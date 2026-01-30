CREATE TABLE IF NOT EXISTS admin_routes (
    admin_route_id TEXT PRIMARY KEY NOT NULL CHECK (length(admin_route_id) = 26),
    slug TEXT NOT NULL
        UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    author_id TEXT NOT NULL
        REFERENCES users
            ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_routes_author ON admin_routes(author_id);

CREATE TRIGGER IF NOT EXISTS update_admin_routes_modified
    AFTER UPDATE ON admin_routes
    FOR EACH ROW
    BEGIN
        UPDATE admin_routes SET date_modified = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
        WHERE admin_route_id = NEW.admin_route_id;
    END;
