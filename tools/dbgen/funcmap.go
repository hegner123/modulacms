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
	}
}
