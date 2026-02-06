package definitions

import "github.com/hegner123/modulacms/internal/db/types"

func init() {
	Register(SchemaDefinition{
		Name:        "sanity-starter",
		Label:       "Sanity Starter",
		Description: "Post, author, and category datatypes matching Sanity document patterns",
		Format:      "sanity",
		Fields: map[string]FieldDef{
			"title": {
				Label: "Title",
				Type:  types.FieldTypeText,
			},
			"slug": {
				Label: "Slug",
				Type:  types.FieldTypeSlug,
			},
			"body": {
				Label: "Body",
				Type:  types.FieldTypeRichText,
			},
			"main_image": {
				Label: "Main Image",
				Type:  types.FieldTypeMedia,
			},
			"published_at": {
				Label: "Published At",
				Type:  types.FieldTypeDatetime,
			},
			"name": {
				Label: "Name",
				Type:  types.FieldTypeText,
			},
			"bio": {
				Label: "Bio",
				Type:  types.FieldTypeTextarea,
			},
			"image": {
				Label: "Image",
				Type:  types.FieldTypeMedia,
			},
			"cat_title": {
				Label: "Category Title",
				Type:  types.FieldTypeText,
			},
			"cat_desc": {
				Label: "Category Description",
				Type:  types.FieldTypeTextarea,
			},
		},
		Datatypes: map[string]DatatypeDef{
			"post": {
				Label:     "Post",
				Type:      "post",
				FieldRefs: []string{"title", "slug", "body", "main_image", "published_at"},
			},
			"author": {
				Label:     "Author",
				Type:      "author",
				FieldRefs: []string{"name", "bio", "image"},
			},
			"category": {
				Label:     "Category",
				Type:      "category",
				FieldRefs: []string{"cat_title", "cat_desc"},
			},
		},
		RootKeys: []string{"post", "author", "category"},
	})
}
