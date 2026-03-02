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
				Name:  "page",
				Label: "Page",
				Type:  types.NewNullableString(string(types.DatatypeTypeRoot)),
				FieldRefs: []FieldDef{
					{Name: "title", Label: "Title", Type: types.FieldTypeText},
					{Name: "slug", Label: "Slug", Type: types.FieldTypeSlug},
					{Name: "body", Label: "Body", Type: types.FieldTypeRichText},
					{Name: "hero_image", Label: "Hero Image", Type: types.FieldTypeMedia},
				},
			},
			"blog_post": {
				Name:  "blog_post",
				Label: "Blog Post",
				Type:  types.NewNullableString(string(types.DatatypeTypeRoot)),
				FieldRefs: []FieldDef{
					{Name: "title", Label: "Title", Type: types.FieldTypeText},
					{Name: "slug", Label: "Slug", Type: types.FieldTypeSlug},
					{Name: "body", Label: "Body", Type: types.FieldTypeRichText},
					{Name: "description", Label: "Description", Type: types.FieldTypeTextarea},
					{Name: "hero_image", Label: "Hero Image", Type: types.FieldTypeMedia},
					{Name: "published_date", Label: "Published Date", Type: types.FieldTypeDatetime},
				},
			},
			"author": {
				Name:  "author",
				Label: "Author",
				Type:  types.NewNullableString(string(types.DatatypeTypeRoot)),
				FieldRefs: []FieldDef{
					{Name: "author_name", Label: "Author Name", Type: types.FieldTypeText},
					{Name: "author_bio", Label: "Author Bio", Type: types.FieldTypeTextarea},
					{Name: "author_avatar", Label: "Author Avatar", Type: types.FieldTypeMedia},
				},
			},
		},
	})
}
