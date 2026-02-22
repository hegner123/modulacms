package db

import (
	"encoding/json"
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

// --- Mapper unit tests ---

func TestMapAuthorView(t *testing.T) {
	t.Parallel()

	user := Users{
		UserID:       types.UserID("01HQXYZ1234567890ABCDEFGH"),
		Username:     "jdoe",
		Name:         "Jane Doe",
		Email:        types.Email("jane@example.com"),
		Hash:         "secret-hash-should-not-appear",
		Role:         "admin",
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	}

	av := MapAuthorView(user)

	if av.UserID != user.UserID {
		t.Errorf("UserID = %v, want %v", av.UserID, user.UserID)
	}
	if av.Username != user.Username {
		t.Errorf("Username = %q, want %q", av.Username, user.Username)
	}
	if av.Name != user.Name {
		t.Errorf("Name = %q, want %q", av.Name, user.Name)
	}
	if av.Email != user.Email {
		t.Errorf("Email = %v, want %v", av.Email, user.Email)
	}
	if av.Role != user.Role {
		t.Errorf("Role = %q, want %q", av.Role, user.Role)
	}

	// Verify hash is excluded: marshal to JSON and confirm no "hash" key
	data, err := json.Marshal(av)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if _, ok := m["hash"]; ok {
		t.Error("AuthorView JSON contains 'hash' key; expected it to be excluded")
	}
}

func TestMapDatatypeView(t *testing.T) {
	t.Parallel()

	dt := Datatypes{
		DatatypeID:   types.NewDatatypeID(),
		Label:        "blog-post",
		Type:         "page",
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	}

	dv := MapDatatypeView(dt)

	if dv.DatatypeID != dt.DatatypeID {
		t.Errorf("DatatypeID = %v, want %v", dv.DatatypeID, dt.DatatypeID)
	}
	if dv.Label != dt.Label {
		t.Errorf("Label = %q, want %q", dv.Label, dt.Label)
	}
	if dv.Type != dt.Type {
		t.Errorf("Type = %q, want %q", dv.Type, dt.Type)
	}
}

func TestMapFieldView(t *testing.T) {
	t.Parallel()

	fieldID := types.NewFieldID()
	cf := ContentFields{
		ContentFieldID: types.NewContentFieldID(),
		FieldID:        types.NullableFieldID{ID: fieldID, Valid: true},
		FieldValue:     "Hello World",
	}
	f := Fields{
		FieldID: fieldID,
		Label:   "title",
		Type:    types.FieldTypeText,
	}

	fv := MapFieldView(cf, f)

	if fv.FieldID != f.FieldID {
		t.Errorf("FieldID = %v, want %v", fv.FieldID, f.FieldID)
	}
	if fv.Label != f.Label {
		t.Errorf("Label = %q, want %q", fv.Label, f.Label)
	}
	if fv.Type != f.Type {
		t.Errorf("Type = %v, want %v", fv.Type, f.Type)
	}
	if fv.Value != cf.FieldValue {
		t.Errorf("Value = %q, want %q", fv.Value, cf.FieldValue)
	}
}

func TestMapFieldViewFromRow(t *testing.T) {
	t.Parallel()

	row := ContentFieldWithFieldRow{
		FFieldID:   types.NewFieldID(),
		FLabel:     "body",
		FType:      types.FieldTypeTextarea,
		FieldValue: "Some content here",
	}

	fv := MapFieldViewFromRow(row)

	if fv.FieldID != row.FFieldID {
		t.Errorf("FieldID = %v, want %v", fv.FieldID, row.FFieldID)
	}
	if fv.Label != row.FLabel {
		t.Errorf("Label = %q, want %q", fv.Label, row.FLabel)
	}
	if fv.Type != row.FType {
		t.Errorf("Type = %v, want %v", fv.Type, row.FType)
	}
	if fv.Value != row.FieldValue {
		t.Errorf("Value = %q, want %q", fv.Value, row.FieldValue)
	}
}

// --- GroupBy tests ---

func TestGroupBy(t *testing.T) {
	t.Parallel()

	items := []struct {
		Group string
		Value int
	}{
		{"a", 1},
		{"b", 2},
		{"a", 3},
		{"c", 4},
		{"b", 5},
	}

	grouped := GroupBy(items, func(item struct {
		Group string
		Value int
	}) string {
		return item.Group
	})

	if len(grouped) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(grouped))
	}
	if len(grouped["a"]) != 2 {
		t.Errorf("group 'a' has %d items, want 2", len(grouped["a"]))
	}
	if len(grouped["b"]) != 2 {
		t.Errorf("group 'b' has %d items, want 2", len(grouped["b"]))
	}
	if len(grouped["c"]) != 1 {
		t.Errorf("group 'c' has %d items, want 1", len(grouped["c"]))
	}
}

