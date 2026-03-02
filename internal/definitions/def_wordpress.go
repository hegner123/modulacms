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
				Name:  "post",
				Label: "Post",
				Type:  types.NewNullableString(string(types.DatatypeTypeRoot)),
				FieldRefs: []FieldDef{
					{Name: "title", Label: "Title", Type: types.FieldTypeText},
					{Name: "slug", Label: "Slug", Type: types.FieldTypeSlug},
					{Name: "content", Label: "Content", Type: types.FieldTypeRichText},
					{Name: "excerpt", Label: "Excerpt", Type: types.FieldTypeTextarea},
					{Name: "featured_image", Label: "Featured Image", Type: types.FieldTypeMedia},
					{Name: "published", Label: "Published", Type: types.FieldTypeBoolean},
					{Name: "category", Label: "Category", Type: types.FieldTypeSelect, Data: types.NewNullableString(`{"options":["uncategorized"]}`)},
					{Name: "tags", Label: "Tags", Type: types.FieldTypeText},
				},
			},
			"page": {
				Name:  "page",
				Label: "Page",
				Type:  types.NewNullableString(string(types.DatatypeTypeRoot)),
				FieldRefs: []FieldDef{
					{Name: "page_title", Label: "Page Title", Type: types.FieldTypeText},
					{Name: "page_slug", Label: "Page Slug", Type: types.FieldTypeSlug},
					{Name: "page_content", Label: "Page Content", Type: types.FieldTypeRichText},
					{Name: "page_featured_image", Label: "Page Featured Image", Type: types.FieldTypeMedia},
					{Name: "published", Label: "Published", Type: types.FieldTypeBoolean},
				},
			},
		},
	})
}
