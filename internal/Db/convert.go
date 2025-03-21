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


func nt(t sql.NullTime) string {
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

func Nb(b bool) sql.NullBool {
	return sql.NullBool{Bool: b, Valid: true}
}

func ns(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}

func ni64(i int64) sql.NullInt64 {
	return sql.NullInt64{Int64: int64(i), Valid: true}
}

func NTT(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: true}
}

func jrS(j json.RawMessage) string {
	s, err := j.MarshalJSON()
	if err != nil {
		return fmt.Sprint(err)
	}
	return string(s)
}

func pString(pS pqtype.NullRawMessage) string {
	return jrS(pS.RawMessage)
}

func sTime(dateStr string) sql.NullTime {
	st := sql.NullTime{}
	err := st.Scan(dateStr)
	if err != nil {
		st.Valid = false
		return st
	}
	return st
}

func NSt(s sql.NullString) time.Time {
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

func sb(s string) bool {
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

func Nsi(s string) sql.NullInt64 {
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

/*
func getColumnValue(column string, s interface{}) string {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if strings.EqualFold(field.Name, column) {
			fieldValue := v.Field(i)
			if fieldValue.Kind() == reflect.String {
				return fieldValue.String()
			}
			return fmt.Sprintf("%v", fieldValue.Interface())
		}
	}

	return ""
}
*/
