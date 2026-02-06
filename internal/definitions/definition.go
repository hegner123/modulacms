package definitions

import "github.com/hegner123/modulacms/internal/db/types"

// FieldDef describes a field to create within a schema definition.
// IDs and timestamps are set at install time, not at definition time.
type FieldDef struct {
	Label string          // Display name: "Title", "Body", "Featured Image"
	Type  types.FieldType // text, richtext, media, boolean, etc.
	Data  string          // JSON config (validation rules, select options, etc.)
}

// DatatypeDef describes a datatype and which fields/children it has.
type DatatypeDef struct {
	Label     string   // Display name: "Page", "Blog Post"
	Type      string   // Category: "page", "post", "ROOT", "GLOBAL"
	FieldRefs []string // Keys into SchemaDefinition.Fields
	ChildRefs []string // Keys into SchemaDefinition.Datatypes (parent-child hierarchy)
}

// SchemaDefinition is the installable unit containing all datatypes and fields
// for a particular CMS schema pattern.
type SchemaDefinition struct {
	Name        string                 // Unique ID: "modulacms-default", "wordpress-blog"
	Label       string                 // Display name: "ModulaCMS Default"
	Description string                 // What this schema provides
	Format      string                 // Source format hint: "modulacms", "wordpress", etc.
	Fields      map[string]FieldDef    // All fields, keyed by local reference
	Datatypes   map[string]DatatypeDef // All datatypes, keyed by local reference
	RootKeys    []string               // Which datatype keys have no parent
}
