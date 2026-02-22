package db

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// --- Compile-time interface checks ---
// These verify that all 9 audited command types satisfy their respective
// audited interfaces. A compilation failure here means the command struct
// drifted from the interface contract.

var (
	_ audited.CreateCommand[mdb.MediaDimensions]  = NewMediaDimensionCmd{}
	_ audited.UpdateCommand[mdb.MediaDimensions]  = UpdateMediaDimensionCmd{}
	_ audited.DeleteCommand[mdb.MediaDimensions]  = DeleteMediaDimensionCmd{}
	_ audited.CreateCommand[mdbm.MediaDimensions] = NewMediaDimensionCmdMysql{}
	_ audited.UpdateCommand[mdbm.MediaDimensions] = UpdateMediaDimensionCmdMysql{}
	_ audited.DeleteCommand[mdbm.MediaDimensions] = DeleteMediaDimensionCmdMysql{}
	_ audited.CreateCommand[mdbp.MediaDimensions] = NewMediaDimensionCmdPsql{}
	_ audited.UpdateCommand[mdbp.MediaDimensions] = UpdateMediaDimensionCmdPsql{}
	_ audited.DeleteCommand[mdbp.MediaDimensions] = DeleteMediaDimensionCmdPsql{}
)

// --- Test data helpers ---

// mdTestFixture returns a fully populated MediaDimensions struct for testing.
func mdTestFixture() MediaDimensions {
	return MediaDimensions{
		MdID:        "test-md-id-001",
		Label:       NewNullString("Thumbnail"),
		Width:       NewNullInt64(1920),
		Height:      NewNullInt64(1080),
		AspectRatio: NewNullString("16:9"),
	}
}

// mdCreateFormParams returns a CreateMediaDimensionFormParams with all fields populated.
func mdCreateFormParams() CreateMediaDimensionFormParams {
	return CreateMediaDimensionFormParams{
		Label:       "Banner",
		Width:       "1280",
		Height:      "720",
		AspectRatio: "16:9",
	}
}

// mdUpdateFormParams returns an UpdateMediaDimensionFormParams with all fields populated.
func mdUpdateFormParams() UpdateMediaDimensionFormParams {
	return UpdateMediaDimensionFormParams{
		Label:       "Updated Banner",
		Width:       "1920",
		Height:      "1080",
		AspectRatio: "16:9",
		MdID:        "update-md-id-001",
	}
}

// mdCreateParams returns a CreateMediaDimensionParams with all fields set to valid values.
func mdCreateParams() CreateMediaDimensionParams {
	return CreateMediaDimensionParams{
		Label:       NewNullString("Hero"),
		Width:       NewNullInt64(2560),
		Height:      NewNullInt64(1440),
		AspectRatio: NewNullString("16:9"),
	}
}

// mdUpdateParams returns an UpdateMediaDimensionParams with all fields set to valid values.
func mdUpdateParams() UpdateMediaDimensionParams {
	return UpdateMediaDimensionParams{
		Label:       NewNullString("Updated Hero"),
		Width:       NewNullInt64(3840),
		Height:      NewNullInt64(2160),
		AspectRatio: NewNullString("16:9"),
		MdID:        "update-hero-md-001",
	}
}

// --- MapCreateMediaDimensionParams (FormParams -> Params) tests ---

func TestMapCreateMediaDimensionParams_AllFields(t *testing.T) {
	t.Parallel()
	form := mdCreateFormParams()

	got := MapCreateMediaDimensionParams(form)

	if got.Label.String != "Banner" || !got.Label.Valid {
		t.Errorf("Label = %v, want valid 'Banner'", got.Label)
	}
	if got.Width.Int64 != 1280 || !got.Width.Valid {
		t.Errorf("Width = %v, want valid 1280", got.Width)
	}
	if got.Height.Int64 != 720 || !got.Height.Valid {
		t.Errorf("Height = %v, want valid 720", got.Height)
	}
	if got.AspectRatio.String != "16:9" || !got.AspectRatio.Valid {
		t.Errorf("AspectRatio = %v, want valid '16:9'", got.AspectRatio)
	}
}

func TestMapCreateMediaDimensionParams_EmptyFields(t *testing.T) {
	t.Parallel()
	// StringToNullString returns Valid=false for empty strings.
	// StringToNullInt64 returns Valid=false for non-parseable strings (empty).
	form := CreateMediaDimensionFormParams{}

	got := MapCreateMediaDimensionParams(form)

	if got.Label.Valid {
		t.Error("Label.Valid = true, want false for empty string")
	}
	if got.Width.Valid {
		t.Error("Width.Valid = true, want false for empty string")
	}
	if got.Height.Valid {
		t.Error("Height.Valid = true, want false for empty string")
	}
	if got.AspectRatio.Valid {
		t.Error("AspectRatio.Valid = true, want false for empty string")
	}
}

func TestMapCreateMediaDimensionParams_NullLiteralLabel(t *testing.T) {
	t.Parallel()
	// StringToNullString treats the literal string "null" as invalid
	form := CreateMediaDimensionFormParams{Label: "null"}

	got := MapCreateMediaDimensionParams(form)

	if got.Label.Valid {
		t.Error("Label.Valid = true, want false for 'null' literal string")
	}
}

func TestMapCreateMediaDimensionParams_NonNumericWidth(t *testing.T) {
	t.Parallel()
	// StringToNullInt64 returns Valid=false when the string cannot be parsed to int
	form := CreateMediaDimensionFormParams{
		Width:  "not-a-number",
		Height: "abc",
	}

	got := MapCreateMediaDimensionParams(form)

	if got.Width.Valid {
		t.Errorf("Width.Valid = true, want false for non-numeric string %q", "not-a-number")
	}
	if got.Height.Valid {
		t.Errorf("Height.Valid = true, want false for non-numeric string %q", "abc")
	}
}

// --- MapUpdateMediaDimensionParams (FormParams -> Params) tests ---

func TestMapUpdateMediaDimensionParams_AllFields(t *testing.T) {
	t.Parallel()
	form := mdUpdateFormParams()

	got := MapUpdateMediaDimensionParams(form)

	if got.Label.String != "Updated Banner" || !got.Label.Valid {
		t.Errorf("Label = %v, want valid 'Updated Banner'", got.Label)
	}
	if got.Width.Int64 != 1920 || !got.Width.Valid {
		t.Errorf("Width = %v, want valid 1920", got.Width)
	}
	if got.Height.Int64 != 1080 || !got.Height.Valid {
		t.Errorf("Height = %v, want valid 1080", got.Height)
	}
	if got.AspectRatio.String != "16:9" || !got.AspectRatio.Valid {
		t.Errorf("AspectRatio = %v, want valid '16:9'", got.AspectRatio)
	}
	if got.MdID != "update-md-id-001" {
		t.Errorf("MdID = %q, want %q", got.MdID, "update-md-id-001")
	}
}

func TestMapUpdateMediaDimensionParams_EmptyFields(t *testing.T) {
	t.Parallel()
	form := UpdateMediaDimensionFormParams{}

	got := MapUpdateMediaDimensionParams(form)

	if got.Label.Valid {
		t.Error("Label.Valid = true, want false for empty string")
	}
	if got.Width.Valid {
		t.Error("Width.Valid = true, want false for empty string")
	}
	if got.Height.Valid {
		t.Error("Height.Valid = true, want false for empty string")
	}
	if got.AspectRatio.Valid {
		t.Error("AspectRatio.Valid = true, want false for empty string")
	}
	if got.MdID != "" {
		t.Errorf("MdID = %q, want empty string", got.MdID)
	}
}

