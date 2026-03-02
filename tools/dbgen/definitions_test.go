package main

import (
	"testing"
)

// These tests validate structural invariants of the Entities and DriverConfigs
// definitions. If a definition is malformed, the code generator produces
// broken output. Catching these at test time is cheaper than debugging a
// template rendering failure.

func TestEntities_AllHaveRequiredFields(t *testing.T) {
	t.Parallel()
	for _, e := range Entities {
		t.Run(e.Name, func(t *testing.T) {
			t.Parallel()
			if e.Name == "" {
				t.Error("Entity.Name is empty")
			}
			if e.Singular == "" {
				t.Errorf("Entity %q: Singular is empty", e.Name)
			}
			if e.Plural == "" {
				t.Errorf("Entity %q: Plural is empty", e.Name)
			}
			if e.TableName == "" {
				t.Errorf("Entity %q: TableName is empty", e.Name)
			}
			if e.IDType == "" {
				t.Errorf("Entity %q: IDType is empty", e.Name)
			}
			if e.IDField == "" {
				t.Errorf("Entity %q: IDField is empty", e.Name)
			}
			if e.NewIDFunc == "" {
				t.Errorf("Entity %q: NewIDFunc is empty", e.Name)
			}
			if e.OutputFile == "" {
				t.Errorf("Entity %q: OutputFile is empty", e.Name)
			}
			if len(e.Fields) == 0 {
				t.Errorf("Entity %q: Fields is empty", e.Name)
			}
		})
	}
}

func TestEntities_AllHaveExactlyOnePrimaryID(t *testing.T) {
	t.Parallel()
	for _, e := range Entities {
		t.Run(e.Name, func(t *testing.T) {
			t.Parallel()
			count := 0
			for _, f := range e.Fields {
				if f.IsPrimaryID {
					count++
				}
			}
			if count != 1 {
				t.Errorf("Entity %q: has %d primary ID fields, want exactly 1", e.Name, count)
			}
		})
	}
}

func TestEntities_PrimaryIDFieldMatchesIDField(t *testing.T) {
	t.Parallel()
	for _, e := range Entities {
		t.Run(e.Name, func(t *testing.T) {
			t.Parallel()
			for _, f := range e.Fields {
				if f.IsPrimaryID {
					if f.AppName != e.IDField {
						t.Errorf("Entity %q: primary ID field AppName=%q does not match IDField=%q",
							e.Name, f.AppName, e.IDField)
					}
					return
				}
			}
			t.Errorf("Entity %q: no field with IsPrimaryID=true", e.Name)
		})
	}
}

func TestEntities_OutputFilesAreUnique(t *testing.T) {
	t.Parallel()
	seen := make(map[string]string) // outputFile -> entityName
	for _, e := range Entities {
		if prev, ok := seen[e.OutputFile]; ok {
			t.Errorf("Entity %q and %q both use OutputFile %q", prev, e.Name, e.OutputFile)
		}
		seen[e.OutputFile] = e.Name
	}
}

func TestEntities_NamesAreUnique(t *testing.T) {
	t.Parallel()
	seen := make(map[string]bool)
	for _, e := range Entities {
		if seen[e.Name] {
			t.Errorf("Duplicate entity Name: %q", e.Name)
		}
		seen[e.Name] = true
	}
}

func TestEntities_FieldsHaveAppName(t *testing.T) {
	t.Parallel()
	for _, e := range Entities {
		t.Run(e.Name, func(t *testing.T) {
			t.Parallel()
			for i, f := range e.Fields {
				if f.AppName == "" {
					t.Errorf("Entity %q: Field[%d] has empty AppName", e.Name, i)
				}
			}
		})
	}
}

func TestEntities_FieldsHaveType(t *testing.T) {
	t.Parallel()
	for _, e := range Entities {
		t.Run(e.Name, func(t *testing.T) {
			t.Parallel()
			for i, f := range e.Fields {
				if f.Type == "" {
					t.Errorf("Entity %q: Field[%d] (%s) has empty Type", e.Name, i, f.AppName)
				}
			}
		})
	}
}

