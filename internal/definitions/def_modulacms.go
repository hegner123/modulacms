package definitions

import "github.com/hegner123/modulacms/internal/db/types"

func init() {
	Register(SchemaDefinition{
		Name:        "modulacms-default",
		Label:       "ModulaCMS Default",
		Description: "Component-based page builder with grid layouts, content blocks, posts, case studies, and documentation",
		Format:      "modulacms",
		Datatypes: map[string]DatatypeDef{

			// ──────────────────────────────────────
			// Root: Page
			// ──────────────────────────────────────

			"page": {
				Label: "Page",
				Type:  types.NewNullableString("ROOT"),
				FieldRefs: []FieldDef{
					{Label: "Title", Type: types.FieldTypeText},
					{Label: "Slug", Type: types.FieldTypeSlug},
					{Label: "Meta Title", Type: types.FieldTypeText},
					{Label: "Meta Description", Type: types.FieldTypeTextarea},
					{Label: "Published", Type: types.FieldTypeBoolean},
				},
			},

			// Layout: Row/Column

			"row": {
				Label:     "Row",
				Type:      types.NewNullableString("layout"),
				ParentRef: "page",
				FieldRefs: []FieldDef{
					{Label: "Full Width", Type: types.FieldTypeBoolean},
				},
			},

			"column": {
				Label:     "Column",
				Type:      types.NewNullableString("layout"),
				ParentRef: "row",
				FieldRefs: []FieldDef{
					{Label: "Span", Type: types.FieldTypeNumber},
				},
			},

			// Layout: Grid/Area

			"grid": {
				Label:     "Grid",
				Type:      types.NewNullableString("layout"),
				ParentRef: "page",
				FieldRefs: []FieldDef{
					{Label: "Columns", Type: types.FieldTypeText},
					{Label: "Rows", Type: types.FieldTypeText},
					{Label: "Gap", Type: types.FieldTypeText},
				},
			},

			"area": {
				Label:     "Area",
				Type:      types.NewNullableString("layout"),
				ParentRef: "grid",
				FieldRefs: []FieldDef{
					{Label: "Column Start", Type: types.FieldTypeNumber},
					{Label: "Column End", Type: types.FieldTypeNumber},
					{Label: "Row Start", Type: types.FieldTypeNumber},
					{Label: "Row End", Type: types.FieldTypeNumber},
				},
			},

			// Settings

			"settings": {
				Label:     "Settings",
				Type:      types.NewNullableString("settings"),
				ParentRef: "page",
				FieldRefs: []FieldDef{
					{Label: "Margin", Type: types.FieldTypeText},
					{Label: "Padding", Type: types.FieldTypeText},
				},
			},

			// Content Blocks

			"cta": {
				Label:     "CTA",
				Type:      types.NewNullableString("content"),
				ParentRef: "page",
				FieldRefs: []FieldDef{
					{Label: "Heading", Type: types.FieldTypeText},
					{Label: "Subheading", Type: types.FieldTypeTextarea},
					{Label: "Button Text", Type: types.FieldTypeText},
					{Label: "Button URL", Type: types.FieldTypeURL},
				},
			},

			"image_block": {
				Label:     "Image",
				Type:      types.NewNullableString("content"),
				ParentRef: "page",
				FieldRefs: []FieldDef{
					{Label: "Image", Type: types.FieldTypeMedia},
					{Label: "Alt Text", Type: types.FieldTypeText},
					{Label: "Caption", Type: types.FieldTypeTextarea},
				},
			},

			"rich_text_block": {
				Label:     "Rich Text",
				Type:      types.NewNullableString("content"),
				ParentRef: "page",
				FieldRefs: []FieldDef{
					{Label: "Content", Type: types.FieldTypeRichText},
				},
			},

			"text_block": {
				Label:     "Text",
				Type:      types.NewNullableString("content"),
				ParentRef: "page",
				FieldRefs: []FieldDef{
					{Label: "Content", Type: types.FieldTypeTextarea},
				},
			},

			"button_block": {
				Label:     "Button",
				Type:      types.NewNullableString("content"),
				ParentRef: "page",
				FieldRefs: []FieldDef{
					{Label: "Label", Type: types.FieldTypeText},
					{Label: "URL", Type: types.FieldTypeURL},
					{Label: "Variant", Type: types.FieldTypeSelect, Data: types.NewNullableString(`{"options":["primary","secondary","outline","ghost"]}`)},
				},
			},

			"card": {
				Label:     "Card",
				Type:      types.NewNullableString("content"),
				ParentRef: "page",
				FieldRefs: []FieldDef{
					{Label: "Title", Type: types.FieldTypeText},
					{Label: "Description", Type: types.FieldTypeTextarea},
					{Label: "Image", Type: types.FieldTypeMedia},
					{Label: "Link URL", Type: types.FieldTypeURL},
				},
			},

			// Animation

			"animation": {
				Label:     "Animation",
				Type:      types.NewNullableString("content"),
				ParentRef: "page",
				FieldRefs: []FieldDef{
					{Label: "Type", Type: types.FieldTypeSelect, Data: types.NewNullableString(`{"options":["fade","slide","scale","rotate"]}`)},
					{Label: "Duration", Type: types.FieldTypeText},
					{Label: "Delay", Type: types.FieldTypeText},
					{Label: "Easing", Type: types.FieldTypeSelect, Data: types.NewNullableString(`{"options":["ease","ease-in","ease-out","ease-in-out","linear"]}`)},
					{Label: "Direction", Type: types.FieldTypeSelect, Data: types.NewNullableString(`{"options":["normal","reverse","alternate"]}`)},
					{Label: "Iterations", Type: types.FieldTypeText},
				},
			},

			// ──────────────────────────────────────
			// Root: Post
			// ──────────────────────────────────────

			"post": {
				Label: "Post",
				Type:  types.NewNullableString("ROOT"),
				FieldRefs: []FieldDef{
					{Label: "Title", Type: types.FieldTypeText},
					{Label: "Slug", Type: types.FieldTypeSlug},
					{Label: "Meta Title", Type: types.FieldTypeText},
					{Label: "Meta Description", Type: types.FieldTypeTextarea},
					{Label: "Published", Type: types.FieldTypeBoolean},
				},
			},

			"post_content": {
				Label:     "Content",
				Type:      types.NewNullableString("content"),
				ParentRef: "post",
				FieldRefs: []FieldDef{
					{Label: "Content", Type: types.FieldTypeRichText},
				},
			},

			// ──────────────────────────────────────
			// Root: Case Study
			// ──────────────────────────────────────

			"case_study": {
				Label: "Case Study",
				Type:  types.NewNullableString("ROOT"),
				FieldRefs: []FieldDef{
					{Label: "Title", Type: types.FieldTypeText},
					{Label: "Slug", Type: types.FieldTypeSlug},
					{Label: "Client Name", Type: types.FieldTypeText},
					{Label: "Description", Type: types.FieldTypeTextarea},
					{Label: "Challenge", Type: types.FieldTypeRichText},
					{Label: "Solution", Type: types.FieldTypeRichText},
					{Label: "Results", Type: types.FieldTypeRichText},
					{Label: "Featured Image", Type: types.FieldTypeMedia},
					{Label: "Published", Type: types.FieldTypeBoolean},
				},
			},

			// ──────────────────────────────────────
			// Root: Documentation
			// ──────────────────────────────────────

			"documentation": {
				Label: "Documentation",
				Type:  types.NewNullableString("ROOT"),
				FieldRefs: []FieldDef{
					{Label: "Title", Type: types.FieldTypeText},
					{Label: "Slug", Type: types.FieldTypeSlug},
					{Label: "Published", Type: types.FieldTypeBoolean},
				},
			},

			"doc_section": {
				Label:     "Section",
				Type:      types.NewNullableString("doc_component"),
				ParentRef: "documentation",
				FieldRefs: []FieldDef{
					{Label: "Heading", Type: types.FieldTypeText},
					{Label: "Content", Type: types.FieldTypeRichText},
				},
			},

			"code_block": {
				Label:     "Code Block",
				Type:      types.NewNullableString("doc_component"),
				ParentRef: "documentation",
				FieldRefs: []FieldDef{
					{Label: "Language", Type: types.FieldTypeSelect, Data: types.NewNullableString(`{"options":["go","javascript","typescript","html","css","bash","sql","json","yaml"]}`)},
					{Label: "Code", Type: types.FieldTypeTextarea},
					{Label: "Caption", Type: types.FieldTypeText},
				},
			},

			"doc_image": {
				Label:     "Image",
				Type:      types.NewNullableString("doc_component"),
				ParentRef: "documentation",
				FieldRefs: []FieldDef{
					{Label: "Image", Type: types.FieldTypeMedia},
					{Label: "Alt Text", Type: types.FieldTypeText},
					{Label: "Caption", Type: types.FieldTypeText},
				},
			},

			"doc_reference": {
				Label:     "Reference",
				Type:      types.NewNullableString("doc_component"),
				ParentRef: "documentation",
				FieldRefs: []FieldDef{
					{Label: "Label", Type: types.FieldTypeText},
					{Label: "URL", Type: types.FieldTypeURL},
					{Label: "Description", Type: types.FieldTypeTextarea},
				},
			},

			"step_header": {
				Label:     "Step Header",
				Type:      types.NewNullableString("doc_component"),
				ParentRef: "documentation",
				FieldRefs: []FieldDef{
					{Label: "Step Number", Type: types.FieldTypeNumber},
					{Label: "Title", Type: types.FieldTypeText},
					{Label: "Description", Type: types.FieldTypeTextarea},
				},
			},

			// ──────────────────────────────────────
			// Root: Menu
			// ──────────────────────────────────────

			"menu": {
				Label: "Menu",
				Type:  types.NewNullableString("ROOT"),
				FieldRefs: []FieldDef{
					{Label: "Title", Type: types.FieldTypeText},
					{Label: "Slug", Type: types.FieldTypeSlug},
					{Label: "Position", Type: types.FieldTypeSelect, Data: types.NewNullableString(`{"options":["header","footer","sidebar"]}`)},
				},
			},

			// Menu Link: direct navigation item under menu

			"menu_link": {
				Label:     "Menu Link",
				Type:      types.NewNullableString("menu_component"),
				ParentRef: "menu",
				FieldRefs: []FieldDef{
					{Label: "Label", Type: types.FieldTypeText},
					{Label: "URL", Type: types.FieldTypeURL},
					{Label: "Target", Type: types.FieldTypeSelect, Data: types.NewNullableString(`{"options":["_self","_blank"]}`)},
					{Label: "Icon", Type: types.FieldTypeText},
				},
			},

			// Menu List: dropdown group under menu

			"menu_list": {
				Label:     "Menu List",
				Type:      types.NewNullableString("menu_component"),
				ParentRef: "menu",
				FieldRefs: []FieldDef{
					{Label: "Label", Type: types.FieldTypeText},
				},
			},

			"menu_list_link": {
				Label:     "Menu List Link",
				Type:      types.NewNullableString("menu_component"),
				ParentRef: "menu_list",
				FieldRefs: []FieldDef{
					{Label: "Label", Type: types.FieldTypeText},
					{Label: "URL", Type: types.FieldTypeURL},
					{Label: "Target", Type: types.FieldTypeSelect, Data: types.NewNullableString(`{"options":["_self","_blank"]}`)},
				},
			},

			// Menu Nested List: sub-group under a menu list

			"menu_nested_list": {
				Label:     "Menu Nested List",
				Type:      types.NewNullableString("menu_component"),
				ParentRef: "menu_list",
				FieldRefs: []FieldDef{
					{Label: "Label", Type: types.FieldTypeText},
				},
			},

			"menu_nested_link": {
				Label:     "Menu Nested Link",
				Type:      types.NewNullableString("menu_component"),
				ParentRef: "menu_nested_list",
				FieldRefs: []FieldDef{
					{Label: "Label", Type: types.FieldTypeText},
					{Label: "URL", Type: types.FieldTypeURL},
					{Label: "Target", Type: types.FieldTypeSelect, Data: types.NewNullableString(`{"options":["_self","_blank"]}`)},
				},
			},
		},
	})
}
