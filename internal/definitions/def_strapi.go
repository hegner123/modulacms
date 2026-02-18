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
				Label: "Article",
				Type:  types.NewNullableString("ROOT"),
				FieldRefs: []FieldDef{
					{Label: "Title", Type: types.FieldTypeText},
					{Label: "Slug", Type: types.FieldTypeSlug},
					{Label: "Description", Type: types.FieldTypeTextarea},
					{Label: "Content", Type: types.FieldTypeRichText},
					{Label: "Cover", Type: types.FieldTypeMedia},
					{Label: "Published At", Type: types.FieldTypeDatetime},
				},
			},
			"page": {
				Label: "Page",
				Type:  types.NewNullableString("ROOT"),
				FieldRefs: []FieldDef{
					{Label: "Page Title", Type: types.FieldTypeText},
					{Label: "Page Slug", Type: types.FieldTypeSlug},
					{Label: "Page Content", Type: types.FieldTypeRichText},
				},
			},
			"global": {
				Label: "Global",
				Type:  types.NewNullableString("ROOT"),
				FieldRefs: []FieldDef{
					{Label: "Site Name", Type: types.FieldTypeText},
					{Label: "Site Description", Type: types.FieldTypeTextarea},
					{Label: "Site Logo", Type: types.FieldTypeMedia},
				},
			},
		},
	})
}
