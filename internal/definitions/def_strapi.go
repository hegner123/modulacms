package definitions

import "github.com/hegner123/modulacms/internal/db/types"

func init() {
	Register(SchemaDefinition{
		Name:        "strapi-starter",
		Label:       "Strapi Starter",
		Description: "Article, page, and global datatypes matching Strapi collection/single type patterns",
		Format:      "strapi",
		Datatypes: map[string]DatatypeDef{
			"article": {
				Name:  "article",
				Label: "Article",
				Type:  types.NewNullableString(string(types.DatatypeTypeRoot)),
				FieldRefs: []FieldDef{
					{Name: "title", Label: "Title", Type: types.FieldTypeText},
					{Name: "slug", Label: "Slug", Type: types.FieldTypeSlug},
					{Name: "description", Label: "Description", Type: types.FieldTypeTextarea},
					{Name: "content", Label: "Content", Type: types.FieldTypeRichText},
					{Name: "cover", Label: "Cover", Type: types.FieldTypeMedia},
					{Name: "published_at", Label: "Published At", Type: types.FieldTypeDatetime},
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
				},
			},
			"global": {
				Name:  "global",
				Label: "Global",
				Type:  types.NewNullableString(string(types.DatatypeTypeRoot)),
				FieldRefs: []FieldDef{
					{Name: "site_name", Label: "Site Name", Type: types.FieldTypeText},
					{Name: "site_description", Label: "Site Description", Type: types.FieldTypeTextarea},
					{Name: "site_logo", Label: "Site Logo", Type: types.FieldTypeMedia},
				},
			},
		},
	})
}
