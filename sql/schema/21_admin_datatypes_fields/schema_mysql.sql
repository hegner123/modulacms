CREATE TABLE IF NOT EXISTS admin_datatypes_fields (
    id VARCHAR(26) NOT NULL
        PRIMARY KEY,
    admin_datatype_id VARCHAR(26) NOT NULL,
    admin_field_id VARCHAR(26) NOT NULL,
    CONSTRAINT fk_df_admin_datatype
        FOREIGN KEY (admin_datatype_id) REFERENCES admin_datatypes (admin_datatype_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_df_admin_field
        FOREIGN KEY (admin_field_id) REFERENCES admin_fields (admin_field_id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX idx_admin_datatypes_fields_datatype ON admin_datatypes_fields(admin_datatype_id);
CREATE INDEX idx_admin_datatypes_fields_field ON admin_datatypes_fields(admin_field_id);
