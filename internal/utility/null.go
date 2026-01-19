package utility

import (
	"database/sql"
	"strconv"
)

// Nullable is a constraint for all sql.Null* types
type Nullable interface {
	sql.NullInt64 | sql.NullInt32 | sql.NullInt16 | sql.NullString | sql.NullByte | sql.NullFloat64 | sql.NullTime | sql.NullBool
}

// IsNull checks if a sql.Null* type has a valid value
// Generic version that works with all nullable database types
func IsNull[T Nullable](value T) bool {
	// This works because all sql.Null* types have a Valid field
	switch v := any(value).(type) {
	case sql.NullInt64:
		return v.Valid
	case sql.NullInt32:
		return v.Valid
	case sql.NullInt16:
		return v.Valid
	case sql.NullString:
		return v.Valid
	case sql.NullByte:
		return v.Valid
	case sql.NullFloat64:
		return v.Valid
	case sql.NullTime:
		return v.Valid
	case sql.NullBool:
		return v.Valid
	default:
		return false
	}
}

// NullToString converts any sql.Null* type to string, returning "null" if invalid
// Generic version that works with all nullable database types
func NullToString[T Nullable](value T) string {
	switch v := any(value).(type) {
	case sql.NullInt64:
		if v.Valid {
			return strconv.FormatInt(v.Int64, 10)
		}
	case sql.NullInt32:
		if v.Valid {
			return strconv.FormatInt(int64(v.Int32), 10)
		}
	case sql.NullInt16:
		if v.Valid {
			return strconv.FormatInt(int64(v.Int16), 10)
		}
	case sql.NullString:
		if v.Valid {
			return v.String
		}
	case sql.NullByte:
		if v.Valid {
			return strconv.FormatUint(uint64(v.Byte), 10)
		}
	case sql.NullFloat64:
		if v.Valid {
			return strconv.FormatFloat(v.Float64, 'f', -1, 64)
		}
	case sql.NullTime:
		if v.Valid {
			return v.Time.Format("2006-01-02 15:04:05")
		}
	case sql.NullBool:
		if v.Valid {
			return strconv.FormatBool(v.Bool)
		}
	}
	return "null"
}

