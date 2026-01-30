CREATE TABLE IF NOT EXISTS admin_routes (
    admin_route_id TEXT PRIMARY KEY NOT NULL,
    slug TEXT NOT NULL
        UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    author_id TEXT NOT NULL
        REFERENCES users
            ON UPDATE CASCADE ON DELETE SET NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_routes_author ON admin_routes(author_id);

CREATE OR REPLACE FUNCTION update_admin_routes_modified()
RETURNS TRIGGER AS $$
BEGIN
    NEW.date_modified = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_admin_routes_modified_trigger
    BEFORE UPDATE ON admin_routes
    FOR EACH ROW
    EXECUTE FUNCTION update_admin_routes_modified();
