
-- name: GetShallowTreeByRouteId :many
    SELECT cd.*, dt.label as datatype_label, dt.type as datatype_type
    FROM content_data cd
    JOIN datatypes dt ON cd.datatype_id = dt.datatype_id  
    WHERE cd.route_id = ? 
    AND (cd.parent_id IS NULL OR cd.parent_id IN (
        SELECT content_data_id FROM content_data 
        WHERE parent_id IS NULL AND route_id = ?
    ))
    ORDER BY cd.parent_id NULLS FIRST, cd.content_data_id;
