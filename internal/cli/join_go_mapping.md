Mapping to Go
Whatever approach you choose, the Go side might look like:

```go
type ContentDataWithFields struct {
  ContentData
  Fields []struct {
    ContentFields
    FieldDef Fields
  }
}

// Example pseudo-code for approach #1:
rows, _ := db.Query(sql, contentDataID)
defer rows.Close()

var parentLoaded bool
var result ContentDataWithFields

for rows.Next() {
  var linkID sql.NullInt64
  var fd ContentFields
  var definition Fields
  // first scan ContentData only once
  if !parentLoaded {
    err := rows.Scan(
      &result.ContentDataID,
      &result.ParentID,
      &result.RouteID,
      &result.DatatypeID,
      &result.AuthorID,
      &result.DateCreated,
      &result.DateModified,
      &result.History,
      &linkID,  /* now into df.ID */
      &definition.FieldID,
      &definition.Label,
      &definition.Data,
      &definition.Type,
      &fd.ContentFieldID,
      &fd.FieldValue,
      &fd.AuthorID,
      &fd.DateCreated,
      &fd.DateModified,
      &fd.History,
    )
    parentLoaded = true
  } else {
    // skip the first 8 columns on subsequent rows
    err := rows.Scan(
      new(interface{}), new(interface{}), new(interface{}), new(interface{}),
      new(interface{}), new(interface{}), new(interface{}), new(interface{}),
      &linkID,
      &definition.FieldID,
      &definition.Label,
      &definition.Data,
      &definition.Type,
      &fd.ContentFieldID,
      &fd.FieldValue,
      &fd.AuthorID,
      &fd.DateCreated,
      &fd.DateModified,
      &fd.History,
    )
  }
  result.Fields = append(result.Fields, struct {
    ContentFields
    FieldDef Fields
  }{fd, definition})
}
```
