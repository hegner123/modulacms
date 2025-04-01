CREATE TABLE datatypes_fields (
    id INTEGER
        PRIMARY KEY,
    datatype_id INTEGER NOT NULL
        CONSTRAINT fk_df_datatype
            REFERENCES datatypes
            ON DELETE CASCADE,
    field_id INTEGER NOT NULL
        CONSTRAINT fk_df_field
            REFERENCES fields
            ON DELETE CASCADE
);

