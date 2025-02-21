CREATE TABLE IF NOT EXISTS content_data (
    content_data_id INTEGER PRIMARY KEY,
    admin_dt_id   INTEGER NOT NULL
        REFERENCES admin_datatypes(admin_dt_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    history TEXT  DEFAULT NULL,
    date_created  TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