func TestGroupBy_Empty(t *testing.T) {
	t.Parallel()

	var items []string
	grouped := GroupBy(items, func(s string) string { return s })

	if len(grouped) != 0 {
		t.Errorf("expected empty map, got %d entries", len(grouped))
	}
}

// --- AssembleFieldViews tests ---

func TestAssembleFieldViews(t *testing.T) {
	t.Parallel()

	rows := []ContentFieldWithFieldRow{
		{FFieldID: types.NewFieldID(), FLabel: "title", FType: types.FieldTypeText, FieldValue: "Hello"},
		{FFieldID: types.NewFieldID(), FLabel: "body", FType: types.FieldTypeTextarea, FieldValue: "World"},
	}

	views := AssembleFieldViews(rows)

	if len(views) != 2 {
		t.Fatalf("expected 2 views, got %d", len(views))
	}
	if views[0].Label != "title" {
		t.Errorf("views[0].Label = %q, want %q", views[0].Label, "title")
	}
	if views[1].Value != "World" {
		t.Errorf("views[1].Value = %q, want %q", views[1].Value, "World")
	}
}

func TestAssembleFieldViews_Empty(t *testing.T) {
	t.Parallel()

	views := AssembleFieldViews([]ContentFieldWithFieldRow{})

	if views == nil {
		t.Fatal("expected non-nil empty slice, got nil")
	}
	if len(views) != 0 {
		t.Errorf("expected 0 views, got %d", len(views))
	}
}

// --- Integration tests ---

