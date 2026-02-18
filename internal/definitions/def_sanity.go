package definitions

import "github.com/hegner123/modulacms/internal/db/types"

func init() {
	Register(SchemaDefinition{
		Name:        "sanity-starter",
		Label:       "Sanity Starter",
		Description: "Post, author, and category datatypes matching Sanity document patterns",
		Format:      "sanity",
		Datatypes: map[string]DatatypeDef{
			"post": {
				Label: "Post",
				Type:  types.NewNullableString("ROOT"),
				FieldRefs: []FieldDef{
					{Label: "Title", Type: types.FieldTypeText},
					{Label: "Slug", Type: types.FieldTypeSlug},
					{Label: "Body", Type: types.FieldTypeRichText},
					{Label: "Main Image", Type: types.FieldTypeMedia},
					{Label: "Published At", Type: types.FieldTypeDatetime},
				},
			},
			"author": {
				Label: "Author",
				Type:  types.NewNullableString("ROOT"),
				FieldRefs: []FieldDef{
					{Label: "Name", Type: types.FieldTypeText},
					{Label: "Bio", Type: types.FieldTypeTextarea},
					{Label: "Image", Type: types.FieldTypeMedia},
				},
			},
			"category": {
				Label: "Category",
				Type:  types.NewNullableString("ROOT"),
				FieldRefs: []FieldDef{
					{Label: "Category Title", Type: types.FieldTypeText},
					{Label: "Category Description", Type: types.FieldTypeTextarea},
				},
			},
		},
	})
}
