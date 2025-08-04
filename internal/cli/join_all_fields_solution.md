1) One‐query approach: all fields (with possible null values)
This returns one row per field defined for the content’s datatype. If that field hasn’t yet been given a value for this content‐instance, cf.* columns will be NULL.

sql
Copy
Edit
SELECT
  cd.ContentDataID,
  cd.ParentID,
  cd.RouteID,
  cd.DatatypeID,
  cd.AuthorID      AS content_author_id,
  cd.DateCreated   AS content_created,
  cd.DateModified  AS content_modified,
  cd.History       AS content_history,

  df.ID            AS datatype_field_link_id,
  df.FieldID       AS field_definition_id,

  f.Label          AS field_label,
  f.Data           AS field_data_schema,
  f.Type           AS field_type,

  cf.ContentFieldID AS content_field_id,
  cf.FieldValue     AS field_value,
  cf.AuthorID       AS field_author_id,
  cf.DateCreated    AS field_created,
  cf.DateModified   AS field_modified,
  cf.History        AS field_history

FROM ContentData cd
  /* only fields defined for this datatype */
  INNER JOIN DatatypeFields df
    ON df.DatatypeID = cd.DatatypeID
  /* grab the field definition */
  INNER JOIN Fields f
    ON f.FieldID = df.FieldID
  /* left‐join any existing value for this content row */
  LEFT JOIN ContentFields cf
    ON cf.ContentDataID = cd.ContentDataID
   AND cf.FieldID       = f.FieldID

WHERE cd.ContentDataID = :contentDataID
ORDER BY f.FieldID;
Use this if you want to see every field that could exist for the datatype, and the value (if any) the user has set.

In Go you’d scan the first six columns once into your ContentData struct, and then for each row append a ContentField+Field pair into a slice.


