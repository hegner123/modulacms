CREATE TABLE IF NOT EXISTS routes (
    route_id TEXT PRIMARY KEY NOT NULL,
    slug TEXT NOT NULL
        UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    author_id TEXT
        REFERENCES users
            ON UPDATE CASCADE ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_routes_author ON routes(author_id);

CREATE OR REPLACE FUNCTION update_routes_modified()
RETURNS TRIGGER AS $$
BEGIN
    NEW.date_modified = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_routes_modified_trigger
    BEFORE UPDATE ON routes
    FOR EACH ROW
    EXECUTE FUNCTION update_routes_modified();
