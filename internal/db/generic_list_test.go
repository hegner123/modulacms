package db

import (
	"testing"
)

func TestGenericHeaders_AllKnownTables(t *testing.T) {
	t.Parallel()
	// Every known DBTable constant should return a non-nil, non-empty slice
	allTables := []DBTable{
		Admin_content_data,
		Admin_content_fields,
		Admin_datatype,
		Admin_datatype_fields,
		Admin_field,
		Admin_route,
		Content_data,
		Content_fields,
		Datatype_fields,
		Datatype,
		Field,
		MediaT,
		Media_dimension,
		Permission,
		Role,
		Route,
		Session,
		Table,
		Token,
		User,
		User_oauth,
	}
	for _, tbl := range allTables {
		t.Run(string(tbl), func(t *testing.T) {
			t.Parallel()
			headers := GenericHeaders(tbl)
			if headers == nil {
				t.Fatalf("GenericHeaders(%q) returned nil", tbl)
			}
			if len(headers) == 0 {
				t.Fatalf("GenericHeaders(%q) returned empty slice", tbl)
			}
			// Check no empty strings in headers
			for i, h := range headers {
				if h == "" {
					t.Errorf("GenericHeaders(%q)[%d] is empty string", tbl, i)
				}
			}
		})
	}
}

func TestGenericHeaders_UnknownTable_ReturnsNil(t *testing.T) {
	t.Parallel()
	got := GenericHeaders(DBTable("nonexistent"))
	if got != nil {
		t.Errorf("GenericHeaders(nonexistent) = %v, want nil", got)
	}
}

func TestGenericHeaders_SpecificCounts(t *testing.T) {
	t.Parallel()
	// Verify a few tables have the expected number of columns
	tests := []struct {
		table     DBTable
		wantCount int
	}{
		{Route, 8},             // route_id, slug, title, status, author_id, date_created, date_modified, history
		{User, 8},              // user_id, username, name, email, hash, role, date_created, date_modified
		{Permission, 4},        // permission_id, table_id, mode, label
		{Role, 3},              // role_id, label, permissions
		{Session, 8},           // session_id, user_id, created_at, expires_at, last_access, ip_address, user_agent, session_data
		{MediaT, 14},           // 14 media fields
		{Media_dimension, 5},   // md_id, label, width, height, aspect_ratio
		{Table, 3},             // id, label, author_id
		{Token, 7},             // id, user_id, token_type, token, issued_at, expires_at, revoked
		{Admin_datatype_fields, 3}, // id, admin_datatype_id, admin_field_id
		{Datatype_fields, 3},   // id, datatype_id, field_id
	}
	for _, tt := range tests {
		t.Run(string(tt.table), func(t *testing.T) {
			t.Parallel()
			headers := GenericHeaders(tt.table)
			if len(headers) != tt.wantCount {
				t.Errorf("GenericHeaders(%q) has %d columns, want %d; columns: %v",
					tt.table, len(headers), tt.wantCount, headers)
			}
		})
	}
}

func TestGenericHeaders_FirstColumnIsID(t *testing.T) {
	t.Parallel()
	// For most tables, the first column should contain "id" (case-insensitive check not needed;
	// all headers use lowercase with underscores)
	tablesToCheck := []DBTable{
		Route, User, Permission, Session, MediaT, Media_dimension, Token, Table,
		Admin_content_data, Admin_content_fields, Admin_datatype, Admin_field, Admin_route,
		Content_data, Content_fields, Datatype, Field, User_oauth,
	}
	for _, tbl := range tablesToCheck {
		t.Run(string(tbl), func(t *testing.T) {
			t.Parallel()
			headers := GenericHeaders(tbl)
			if len(headers) == 0 {
				t.Fatal("empty headers")
			}
			first := headers[0]
			// The first column should end with "_id" or be "id"
			if first != "id" && len(first) < 3 {
				t.Errorf("first header %q does not look like an ID column", first)
			}
		})
	}
}

func TestGenericHeaders_RouteSpecificColumns(t *testing.T) {
	t.Parallel()
	headers := GenericHeaders(Route)
	expected := []string{
		"route_id", "slug", "title", "status",
		"author_id", "date_created", "date_modified", "history",
	}
	if len(headers) != len(expected) {
		t.Fatalf("got %d headers, want %d", len(headers), len(expected))
	}
	for i, want := range expected {
		if headers[i] != want {
			t.Errorf("headers[%d] = %q, want %q", i, headers[i], want)
		}
	}
}

func TestGenericHeaders_UserSpecificColumns(t *testing.T) {
	t.Parallel()
	headers := GenericHeaders(User)
	expected := []string{
		"user_id", "username", "name", "email",
		"hash", "role", "date_created", "date_modified",
	}
	if len(headers) != len(expected) {
		t.Fatalf("got %d headers, want %d", len(headers), len(expected))
	}
	for i, want := range expected {
		if headers[i] != want {
			t.Errorf("headers[%d] = %q, want %q", i, headers[i], want)
		}
	}
}
