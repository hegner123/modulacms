package cli

import "database/sql"

func IsNullInt64(value sql.NullInt64) bool {
	return value.Valid
}
func IsNullInt32(value sql.NullInt32) bool {
	return value.Valid
}
func IsNullInt16(value sql.NullInt16) bool {
	return value.Valid
}
func IsNullString(value sql.NullString) bool {
	return value.Valid
}
func IsNullByte(value sql.NullByte) bool {
	return value.Valid
}
func IsNullFloat64(value sql.NullFloat64) bool {
	return value.Valid
}
func IsNullTime(value sql.NullTime) bool {
	return value.Valid
}
func IsNullBool(value sql.NullBool) bool {
	return value.Valid
}
