CREATE TABLE IF NOT EXISTS datatypes_fields (
    id SERIAL
        PRIMARY KEY,
    datatype_id INTEGER NOT NULL
        CONSTRAINT fk_df_datatype
            REFERENCES datatypes
            ON UPDATE CASCADE ON DELETE CASCADE,
    field_id INTEGER NOT NULL
        CONSTRAINT fk_df_field
            REFERENCES fields
            ON UPDATE CASCADE ON DELETE CASCADE
);
