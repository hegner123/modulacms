package definitions

import "github.com/hegner123/modulacms/internal/db/types"

func init() {
	Register(SchemaDefinition{
		Name:        "strapi-starter",
		Label:       "Strapi Starter",
		Description: "Article, page, and global datatypes matching Strapi collection/single type patterns",
		Format:      "strapi",
		Fields: map[string]FieldDef{
			"title": {
				Label: "Title",
				Type:  types.FieldTypeText,
			},
			"slug": {
				Label: "Slug",
				Type:  types.FieldTypeSlug,
			},
			"description": {
				Label: "Description",
				Type:  types.FieldTypeTextarea,
			},
			"content": {
				Label: "Content",
				Type:  types.FieldTypeRichText,
			},
			"cover": {
				Label: "Cover",
				Type:  types.FieldTypeMedia,
			},
			"published_at": {
				Label: "Published At",
				Type:  types.FieldTypeDatetime,
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
			"site_name": {
				Label: "Site Name",
				Type:  types.FieldTypeText,
			},
			"site_desc": {
				Label: "Site Description",
				Type:  types.FieldTypeTextarea,
			},
			"site_logo": {
				Label: "Site Logo",
				Type:  types.FieldTypeMedia,
			},
		},
		Datatypes: map[string]DatatypeDef{
			"article": {
				Label:     "Article",
				Type:      "article",
				FieldRefs: []string{"title", "slug", "description", "content", "cover", "published_at"},
			},
			"page": {
				Label:     "Page",
				Type:      "page",
				FieldRefs: []string{"page_title", "page_slug", "page_content"},
			},
			"global": {
				Label:     "Global",
				Type:      "GLOBAL",
				FieldRefs: []string{"site_name", "site_desc", "site_logo"},
			},
		},
		RootKeys: []string{"article", "page", "global"},
	})
}
