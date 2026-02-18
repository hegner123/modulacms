package definitions

import "github.com/hegner123/modulacms/internal/db/types"

func init() {
	Register(SchemaDefinition{
		Name:        "contentful-starter",
		Label:       "Contentful Starter",
		Description: "Page, blog post, and author datatypes matching Contentful entry structure",
		Format:      "contentful",
		Datatypes: map[string]DatatypeDef{
			"page": {
				Label: "Page",
				Type:  types.NewNullableString("ROOT"),
				FieldRefs: []FieldDef{
					{Label: "Title", Type: types.FieldTypeText},
					{Label: "Slug", Type: types.FieldTypeSlug},
					{Label: "Body", Type: types.FieldTypeRichText},
					{Label: "Hero Image", Type: types.FieldTypeMedia},
				},
			},
			"blog_post": {
				Label: "Blog Post",
				Type:  types.NewNullableString("ROOT"),
				FieldRefs: []FieldDef{
					{Label: "Title", Type: types.FieldTypeText},
					{Label: "Slug", Type: types.FieldTypeSlug},
					{Label: "Body", Type: types.FieldTypeRichText},
					{Label: "Description", Type: types.FieldTypeTextarea},
					{Label: "Hero Image", Type: types.FieldTypeMedia},
					{Label: "Published Date", Type: types.FieldTypeDatetime},
				},
			},
			"author": {
				Label: "Author",
				Type:  types.NewNullableString("ROOT"),
				FieldRefs: []FieldDef{
					{Label: "Author Name", Type: types.FieldTypeText},
					{Label: "Author Bio", Type: types.FieldTypeTextarea},
					{Label: "Author Avatar", Type: types.FieldTypeMedia},
				},
			},
		},
	})
}
