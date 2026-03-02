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
				Name:  "post",
				Label: "Post",
				Type:  types.NewNullableString(string(types.DatatypeTypeRoot)),
				FieldRefs: []FieldDef{
					{Name: "title", Label: "Title", Type: types.FieldTypeText},
					{Name: "slug", Label: "Slug", Type: types.FieldTypeSlug},
					{Name: "body", Label: "Body", Type: types.FieldTypeRichText},
					{Name: "main_image", Label: "Main Image", Type: types.FieldTypeMedia},
					{Name: "published_at", Label: "Published At", Type: types.FieldTypeDatetime},
				},
			},
			"author": {
				Name:  "author",
				Label: "Author",
				Type:  types.NewNullableString(string(types.DatatypeTypeRoot)),
				FieldRefs: []FieldDef{
					{Name: "name", Label: "Name", Type: types.FieldTypeText},
					{Name: "bio", Label: "Bio", Type: types.FieldTypeTextarea},
					{Name: "image", Label: "Image", Type: types.FieldTypeMedia},
				},
			},
			"category": {
				Name:  "category",
				Label: "Category",
				Type:  types.NewNullableString(string(types.DatatypeTypeRoot)),
				FieldRefs: []FieldDef{
					{Name: "category_title", Label: "Category Title", Type: types.FieldTypeText},
					{Name: "category_description", Label: "Category Description", Type: types.FieldTypeTextarea},
				},
			},
		},
	})
}