func TestAssembleContentDataView(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	now := types.TimestampNow()

	routeID := types.NullableRouteID{ID: seed.Route.RouteID, Valid: true}
	datatypeID := types.NullableDatatypeID{ID: seed.Datatype.DatatypeID, Valid: true}
	cdAuthorID := seed.User.UserID // ContentData uses types.UserID (non-nullable)
	authorID := seed.User.UserID   // ContentFields also uses types.UserID (non-nullable)

	// Link the field to the datatype
	_, err := d.CreateDatatypeField(ctx, ac, CreateDatatypeFieldParams{
		DatatypeID: seed.Datatype.DatatypeID,
		FieldID:    seed.Field.FieldID,
		SortOrder:  1,
	})
	if err != nil {
		t.Fatalf("CreateDatatypeField: %v", err)
	}

	// Create content data
	cd, err := d.CreateContentData(ctx, ac, CreateContentDataParams{
		RouteID:       routeID,
		ParentID:      types.NullableContentID{},
		FirstChildID:  types.NullableContentID{},
		NextSiblingID: types.NullableContentID{},
		PrevSiblingID: types.NullableContentID{},
		DatatypeID:    datatypeID,
		AuthorID:      cdAuthorID,
		Status:        types.ContentStatusDraft,
		DateCreated:   now,
		DateModified:  now,
	})
	if err != nil {
		t.Fatalf("CreateContentData: %v", err)
	}

	contentDataID := types.NullableContentID{ID: cd.ContentDataID, Valid: true}
	fieldID := types.NullableFieldID{ID: seed.Field.FieldID, Valid: true}

	// Create a content field
	_, err = d.CreateContentField(ctx, ac, CreateContentFieldParams{
		RouteID:       routeID,
		ContentDataID: contentDataID,
		FieldID:       fieldID,
		FieldValue:    "Test Value",
		AuthorID:      authorID,
		DateCreated:   now,
		DateModified:  now,
	})
	if err != nil {
		t.Fatalf("CreateContentField: %v", err)
	}

	// Assemble the view
	view, err := AssembleContentDataView(d, cd.ContentDataID)
	if err != nil {
		t.Fatalf("AssembleContentDataView: %v", err)
	}

	// Verify main fields
	if view.ContentDataID != cd.ContentDataID {
		t.Errorf("ContentDataID = %v, want %v", view.ContentDataID, cd.ContentDataID)
	}
	if view.Status != types.ContentStatusDraft {
		t.Errorf("Status = %v, want %v", view.Status, types.ContentStatusDraft)
	}

	// Verify author is embedded
	if view.Author == nil {
		t.Fatal("Author is nil, expected embedded AuthorView")
	}
	if view.Author.UserID != seed.User.UserID {
		t.Errorf("Author.UserID = %v, want %v", view.Author.UserID, seed.User.UserID)
	}
	if view.Author.Username != "testuser" {
		t.Errorf("Author.Username = %q, want %q", view.Author.Username, "testuser")
	}

	// Verify datatype is embedded
	if view.Datatype == nil {
		t.Fatal("Datatype is nil, expected embedded DatatypeView")
	}
	if view.Datatype.DatatypeID != seed.Datatype.DatatypeID {
		t.Errorf("Datatype.DatatypeID = %v, want %v", view.Datatype.DatatypeID, seed.Datatype.DatatypeID)
	}
	if view.Datatype.Label != "test-datatype" {
		t.Errorf("Datatype.Label = %q, want %q", view.Datatype.Label, "test-datatype")
	}

	// Verify fields
	if len(view.Fields) != 1 {
		t.Fatalf("Fields length = %d, want 1", len(view.Fields))
	}
	if view.Fields[0].Value != "Test Value" {
		t.Errorf("Fields[0].Value = %q, want %q", view.Fields[0].Value, "Test Value")
	}
	if view.Fields[0].Label != "test-field" {
		t.Errorf("Fields[0].Label = %q, want %q", view.Fields[0].Label, "test-field")
	}

	// Verify JSON structure has nested objects
	data, err := json.Marshal(view)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if _, ok := m["author"]; !ok {
		t.Error("JSON missing 'author' key")
	}
	if _, ok := m["datatype"]; !ok {
		t.Error("JSON missing 'datatype' key")
	}
	if _, ok := m["fields"]; !ok {
		t.Error("JSON missing 'fields' key")
	}
}

func TestContentDataView_NilAuthor_OmittedFromJSON(t *testing.T) {
	// content_data.author_id is NOT NULL in the schema, so a true null-author
	// integration test is impossible. This unit test verifies the omitempty
	// behavior on ContentDataView when Author is nil.
	t.Parallel()

	view := ContentDataView{
		ContentDataID: types.ContentID("01HQXYZ1234567890ABCDEFGH"),
		Status:        types.ContentStatusDraft,
		DateCreated:   types.TimestampNow(),
		DateModified:  types.TimestampNow(),
		Author:        nil,
		Datatype:      nil,
		Fields:        []FieldView{},
	}

	data, err := json.Marshal(view)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if _, ok := m["author"]; ok {
		t.Error("JSON contains 'author' key; expected omitempty to exclude nil author")
	}
	if _, ok := m["datatype"]; ok {
		t.Error("JSON contains 'datatype' key; expected omitempty to exclude nil datatype")
	}
	// fields should still be present as empty array
	fieldsVal, ok := m["fields"]
	if !ok {
		t.Fatal("JSON missing 'fields' key")
	}
	arr, ok := fieldsVal.([]any)
	if !ok {
		t.Fatalf("fields is %T, want []any", fieldsVal)
	}
	if len(arr) != 0 {
		t.Errorf("fields array length = %d, want 0", len(arr))
	}
}

