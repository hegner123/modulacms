CREATE TABLE IF NOT EXISTS content_data (
    content_data_id SERIAL PRIMARY KEY,
    admin_dt_id INTEGER,
    history TEXT DEFAULT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_admin_datatypes FOREIGN KEY (admin_dt_id)
        REFERENCES admin_datatypes(admin_dt_id)
        ON UPDATE CASCADE ON DELETE SET NULL
);

