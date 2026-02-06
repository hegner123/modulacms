package definitions

import "github.com/hegner123/modulacms/internal/db/types"

func init() {
	Register(SchemaDefinition{
		Name:        "modulacms-default",
		Label:       "ModulaCMS Default",
		Description: "Standard page and section datatypes with common content fields",
		Format:      "modulacms",
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
			"meta_title": {
				Label: "Meta Title",
				Type:  types.FieldTypeText,
			},
			"meta_description": {
				Label: "Meta Description",
				Type:  types.FieldTypeTextarea,
			},
			"section_title": {
				Label: "Section Title",
				Type:  types.FieldTypeText,
			},
			"section_content": {
				Label: "Section Content",
				Type:  types.FieldTypeRichText,
			},
		},
		Datatypes: map[string]DatatypeDef{
			"page": {
				Label:     "Page",
				Type:      "page",
				FieldRefs: []string{"title", "slug", "body", "excerpt", "featured_image", "published", "meta_title", "meta_description"},
				ChildRefs: []string{"section"},
			},
			"section": {
				Label:     "Section",
				Type:      "section",
				FieldRefs: []string{"section_title", "section_content"},
			},
		},
		RootKeys: []string{"page"},
	})
}