func TestAssembleContentDataView_NoFields(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	now := types.TimestampNow()

	routeID := types.NullableRouteID{ID: seed.Route.RouteID, Valid: true}
	datatypeID := types.NullableDatatypeID{ID: seed.Datatype.DatatypeID, Valid: true}
	authorID := seed.User.UserID

	// Create content data with no content fields
	cd, err := d.CreateContentData(ctx, ac, CreateContentDataParams{
		RouteID:       routeID,
		ParentID:      types.NullableContentID{},
		FirstChildID:  types.NullableContentID{},
		NextSiblingID: types.NullableContentID{},
		PrevSiblingID: types.NullableContentID{},
		DatatypeID:    datatypeID,
		AuthorID:      authorID,
		Status:        types.ContentStatusDraft,
		DateCreated:   now,
		DateModified:  now,
	})
	if err != nil {
		t.Fatalf("CreateContentData: %v", err)
	}

	view, err := AssembleContentDataView(d, cd.ContentDataID)
	if err != nil {
		t.Fatalf("AssembleContentDataView: %v", err)
	}

	// Fields should be empty slice, not nil
	if view.Fields == nil {
		t.Fatal("Fields is nil, expected empty slice")
	}
	if len(view.Fields) != 0 {
		t.Errorf("Fields length = %d, want 0", len(view.Fields))
	}

	// Verify JSON produces [] not null for fields
	data, err := json.Marshal(view)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	fieldsVal, ok := m["fields"]
	if !ok {
		t.Fatal("JSON missing 'fields' key")
	}
	arr, ok := fieldsVal.([]any)
	if !ok {
		t.Fatalf("fields is %T, want []any", fieldsVal)
	}
	if len(arr) != 0 {
		t.Errorf("fields array length = %d, want 0", len(arr))
	}
}

// --- DatatypeFullView tests ---

func TestMapDatatypeFieldView(t *testing.T) {
	t.Parallel()

	row := FieldWithSortOrderRow{
		SortOrder:  3,
		FieldID:    types.NewFieldID(),
		Label:      "title",
		Type:       types.FieldTypeText,
		Data:       "{}",
		Validation: "{}",
		UIConfig:   "{}",
	}

	fv := MapDatatypeFieldView(row)

	if fv.FieldID != row.FieldID {
		t.Errorf("FieldID = %v, want %v", fv.FieldID, row.FieldID)
	}
	if fv.Label != row.Label {
		t.Errorf("Label = %q, want %q", fv.Label, row.Label)
	}
	if fv.Type != row.Type {
		t.Errorf("Type = %v, want %v", fv.Type, row.Type)
	}
	if fv.Data != row.Data {
		t.Errorf("Data = %q, want %q", fv.Data, row.Data)
	}
	if fv.Validation != row.Validation {
		t.Errorf("Validation = %q, want %q", fv.Validation, row.Validation)
	}
	if fv.UIConfig != row.UIConfig {
		t.Errorf("UIConfig = %q, want %q", fv.UIConfig, row.UIConfig)
	}
	if fv.SortOrder != row.SortOrder {
		t.Errorf("SortOrder = %d, want %d", fv.SortOrder, row.SortOrder)
	}
}

