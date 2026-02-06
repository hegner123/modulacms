package definitions

import "github.com/hegner123/modulacms/internal/db/types"

func init() {
	Register(SchemaDefinition{
		Name:        "wordpress-blog",
		Label:       "WordPress Blog",
		Description: "Post and page datatypes matching WordPress content structure",
		Format:      "wordpress",
		Fields: map[string]FieldDef{
			"title": {
				Label: "Title",
				Type:  types.FieldTypeText,
			},
			"slug": {
				Label: "Slug",
				Type:  types.FieldTypeSlug,
			},
			"content": {
				Label: "Content",
				Type:  types.FieldTypeRichText,
			},
			"excerpt": {
				Label: "Excerpt",
				Type:  types.FieldTypeTextarea,
			},
			"featured_image": {
				Label: "Featured Image",
				Type:  types.FieldTypeMedia,
			},
			"published": {
				Label: "Published",
				Type:  types.FieldTypeBoolean,
			},
			"category": {
				Label: "Category",
				Type:  types.FieldTypeSelect,
				Data:  `{"options":["uncategorized"]}`,
			},
			"tags": {
				Label: "Tags",
				Type:  types.FieldTypeText,
			},
			"page_title": {
				Label: "Page Title",
				Type:  types.FieldTypeText,
			},
			"page_slug": {
				Label: "Page Slug",
				Type:  types.FieldTypeSlug,
			},
			"page_content": {
				Label: "Page Content",
				Type:  types.FieldTypeRichText,
			},
			"page_featured_image": {
				Label: "Page Featured Image",
				Type:  types.FieldTypeMedia,
			},
		},
		Datatypes: map[string]DatatypeDef{
			"post": {
				Label:     "Post",
				Type:      "post",
				FieldRefs: []string{"title", "slug", "content", "excerpt", "featured_image", "published", "category", "tags"},
			},
			"page": {
				Label:     "Page",
				Type:      "page",
				FieldRefs: []string{"page_title", "page_slug", "page_content", "page_featured_image", "published"},
			},
		},
		RootKeys: []string{"post", "page"},
	})
}
