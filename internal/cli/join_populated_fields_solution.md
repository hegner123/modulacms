2) Two‐query approach: only populated fields
If you only care about fields that already have values in ContentFields, you can skip the DatatypeFields join entirely:

sql
Copy
Edit
-- (a) load the content row
SELECT
  ContentDataID,
  ParentID,
  RouteID,
  DatatypeID,
  AuthorID,
  DateCreated,
  DateModified,
  History
FROM ContentData
WHERE ContentDataID = :contentDataID;

-- (b) load its fields & definitions
SELECT
  cf.ContentFieldID,
  cf.FieldID,
  cf.FieldValue,
  cf.AuthorID      AS field_author_id,
  cf.DateCreated   AS field_created,
  cf.DateModified  AS field_modified,
  cf.History       AS field_history,

  f.ParentID       AS field_parent_id,
  f.Label          AS field_label,
  f.Data           AS field_data_schema,
  f.Type           AS field_type

FROM ContentFields cf
  INNER JOIN Fields f
    ON f.FieldID = cf.FieldID
WHERE cf.ContentDataID = :contentDataID
ORDER BY cf.ContentFieldID;
Use this if you only want the fields the user has actually populated, rather than the full datatype definition.
