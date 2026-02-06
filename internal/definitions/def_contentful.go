package definitions

import "github.com/hegner123/modulacms/internal/db/types"

func init() {
	Register(SchemaDefinition{
		Name:        "contentful-starter",
		Label:       "Contentful Starter",
		Description: "Page, blog post, and author datatypes matching Contentful entry structure",
		Format:      "contentful",
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
			"description": {
				Label: "Description",
				Type:  types.FieldTypeTextarea,
			},
			"hero_image": {
				Label: "Hero Image",
				Type:  types.FieldTypeMedia,
			},
			"published_date": {
				Label: "Published Date",
				Type:  types.FieldTypeDatetime,
			},
			"author_name": {
				Label: "Author Name",
				Type:  types.FieldTypeText,
			},
			"author_bio": {
				Label: "Author Bio",
				Type:  types.FieldTypeTextarea,
			},
			"author_avatar": {
				Label: "Author Avatar",
				Type:  types.FieldTypeMedia,
			},
		},
		Datatypes: map[string]DatatypeDef{
			"page": {
				Label:     "Page",
				Type:      "page",
				FieldRefs: []string{"title", "slug", "body", "hero_image"},
			},
			"blog_post": {
				Label:     "Blog Post",
				Type:      "post",
				FieldRefs: []string{"title", "slug", "body", "description", "hero_image", "published_date"},
			},
			"author": {
				Label:     "Author",
				Type:      "author",
				FieldRefs: []string{"author_name", "author_bio", "author_avatar"},
			},
		},
		RootKeys: []string{"page", "blog_post", "author"},
	})
}