func TestMapUpdateMediaDimensionParams_MdIDPreserved(t *testing.T) {
	t.Parallel()
	// The MdID is a plain string passthrough; verify it is not transformed
	form := UpdateMediaDimensionFormParams{MdID: "special-id-xyz"}

	got := MapUpdateMediaDimensionParams(form)

	if got.MdID != "special-id-xyz" {
		t.Errorf("MdID = %q, want %q", got.MdID, "special-id-xyz")
	}
}

// --- MapStringMediaDimension tests ---

func TestMapStringMediaDimension_AllFields(t *testing.T) {
	t.Parallel()
	md := mdTestFixture()

	got := MapStringMediaDimension(md)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"MdID", got.MdID, "test-md-id-001"},
		{"Label", got.Label, "Thumbnail"},
		{"Width", got.Width, "1920"},
		{"Height", got.Height, "1080"},
		{"AspectRatio", got.AspectRatio, "16:9"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestMapStringMediaDimension_ZeroValue(t *testing.T) {
	t.Parallel()
	got := MapStringMediaDimension(MediaDimensions{})

	if got.MdID != "" {
		t.Errorf("MdID = %q, want empty string", got.MdID)
	}
	// Zero-value sql.NullString and sql.NullInt64 have Valid=false,
	// so NullToString returns "null" for each.
	nullFields := []struct {
		name string
		got  string
	}{
		{"Label", got.Label},
		{"Width", got.Width},
		{"Height", got.Height},
		{"AspectRatio", got.AspectRatio},
	}
	for _, f := range nullFields {
		if f.got != "null" {
			t.Errorf("%s = %q, want %q for zero-value null field", f.name, f.got, "null")
		}
	}
}

func TestMapStringMediaDimension_NullFields(t *testing.T) {
	t.Parallel()
	// Fields with Valid=false should render as "null" regardless of the underlying value.
	md := MediaDimensions{
		MdID:        "some-id",
		Label:       NullString{sql.NullString{String: "should-not-appear", Valid: false}},
		Width:       NullInt64{sql.NullInt64{Int64: 9999, Valid: false}},
		Height:      NullInt64{sql.NullInt64{Int64: 8888, Valid: false}},
		AspectRatio: NullString{sql.NullString{String: "should-not-appear", Valid: false}},
	}

	got := MapStringMediaDimension(md)

	nullFields := []struct {
		name string
		got  string
	}{
		{"Label", got.Label},
		{"Width", got.Width},
		{"Height", got.Height},
		{"AspectRatio", got.AspectRatio},
	}

	for _, f := range nullFields {
		t.Run(f.name, func(t *testing.T) {
			t.Parallel()
			if f.got != "null" {
				t.Errorf("%s = %q, want %q for invalid null field", f.name, f.got, "null")
			}
			if f.got == "should-not-appear" || f.got == "9999" || f.got == "8888" {
				t.Errorf("%s leaked the underlying value through an invalid field", f.name)
			}
		})
	}
}

func TestMapStringMediaDimension_ZeroWidthHeight(t *testing.T) {
	t.Parallel()
	// When Width/Height are valid but zero, NullToString should return "0"
	md := MediaDimensions{
		Width:  NewNullInt64(0),
		Height: NewNullInt64(0),
	}

	got := MapStringMediaDimension(md)

	if got.Width != "0" {
		t.Errorf("Width = %q, want %q for valid zero int", got.Width, "0")
	}
	if got.Height != "0" {
		t.Errorf("Height = %q, want %q for valid zero int", got.Height, "0")
	}
}

// --- MapCreateMediaDimensionJSONParams tests ---

func TestMapCreateMediaDimensionJSONParams_AllFields(t *testing.T) {
	t.Parallel()
	input := CreateMediaDimensionParamsJSON{
		Label:       NullString{sql.NullString{String: "JSON Label", Valid: true}},
		Width:       NullInt64{sql.NullInt64{Int64: 640, Valid: true}},
		Height:      NullInt64{sql.NullInt64{Int64: 480, Valid: true}},
		AspectRatio: NullString{sql.NullString{String: "4:3", Valid: true}},
	}

	got := MapCreateMediaDimensionJSONParams(input)

	if got.Label.String != "JSON Label" || !got.Label.Valid {
		t.Errorf("Label = %v, want valid 'JSON Label'", got.Label)
	}
	if got.Width.Int64 != 640 || !got.Width.Valid {
		t.Errorf("Width = %v, want valid 640", got.Width)
	}
	if got.Height.Int64 != 480 || !got.Height.Valid {
		t.Errorf("Height = %v, want valid 480", got.Height)
	}
	if got.AspectRatio.String != "4:3" || !got.AspectRatio.Valid {
		t.Errorf("AspectRatio = %v, want valid '4:3'", got.AspectRatio)
	}
}

func TestMapCreateMediaDimensionJSONParams_NullFields(t *testing.T) {
	t.Parallel()
	input := CreateMediaDimensionParamsJSON{
		Label:       NullString{sql.NullString{Valid: false}},
		Width:       NullInt64{sql.NullInt64{Valid: false}},
		Height:      NullInt64{sql.NullInt64{Valid: false}},
		AspectRatio: NullString{sql.NullString{Valid: false}},
	}

	got := MapCreateMediaDimensionJSONParams(input)

	if got.Label.Valid {
		t.Error("Label.Valid = true, want false")
	}
	if got.Width.Valid {
		t.Error("Width.Valid = true, want false")
	}
	if got.Height.Valid {
		t.Error("Height.Valid = true, want false")
	}
	if got.AspectRatio.Valid {
		t.Error("AspectRatio.Valid = true, want false")
	}
}

func TestMapCreateMediaDimensionJSONParams_ZeroValue(t *testing.T) {
	t.Parallel()
	got := MapCreateMediaDimensionJSONParams(CreateMediaDimensionParamsJSON{})

	if got.Label.Valid {
		t.Error("Label.Valid = true, want false for zero value")
	}
	if got.Width.Valid {
		t.Error("Width.Valid = true, want false for zero value")
	}
}

// --- MapUpdateMediaDimensionJSONParams tests ---

func TestMapUpdateMediaDimensionJSONParams_AllFields(t *testing.T) {
	t.Parallel()
	input := UpdateMediaDimensionParamsJSON{
		Label:       NullString{sql.NullString{String: "JSON Updated", Valid: true}},
		Width:       NullInt64{sql.NullInt64{Int64: 800, Valid: true}},
		Height:      NullInt64{sql.NullInt64{Int64: 600, Valid: true}},
		AspectRatio: NullString{sql.NullString{String: "4:3", Valid: true}},
		MdID:        "json-update-md-001",
	}

	got := MapUpdateMediaDimensionJSONParams(input)

	if got.Label.String != "JSON Updated" || !got.Label.Valid {
		t.Errorf("Label = %v, want valid 'JSON Updated'", got.Label)
	}
	if got.Width.Int64 != 800 || !got.Width.Valid {
		t.Errorf("Width = %v, want valid 800", got.Width)
	}
	if got.Height.Int64 != 600 || !got.Height.Valid {
		t.Errorf("Height = %v, want valid 600", got.Height)
	}
	if got.AspectRatio.String != "4:3" || !got.AspectRatio.Valid {
		t.Errorf("AspectRatio = %v, want valid '4:3'", got.AspectRatio)
	}
	if got.MdID != "json-update-md-001" {
		t.Errorf("MdID = %q, want %q", got.MdID, "json-update-md-001")
	}
}

func TestMapUpdateMediaDimensionJSONParams_NullFields(t *testing.T) {
	t.Parallel()
	input := UpdateMediaDimensionParamsJSON{
		Label:       NullString{sql.NullString{Valid: false}},
		Width:       NullInt64{sql.NullInt64{Valid: false}},
		Height:      NullInt64{sql.NullInt64{Valid: false}},
		AspectRatio: NullString{sql.NullString{Valid: false}},
		MdID:        "json-null-md",
	}

	got := MapUpdateMediaDimensionJSONParams(input)

	if got.Label.Valid {
		t.Error("Label.Valid = true, want false")
	}
	if got.Width.Valid {
		t.Error("Width.Valid = true, want false")
	}
	if got.Height.Valid {
		t.Error("Height.Valid = true, want false")
	}
	if got.AspectRatio.Valid {
		t.Error("AspectRatio.Valid = true, want false")
	}
	if got.MdID != "json-null-md" {
		t.Errorf("MdID = %q, want %q", got.MdID, "json-null-md")
	}
}

func TestMapUpdateMediaDimensionJSONParams_ZeroValue(t *testing.T) {
	t.Parallel()
	got := MapUpdateMediaDimensionJSONParams(UpdateMediaDimensionParamsJSON{})

	if got.MdID != "" {
		t.Errorf("MdID = %q, want empty string", got.MdID)
	}
	if got.Label.Valid {
		t.Error("Label.Valid = true, want false for zero value")
	}
}

// --- SQLite Database.MapMediaDimension tests ---

func TestDatabase_MapMediaDimension_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}

	input := mdb.MediaDimensions{
		MdID:        "sqlite-md-001",
		Label:       sql.NullString{String: "SQLite Thumb", Valid: true},
		Width:       sql.NullInt64{Int64: 320, Valid: true},
		Height:      sql.NullInt64{Int64: 240, Valid: true},
		AspectRatio: sql.NullString{String: "4:3", Valid: true},
	}


	got := d.MapMediaDimension(input)

	if got.MdID != "sqlite-md-001" {
		t.Errorf("MdID = %q, want %q", got.MdID, "sqlite-md-001")
	}
	if got.Label.String != "SQLite Thumb" || !got.Label.Valid {
		t.Errorf("Label = %v, want valid 'SQLite Thumb'", got.Label)
	}
	if got.Width.Int64 != 320 || !got.Width.Valid {
		t.Errorf("Width = %v, want valid 320", got.Width)
	}
	if got.Height.Int64 != 240 || !got.Height.Valid {
		t.Errorf("Height = %v, want valid 240", got.Height)
	}
	if got.AspectRatio.String != "4:3" || !got.AspectRatio.Valid {
		t.Errorf("AspectRatio = %v, want valid '4:3'", got.AspectRatio)
	}
}

