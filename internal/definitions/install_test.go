package definitions

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// callRecord tracks which create method was called and in what order.
type callRecord struct {
	kind string // "field", "datatype"
}

type mockInstaller struct {
	calls     []callRecord
	datatypes []db.Datatypes
	fields    []db.Fields
}

func (m *mockInstaller) CreateField(p db.CreateFieldParams) (db.Fields, error) {
	m.calls = append(m.calls, callRecord{kind: "field"})
	f := db.Fields{
		FieldID:      p.FieldID,
		ParentID:     p.ParentID,
		Name:         p.Name,
		Label:        p.Label,
		Data:         p.Data,
		Validation:   p.Validation,
		UIConfig:     p.UIConfig,
		Type:         p.Type,
		AuthorID:     p.AuthorID,
		DateCreated:  p.DateCreated,
		DateModified: p.DateModified,
	}
	m.fields = append(m.fields, f)
	return f, nil
}

func (m *mockInstaller) CreateDatatype(p db.CreateDatatypeParams) (db.Datatypes, error) {
	m.calls = append(m.calls, callRecord{kind: "datatype"})
	d := db.Datatypes{
		DatatypeID:   p.DatatypeID,
		ParentID:     p.ParentID,
		Name:         p.Name,
		Label:        p.Label,
		Type:         p.Type,
		AuthorID:     p.AuthorID,
		DateCreated:  p.DateCreated,
		DateModified: p.DateModified,
	}
	m.datatypes = append(m.datatypes, d)
	return d, nil
}

// totalFieldDefs counts all inline FieldDefs across all datatypes.
func totalFieldDefs(def SchemaDefinition) int {
	count := 0
	for _, dt := range def.Datatypes {
		count += len(dt.FieldRefs)
	}
	return count
}

func TestInstall_DefaultSchema(t *testing.T) {
	def, ok := Get("modula-default")
	if !ok {
		t.Fatal("modula-default definition not found")
	}

	mock := &mockInstaller{}
	authorID := types.NewUserID()
	result, err := Install(mock, def, authorID)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if result.DefinitionName != "modula-default" {
		t.Errorf("expected definition name %q, got %q", "modula-default", result.DefinitionName)
	}

	expectedFields := totalFieldDefs(def)
	if result.Fields != expectedFields {
		t.Errorf("expected %d fields, got %d", expectedFields, result.Fields)
	}

	expectedDatatypes := len(def.Datatypes)
	if result.Datatypes != expectedDatatypes {
		t.Errorf("expected %d datatypes, got %d", expectedDatatypes, result.Datatypes)
	}
}

func TestInstall_FieldsHaveParentID(t *testing.T) {
	def, ok := Get("modula-default")
	if !ok {
		t.Fatal("modula-default definition not found")
	}

	mock := &mockInstaller{}
	authorID := types.NewUserID()
	_, err := Install(mock, def, authorID)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	for _, f := range mock.fields {
		if !f.ParentID.Valid {
			t.Errorf("field %q has no parent_id, expected a datatype reference", f.Label)
		}
		if f.ParentID.ID.IsZero() {
			t.Errorf("field %q has zero parent_id", f.Label)
		}
	}
}

func TestInstall_PhaseOrdering(t *testing.T) {
	def, ok := Get("modula-default")
	if !ok {
		t.Fatal("modula-default definition not found")
	}

	mock := &mockInstaller{}
	authorID := types.NewUserID()
	_, err := Install(mock, def, authorID)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Verify ordering: all datatypes first, then fields.
	seenDatatype := false
	for _, call := range mock.calls {
		if call.kind == "datatype" {
			seenDatatype = true
		}
		if !seenDatatype && call.kind == "field" {
			t.Fatal("field created before any datatype")
		}
	}
}

