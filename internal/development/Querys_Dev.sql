-- Get Content Data Rows By Route
SELECT
    cd.content_data_id ContentDataID,
    dt.label Label,
    f.label FieldLabel,
    cf.field_value FieldValue
FROM content_data cd, content_fields cf,datatypes dt,fields f
INNER JOIN datatypes d ON f.parent_id = d.datatype_id
LEFT JOIN main.content_data c ON d.datatype_id = c.datatype_id
WHERE cd.route_id=1;

-- Match Content Data To Datatype Definitions
SELECT
    d.label as Label
FROM content_data
    LEFT JOIN main.datatypes d
    ON d.datatype_id = content_data.datatype_id;

-- Match Content Fields to Field Definitions
SELECT
    f.label as Label,
    field_value as Value
    FROM content_fields
    LEFT JOIN main.fields f
    ON f.field_id = content_fields.field_id;

-- Comprehensive query: Get all content structure and values for a route
SELECT 
    cd.content_data_id,
    cd.parent_id,
    dt.label AS datatype_label,
    dt.type AS datatype_type,
    f.label AS field_label,
    f.type AS field_type,
    cf.field_value
FROM content_data cd
    INNER JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
    INNER JOIN datatypes_fields df ON dt.datatype_id = df.datatype_id
    INNER JOIN fields f ON df.field_id = f.field_id
    LEFT JOIN content_fields cf ON cd.content_data_id = cf.content_data_id 
        AND f.field_id = cf.field_id
WHERE cd.route_id = 1
ORDER BY cd.content_data_id, f.field_id;