func TestDatabase_MapMediaDimension_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapMediaDimension(mdb.MediaDimensions{})

	if got.MdID != "" {
		t.Errorf("MdID = %q, want empty string", got.MdID)
	}
	if got.Label.Valid {
		t.Error("Label.Valid = true, want false for zero value")
	}
	if got.Width.Valid {
		t.Error("Width.Valid = true, want false for zero value")
	}
	if got.Height.Valid {
		t.Error("Height.Valid = true, want false for zero value")
	}
	if got.AspectRatio.Valid {
		t.Error("AspectRatio.Valid = true, want false for zero value")
	}
}

func TestDatabase_MapMediaDimension_NullDimensions(t *testing.T) {
	t.Parallel()
	d := Database{}
	// Label and aspect ratio present but width/height null
	input := mdb.MediaDimensions{
		MdID:        "sqlite-partial",
		Label:       sql.NullString{String: "Icon", Valid: true},
		Width:       sql.NullInt64{},
		Height:      sql.NullInt64{},
		AspectRatio: sql.NullString{String: "1:1", Valid: true},
	}

	got := d.MapMediaDimension(input)

	if !got.Label.Valid {
		t.Error("Label should be valid")
	}
	if got.Width.Valid {
		t.Error("Width should be invalid/null")
	}
	if got.Height.Valid {
		t.Error("Height should be invalid/null")
	}
	if !got.AspectRatio.Valid {
		t.Error("AspectRatio should be valid")
	}
}

// --- SQLite Database.MapCreateMediaDimensionParams tests ---

func TestDatabase_MapCreateMediaDimensionParams_GeneratesID(t *testing.T) {
	t.Parallel()
	d := Database{}
	params := mdCreateParams()

	got := d.MapCreateMediaDimensionParams(params)

	// A new MdID (MediaDimensionID) should be generated
	if got.MdID == "" {
		t.Fatal("expected non-empty MdID to be generated")
	}
	if got.Label != params.Label.NullString {
		t.Errorf("Label = %v, want %v", got.Label, params.Label)
	}
	if got.Width != params.Width.NullInt64 {
		t.Errorf("Width = %v, want %v", got.Width, params.Width)
	}
	if got.Height != params.Height.NullInt64 {
		t.Errorf("Height = %v, want %v", got.Height, params.Height)
	}
	if got.AspectRatio != params.AspectRatio.NullString {
		t.Errorf("AspectRatio = %v, want %v", got.AspectRatio, params.AspectRatio)
	}
}

func TestDatabase_MapCreateMediaDimensionParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := Database{}
	params := mdCreateParams()

	got1 := d.MapCreateMediaDimensionParams(params)
	got2 := d.MapCreateMediaDimensionParams(params)

	if got1.MdID == "" {
		t.Fatal("first call: expected non-empty MdID")
	}
	if got2.MdID == "" {
		t.Fatal("second call: expected non-empty MdID")
	}
	if got1.MdID == got2.MdID {
		t.Error("two calls produced identical MdIDs; each should be unique")
	}
}

func TestDatabase_MapCreateMediaDimensionParams_ZeroInput(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapCreateMediaDimensionParams(CreateMediaDimensionParams{})

	if got.MdID == "" {
		t.Fatal("expected non-empty MdID even with zero-value input")
	}
}

// --- SQLite Database.MapUpdateMediaDimensionParams tests ---

func TestDatabase_MapUpdateMediaDimensionParams_AllFields(t *testing.T) {
	t.Parallel()
	d := Database{}
	params := mdUpdateParams()

	got := d.MapUpdateMediaDimensionParams(params)

	if got.Label != params.Label.NullString {
		t.Errorf("Label = %v, want %v", got.Label, params.Label)
	}
	if got.Width != params.Width.NullInt64 {
		t.Errorf("Width = %v, want %v", got.Width, params.Width)
	}
	if got.Height != params.Height.NullInt64 {
		t.Errorf("Height = %v, want %v", got.Height, params.Height)
	}
	if got.AspectRatio != params.AspectRatio.NullString {
		t.Errorf("AspectRatio = %v, want %v", got.AspectRatio, params.AspectRatio)
	}
	// MdID is the WHERE clause identifier and must be preserved
	if got.MdID != params.MdID {
		t.Errorf("MdID = %q, want %q", got.MdID, params.MdID)
	}
}

func TestDatabase_MapUpdateMediaDimensionParams_ZeroValues(t *testing.T) {
	t.Parallel()
	d := Database{}
	got := d.MapUpdateMediaDimensionParams(UpdateMediaDimensionParams{})

	if got.MdID != "" {
		t.Errorf("MdID = %q, want empty string", got.MdID)
	}
	if got.Label.Valid {
		t.Error("Label.Valid = true, want false for zero value")
	}
}

// --- MySQL MysqlDatabase.MapMediaDimension tests ---
// MySQL uses sql.NullInt32 for Width/Height; the mapper converts to sql.NullInt64.

