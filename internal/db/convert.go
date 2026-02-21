package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	utility "github.com/hegner123/modulacms/internal/utility"
	"github.com/sqlc-dev/pqtype"
)

// NullStringToEmpty returns the string value if valid, or "" if null.
func NullStringToEmpty(n sql.NullString) string {
	if n.Valid {
		return n.String
	}
	return ""
}

// nullableStringer is implemented by nullable ID types that can report validity.
type nullableStringer interface {
	IsZero() bool
	String() string
}

// nullableIDToEmpty returns "" if the nullable ID is zero/invalid, or the string value otherwise.
func nullableIDToEmpty(n nullableStringer) string {
	if n.IsZero() {
		return ""
	}
	return n.String()
}

// DBTableString converts a DBTable to its string representation.
func DBTableString(t DBTable) string {
	return string(t)
}

// StringDBTable converts a string to a DBTable.
func StringDBTable(t string) DBTable {
	return DBTable(t)
}

// TimeToNullString converts a time.Time to a sql.NullString.
func TimeToNullString(s time.Time) sql.NullString {
	return sql.NullString{String: s.String(), Valid: true}
}

// IntToNullInt64 converts an int to a sql.NullInt64.
func IntToNullInt64(i int) sql.NullInt64 {
	return sql.NullInt64{Int64: int64(i), Valid: true}
}

// Int64ToNullInt64 converts an int64 to a sql.NullInt64.
func Int64ToNullInt64(i int64) sql.NullInt64 {
	return sql.NullInt64{Int64: int64(i), Valid: true}
}

// Int64ToNullInt32 converts an int64 to a sql.NullInt32.
func Int64ToNullInt32(i int64) sql.NullInt32 {
	return sql.NullInt32{Int32: int32(i), Valid: true}
}

// NullInt32ToNullInt64 converts a sql.NullInt32 to a sql.NullInt64.
func NullInt32ToNullInt64(i sql.NullInt32) sql.NullInt64 {
	if i.Valid {
		return sql.NullInt64{
			Int64: int64(i.Int32),
			Valid: true,
		}
	} else {
		return sql.NullInt64{
			Int64: 0,
			Valid: false,
		}
	}
}

// NullInt64ToNullInt32 converts a sql.NullInt64 to a sql.NullInt32.
func NullInt64ToNullInt32(i sql.NullInt64) sql.NullInt32 {
	if i.Valid {
		return sql.NullInt32{
			Int32: int32(i.Int64),
			Valid: true,
		}
	} else {
		return sql.NullInt32{
			Int32: 0,
			Valid: false,
		}
	}
}

// BoolToNullBool converts a bool to a sql.NullBool.
func BoolToNullBool(b bool) sql.NullBool {
	return sql.NullBool{Bool: b, Valid: true}
}

// TimeToNullTime converts a time.Time to a sql.NullTime.
func TimeToNullTime(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: true}
}

// JSONRawToString converts a json.RawMessage to a string.
func JSONRawToString(j json.RawMessage) string {
	s, err := j.MarshalJSON()
	if err != nil {
		return fmt.Sprint(err)
	}
	return string(s)
}

// PSQLstring converts a pqtype.NullRawMessage to a string.
func PSQLstring(pS pqtype.NullRawMessage) string {
	return JSONRawToString(pS.RawMessage)
}

// StringToNTime converts a string to a sql.NullTime, supporting Unix timestamps and RFC3339 format.
func StringToNTime(dateStr string) sql.NullTime {
	if dateStr == "" {
		return sql.NullTime{Valid: false}
	}
	// Try parsing as Unix timestamp (numeric string)
	if unix, err := strconv.ParseInt(dateStr, 10, 64); err == nil {
		return sql.NullTime{Time: time.Unix(unix, 0).UTC(), Valid: true}
	}
	// Try RFC3339
	if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return sql.NullTime{Time: t, Valid: true}
	}
	// Try common datetime format
	if t, err := time.Parse("2006-01-02 15:04:05", dateStr); err == nil {
		return sql.NullTime{Time: t, Valid: true}
	}
	return sql.NullTime{Valid: false}
}

// NullTimeToString converts a sql.NullTime to a string.
func NullTimeToString(t sql.NullTime) string {
	v, err := t.Value()
	if err != nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

// NullStringToTime converts a sql.NullString to a time.Time, defaulting to now on parse error.
func NullStringToTime(s sql.NullString) time.Time {
	t, err := time.Parse(time.RFC3339, s.String)
	if err != nil {
		return time.Now()
	}
	return t
}

// NullStringToNullTime converts a sql.NullString to a sql.NullTime.
func NullStringToNullTime(s sql.NullString) sql.NullTime {
	ns := ReadNullString(s)
	t := utility.ParseTimeReadable(ns)
	nt := sql.NullTime{}
	if t != nil {
		nt = sql.NullTime{Time: *t, Valid: true}
	} else {
		nt = sql.NullTime{Time: time.Time{}, Valid: false}
	}
	return nt
}

// StringToInt64 converts a string to an int64, returning 0 on parse error.
func StringToInt64(s string) int64 {
	res, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	} else {
		return res
	}
}

