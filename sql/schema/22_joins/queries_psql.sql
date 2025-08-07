

-- name: GetRouteTreeByRouteID :many
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
WHERE cd.route_id = $1
ORDER BY cd.content_data_id, f.field_id;

