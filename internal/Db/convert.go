package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/sqlc-dev/pqtype"
)

func DBTableString(t DBTable) string {
	return string(t)
}

func StringDBTable(t string) DBTable {
	return DBTable(t)
}

func Ns(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}

func Ni(i int) sql.NullInt64 {
	return sql.NullInt64{Int64: int64(i), Valid: true}
}

func Ni64(i int64) sql.NullInt64 {
	return sql.NullInt64{Int64: int64(i), Valid: true}
}

func Ni32(i int64) sql.NullInt32 {
	return sql.NullInt32{Int32: int32(i), Valid: true}
}

func Ni32Ni64(i sql.NullInt32) sql.NullInt64 {
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
func Ni64Ni32(i sql.NullInt64) sql.NullInt32 {
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

func Nb(b bool) sql.NullBool {
	return sql.NullBool{Bool: b, Valid: true}
}

func NTT(t time.Time) sql.NullTime {
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
	st := sql.NullTime{}
	err := st.Scan(dateStr)
	if err != nil {
		st.Valid = false
		return st
	}
	return st
}

func Nt(t sql.NullTime) string {
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

func NStringToTime(s sql.NullString) time.Time {
	t, err := time.Parse(time.RFC3339, s.String)
	if err != nil {
		return time.Now()
	}
	return t
}

func Si(s string) int64 {
	res, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	} else {
		return res
	}
}

func Sb(s string) bool {
	res, err := strconv.ParseBool(s)
	if err != nil {
		return false
	} else {
		return res
	}
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

func SNi64(s string) sql.NullInt64 {
	var res sql.NullInt64
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		res.Int64 = 0
		res.Valid = false
		return res
	} else {
		res.Int64 = i
		res.Valid = true
		return res
	}

}
func SNi32(s string) sql.NullInt32 {
	var res sql.NullInt32
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		res.Int32 = 0
		res.Valid = false
		return res
	} else {
		res.Int32 = int32(i)
		res.Valid = true
		return res
	}

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
