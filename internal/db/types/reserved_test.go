package types

import "testing"

func TestDatatypeType_IsReserved(t *testing.T) {
	tests := []struct {
		name string
		dt   DatatypeType
		want bool
	}{
		{"root is reserved", DatatypeTypeRoot, true},
		{"reference is reserved", DatatypeTypeReference, true},
		{"nested_root is reserved", DatatypeTypeNestedRoot, true},
		{"system_log is reserved", DatatypeTypeSystemLog, true},
		{"page is not reserved", DatatypeType("page"), false},
		{"empty is not reserved", DatatypeType(""), false},
		{"unknown underscore is not reserved", DatatypeType("_unknown"), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.dt.IsReserved(); got != tc.want {
				t.Errorf("DatatypeType(%q).IsReserved() = %v, want %v", tc.dt, got, tc.want)
			}
		})
	}
}

func TestIsReservedPrefix(t *testing.T) {
	tests := []struct {
		name string
		val  string
		want bool
	}{
		{"underscore prefix", "_root", true},
		{"underscore only", "_", true},
		{"unknown underscore", "_foo", true},
		{"normal string", "page", false},
		{"empty string", "", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsReservedPrefix(tc.val); got != tc.want {
				t.Errorf("IsReservedPrefix(%q) = %v, want %v", tc.val, got, tc.want)
			}
		})
	}
}

func TestDatatypeType_IsRootType(t *testing.T) {
	tests := []struct {
		name string
		dt   DatatypeType
		want bool
	}{
		{"_root is root type", DatatypeTypeRoot, true},
		{"_nested_root is root type", DatatypeTypeNestedRoot, true},
		{"_reference is not root type", DatatypeTypeReference, false},
		{"_system_log is not root type", DatatypeTypeSystemLog, false},
		{"page is not root type", DatatypeType("page"), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.dt.IsRootType(); got != tc.want {
				t.Errorf("DatatypeType(%q).IsRootType() = %v, want %v", tc.dt, got, tc.want)
			}
		})
	}
}

func TestValidateDatatypeType(t *testing.T) {
	tests := []struct {
		name    string
		val     string
		wantErr bool
	}{
		{"empty is error", "", true},
		{"_root is valid", "_root", false},
		{"_reference is valid", "_reference", false},
		{"_nested_root is valid", "_nested_root", false},
		{"_system_log is valid", "_system_log", false},
		{"_unknown is error", "_unknown", true},
		{"page is valid", "page", false},
		{"hero_banner is valid", "hero_banner", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateDatatypeType(tc.val)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateDatatypeType(%q) error = %v, wantErr %v", tc.val, err, tc.wantErr)
			}
		})
	}
}

func TestValidateUserDatatypeType(t *testing.T) {
	tests := []struct {
		name    string
		val     string
		wantErr bool
	}{
		{"empty is error", "", true},
		{"underscore prefix is error", "_root", true},
		{"unknown underscore is error", "_foo", true},
		{"page is valid", "page", false},
		{"blog_post is valid", "blog_post", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateUserDatatypeType(tc.val)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateUserDatatypeType(%q) error = %v, wantErr %v", tc.val, err, tc.wantErr)
			}
		})
	}
}

func TestReservedTypes_ReturnsCopy(t *testing.T) {
	rt := ReservedTypes()
	if len(rt) != 4 {
		t.Fatalf("expected 4 reserved types, got %d", len(rt))
	}

	// Mutating the copy should not affect the original
	rt[DatatypeType("_test")] = "test"
	if len(ReservedTypes()) != 4 {
		t.Fatal("mutating returned map affected the internal registry")
	}
}
