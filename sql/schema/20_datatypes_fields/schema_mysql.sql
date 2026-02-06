CREATE TABLE IF NOT EXISTS datatypes_fields (
    id VARCHAR(26) NOT NULL
        PRIMARY KEY,
    datatype_id VARCHAR(26) NOT NULL,
    field_id VARCHAR(26) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    CONSTRAINT fk_df_datatype
        FOREIGN KEY (datatype_id) REFERENCES datatypes (datatype_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_df_field
        FOREIGN KEY (field_id) REFERENCES fields (field_id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX idx_datatypes_fields_datatype ON datatypes_fields(datatype_id);
CREATE INDEX idx_datatypes_fields_field ON datatypes_fields(field_id);
