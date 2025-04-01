CREATE TABLE admin_content_data (
    admin_content_data_id SERIAL
        PRIMARY KEY,
    admin_route_id INTEGER
        CONSTRAINT fk_admin_routes
            REFERENCES admin_routes
            ON UPDATE CASCADE ON DELETE SET NULL,
    parent_id INTEGER
        CONSTRAINT fk_parent_id
            REFERENCES admin_content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    admin_datatype_id INTEGER
        CONSTRAINT fk_admin_datatypes
            REFERENCES admin_datatypes
            ON UPDATE CASCADE ON DELETE SET NULL,
    author_id INTEGER NOT NULL
        CONSTRAINT fk_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);

ALTER TABLE admin_content_data
    OWNER TO modula_u;

CREATE INDEX admin_content_data_admin_datatype_id_index
    ON admin_content_data (admin_datatype_id);

CREATE INDEX admin_content_data_admin_route_id_index
    ON admin_content_data (admin_route_id);

CREATE INDEX admin_content_data_parent_id_index
    ON admin_content_data (parent_id);