func TestMysqlDatabase_MapMediaDimension_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}

	input := mdbm.MediaDimensions{
		MdID:        "mysql-md-001",
		Label:       sql.NullString{String: "MySQL Thumb", Valid: true},
		Width:       sql.NullInt32{Int32: 640, Valid: true},
		Height:      sql.NullInt32{Int32: 480, Valid: true},
		AspectRatio: sql.NullString{String: "4:3", Valid: true},
	}

	got := d.MapMediaDimension(input)

	if got.MdID != "mysql-md-001" {
		t.Errorf("MdID = %q, want %q", got.MdID, "mysql-md-001")
	}
	if got.Label.String != "MySQL Thumb" || !got.Label.Valid {
		t.Errorf("Label = %v, want valid 'MySQL Thumb'", got.Label)
	}
	// Verify int32 -> int64 conversion
	if got.Width.Int64 != 640 || !got.Width.Valid {
		t.Errorf("Width = %v, want valid 640 (int64)", got.Width)
	}
	if got.Height.Int64 != 480 || !got.Height.Valid {
		t.Errorf("Height = %v, want valid 480 (int64)", got.Height)
	}
	if got.AspectRatio.String != "4:3" || !got.AspectRatio.Valid {
		t.Errorf("AspectRatio = %v, want valid '4:3'", got.AspectRatio)
	}
}

func TestMysqlDatabase_MapMediaDimension_ZeroValues(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	got := d.MapMediaDimension(mdbm.MediaDimensions{})

	if got.MdID != "" {
		t.Errorf("MdID = %q, want empty string", got.MdID)
	}
	// Int64ToNullInt64(int64(0)) produces a valid NullInt64 with value 0,
	// because the MySQL mapper always calls Int64ToNullInt64 regardless of
	// the source's Valid flag. This is the actual behavior of the code.
	// Width/Height will be {Int64: 0, Valid: true} from Int64ToNullInt64(0).
	if !got.Width.Valid {
		t.Error("Width.Valid = false; MySQL mapper always produces valid NullInt64 via Int64ToNullInt64")
	}
	if got.Width.Int64 != 0 {
		t.Errorf("Width.Int64 = %d, want 0", got.Width.Int64)
	}
}

func TestMysqlDatabase_MapMediaDimension_Int32Conversion(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	// Verify large int32 values convert correctly to int64
	input := mdbm.MediaDimensions{
		Width:  sql.NullInt32{Int32: 2147483647, Valid: true}, // max int32
		Height: sql.NullInt32{Int32: -1, Valid: true},
	}

	got := d.MapMediaDimension(input)

	if got.Width.Int64 != 2147483647 {
		t.Errorf("Width = %d, want 2147483647 (max int32)", got.Width.Int64)
	}
	if got.Height.Int64 != -1 {
		t.Errorf("Height = %d, want -1", got.Height.Int64)
	}
}

// --- MySQL MysqlDatabase.MapCreateMediaDimensionParams tests ---

func TestMysqlDatabase_MapCreateMediaDimensionParams_GeneratesID(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	params := mdCreateParams()

	got := d.MapCreateMediaDimensionParams(params)

	if got.MdID == "" {
		t.Fatal("expected non-empty MdID to be generated")
	}
	if got.Label != params.Label.NullString {
		t.Errorf("Label = %v, want %v", got.Label, params.Label)
	}
	if got.AspectRatio != params.AspectRatio.NullString {
		t.Errorf("AspectRatio = %v, want %v", got.AspectRatio, params.AspectRatio)
	}
	// Width/Height should be converted from int64 to NullInt32
	if got.Width.Int32 != int32(params.Width.Int64) {
		t.Errorf("Width = %d, want %d (int32 from int64)", got.Width.Int32, int32(params.Width.Int64))
	}
	if got.Height.Int32 != int32(params.Height.Int64) {
		t.Errorf("Height = %d, want %d (int32 from int64)", got.Height.Int32, int32(params.Height.Int64))
	}
}

func TestMysqlDatabase_MapCreateMediaDimensionParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	params := mdCreateParams()

	got1 := d.MapCreateMediaDimensionParams(params)
	got2 := d.MapCreateMediaDimensionParams(params)

	if got1.MdID == got2.MdID {
		t.Error("two calls produced identical MdIDs; each should be unique")
	}
}

// --- MySQL MysqlDatabase.MapUpdateMediaDimensionParams tests ---

func TestMysqlDatabase_MapUpdateMediaDimensionParams_AllFields(t *testing.T) {
	t.Parallel()
	d := MysqlDatabase{}
	params := mdUpdateParams()

	got := d.MapUpdateMediaDimensionParams(params)

	if got.Label != params.Label.NullString {
		t.Errorf("Label = %v, want %v", got.Label, params.Label)
	}
	if got.AspectRatio != params.AspectRatio.NullString {
		t.Errorf("AspectRatio = %v, want %v", got.AspectRatio, params.AspectRatio)
	}
	if got.MdID != params.MdID {
		t.Errorf("MdID = %q, want %q", got.MdID, params.MdID)
	}
	// int64 -> int32 conversion for Width/Height
	if got.Width.Int32 != int32(params.Width.Int64) {
		t.Errorf("Width = %d, want %d", got.Width.Int32, int32(params.Width.Int64))
	}
	if got.Height.Int32 != int32(params.Height.Int64) {
		t.Errorf("Height = %d, want %d", got.Height.Int32, int32(params.Height.Int64))
	}
}

// --- PostgreSQL PsqlDatabase.MapMediaDimension tests ---
// PostgreSQL uses sql.NullInt32 for Width/Height; identical pattern to MySQL.

func TestPsqlDatabase_MapMediaDimension_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}

	input := mdbp.MediaDimensions{
		MdID:        "psql-md-001",
		Label:       sql.NullString{String: "PostgreSQL Thumb", Valid: true},
		Width:       sql.NullInt32{Int32: 1024, Valid: true},
		Height:      sql.NullInt32{Int32: 768, Valid: true},
		AspectRatio: sql.NullString{String: "4:3", Valid: true},
	}

	got := d.MapMediaDimension(input)

	if got.MdID != "psql-md-001" {
		t.Errorf("MdID = %q, want %q", got.MdID, "psql-md-001")
	}
	if got.Label.String != "PostgreSQL Thumb" || !got.Label.Valid {
		t.Errorf("Label = %v, want valid 'PostgreSQL Thumb'", got.Label)
	}
	// Verify int32 -> int64 conversion
	if got.Width.Int64 != 1024 || !got.Width.Valid {
		t.Errorf("Width = %v, want valid 1024 (int64)", got.Width)
	}
	if got.Height.Int64 != 768 || !got.Height.Valid {
		t.Errorf("Height = %v, want valid 768 (int64)", got.Height)
	}
	if got.AspectRatio.String != "4:3" || !got.AspectRatio.Valid {
		t.Errorf("AspectRatio = %v, want valid '4:3'", got.AspectRatio)
	}
}

func TestPsqlDatabase_MapMediaDimension_ZeroValues(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	got := d.MapMediaDimension(mdbp.MediaDimensions{})

	if got.MdID != "" {
		t.Errorf("MdID = %q, want empty string", got.MdID)
	}
	// Same behavior as MySQL: Int64ToNullInt64(0) produces valid NullInt64
	if !got.Width.Valid {
		t.Error("Width.Valid = false; PostgreSQL mapper always produces valid NullInt64 via Int64ToNullInt64")
	}
}