func TestInstall_ChildDatatypeReceivesParentID(t *testing.T) {
	def, ok := Get("modula-default")
	if !ok {
		t.Fatal("modula-default definition not found")
	}

	mock := &mockInstaller{}
	authorID := types.NewUserID()
	_, err := Install(mock, def, authorID)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// The "section" datatype should have a non-empty ParentID
	foundSection := false
	for _, dt := range mock.datatypes {
		if dt.Label == "Section" {
			foundSection = true
			if !dt.ParentID.Valid {
				t.Error("section datatype should have a valid ParentID")
			}
		}
	}
	if !foundSection {
		t.Error("section datatype not found in created datatypes")
	}

	// The "page" datatype should have no parent
	for _, dt := range mock.datatypes {
		if dt.Label == "Page" {
			if dt.ParentID.Valid {
				t.Error("page datatype should not have a ParentID")
			}
		}
	}
}

func TestInstall_EmptyAuthorID(t *testing.T) {
	def, ok := Get("modula-default")
	if !ok {
		t.Fatal("modula-default definition not found")
	}

	mock := &mockInstaller{}
	_, err := Install(mock, def, types.UserID(""))
	if err == nil {
		t.Fatal("expected error for empty authorID")
	}
	if !strings.Contains(err.Error(), "authorID cannot be empty") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestInstall_InvalidDefinition(t *testing.T) {
	mock := &mockInstaller{}
	authorID := types.NewUserID()

	invalid := SchemaDefinition{
		Name: "", // will fail validation
	}
	_, err := Install(mock, invalid, authorID)
	if err == nil {
		t.Fatal("expected error for invalid definition")
	}
}

func TestInstall_AllRegisteredDefinitions(t *testing.T) {
	for _, def := range List() {
		t.Run(def.Name, func(t *testing.T) {
			mock := &mockInstaller{}
			authorID := types.NewUserID()
			result, err := Install(mock, def, authorID)
			if err != nil {
				t.Fatalf("Install(%q) failed: %v", def.Name, err)
			}

			expectedFields := totalFieldDefs(def)
			if result.Fields != expectedFields {
				t.Errorf("expected %d fields, got %d", expectedFields, result.Fields)
			}
			if result.Datatypes != len(def.Datatypes) {
				t.Errorf("expected %d datatypes, got %d", len(def.Datatypes), result.Datatypes)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// mockCleaner: in-memory implementation of Cleaner for Reinstall tests
// ---------------------------------------------------------------------------

type mockCleaner struct {
	users         []db.Users
	datatypes     []db.Datatypes
	fields        []db.Fields
	contentData   []db.ContentData
	contentFields []db.ContentFields
	routes        []db.Routes

	deletedContentFieldIDs []types.ContentFieldID
	deletedContentDataIDs  []types.ContentID
	deletedFieldIDs        []types.FieldID
	deletedDatatypeIDs     []types.DatatypeID
	deletedRouteIDs        []types.RouteID
}

func (m *mockCleaner) GetUserByEmail(email types.Email) (*db.Users, error) {
	for _, u := range m.users {
		if u.Email == email {
			return &u, nil
		}
	}
	return nil, fmt.Errorf("user with email %s not found", email)
}

func (m *mockCleaner) ListDatatypes() (*[]db.Datatypes, error) {
	return &m.datatypes, nil
}

func (m *mockCleaner) ListFields() (*[]db.Fields, error) {
	return &m.fields, nil
}

func (m *mockCleaner) ListContentData() (*[]db.ContentData, error) {
	return &m.contentData, nil
}

func (m *mockCleaner) ListContentFields() (*[]db.ContentFields, error) {
	return &m.contentFields, nil
}

func (m *mockCleaner) ListRoutes() (*[]db.Routes, error) {
	return &m.routes, nil
}

func (m *mockCleaner) DeleteContentField(_ context.Context, _ audited.AuditContext, id types.ContentFieldID) error {
	m.deletedContentFieldIDs = append(m.deletedContentFieldIDs, id)
	return nil
}

func (m *mockCleaner) DeleteContentData(_ context.Context, _ audited.AuditContext, id types.ContentID) error {
	m.deletedContentDataIDs = append(m.deletedContentDataIDs, id)
	return nil
}

func (m *mockCleaner) DeleteField(_ context.Context, _ audited.AuditContext, id types.FieldID) error {
	m.deletedFieldIDs = append(m.deletedFieldIDs, id)
	return nil
}

func (m *mockCleaner) DeleteDatatype(_ context.Context, _ audited.AuditContext, id types.DatatypeID) error {
	m.deletedDatatypeIDs = append(m.deletedDatatypeIDs, id)
	return nil
}

func (m *mockCleaner) DeleteRoute(_ context.Context, _ audited.AuditContext, id types.RouteID) error {
	m.deletedRouteIDs = append(m.deletedRouteIDs, id)
	return nil
}

// helpers for building test data

func systemUser() db.Users {
	return db.Users{
		UserID:   types.NewUserID(),
		Username: "system",
		Email:    types.Email("system@modula.local"),
	}
}

func TestReinstall_CleansBootstrappedRecords(t *testing.T) {
	su := systemUser()
	userAuthor := types.NewUserID()

	sysDt := db.Datatypes{DatatypeID: types.NewDatatypeID(), Label: "Page", Type: "root", AuthorID: su.UserID}
	sysField := db.Fields{FieldID: types.NewFieldID(), Label: "Title", AuthorID: types.NullableUserID{Valid: true, ID: su.UserID}, ParentID: types.NullableDatatypeID{Valid: true, ID: sysDt.DatatypeID}}
	sysContent := db.ContentData{ContentDataID: types.NewContentID(), AuthorID: su.UserID}
	sysCF := db.ContentFields{ContentFieldID: types.NewContentFieldID(), AuthorID: su.UserID}
	sysRoute := db.Routes{RouteID: types.NewRouteID(), AuthorID: types.NullableUserID{Valid: true, ID: su.UserID}}

	cleaner := &mockCleaner{
		users:         []db.Users{su},
		datatypes:     []db.Datatypes{sysDt},
		fields:        []db.Fields{sysField},
		contentData:   []db.ContentData{sysContent},
		contentFields: []db.ContentFields{sysCF},
		routes:        []db.Routes{sysRoute},
	}
	installer := &mockInstaller{}

	def, ok := Get("modula-default")
	if !ok {
		t.Fatal("modula-default definition not found")
	}

	_, err := Reinstall(context.Background(), cleaner, installer, def, userAuthor)
	if err != nil {
		t.Fatalf("Reinstall failed: %v", err)
	}

	if len(cleaner.deletedContentFieldIDs) != 1 {
		t.Errorf("expected 1 deleted content_field, got %d", len(cleaner.deletedContentFieldIDs))
	}
	if len(cleaner.deletedContentDataIDs) != 1 {
		t.Errorf("expected 1 deleted content_data, got %d", len(cleaner.deletedContentDataIDs))
	}
	if len(cleaner.deletedFieldIDs) != 1 {
		t.Errorf("expected 1 deleted field, got %d", len(cleaner.deletedFieldIDs))
	}
	if len(cleaner.deletedDatatypeIDs) != 1 {
		t.Errorf("expected 1 deleted datatype, got %d", len(cleaner.deletedDatatypeIDs))
	}
	if len(cleaner.deletedRouteIDs) != 1 {
		t.Errorf("expected 1 deleted route, got %d", len(cleaner.deletedRouteIDs))
	}
}

func TestReinstall_PreservesUserRecords(t *testing.T) {
	su := systemUser()
	userAuthor := types.NewUserID()

	userDt := db.Datatypes{DatatypeID: types.NewDatatypeID(), Label: "Custom", Type: "custom", AuthorID: userAuthor}
	userField := db.Fields{FieldID: types.NewFieldID(), Label: "Custom Field", AuthorID: types.NullableUserID{Valid: true, ID: userAuthor}}
	userContent := db.ContentData{ContentDataID: types.NewContentID(), AuthorID: userAuthor}
	userCF := db.ContentFields{ContentFieldID: types.NewContentFieldID(), AuthorID: userAuthor}
	userRoute := db.Routes{RouteID: types.NewRouteID(), AuthorID: types.NullableUserID{Valid: true, ID: userAuthor}}

	cleaner := &mockCleaner{
		users:         []db.Users{su},
		datatypes:     []db.Datatypes{userDt},
		fields:        []db.Fields{userField},
		contentData:   []db.ContentData{userContent},
		contentFields: []db.ContentFields{userCF},
		routes:        []db.Routes{userRoute},
	}
	installer := &mockInstaller{}

	def, ok := Get("modula-default")
	if !ok {
		t.Fatal("modula-default definition not found")
	}

	_, err := Reinstall(context.Background(), cleaner, installer, def, types.NewUserID())
	if err != nil {
		t.Fatalf("Reinstall failed: %v", err)
	}

	if len(cleaner.deletedContentFieldIDs) != 0 {
		t.Errorf("expected 0 deleted content_fields, got %d", len(cleaner.deletedContentFieldIDs))
	}
	if len(cleaner.deletedContentDataIDs) != 0 {
		t.Errorf("expected 0 deleted content_data, got %d", len(cleaner.deletedContentDataIDs))
	}
	if len(cleaner.deletedFieldIDs) != 0 {
		t.Errorf("expected 0 deleted fields, got %d", len(cleaner.deletedFieldIDs))
	}
	if len(cleaner.deletedDatatypeIDs) != 0 {
		t.Errorf("expected 0 deleted datatypes, got %d", len(cleaner.deletedDatatypeIDs))
	}
	if len(cleaner.deletedRouteIDs) != 0 {
		t.Errorf("expected 0 deleted routes, got %d", len(cleaner.deletedRouteIDs))
	}
}

func TestReinstall_DeletesReservedPrefixDatatypes(t *testing.T) {
	su := systemUser()

	refDt := db.Datatypes{DatatypeID: types.NewDatatypeID(), Label: "Reference", Type: "_reference", AuthorID: su.UserID}
	refField := db.Fields{FieldID: types.NewFieldID(), Label: "Target", AuthorID: types.NullableUserID{Valid: true, ID: su.UserID}, ParentID: types.NullableDatatypeID{Valid: true, ID: refDt.DatatypeID}}

	cleaner := &mockCleaner{
		users:     []db.Users{su},
		datatypes: []db.Datatypes{refDt},
		fields:    []db.Fields{refField},
	}
	installer := &mockInstaller{}

	def, ok := Get("modula-default")
	if !ok {
		t.Fatal("modula-default definition not found")
	}

	_, err := Reinstall(context.Background(), cleaner, installer, def, types.NewUserID())
	if err != nil {
		t.Fatalf("Reinstall failed: %v", err)
	}

	if len(cleaner.deletedDatatypeIDs) != 1 {
		t.Errorf("expected 1 deleted datatype, got %d", len(cleaner.deletedDatatypeIDs))
	}
	if len(cleaner.deletedFieldIDs) != 1 {
		t.Errorf("expected 1 deleted field, got %d", len(cleaner.deletedFieldIDs))
	}
}

func TestReinstall_ThenInstalls(t *testing.T) {
	su := systemUser()
	authorID := types.NewUserID()

	cleaner := &mockCleaner{
		users: []db.Users{su},
	}
	installer := &mockInstaller{}

	def, ok := Get("modula-default")
	if !ok {
		t.Fatal("modula-default definition not found")
	}

	result, err := Reinstall(context.Background(), cleaner, installer, def, authorID)
	if err != nil {
		t.Fatalf("Reinstall failed: %v", err)
	}

	if result.DefinitionName != "modula-default" {
		t.Errorf("expected definition name %q, got %q", "modula-default", result.DefinitionName)
	}

	expectedDatatypes := len(def.Datatypes)
	if result.Datatypes != expectedDatatypes {
		t.Errorf("expected %d datatypes, got %d", expectedDatatypes, result.Datatypes)
	}

	expectedFields := totalFieldDefs(def)
	if result.Fields != expectedFields {
		t.Errorf("expected %d fields, got %d", expectedFields, result.Fields)
	}

	if len(installer.datatypes) != expectedDatatypes {
		t.Errorf("expected %d created datatypes on installer, got %d", expectedDatatypes, len(installer.datatypes))
	}
	if len(installer.fields) != expectedFields {
		t.Errorf("expected %d created fields on installer, got %d", expectedFields, len(installer.fields))
	}
}