func TestEntities_FieldsHaveJSONTag(t *testing.T) {
	t.Parallel()
	for _, e := range Entities {
		t.Run(e.Name, func(t *testing.T) {
			t.Parallel()
			for i, f := range e.Fields {
				// UpdateParamsOnly fields may not need JSON tags, but struct fields should
				if !f.UpdateParamsOnly && f.JSONTag == "" {
					t.Errorf("Entity %q: Field[%d] (%s) has empty JSONTag", e.Name, i, f.AppName)
				}
			}
		})
	}
}

func TestEntities_ExtraQueriesHaveMethodName(t *testing.T) {
	t.Parallel()
	for _, e := range Entities {
		for i, eq := range e.ExtraQueries {
			if eq.MethodName == "" {
				t.Errorf("Entity %q: ExtraQueries[%d] has empty MethodName", e.Name, i)
			}
		}
	}
}

func TestEntities_ExtraQueriesHaveAtLeastOneParam(t *testing.T) {
	t.Parallel()
	// Paramless ExtraQueries (e.g. ListActiveWebhooks) are valid —
	// the template generates correct code without params.
	// This list tracks intentionally paramless queries so new ones
	// are added consciously.
	allowParamless := map[string]bool{
		"ListActiveWebhooks": true,
	}
	for _, e := range Entities {
		for i, eq := range e.ExtraQueries {
			if len(eq.Params) == 0 && !allowParamless[eq.MethodName] {
				t.Errorf("Entity %q: ExtraQueries[%d] (%s) has no Params", e.Name, i, eq.MethodName)
			}
		}
	}
}

func TestEntities_PaginatedExtraQueriesHaveFilterFields(t *testing.T) {
	t.Parallel()
	for _, e := range Entities {
		for i, pq := range e.PaginatedExtraQueries {
			if pq.MethodName == "" {
				t.Errorf("Entity %q: PaginatedExtraQueries[%d] has empty MethodName", e.Name, i)
			}
			if len(pq.FilterFields) == 0 {
				t.Errorf("Entity %q: PaginatedExtraQueries[%d] (%s) has no FilterFields",
					e.Name, i, pq.MethodName)
			}
			if pq.AppParamsType == "" {
				t.Errorf("Entity %q: PaginatedExtraQueries[%d] (%s) has empty AppParamsType",
					e.Name, i, pq.MethodName)
			}
		}
	}
}

func TestEntities_StringTypeNameConsistency(t *testing.T) {
	t.Parallel()
	// If StringTypeName is set, the entity should have at least one field with a non-empty StringConvert
	for _, e := range Entities {
		if e.StringTypeName == "" {
			continue
		}
		t.Run(e.Name, func(t *testing.T) {
			t.Parallel()
			if !e.HasStringConvertFields() {
				t.Errorf("Entity %q: StringTypeName=%q but no fields have StringConvert set",
					e.Name, e.StringTypeName)
			}
		})
	}
}

func TestEntities_CallerSuppliedIDImpliesPrimaryIDInCreate(t *testing.T) {
	t.Parallel()
	// When CallerSuppliedID is true, the primary ID field should have InCreate=true
	// because the caller provides the ID
	for _, e := range Entities {
		if !e.CallerSuppliedID {
			continue
		}
		t.Run(e.Name, func(t *testing.T) {
			t.Parallel()
			for _, f := range e.Fields {
				if f.IsPrimaryID && !f.InCreate {
					t.Errorf("Entity %q: CallerSuppliedID=true but primary ID field %q has InCreate=false",
						e.Name, f.AppName)
				}
			}
		})
	}
}

func TestEntities_SqlcTypeNameFallback(t *testing.T) {
	t.Parallel()
	// SqlcTypeName should always be set; if it's empty, the template will generate
	// invalid code referencing an empty type
	for _, e := range Entities {
		if e.SqlcTypeName == "" {
			t.Errorf("Entity %q: SqlcTypeName is empty (template will generate invalid code)", e.Name)
		}
	}
}