// StringToBool converts a string to a bool, returning false on parse error.
func StringToBool(s string) bool {
	res, err := strconv.ParseBool(s)
	if err != nil {
		return false
	} else {
		return res
	}
}

// ParseBool parses a string to a bool, returning false on parse error.
func ParseBool(s string) bool {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return false
	}
	return b

}

// ParseTime parses a string to a time.Time in RFC3339 format, defaulting to now on parse error.
func ParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Now()
	}
	return t

}

// AssertString asserts that a value is a string, converting to string if not.
func AssertString(i any) string {
	s, ok := i.(string)
	if !ok {
		return fmt.Sprint(i)
	}
	return s
}

// AssertInteger asserts that a value is an int, returning 0 if not.
func AssertInteger(i any) int {
	d, ok := i.(int)
	if !ok {
		return 0
	}
	return d
}

// AssertInt32 asserts that a value is an int32, returning 0 if not.
func AssertInt32(i any) int32 {
	d, ok := i.(int32)
	if !ok {
		return 0
	}
	return d
}

// AssertInt64 asserts that a value is an int64, returning 0 if not.
func AssertInt64(i any) int64 {
	d, ok := i.(int64)
	if !ok {
		return 0
	}
	return d
}

// StringToNullString converts a string to a sql.NullString, treating empty string and "null" as invalid.
func StringToNullString(s string) sql.NullString {
	switch s {
	case "":
		return sql.NullString{String: s, Valid: false}
	case "null":
		return sql.NullString{String: s, Valid: false}
	default:
		return sql.NullString{String: s, Valid: true}
	}
}

// StringToNullInt64 converts a string to a sql.NullInt64, returning invalid on parse error.
func StringToNullInt64(s string) sql.NullInt64 {
	var res sql.NullInt64
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		res.Valid = false
		return res
	} else {
		res.Int64 = i
		res.Valid = true
		return res
	}

}

// StringToNullInt32 converts a string to a sql.NullInt32, returning invalid on parse error.
func StringToNullInt32(s string) sql.NullInt32 {
	var res sql.NullInt32
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		res.Valid = false
		return res
	} else {
		res.Int32 = int32(i)
		res.Valid = true
		return res
	}

}

// StringToNullBool converts a string to a sql.NullBool, returning invalid on parse error.
func StringToNullBool(s string) sql.NullBool {
	var nb sql.NullBool
	if b, err := strconv.ParseBool(s); err == nil {
		nb = sql.NullBool{Bool: b, Valid: true}
	} else {
		nb = sql.NullBool{Bool: false, Valid: false}
	}
	return nb
}

// StringToNullTime converts a string to a sql.NullTime using readable time parsing.
func StringToNullTime(s string) sql.NullTime {
	var nt sql.NullTime
	t := utility.ParseTimeReadable(s)
	if t != nil {
		nt = sql.NullTime{Time: *t, Valid: true}
	} else {
		nt = sql.NullTime{Time: time.Time{}, Valid: false}
	}
	return nt
}

// ReadNullString reads a sql.NullString as a string, returning "null" if invalid.
func ReadNullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	} else {
		return "null"
	}
}

// ReadNullInt64 reads a sql.NullInt64 as a string, returning "null" if invalid.
func ReadNullInt64(ns sql.NullInt64) string {
	if ns.Valid {
		return strconv.FormatInt(ns.Int64, 10)
	} else {
		return "null"
	}
}

// ReadNullInt32 reads a sql.NullInt32 as a string, returning "null" if invalid.
func ReadNullInt32(ns sql.NullInt32) string {
	if ns.Valid {
		return strconv.FormatInt(int64(ns.Int32), 10)
	} else {
		return "null"
	}
}

// ReadNullTime reads a sql.NullTime as a string, returning "null" if invalid.
func ReadNullTime(ns sql.NullTime) string {
	if ns.Valid {
		return ns.Time.String()
	} else {
		return "null"
	}
}

// ReadNullBool reads a sql.NullBool as a string, returning "null" if invalid.
func ReadNullBool(ns sql.NullBool) string {
	if ns.Valid {
		return strconv.FormatBool(ns.Bool)
	} else {
		return "null"
	}
}
