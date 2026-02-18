package definitions

import "github.com/hegner123/modulacms/internal/db/types"

func init() {
	Register(SchemaDefinition{
		Name:        "wordpress-blog",
		Label:       "WordPress Blog",
		Description: "Post and page datatypes matching WordPress content structure",
		Format:      "wordpress",
		Datatypes: map[string]DatatypeDef{
			"post": {
				Label: "Post",
				Type:  types.NewNullableString("ROOT"),
				FieldRefs: []FieldDef{
					{Label: "Title", Type: types.FieldTypeText},
					{Label: "Slug", Type: types.FieldTypeSlug},
					{Label: "Content", Type: types.FieldTypeRichText},
					{Label: "Excerpt", Type: types.FieldTypeTextarea},
					{Label: "Featured Image", Type: types.FieldTypeMedia},
					{Label: "Published", Type: types.FieldTypeBoolean},
					{Label: "Category", Type: types.FieldTypeSelect, Data: types.NewNullableString(`{"options":["uncategorized"]}`)},
					{Label: "Tags", Type: types.FieldTypeText},
				},
			},
			"page": {
				Label: "Page",
				Type:  types.NewNullableString("ROOT"),
				FieldRefs: []FieldDef{
					{Label: "Page Title", Type: types.FieldTypeText},
					{Label: "Page Slug", Type: types.FieldTypeSlug},
					{Label: "Page Content", Type: types.FieldTypeRichText},
					{Label: "Page Featured Image", Type: types.FieldTypeMedia},
					{Label: "Published", Type: types.FieldTypeBoolean},
				},
			},
		},
	})
}