// ========== DriverConfigs ==========

func TestDriverConfigs_HasThreeDrivers(t *testing.T) {
	t.Parallel()
	if len(DriverConfigs) != 3 {
		t.Fatalf("DriverConfigs has %d entries, want 3", len(DriverConfigs))
	}
}

func TestDriverConfigs_AllHaveRequiredFields(t *testing.T) {
	t.Parallel()
	for _, d := range DriverConfigs {
		t.Run(d.Name, func(t *testing.T) {
			t.Parallel()
			if d.Name == "" {
				t.Error("DriverConfig.Name is empty")
			}
			if d.Struct == "" {
				t.Errorf("DriverConfig %q: Struct is empty", d.Name)
			}
			if d.Package == "" {
				t.Errorf("DriverConfig %q: Package is empty", d.Name)
			}
			if d.Recorder == "" {
				t.Errorf("DriverConfig %q: Recorder is empty", d.Name)
			}
		})
	}
}

func TestDriverConfigs_NamesAreUnique(t *testing.T) {
	t.Parallel()
	seen := make(map[string]bool)
	for _, d := range DriverConfigs {
		if seen[d.Name] {
			t.Errorf("Duplicate DriverConfig Name: %q", d.Name)
		}
		seen[d.Name] = true
	}
}

func TestDriverConfigs_SpecificValues(t *testing.T) {
	t.Parallel()

	// Verify expected driver configs match what the template relies on
	tests := []struct {
		name              string
		wantStruct        string
		wantPackage       string
		wantRecorder      string
		wantCmdSuffix     string
		wantMysqlGap      bool
		wantInt32Paginate bool
	}{
		{
			name:              "sqlite",
			wantStruct:        "Database",
			wantPackage:       "mdb",
			wantRecorder:      "SQLiteRecorder",
			wantCmdSuffix:     "",
			wantMysqlGap:      false,
			wantInt32Paginate: false,
		},
		{
			name:              "mysql",
			wantStruct:        "MysqlDatabase",
			wantPackage:       "mdbm",
			wantRecorder:      "MysqlRecorder",
			wantCmdSuffix:     "Mysql",
			wantMysqlGap:      true,
			wantInt32Paginate: true,
		},
		{
			name:              "psql",
			wantStruct:        "PsqlDatabase",
			wantPackage:       "mdbp",
			wantRecorder:      "PsqlRecorder",
			wantCmdSuffix:     "Psql",
			wantMysqlGap:      false,
			wantInt32Paginate: true,
		},
	}

	driverMap := make(map[string]DriverConfig)
	for _, d := range DriverConfigs {
		driverMap[d.Name] = d
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d, ok := driverMap[tt.name]
			if !ok {
				t.Fatalf("DriverConfig %q not found", tt.name)
			}
			if d.Struct != tt.wantStruct {
				t.Errorf("Struct = %q, want %q", d.Struct, tt.wantStruct)
			}
			if d.Package != tt.wantPackage {
				t.Errorf("Package = %q, want %q", d.Package, tt.wantPackage)
			}
			if d.Recorder != tt.wantRecorder {
				t.Errorf("Recorder = %q, want %q", d.Recorder, tt.wantRecorder)
			}
			if d.CmdSuffix != tt.wantCmdSuffix {
				t.Errorf("CmdSuffix = %q, want %q", d.CmdSuffix, tt.wantCmdSuffix)
			}
			if d.MysqlReturningGap != tt.wantMysqlGap {
				t.Errorf("MysqlReturningGap = %v, want %v", d.MysqlReturningGap, tt.wantMysqlGap)
			}
			if d.Int32Pagination != tt.wantInt32Paginate {
				t.Errorf("Int32Pagination = %v, want %v", d.Int32Pagination, tt.wantInt32Paginate)
			}
		})
	}
}