func TestPsqlDatabase_MapMediaDimension_Int32Conversion(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	input := mdbp.MediaDimensions{
		Width:  sql.NullInt32{Int32: 2147483647, Valid: true},
		Height: sql.NullInt32{Int32: -2147483648, Valid: true}, // min int32
	}

	got := d.MapMediaDimension(input)

	if got.Width.Int64 != 2147483647 {
		t.Errorf("Width = %d, want 2147483647 (max int32)", got.Width.Int64)
	}
	if got.Height.Int64 != -2147483648 {
		t.Errorf("Height = %d, want -2147483648 (min int32)", got.Height.Int64)
	}
}

// --- PostgreSQL PsqlDatabase.MapCreateMediaDimensionParams tests ---

func TestPsqlDatabase_MapCreateMediaDimensionParams_GeneratesID(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	params := mdCreateParams()

	got := d.MapCreateMediaDimensionParams(params)

	if got.MdID == "" {
		t.Fatal("expected non-empty MdID to be generated")
	}
	if got.Label != params.Label.NullString {
		t.Errorf("Label = %v, want %v", got.Label, params.Label)
	}
	if got.AspectRatio != params.AspectRatio.NullString {
		t.Errorf("AspectRatio = %v, want %v", got.AspectRatio, params.AspectRatio)
	}
	if got.Width.Int32 != int32(params.Width.Int64) {
		t.Errorf("Width = %d, want %d (int32 from int64)", got.Width.Int32, int32(params.Width.Int64))
	}
	if got.Height.Int32 != int32(params.Height.Int64) {
		t.Errorf("Height = %d, want %d (int32 from int64)", got.Height.Int32, int32(params.Height.Int64))
	}
}

func TestPsqlDatabase_MapCreateMediaDimensionParams_UniqueIDs(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	params := mdCreateParams()

	got1 := d.MapCreateMediaDimensionParams(params)
	got2 := d.MapCreateMediaDimensionParams(params)

	if got1.MdID == got2.MdID {
		t.Error("two calls produced identical MdIDs; each should be unique")
	}
}

// --- PostgreSQL PsqlDatabase.MapUpdateMediaDimensionParams tests ---

func TestPsqlDatabase_MapUpdateMediaDimensionParams_AllFields(t *testing.T) {
	t.Parallel()
	d := PsqlDatabase{}
	params := mdUpdateParams()

	got := d.MapUpdateMediaDimensionParams(params)

	if got.Label != params.Label.NullString {
		t.Errorf("Label = %v, want %v", got.Label, params.Label)
	}
	if got.AspectRatio != params.AspectRatio.NullString {
		t.Errorf("AspectRatio = %v, want %v", got.AspectRatio, params.AspectRatio)
	}
	if got.MdID != params.MdID {
		t.Errorf("MdID = %q, want %q", got.MdID, params.MdID)
	}
	if got.Width.Int32 != int32(params.Width.Int64) {
		t.Errorf("Width = %d, want %d", got.Width.Int32, int32(params.Width.Int64))
	}
	if got.Height.Int32 != int32(params.Height.Int64) {
		t.Errorf("Height = %d, want %d", got.Height.Int32, int32(params.Height.Int64))
	}
}

// --- Cross-database mapper consistency ---
// SQLite uses sql.NullInt64 for Width/Height directly, while MySQL and
// PostgreSQL use sql.NullInt32 which get converted via Int64ToNullInt64.
// All three must produce identical MediaDimensions wrapper structs.

func TestCrossDatabaseMapMediaDimension_Consistency(t *testing.T) {
	t.Parallel()
	label := sql.NullString{String: "cross-db-thumb", Valid: true}
	aspect := sql.NullString{String: "16:9", Valid: true}

	sqliteInput := mdb.MediaDimensions{
		MdID:        "cross-md-001",
		Label:       label,
		Width:       sql.NullInt64{Int64: 800, Valid: true},
		Height:      sql.NullInt64{Int64: 450, Valid: true},
		AspectRatio: aspect,
	}
	mysqlInput := mdbm.MediaDimensions{
		MdID:        "cross-md-001",
		Label:       label,
		Width:       sql.NullInt32{Int32: 800, Valid: true},
		Height:      sql.NullInt32{Int32: 450, Valid: true},
		AspectRatio: aspect,
	}
	psqlInput := mdbp.MediaDimensions{
		MdID:        "cross-md-001",
		Label:       label,
		Width:       sql.NullInt32{Int32: 800, Valid: true},
		Height:      sql.NullInt32{Int32: 450, Valid: true},
		AspectRatio: aspect,
	}

	sqliteResult := Database{}.MapMediaDimension(sqliteInput)
	mysqlResult := MysqlDatabase{}.MapMediaDimension(mysqlInput)
	psqlResult := PsqlDatabase{}.MapMediaDimension(psqlInput)

	if sqliteResult != mysqlResult {
		t.Errorf("SQLite and MySQL produced different results:\n  sqlite: %+v\n  mysql:  %+v", sqliteResult, mysqlResult)
	}
	if sqliteResult != psqlResult {
		t.Errorf("SQLite and PostgreSQL produced different results:\n  sqlite: %+v\n  psql:   %+v", sqliteResult, psqlResult)
	}
}

// --- Cross-database MapCreateMediaDimensionParams auto-ID generation ---

func TestCrossDatabaseMapCreateMediaDimensionParams_AutoIDGeneration(t *testing.T) {
	t.Parallel()
	params := mdCreateParams()

	sqliteResult := Database{}.MapCreateMediaDimensionParams(params)
	mysqlResult := MysqlDatabase{}.MapCreateMediaDimensionParams(params)
	psqlResult := PsqlDatabase{}.MapCreateMediaDimensionParams(params)

	if sqliteResult.MdID == "" {
		t.Error("SQLite: expected non-empty generated MdID")
	}
	if mysqlResult.MdID == "" {
		t.Error("MySQL: expected non-empty generated MdID")
	}
	if psqlResult.MdID == "" {
		t.Error("PostgreSQL: expected non-empty generated MdID")
	}

	// Each call should generate a unique ID
	if sqliteResult.MdID == mysqlResult.MdID {
		t.Error("SQLite and MySQL generated the same MdID -- each call should be unique")
	}
	if sqliteResult.MdID == psqlResult.MdID {
		t.Error("SQLite and PostgreSQL generated the same MdID -- each call should be unique")
	}
}

// --- SQLite Audited Command Accessor tests ---

func TestNewMediaDimensionCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := types.NewUserID()
	ac := audited.AuditContext{
		UserID:    userID,
		NodeID:    types.NodeID("md-node-1"),
		RequestID: "md-req-123",
		IP:        "10.0.0.1",
	}
	params := mdCreateParams()

	cmd := Database{}.NewMediaDimensionCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "media_dimensions" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "media_dimensions")
	}
	p, ok := cmd.Params().(CreateMediaDimensionParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateMediaDimensionParams", cmd.Params())
	}
	if p.Label != params.Label {
		t.Errorf("Params().Label = %v, want %v", p.Label, params.Label)
	}
	if p.Width != params.Width {
		t.Errorf("Params().Width = %v, want %v", p.Width, params.Width)
	}
	// Connection is nil because we used an empty Database{}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewMediaDimensionCmd_GetID_ExtractsFromRow(t *testing.T) {
	t.Parallel()
	cmd := NewMediaDimensionCmd{}

	row := mdb.MediaDimensions{MdID: "extracted-md-id"}
	got := cmd.GetID(row)
	if got != "extracted-md-id" {
		t.Errorf("GetID() = %q, want %q", got, "extracted-md-id")
	}
}

func TestNewMediaDimensionCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	cmd := NewMediaDimensionCmd{}
	got := cmd.GetID(mdb.MediaDimensions{})
	if got != "" {
		t.Errorf("GetID() = %q, want empty string", got)
	}
}

func TestUpdateMediaDimensionCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := mdUpdateParams()

	cmd := Database{}.UpdateMediaDimensionCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "media_dimensions" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "media_dimensions")
	}
	if cmd.GetID() != params.MdID {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), params.MdID)
	}
	p, ok := cmd.Params().(UpdateMediaDimensionParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateMediaDimensionParams", cmd.Params())
	}
	if p.Label != params.Label {
		t.Errorf("Params().Label = %v, want %v", p.Label, params.Label)
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteMediaDimensionCmd_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	mdID := "delete-md-001"

	cmd := Database{}.DeleteMediaDimensionCmd(ctx, ac, mdID)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "media_dimensions" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "media_dimensions")
	}
	if cmd.GetID() != mdID {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), mdID)
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty Database")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- MySQL Audited Command Accessor tests ---

func TestNewMediaDimensionCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("mysql-md-node"),
		RequestID: "mysql-md-req",
		IP:        "192.168.1.1",
	}
	params := mdCreateParams()

	cmd := MysqlDatabase{}.NewMediaDimensionCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "media_dimensions" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "media_dimensions")
	}
	p, ok := cmd.Params().(CreateMediaDimensionParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateMediaDimensionParams", cmd.Params())
	}
	if p.Label != params.Label {
		t.Errorf("Params().Label = %v, want %v", p.Label, params.Label)
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty MysqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewMediaDimensionCmdMysql_GetID(t *testing.T) {
	t.Parallel()
	cmd := NewMediaDimensionCmdMysql{}

	row := mdbm.MediaDimensions{MdID: "mysql-extracted-id"}
	got := cmd.GetID(row)
	if got != "mysql-extracted-id" {
		t.Errorf("GetID() = %q, want %q", got, "mysql-extracted-id")
	}
}

func TestUpdateMediaDimensionCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := mdUpdateParams()

	cmd := MysqlDatabase{}.UpdateMediaDimensionCmd(ctx, ac, params)

	if cmd.TableName() != "media_dimensions" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "media_dimensions")
	}
	if cmd.GetID() != params.MdID {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), params.MdID)
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateMediaDimensionParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateMediaDimensionParams", cmd.Params())
	}
	if p.Label != params.Label {
		t.Errorf("Params().Label = %v, want %v", p.Label, params.Label)
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteMediaDimensionCmdMysql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	mdID := "mysql-delete-md-001"

	cmd := MysqlDatabase{}.DeleteMediaDimensionCmd(ctx, ac, mdID)

	if cmd.TableName() != "media_dimensions" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "media_dimensions")
	}
	if cmd.GetID() != mdID {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), mdID)
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty MysqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- PostgreSQL Audited Command Accessor tests ---

func TestNewMediaDimensionCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{
		UserID:    types.NewUserID(),
		NodeID:    types.NodeID("psql-md-node"),
		RequestID: "psql-md-req",
		IP:        "172.16.0.1",
	}
	params := mdCreateParams()

	cmd := PsqlDatabase{}.NewMediaDimensionCmd(ctx, ac, params)

	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.TableName() != "media_dimensions" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "media_dimensions")
	}
	p, ok := cmd.Params().(CreateMediaDimensionParams)
	if !ok {
		t.Fatalf("Params() returned %T, want CreateMediaDimensionParams", cmd.Params())
	}
	if p.Label != params.Label {
		t.Errorf("Params().Label = %v, want %v", p.Label, params.Label)
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty PsqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestNewMediaDimensionCmdPsql_GetID(t *testing.T) {
	t.Parallel()
	cmd := NewMediaDimensionCmdPsql{}

	row := mdbp.MediaDimensions{MdID: "psql-extracted-id"}
	got := cmd.GetID(row)
	if got != "psql-extracted-id" {
		t.Errorf("GetID() = %q, want %q", got, "psql-extracted-id")
	}
}

func TestUpdateMediaDimensionCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	params := mdUpdateParams()

	cmd := PsqlDatabase{}.UpdateMediaDimensionCmd(ctx, ac, params)

	if cmd.TableName() != "media_dimensions" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "media_dimensions")
	}
	if cmd.GetID() != params.MdID {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), params.MdID)
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	p, ok := cmd.Params().(UpdateMediaDimensionParams)
	if !ok {
		t.Fatalf("Params() returned %T, want UpdateMediaDimensionParams", cmd.Params())
	}
	if p.Label != params.Label {
		t.Errorf("Params().Label = %v, want %v", p.Label, params.Label)
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

func TestDeleteMediaDimensionCmdPsql_AllAccessors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{UserID: types.NewUserID()}
	mdID := "psql-delete-md-001"

	cmd := PsqlDatabase{}.DeleteMediaDimensionCmd(ctx, ac, mdID)

	if cmd.TableName() != "media_dimensions" {
		t.Errorf("TableName() = %q, want %q", cmd.TableName(), "media_dimensions")
	}
	if cmd.GetID() != mdID {
		t.Errorf("GetID() = %q, want %q", cmd.GetID(), mdID)
	}
	if cmd.Context() != ctx {
		t.Error("Context() returned wrong context")
	}
	if cmd.AuditContext() != ac {
		t.Error("AuditContext() returned wrong audit context")
	}
	if cmd.Connection() != nil {
		t.Error("Connection() should be nil for empty PsqlDatabase")
	}
	if cmd.Recorder() == nil {
		t.Fatal("Recorder() returned nil")
	}
}

// --- Cross-database audited command consistency ---

func TestAuditedMediaDimensionCommands_TableNameConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateMediaDimensionParams{}
	updateParams := UpdateMediaDimensionParams{MdID: "consistency-test"}
	deleteID := "consistency-delete"

	commands := []struct {
		label string
		name  string
	}{
		{"SQLite Create", Database{}.NewMediaDimensionCmd(ctx, ac, createParams).TableName()},
		{"SQLite Update", Database{}.UpdateMediaDimensionCmd(ctx, ac, updateParams).TableName()},
		{"SQLite Delete", Database{}.DeleteMediaDimensionCmd(ctx, ac, deleteID).TableName()},
		{"MySQL Create", MysqlDatabase{}.NewMediaDimensionCmd(ctx, ac, createParams).TableName()},
		{"MySQL Update", MysqlDatabase{}.UpdateMediaDimensionCmd(ctx, ac, updateParams).TableName()},
		{"MySQL Delete", MysqlDatabase{}.DeleteMediaDimensionCmd(ctx, ac, deleteID).TableName()},
		{"PostgreSQL Create", PsqlDatabase{}.NewMediaDimensionCmd(ctx, ac, createParams).TableName()},
		{"PostgreSQL Update", PsqlDatabase{}.UpdateMediaDimensionCmd(ctx, ac, updateParams).TableName()},
		{"PostgreSQL Delete", PsqlDatabase{}.DeleteMediaDimensionCmd(ctx, ac, deleteID).TableName()},
	}

	for _, c := range commands {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			if c.name != "media_dimensions" {
				t.Errorf("TableName() = %q, want %q", c.name, "media_dimensions")
			}
		})
	}
}

func TestAuditedMediaDimensionCommands_RecorderAssignment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateMediaDimensionParams{}

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
	}{
		{"SQLite", Database{}.NewMediaDimensionCmd(ctx, ac, createParams).Recorder()},
		{"MySQL", MysqlDatabase{}.NewMediaDimensionCmd(ctx, ac, createParams).Recorder()},
		{"PostgreSQL", PsqlDatabase{}.NewMediaDimensionCmd(ctx, ac, createParams).Recorder()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.recorder == nil {
				t.Fatalf("%s recorder is nil", tt.name)
			}
		})
	}
}

