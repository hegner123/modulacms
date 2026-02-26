package utility

import (
	"testing"
)

// ============================================================
// StorageUnit constants
// ============================================================

func TestStorageUnit_Values(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		unit StorageUnit
		want int64
	}{
		{name: "KB is 1024", unit: KB, want: 1024},
		{name: "MB is 1048576", unit: MB, want: 1048576},
		{name: "GB is 1073741824", unit: GB, want: 1073741824},
		{name: "TB is 1099511627776", unit: TB, want: 1099511627776},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if int64(tt.unit) != tt.want {
				t.Errorf("%s = %d, want %d", tt.name, tt.unit, tt.want)
			}
		})
	}
}

// Verify units are powers of 2 and relate correctly
func TestStorageUnit_Relationships(t *testing.T) {
	t.Parallel()
	if MB != KB*1024 {
		t.Errorf("MB (%d) should be KB*1024 (%d)", MB, KB*1024)
	}
	if GB != MB*1024 {
		t.Errorf("GB (%d) should be MB*1024 (%d)", GB, MB*1024)
	}
	if TB != GB*1024 {
		t.Errorf("TB (%d) should be GB*1024 (%d)", TB, GB*1024)
	}
}

// ============================================================
// SizeInBytes
// ============================================================

func TestSizeInBytes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value int64
		unit  StorageUnit
		want  int64
	}{
		{name: "1 KB", value: 1, unit: KB, want: 1024},
		{name: "5 MB", value: 5, unit: MB, want: 5 * 1048576},
		{name: "2 GB", value: 2, unit: GB, want: 2 * 1073741824},
		{name: "1 TB", value: 1, unit: TB, want: 1099511627776},
		{name: "0 of any unit", value: 0, unit: GB, want: 0},
		{name: "10 KB", value: 10, unit: KB, want: 10240},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := SizeInBytes(tt.value, tt.unit)
			if got != tt.want {
				t.Errorf("SizeInBytes(%d, %d) = %d, want %d", tt.value, tt.unit, got, tt.want)
			}
		})
	}
}

// ============================================================
// MIME type constants
// ============================================================

func TestAppJson_Value(t *testing.T) {
	t.Parallel()
	if AppJson != "application/json" {
		t.Errorf("AppJson = %q, want %q", AppJson, "application/json")
	}
}