func TestAssembleDatatypeFullView(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)
	ctx := d.Context
	ac := testAuditCtxWithUser(d, seed.User.UserID)
	now := types.TimestampNow()
	authorID := types.NullableUserID{ID: seed.User.UserID, Valid: true}

	// Create a second field
	field2, err := d.CreateField(ctx, ac, CreateFieldParams{
		FieldID:      types.NewFieldID(),
		ParentID:     types.NullableDatatypeID{},
		Label:        "body",
		Data:         "{}",
		Validation:   "{}",
		UIConfig:     "{}",
		Type:         types.FieldTypeTextarea,
		AuthorID:     authorID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("CreateField: %v", err)
	}

	// Link both fields to the datatype
	_, err = d.CreateDatatypeField(ctx, ac, CreateDatatypeFieldParams{
		DatatypeID: seed.Datatype.DatatypeID,
		FieldID:    seed.Field.FieldID,
		SortOrder:  1,
	})
	if err != nil {
		t.Fatalf("CreateDatatypeField[0]: %v", err)
	}
	_, err = d.CreateDatatypeField(ctx, ac, CreateDatatypeFieldParams{
		DatatypeID: seed.Datatype.DatatypeID,
		FieldID:    field2.FieldID,
		SortOrder:  2,
	})
	if err != nil {
		t.Fatalf("CreateDatatypeField[1]: %v", err)
	}

	view, err := AssembleDatatypeFullView(d, seed.Datatype.DatatypeID)
	if err != nil {
		t.Fatalf("AssembleDatatypeFullView: %v", err)
	}

	// Verify main fields
	if view.DatatypeID != seed.Datatype.DatatypeID {
		t.Errorf("DatatypeID = %v, want %v", view.DatatypeID, seed.Datatype.DatatypeID)
	}
	if view.Label != "test-datatype" {
		t.Errorf("Label = %q, want %q", view.Label, "test-datatype")
	}
	if view.Type != "page" {
		t.Errorf("Type = %q, want %q", view.Type, "page")
	}

	// Verify author
	if view.Author == nil {
		t.Fatal("Author is nil, expected embedded AuthorView")
	}
	if view.Author.UserID != seed.User.UserID {
		t.Errorf("Author.UserID = %v, want %v", view.Author.UserID, seed.User.UserID)
	}
	if view.Author.Username != "testuser" {
		t.Errorf("Author.Username = %q, want %q", view.Author.Username, "testuser")
	}

	// Verify fields
	if len(view.Fields) != 2 {
		t.Fatalf("Fields length = %d, want 2", len(view.Fields))
	}
	if view.Fields[0].Label != "test-field" {
		t.Errorf("Fields[0].Label = %q, want %q", view.Fields[0].Label, "test-field")
	}
	if view.Fields[0].SortOrder != 1 {
		t.Errorf("Fields[0].SortOrder = %d, want 1", view.Fields[0].SortOrder)
	}
	if view.Fields[1].Label != "body" {
		t.Errorf("Fields[1].Label = %q, want %q", view.Fields[1].Label, "body")
	}
	if view.Fields[1].SortOrder != 2 {
		t.Errorf("Fields[1].SortOrder = %d, want 2", view.Fields[1].SortOrder)
	}

	// Verify JSON structure
	data, err := json.Marshal(view)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if _, ok := m["author"]; !ok {
		t.Error("JSON missing 'author' key")
	}
	if _, ok := m["fields"]; !ok {
		t.Error("JSON missing 'fields' key")
	}
	if _, ok := m["datatype_id"]; !ok {
		t.Error("JSON missing 'datatype_id' key")
	}
}

func TestAssembleDatatypeFullView_EmptyFields(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)

	// Datatype exists but has no fields linked via datatypes_fields
	view, err := AssembleDatatypeFullView(d, seed.Datatype.DatatypeID)
	if err != nil {
		t.Fatalf("AssembleDatatypeFullView: %v", err)
	}

	if view.Fields == nil {
		t.Fatal("Fields is nil, expected empty slice")
	}
	if len(view.Fields) != 0 {
		t.Errorf("Fields length = %d, want 0", len(view.Fields))
	}

	// Verify JSON produces [] not null for fields
	data, err := json.Marshal(view)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	fieldsVal, ok := m["fields"]
	if !ok {
		t.Fatal("JSON missing 'fields' key")
	}
	arr, ok := fieldsVal.([]any)
	if !ok {
		t.Fatalf("fields is %T, want []any", fieldsVal)
	}
	if len(arr) != 0 {
		t.Errorf("fields array length = %d, want 0", len(arr))
	}
}

func TestDatatypeFullView_NilAuthor_OmittedFromJSON(t *testing.T) {
	t.Parallel()

	view := DatatypeFullView{
		DatatypeID:   types.NewDatatypeID(),
		Label:        "orphan-type",
		Type:         "page",
		ParentID:     types.NullableDatatypeID{},
		Author:       nil,
		Fields:       []DatatypeFieldView{},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	}

	data, err := json.Marshal(view)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if _, ok := m["author"]; ok {
		t.Error("JSON contains 'author' key; expected omitempty to exclude nil author")
	}
	// fields should still be present as empty array
	fieldsVal, ok := m["fields"]
	if !ok {
		t.Fatal("JSON missing 'fields' key")
	}
	arr, ok := fieldsVal.([]any)
	if !ok {
		t.Fatalf("fields is %T, want []any", fieldsVal)
	}
	if len(arr) != 0 {
		t.Errorf("fields array length = %d, want 0", len(arr))
	}
}

// --- UserFullView tests ---

