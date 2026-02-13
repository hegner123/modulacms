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

func DBTableString(t DBTable) string {
	return string(t)
}

func StringDBTable(t string) DBTable {
	return DBTable(t)
}
func TimeToNullString(s time.Time) sql.NullString {
	return sql.NullString{String: s.String(), Valid: true}
}

func IntToNullInt64(i int) sql.NullInt64 {
	return sql.NullInt64{Int64: int64(i), Valid: true}
}

func Int64ToNullInt64(i int64) sql.NullInt64 {
	return sql.NullInt64{Int64: int64(i), Valid: true}
}

func Int64ToNullInt32(i int64) sql.NullInt32 {
	return sql.NullInt32{Int32: int32(i), Valid: true}
}

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

func BoolToNullBool(b bool) sql.NullBool {
	return sql.NullBool{Bool: b, Valid: true}
}

func TimeToNullTime(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: true}
}

func JSONRawToString(j json.RawMessage) string {
	s, err := j.MarshalJSON()
	if err != nil {
		return fmt.Sprint(err)
	}
	return string(s)
}

func PSQLstring(pS pqtype.NullRawMessage) string {
	return JSONRawToString(pS.RawMessage)
}

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

func NullStringToTime(s sql.NullString) time.Time {
	t, err := time.Parse(time.RFC3339, s.String)
	if err != nil {
		return time.Now()
	}
	return t
}
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

func StringToInt64(s string) int64 {
	res, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	} else {
		return res
	}
}

func StringToBool(s string) bool {
	res, err := strconv.ParseBool(s)
	if err != nil {
		return false
	} else {
		return res
	}
}

func ParseBool(s string) bool {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return false
	}
	return b

}

func ParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Now()
	}
	return t

}

func AssertString(i any) string {
	s, ok := i.(string)
	if !ok {
		return fmt.Sprint(i)
	}
	return s
}

func AssertInteger(i any) int {
	d, ok := i.(int)
	if !ok {
		return 0
	}
	return d
}

func AssertInt32(i any) int32 {
	d, ok := i.(int32)
	if !ok {
		return 0
	}
	return d
}

func AssertInt64(i any) int64 {
	d, ok := i.(int64)
	if !ok {
		return 0
	}
	return d
}

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

func StringToNullBool(s string) sql.NullBool {
	var nb sql.NullBool
	if b, err := strconv.ParseBool(s); err == nil {
		nb = sql.NullBool{Bool: b, Valid: true}
	} else {
		nb = sql.NullBool{Bool: false, Valid: false}
	}
	return nb
}

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

func ReadNullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	} else {
		return "null"
	}
}
func ReadNullInt64(ns sql.NullInt64) string {
	if ns.Valid {
		return strconv.FormatInt(ns.Int64, 10)
	} else {
		return "null"
	}
}
func ReadNullInt32(ns sql.NullInt32) string {
	if ns.Valid {
		return strconv.FormatInt(int64(ns.Int32), 10)
	} else {
		return "null"
	}
}
func ReadNullTime(ns sql.NullTime) string {
	if ns.Valid {
		return ns.Time.String()
	} else {
		return "null"
	}
}
func ReadNullBool(ns sql.NullBool) string {
	if ns.Valid {
		return strconv.FormatBool(ns.Bool)
	} else {
		return "null"
	}
}
