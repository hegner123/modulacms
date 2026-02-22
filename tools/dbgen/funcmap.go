package main

import (
	"strings"
	"text/template"
)

// templateFuncMap returns the FuncMap for code generation templates.
func templateFuncMap() template.FuncMap {
	return template.FuncMap{
		// lower returns the string with its first character lowercased.
		"lower": func(s string) string {
			if s == "" {
				return s
			}
			return strings.ToLower(s[:1]) + s[1:]
		},

		// driverEntity combines a driver config with entity data for sub-templates.
		"driverEntity": func(e Entity, d DriverConfig) DriverEntityData {
			return DriverEntityData{Entity: e, Driver: d}
		},

		// driverLabel returns a human-readable driver label for comments.
		"driverLabel": func(d DriverConfig) string {
			switch d.Name {
			case "sqlite":
				return "SQLITE"
			case "mysql":
				return "MYSQL"
			case "psql":
				return "POSTGRES"
			}
			return strings.ToUpper(d.Name)
		},

		// driverCommentLabel returns a label for audited command comments.
		"driverCommentLabel": func(d DriverConfig) string {
			switch d.Name {
			case "sqlite":
				return "SQLite"
			case "mysql":
				return "MySQL"
			case "psql":
				return "PostgreSQL"
			}
			return d.Name
		},

		// paginationCast wraps a value in int32() if the driver uses int32 pagination.
		"paginationCast": func(d DriverConfig, value string) string {
			if d.Int32Pagination {
				return "int32(" + value + ")"
			}
			return value
		},

		// idFieldInGetParams returns the ID field name used in Get/Delete params.
		// For most entities this is the same as IDField, but some sqlc params use different names.
		"idFieldInGetParams": func(e Entity) string {
			return e.IDField
		},

		// isNullWrapper returns true if the type is a JSON-aware null wrapper
		// (NullString, NullInt32, NullInt64, NullTime) that embeds its sql.Null* counterpart.
		"isNullWrapper": func(typ string) bool {
			switch typ {
			case "NullString", "NullInt32", "NullInt64", "NullTime":
				return true
			}
			return false
		},

		// stringExpr returns the Go expression to convert a field to string based on StringConvert value.
		"stringExpr": func(convert, expr string) string {
			switch convert {
			case "toString":
				return expr + ".String()"
			case "string":
				return expr
			case "sprintf":
				return "fmt.Sprintf(\"%d\", " + expr + ")"
			case "cast":
				return "string(" + expr + ")"
			case "nullToString":
				return "utility.NullToString(" + expr + ")"
			case "wrapperNullToString":
				return "utility.NullToString(" + expr + ".NullString)"
			case "nullToEmpty":
				return "NullStringToEmpty(" + expr + ")"
			case "wrapperNullToEmpty":
				return "NullStringToEmpty(" + expr + ".NullString)"
			case "wrapperNullInt64ToString":
				return "utility.NullToString(" + expr + ".NullInt64)"
			case "nullableIDToEmpty":
				return "nullableIDToEmpty(" + expr + ")"
			case "sprintfBool":
				return "fmt.Sprintf(\"%t\", " + expr + ")"
			case "sprintfFloat64":
				return "fmt.Sprintf(\"%v\", " + expr + ".Float64)"
			}
			return expr
		},

		// wrapParam applies an optional wrapping expression to a parameter value.
		// If wrapExpr contains %s, it is replaced with the value; otherwise value is returned as-is.
		"wrapParam": func(wrapExpr, value string) string {
			if wrapExpr == "" {
				return value
			}
			return strings.Replace(wrapExpr, "%s", value, 1)
		},

		// sqlcExtraQueryName returns the sqlc function name for an ExtraQuery.
		"sqlcExtraQueryName": func(e Entity, eq ExtraQuery) string {
			return e.SqlcExtraQueryName(eq)
		},

		// sqlcPaginatedQueryName returns the sqlc function name for a PaginatedExtraQuery.
		"sqlcPaginatedQueryName": func(e Entity, pq PaginatedExtraQuery) string {
			return e.SqlcPaginatedQueryName(pq)
		},
	}
}
