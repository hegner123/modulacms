package definitions

import "github.com/hegner123/modulacms/internal/db/types"

// FieldDef describes a field to create within a schema definition.
// IDs and timestamps are set at install time, not at definition time.
type FieldDef struct {
	Label      string               // Display name: "Title", "Content", "Featured Image"
	Type       types.FieldType      //
	Data       types.NullableString // expected
	UiConfig   types.UIConfig
	Validation types.ValidationConfig
}

// DatatypeDef describes a datatype and which fields/children it has.
type DatatypeDef struct {
	Label     string               // Display name: "Page", "Blog Post"
	Type      types.NullableString // Category: "page", "post", "ROOT", "GLOBAL"
	ParentRef string               // Keys into SchemaDefinition.Datatypes (parent-child hierarchy)
	FieldRefs []FieldDef           // Keys into SchemaDefinition.Fields
}

// SchemaDefinition is the installable unit containing all datatypes and fields
// for a particular CMS schema pattern.
type SchemaDefinition struct {
	Name        string                 // Unique ID: "modulacms-default", "wordpress-blog"
	Label       string                 // Display name: "ModulaCMS Default"
	Description string                 // What this schema provides
	Format      string                 // Source format hint: "modulacms", "wordpress", etc.
	Datatypes   map[string]DatatypeDef // All datatypes, keyed by local reference
}
