package db

import (
	"testing"
)

func TestIsValidPluginTableName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid plugin table", "plugin_blog_posts", true},
		{"valid short name", "plugin_xy", true},
		{"too short", "plugin_x", false},
		{"no prefix", "blog_posts", false},
		{"core table", "users", false},
		{"empty string", "", false},
		{"just prefix", "plugin_", false},
		{"system plugin table", "plugin_routes", true}, // valid name format, system filtering is at discovery level
		{"contains uppercase", "plugin_Blog_Posts", true},
		{"contains numbers", "plugin_v2_items", true},
		{"contains hyphen", "plugin_my-table", false}, // hyphens not valid in SQL identifiers
		{"contains space", "plugin_ table", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidPluginTableName(tt.input)
			if got != tt.want {
				t.Errorf("IsValidPluginTableName(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidTable_AcceptsPluginTables(t *testing.T) {
	// Core tables should pass.
	if !IsValidTable(User) {
		t.Error("IsValidTable should accept core table 'users'")
	}

	// Plugin tables should pass.
	if !IsValidTable(DBTable("plugin_blog_posts")) {
		t.Error("IsValidTable should accept valid plugin table 'plugin_blog_posts'")
	}

	// Invalid names should fail.
	if IsValidTable(DBTable("nonexistent_table")) {
		t.Error("IsValidTable should reject unknown non-plugin table")
	}
}

func TestValidateTableName_AcceptsPluginTables(t *testing.T) {
	// Core table.
	dt, err := ValidateTableName("users")
	if err != nil {
		t.Fatalf("ValidateTableName('users') error: %v", err)
	}
	if dt != User {
		t.Errorf("ValidateTableName('users') = %q, want %q", dt, User)
	}

	// Plugin table.
	dt, err = ValidateTableName("plugin_blog_posts")
	if err != nil {
		t.Fatalf("ValidateTableName('plugin_blog_posts') error: %v", err)
	}
	if dt != DBTable("plugin_blog_posts") {
		t.Errorf("got %q, want %q", dt, "plugin_blog_posts")
	}

	// Invalid.
	_, err = ValidateTableName("nonexistent")
	if err == nil {
		t.Error("ValidateTableName('nonexistent') should return error")
	}
}

func TestSystemPluginTables(t *testing.T) {
	expected := []string{"plugin_routes", "plugin_hooks", "plugin_requests"}
	for _, name := range expected {
		if !SystemPluginTables[name] {
			t.Errorf("SystemPluginTables should contain %q", name)
		}
	}
	if SystemPluginTables["plugin_blog_posts"] {
		t.Error("SystemPluginTables should not contain user plugin tables")
	}
}