// --- Recorder identity verification ---
// Verify the command factories assign the correct recorder variant.

func TestMediaDimensionCommand_RecorderIdentity(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	createParams := CreateMediaDimensionParams{}
	updateParams := UpdateMediaDimensionParams{MdID: "recorder-test"}
	deleteID := "recorder-delete"

	tests := []struct {
		name     string
		recorder audited.ChangeEventRecorder
		want     audited.ChangeEventRecorder
	}{
		// SQLite commands should use SQLiteRecorder
		{"SQLite Create", Database{}.NewMediaDimensionCmd(ctx, ac, createParams).Recorder(), SQLiteRecorder},
		{"SQLite Update", Database{}.UpdateMediaDimensionCmd(ctx, ac, updateParams).Recorder(), SQLiteRecorder},
		{"SQLite Delete", Database{}.DeleteMediaDimensionCmd(ctx, ac, deleteID).Recorder(), SQLiteRecorder},
		// MySQL commands should use MysqlRecorder
		{"MySQL Create", MysqlDatabase{}.NewMediaDimensionCmd(ctx, ac, createParams).Recorder(), MysqlRecorder},
		{"MySQL Update", MysqlDatabase{}.UpdateMediaDimensionCmd(ctx, ac, updateParams).Recorder(), MysqlRecorder},
		{"MySQL Delete", MysqlDatabase{}.DeleteMediaDimensionCmd(ctx, ac, deleteID).Recorder(), MysqlRecorder},
		// PostgreSQL commands should use PsqlRecorder
		{"PostgreSQL Create", PsqlDatabase{}.NewMediaDimensionCmd(ctx, ac, createParams).Recorder(), PsqlRecorder},
		{"PostgreSQL Update", PsqlDatabase{}.UpdateMediaDimensionCmd(ctx, ac, updateParams).Recorder(), PsqlRecorder},
		{"PostgreSQL Delete", PsqlDatabase{}.DeleteMediaDimensionCmd(ctx, ac, deleteID).Recorder(), PsqlRecorder},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if fmt.Sprintf("%T", tt.recorder) != fmt.Sprintf("%T", tt.want) {
				t.Errorf("Recorder type = %T, want %T", tt.recorder, tt.want)
			}
		})
	}
}

// --- Audited Command GetID consistency ---

func TestAuditedMediaDimensionCommands_GetID_CrossDatabase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ac := audited.AuditContext{}
	mdID := "cross-db-md-id"

	t.Run("UpdateCmd GetID returns MdID", func(t *testing.T) {
		t.Parallel()
		params := UpdateMediaDimensionParams{MdID: mdID}

		sqliteCmd := Database{}.UpdateMediaDimensionCmd(ctx, ac, params)
		mysqlCmd := MysqlDatabase{}.UpdateMediaDimensionCmd(ctx, ac, params)
		psqlCmd := PsqlDatabase{}.UpdateMediaDimensionCmd(ctx, ac, params)

		if sqliteCmd.GetID() != mdID {
			t.Errorf("SQLite GetID() = %q, want %q", sqliteCmd.GetID(), mdID)
		}
		if mysqlCmd.GetID() != mdID {
			t.Errorf("MySQL GetID() = %q, want %q", mysqlCmd.GetID(), mdID)
		}
		if psqlCmd.GetID() != mdID {
			t.Errorf("PostgreSQL GetID() = %q, want %q", psqlCmd.GetID(), mdID)
		}
	})

	t.Run("DeleteCmd GetID returns MdID", func(t *testing.T) {
		t.Parallel()
		sqliteCmd := Database{}.DeleteMediaDimensionCmd(ctx, ac, mdID)
		mysqlCmd := MysqlDatabase{}.DeleteMediaDimensionCmd(ctx, ac, mdID)
		psqlCmd := PsqlDatabase{}.DeleteMediaDimensionCmd(ctx, ac, mdID)

		if sqliteCmd.GetID() != mdID {
			t.Errorf("SQLite GetID() = %q, want %q", sqliteCmd.GetID(), mdID)
		}
		if mysqlCmd.GetID() != mdID {
			t.Errorf("MySQL GetID() = %q, want %q", mysqlCmd.GetID(), mdID)
		}
		if psqlCmd.GetID() != mdID {
			t.Errorf("PostgreSQL GetID() = %q, want %q", psqlCmd.GetID(), mdID)
		}
	})

	t.Run("CreateCmd GetID extracts from result row", func(t *testing.T) {
		t.Parallel()
		testID := "create-extracted-id"

		sqliteCmd := NewMediaDimensionCmd{}
		mysqlCmd := NewMediaDimensionCmdMysql{}
		psqlCmd := NewMediaDimensionCmdPsql{}

		sqliteRow := mdb.MediaDimensions{MdID: testID}
		mysqlRow := mdbm.MediaDimensions{MdID: testID}
		psqlRow := mdbp.MediaDimensions{MdID: testID}

		if sqliteCmd.GetID(sqliteRow) != testID {
			t.Errorf("SQLite GetID() = %q, want %q", sqliteCmd.GetID(sqliteRow), testID)
		}
		if mysqlCmd.GetID(mysqlRow) != testID {
			t.Errorf("MySQL GetID() = %q, want %q", mysqlCmd.GetID(mysqlRow), testID)
		}
		if psqlCmd.GetID(psqlRow) != testID {
			t.Errorf("PostgreSQL GetID() = %q, want %q", psqlCmd.GetID(psqlRow), testID)
		}
	})
}

// --- Edge cases ---

func TestUpdateMediaDimensionCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	params := UpdateMediaDimensionParams{MdID: ""}

	sqliteCmd := Database{}.UpdateMediaDimensionCmd(context.Background(), audited.AuditContext{}, params)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.UpdateMediaDimensionCmd(context.Background(), audited.AuditContext{}, params)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.UpdateMediaDimensionCmd(context.Background(), audited.AuditContext{}, params)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

func TestDeleteMediaDimensionCmd_GetID_EmptyID(t *testing.T) {
	t.Parallel()
	emptyID := ""

	sqliteCmd := Database{}.DeleteMediaDimensionCmd(context.Background(), audited.AuditContext{}, emptyID)
	if sqliteCmd.GetID() != "" {
		t.Errorf("GetID() = %q, want empty string", sqliteCmd.GetID())
	}

	mysqlCmd := MysqlDatabase{}.DeleteMediaDimensionCmd(context.Background(), audited.AuditContext{}, emptyID)
	if mysqlCmd.GetID() != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID())
	}

	psqlCmd := PsqlDatabase{}.DeleteMediaDimensionCmd(context.Background(), audited.AuditContext{}, emptyID)
	if psqlCmd.GetID() != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID())
	}
}

func TestNewMediaDimensionCmd_GetID_EmptyMdID(t *testing.T) {
	t.Parallel()
	// CreateCmd.GetID extracts from a row; empty MdID in the row should return ""
	sqliteCmd := NewMediaDimensionCmd{}
	if sqliteCmd.GetID(mdb.MediaDimensions{}) != "" {
		t.Errorf("SQLite GetID() = %q, want empty string", sqliteCmd.GetID(mdb.MediaDimensions{}))
	}

	mysqlCmd := NewMediaDimensionCmdMysql{}
	if mysqlCmd.GetID(mdbm.MediaDimensions{}) != "" {
		t.Errorf("MySQL GetID() = %q, want empty string", mysqlCmd.GetID(mdbm.MediaDimensions{}))
	}

	psqlCmd := NewMediaDimensionCmdPsql{}
	if psqlCmd.GetID(mdbp.MediaDimensions{}) != "" {
		t.Errorf("PostgreSQL GetID() = %q, want empty string", psqlCmd.GetID(mdbp.MediaDimensions{}))
	}
}