func TestMapUserOauthView(t *testing.T) {
	t.Parallel()

	o := UserOauth{
		UserOauthID:         types.NewUserOauthID(),
		OauthProvider:       "google",
		OauthProviderUserID: "12345",
		AccessToken:         "secret-access-token",
		RefreshToken:        "secret-refresh-token",
		TokenExpiresAt:      "2026-12-31T23:59:59Z",
		DateCreated:         types.TimestampNow(),
	}

	ov := MapUserOauthView(o)

	if ov.UserOauthID != o.UserOauthID {
		t.Errorf("UserOauthID = %v, want %v", ov.UserOauthID, o.UserOauthID)
	}
	if ov.OauthProvider != "google" {
		t.Errorf("OauthProvider = %q, want %q", ov.OauthProvider, "google")
	}

	// Verify tokens are excluded from JSON
	data, err := json.Marshal(ov)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if _, ok := m["access_token"]; ok {
		t.Error("UserOauthView JSON contains 'access_token'; expected exclusion")
	}
	if _, ok := m["refresh_token"]; ok {
		t.Error("UserOauthView JSON contains 'refresh_token'; expected exclusion")
	}
}

func TestMapTokenView(t *testing.T) {
	t.Parallel()

	tok := Tokens{
		ID:        "tok-1",
		TokenType: "api",
		Token:     "secret-token-value",
		IssuedAt:  "2026-01-01T00:00:00Z",
		ExpiresAt: types.TimestampNow(),
		Revoked:   false,
	}

	tv := MapTokenView(tok)

	if tv.ID != tok.ID {
		t.Errorf("ID = %q, want %q", tv.ID, tok.ID)
	}
	if tv.TokenType != "api" {
		t.Errorf("TokenType = %q, want %q", tv.TokenType, "api")
	}

	// Verify token value is excluded
	data, err := json.Marshal(tv)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if _, ok := m["token"]; ok {
		t.Error("TokenView JSON contains 'token'; expected exclusion")
	}
}

func TestAssembleUserFullView(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)

	view, err := AssembleUserFullView(d, seed.User.UserID)
	if err != nil {
		t.Fatalf("AssembleUserFullView: %v", err)
	}

	if view.UserID != seed.User.UserID {
		t.Errorf("UserID = %v, want %v", view.UserID, seed.User.UserID)
	}
	if view.Username != "testuser" {
		t.Errorf("Username = %q, want %q", view.Username, "testuser")
	}
	if view.RoleLabel != "test-role" {
		t.Errorf("RoleLabel = %q, want %q", view.RoleLabel, "test-role")
	}

	// No oauth, ssh keys, sessions, or tokens in seed data
	if view.Oauth != nil {
		t.Error("Oauth should be nil for seed user")
	}
	if view.Sessions != nil {
		t.Error("Sessions should be nil for seed user")
	}
	if view.SshKeys == nil {
		t.Fatal("SshKeys is nil, expected empty slice")
	}
	if len(view.SshKeys) != 0 {
		t.Errorf("SshKeys length = %d, want 0", len(view.SshKeys))
	}
	if view.Tokens == nil {
		t.Fatal("Tokens is nil, expected empty slice")
	}
	if len(view.Tokens) != 0 {
		t.Errorf("Tokens length = %d, want 0", len(view.Tokens))
	}
}

func TestAssembleUserFullView_EmptyArraysInJSON(t *testing.T) {
	t.Parallel()
	d, seed := testSeededDB(t)

	view, err := AssembleUserFullView(d, seed.User.UserID)
	if err != nil {
		t.Fatalf("AssembleUserFullView: %v", err)
	}

	data, err := json.Marshal(view)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	// oauth and sessions should be omitted (nil + omitempty)
	if _, ok := m["oauth"]; ok {
		t.Error("JSON contains 'oauth' key; expected omitempty to exclude nil")
	}
	if _, ok := m["sessions"]; ok {
		t.Error("JSON contains 'sessions' key; expected omitempty to exclude nil")
	}

	// ssh_keys and tokens should be empty arrays, not null
	for _, key := range []string{"ssh_keys", "tokens"} {
		val, ok := m[key]
		if !ok {
			t.Errorf("JSON missing %q key", key)
			continue
		}
		arr, ok := val.([]any)
		if !ok {
			t.Errorf("%s is %T, want []any", key, val)
			continue
		}
		if len(arr) != 0 {
			t.Errorf("%s array length = %d, want 0", key, len(arr))
		}
	}

	// hash should not appear
	if _, ok := m["hash"]; ok {
		t.Error("JSON contains 'hash' key; expected exclusion")
	}
}