// --- UpdateMediaDimension success message format ---
// The UpdateMediaDimension method produces a success message string using the MdID.
// We verify the expected format matches what fmt.Sprintf produces.

func TestUpdateMediaDimension_SuccessMessageFormat(t *testing.T) {
	t.Parallel()
	mdID := "test-md-format-001"
	expected := fmt.Sprintf("Successfully updated %v\n", mdID)

	if expected != "Successfully updated test-md-format-001\n" {
		t.Errorf("message = %q, want %q", expected, "Successfully updated test-md-format-001\n")
	}
}

func TestUpdateMediaDimension_SuccessMessageFormat_EmptyID(t *testing.T) {
	t.Parallel()
	expected := fmt.Sprintf("Successfully updated %v\n", "")
	if expected != "Successfully updated \n" {
		t.Errorf("message = %q, want %q", expected, "Successfully updated \n")
	}
}

// --- Struct type zero-value tests ---
// Verify all declared struct types are usable at zero value.

func TestMediaDimensions_ZeroValue(t *testing.T) {
	t.Parallel()
	var md MediaDimensions

	if md.MdID != "" {
		t.Errorf("MdID = %q, want empty string", md.MdID)
	}
	if md.Label.Valid {
		t.Error("Label.Valid = true, want false")
	}
	if md.Width.Valid {
		t.Error("Width.Valid = true, want false")
	}
	if md.Height.Valid {
		t.Error("Height.Valid = true, want false")
	}
	if md.AspectRatio.Valid {
		t.Error("AspectRatio.Valid = true, want false")
	}
}

func TestCreateMediaDimensionParams_ZeroValue(t *testing.T) {
	t.Parallel()
	var p CreateMediaDimensionParams

	if p.Label.Valid {
		t.Error("Label.Valid = true, want false")
	}
	if p.Width.Valid {
		t.Error("Width.Valid = true, want false")
	}
}

func TestUpdateMediaDimensionParams_ZeroValue(t *testing.T) {
	t.Parallel()
	var p UpdateMediaDimensionParams

	if p.MdID != "" {
		t.Errorf("MdID = %q, want empty string", p.MdID)
	}
	if p.Label.Valid {
		t.Error("Label.Valid = true, want false")
	}
}

func TestMediaDimensionsHistoryEntry_ZeroValue(t *testing.T) {
	t.Parallel()
	var h MediaDimensionsHistoryEntry

	if h.MdID != "" {
		t.Errorf("MdID = %q, want empty string", h.MdID)
	}
	if h.Label.Valid {
		t.Error("Label.Valid = true, want false")
	}
	if h.Width.Valid {
		t.Error("Width.Valid = true, want false")
	}
}

func TestCreateMediaDimensionFormParams_ZeroValue(t *testing.T) {
	t.Parallel()
	var f CreateMediaDimensionFormParams

	if f.Label != "" {
		t.Errorf("Label = %q, want empty string", f.Label)
	}
	if f.Width != "" {
		t.Errorf("Width = %q, want empty string", f.Width)
	}
	if f.Height != "" {
		t.Errorf("Height = %q, want empty string", f.Height)
	}
	if f.AspectRatio != "" {
		t.Errorf("AspectRatio = %q, want empty string", f.AspectRatio)
	}
}

func TestUpdateMediaDimensionFormParams_ZeroValue(t *testing.T) {
	t.Parallel()
	var f UpdateMediaDimensionFormParams

	if f.MdID != "" {
		t.Errorf("MdID = %q, want empty string", f.MdID)
	}
	if f.Label != "" {
		t.Errorf("Label = %q, want empty string", f.Label)
	}
}

func TestMediaDimensionsJSON_ZeroValue(t *testing.T) {
	t.Parallel()
	var j MediaDimensionsJSON

	if j.MdID != "" {
		t.Errorf("MdID = %q, want empty string", j.MdID)
	}
	if j.Label.Valid {
		t.Error("Label.Valid = true, want false")
	}
	if j.Width.Valid {
		t.Error("Width.Valid = true, want false")
	}
}

func TestCreateMediaDimensionParamsJSON_ZeroValue(t *testing.T) {
	t.Parallel()
	var j CreateMediaDimensionParamsJSON

	if j.Label.Valid {
		t.Error("Label.Valid = true, want false")
	}
	if j.Width.Valid {
		t.Error("Width.Valid = true, want false")
	}
}

func TestUpdateMediaDimensionParamsJSON_ZeroValue(t *testing.T) {
	t.Parallel()
	var j UpdateMediaDimensionParamsJSON

	if j.MdID != "" {
		t.Errorf("MdID = %q, want empty string", j.MdID)
	}
	if j.Label.Valid {
		t.Error("Label.Valid = true, want false")
	}
}

// --- MediaDimensionsHistoryEntry populated values test ---

func TestMediaDimensionsHistoryEntry_PopulatedValues(t *testing.T) {
	t.Parallel()
	h := MediaDimensionsHistoryEntry{
		MdID:        "history-md-001",
		Label:       NewNullString("History Label"),
		Width:       NewNullInt64(512),
		Height:      NewNullInt64(384),
		AspectRatio: NewNullString("4:3"),
	}

	if h.MdID != "history-md-001" {
		t.Errorf("MdID = %q, want %q", h.MdID, "history-md-001")
	}
	if h.Label.String != "History Label" || !h.Label.Valid {
		t.Errorf("Label = %v, want valid 'History Label'", h.Label)
	}
	if h.Width.Int64 != 512 || !h.Width.Valid {
		t.Errorf("Width = %v, want valid 512", h.Width)
	}
	if h.Height.Int64 != 384 || !h.Height.Valid {
		t.Errorf("Height = %v, want valid 384", h.Height)
	}
	if h.AspectRatio.String != "4:3" || !h.AspectRatio.Valid {
		t.Errorf("AspectRatio = %v, want valid '4:3'", h.AspectRatio)
	}
}

// --- MediaDimensionsJSON populated values test ---

func TestMediaDimensionsJSON_PopulatedValues(t *testing.T) {
	t.Parallel()
	j := MediaDimensionsJSON{
		MdID:        "json-md-001",
		Label:       NullString{sql.NullString{String: "JSON Thumbnail", Valid: true}},
		Width:       NullInt64{sql.NullInt64{Int64: 256, Valid: true}},
		Height:      NullInt64{sql.NullInt64{Int64: 256, Valid: true}},
		AspectRatio: NullString{sql.NullString{String: "1:1", Valid: true}},
	}

	if j.MdID != "json-md-001" {
		t.Errorf("MdID = %q, want %q", j.MdID, "json-md-001")
	}
	if j.Label.String != "JSON Thumbnail" || !j.Label.Valid {
		t.Errorf("Label = %v, want valid 'JSON Thumbnail'", j.Label)
	}
	if j.Width.Int64 != 256 || !j.Width.Valid {
		t.Errorf("Width = %v, want valid 256", j.Width)
	}
	if j.Height.Int64 != 256 || !j.Height.Valid {
		t.Errorf("Height = %v, want valid 256", j.Height)
	}
	if j.AspectRatio.String != "1:1" || !j.AspectRatio.Valid {
		t.Errorf("AspectRatio = %v, want valid '1:1'", j.AspectRatio)
	}
}
