package definitions

import "github.com/hegner123/modulacms/internal/db/types"

func init() {
	Register(SchemaDefinition{
		Name:        "modula-default",
		Label:       "Modula Default",
		Description: "The default Modula schema. Component-based page builder with grid layouts, content blocks, posts, case studies, and documentation. Fully compatible with the built-in admin panel and designed to be extended with your own datatypes and fields.",
		Format:      "modula",
		Datatypes: map[string]DatatypeDef{

			// ──────────────────────────────────────
			// Root: Page
			// ──────────────────────────────────────

			"page": {
				Name:  "page",
				Label: "Page",
				Type:  types.NewNullableString(string(types.DatatypeTypeRoot)),
				FieldRefs: []FieldDef{
					{Name: "title", Label: "Title", Type: types.FieldTypeTitle},
					{Name: "slug", Label: "Slug", Type: types.FieldTypeSlug},
					{Name: "description", Label: "Description", Type: types.FieldTypeTextarea},
					{Name: "featured_image", Label: "Featured Image", Type: types.FieldTypeMedia},
					{Name: "meta_title", Label: "Meta Title", Type: types.FieldTypeText},
					{Name: "meta_description", Label: "Meta Description", Type: types.FieldTypeTextarea},
					{Name: "og_image", Label: "OG Image", Type: types.FieldTypeMedia},
					{Name: "published", Label: "Published", Type: types.FieldTypeBoolean},
				},
			},

			// Layout: Row/Column

			"row": {
				Name:      "row",
				Label:     "Row",
				Type:      types.NewNullableString("layout"),
				ParentRef: "page",
				FieldRefs: []FieldDef{
					{Name: "full_width", Label: "Full Width", Type: types.FieldTypeBoolean},
				},
			},

			"columns": {
				Name:      "columns",
				Label:     "Columns",
				Type:      types.NewNullableString("layout"),
				ParentRef: "row",
				FieldRefs: []FieldDef{
					{Name: "count", Label: "Count", Type: types.FieldTypeNumber},
				},
			},

			// Layout: Grid/Area

			"grid": {
				Name:      "grid",
				Label:     "Grid",
				Type:      types.NewNullableString("layout"),
				ParentRef: "page",
				FieldRefs: []FieldDef{
					{Name: "columns", Label: "Columns", Type: types.FieldTypeText},
					{Name: "rows", Label: "Rows", Type: types.FieldTypeText},
					{Name: "gap", Label: "Gap", Type: types.FieldTypeText},
				},
			},

			"area": {
				Name:      "area",
				Label:     "Area",
				Type:      types.NewNullableString("layout"),
				ParentRef: "grid",
				FieldRefs: []FieldDef{
					{Name: "column_start", Label: "Column Start", Type: types.FieldTypeNumber},
					{Name: "column_end", Label: "Column End", Type: types.FieldTypeNumber},
					{Name: "row_start", Label: "Row Start", Type: types.FieldTypeNumber},
					{Name: "row_end", Label: "Row End", Type: types.FieldTypeNumber},
				},
			},

			// Settings

			"settings": {
				Name:      "settings",
				Label:     "Settings",
				Type:      types.NewNullableString("settings"),
				ParentRef: "page",
				FieldRefs: []FieldDef{
					{Name: "margin", Label: "Margin", Type: types.FieldTypeText},
					{Name: "padding", Label: "Padding", Type: types.FieldTypeText},
				},
			},

			// Content Blocks

			"cta": {
				Name:      "cta",
				Label:     "CTA",
				Type:      types.NewNullableString("content"),
				ParentRef: "page",
				FieldRefs: []FieldDef{
					{Name: "heading", Label: "Heading", Type: types.FieldTypeText},
					{Name: "subheading", Label: "Subheading", Type: types.FieldTypeTextarea},
					{Name: "button_text", Label: "Button Text", Type: types.FieldTypeText},
					{Name: "button_url", Label: "Button URL", Type: types.FieldTypeURL},
				},
			},

			"image_block": {
				Name:      "image_block",
				Label:     "Image",
				Type:      types.NewNullableString("content"),
				ParentRef: "page",
				FieldRefs: []FieldDef{
					{Name: "image", Label: "Image", Type: types.FieldTypeMedia},
					{Name: "alt_text", Label: "Alt Text", Type: types.FieldTypeText},
					{Name: "caption", Label: "Caption", Type: types.FieldTypeTextarea},
				},
			},

			"rich_text_block": {
				Name:      "rich_text_block",
				Label:     "Rich Text",
				Type:      types.NewNullableString("content"),
				ParentRef: "page",
				FieldRefs: []FieldDef{
					{Name: "content", Label: "Content", Type: types.FieldTypeRichText},
				},
			},

			"text_block": {
				Name:      "text_block",
				Label:     "Text",
				Type:      types.NewNullableString("content"),
				ParentRef: "page",
				FieldRefs: []FieldDef{
					{Name: "content", Label: "Content", Type: types.FieldTypeTextarea},
				},
			},

			"button_block": {
				Name:      "button_block",
				Label:     "Button",
				Type:      types.NewNullableString("content"),
				ParentRef: "page",
				FieldRefs: []FieldDef{
					{Name: "label", Label: "Label", Type: types.FieldTypeText},
					{Name: "url", Label: "URL", Type: types.FieldTypeURL},
					{Name: "variant", Label: "Variant", Type: types.FieldTypeSelect, Data: types.NewNullableString(`{"options":["primary","secondary","outline","ghost"]}`)},
				},
			},

			"card": {
				Name:      "card",
				Label:     "Card",
				Type:      types.NewNullableString("content"),
				ParentRef: "page",
				FieldRefs: []FieldDef{
					{Name: "title", Label: "Title", Type: types.FieldTypeText},
					{Name: "description", Label: "Description", Type: types.FieldTypeTextarea},
					{Name: "image", Label: "Image", Type: types.FieldTypeMedia},
					{Name: "link_url", Label: "Link URL", Type: types.FieldTypeURL},
				},
			},

			// Animation

			"animation": {
				Name:      "animation",
				Label:     "Animation",
				Type:      types.NewNullableString("content"),
				ParentRef: "page",
				FieldRefs: []FieldDef{
					{Name: "type", Label: "Type", Type: types.FieldTypeSelect, Data: types.NewNullableString(`{"options":["fade","slide","scale","rotate"]}`)},
					{Name: "duration", Label: "Duration", Type: types.FieldTypeText},
					{Name: "delay", Label: "Delay", Type: types.FieldTypeText},
					{Name: "easing", Label: "Easing", Type: types.FieldTypeSelect, Data: types.NewNullableString(`{"options":["ease","ease-in","ease-out","ease-in-out","linear"]}`)},
					{Name: "direction", Label: "Direction", Type: types.FieldTypeSelect, Data: types.NewNullableString(`{"options":["normal","reverse","alternate"]}`)},
					{Name: "iterations", Label: "Iterations", Type: types.FieldTypeText},
				},
			},

			// ──────────────────────────────────────
			// Root: Post
			// ──────────────────────────────────────

			"post": {
				Name:  "post",
				Label: "Post",
				Type:  types.NewNullableString(string(types.DatatypeTypeRoot)),
				FieldRefs: []FieldDef{
					{Name: "title", Label: "Title", Type: types.FieldTypeTitle},
					{Name: "slug", Label: "Slug", Type: types.FieldTypeSlug},
					{Name: "excerpt", Label: "Excerpt", Type: types.FieldTypeTextarea},
					{Name: "featured_image", Label: "Featured Image", Type: types.FieldTypeMedia},
					{Name: "category", Label: "Category", Type: types.FieldTypeSelect, Data: types.NewNullableString(`{"options":["uncategorized","news","tutorial","opinion","review","announcement"]}`)},
					{Name: "tags", Label: "Tags", Type: types.FieldTypeText},
					{Name: "publish_date", Label: "Publish Date", Type: types.FieldTypeDatetime},
					{Name: "meta_title", Label: "Meta Title", Type: types.FieldTypeText},
					{Name: "meta_description", Label: "Meta Description", Type: types.FieldTypeTextarea},
					{Name: "published", Label: "Published", Type: types.FieldTypeBoolean},
				},
			},

			"post_content": {
				Name:      "post_content",
				Label:     "Content",
				Type:      types.NewNullableString("content"),
				ParentRef: "post",
				FieldRefs: []FieldDef{
					{Name: "content", Label: "Content", Type: types.FieldTypeRichText},
				},
			},

			// ──────────────────────────────────────
			// Root: Case Study
			// ──────────────────────────────────────

			"case_study": {
				Name:  "case_study",
				Label: "Case Study",
				Type:  types.NewNullableString(string(types.DatatypeTypeRoot)),
				FieldRefs: []FieldDef{
					{Name: "title", Label: "Title", Type: types.FieldTypeTitle},
					{Name: "slug", Label: "Slug", Type: types.FieldTypeSlug},
					{Name: "company_name", Label: "Company Name", Type: types.FieldTypeText},
					{Name: "company_logo", Label: "Company Logo", Type: types.FieldTypeMedia},
					{Name: "company_url", Label: "Company URL", Type: types.FieldTypeURL},
					{Name: "industry", Label: "Industry", Type: types.FieldTypeText},
					{Name: "description", Label: "Description", Type: types.FieldTypeTextarea},
					{Name: "challenge", Label: "Challenge", Type: types.FieldTypeRichText},
					{Name: "solution", Label: "Solution", Type: types.FieldTypeRichText},
					{Name: "results", Label: "Results", Type: types.FieldTypeRichText},
					{Name: "testimonial", Label: "Testimonial", Type: types.FieldTypeTextarea},
					{Name: "featured_image", Label: "Featured Image", Type: types.FieldTypeMedia},
					{Name: "published", Label: "Published", Type: types.FieldTypeBoolean},
				},
			},

			// ──────────────────────────────────────
			// Root: Documentation
			// ──────────────────────────────────────

			"documentation": {
				Name:  "documentation",
				Label: "Documentation",
				Type:  types.NewNullableString(string(types.DatatypeTypeRoot)),
				FieldRefs: []FieldDef{
					{Name: "title", Label: "Title", Type: types.FieldTypeTitle},
					{Name: "slug", Label: "Slug", Type: types.FieldTypeSlug},
					{Name: "chapter", Label: "Chapter", Type: types.FieldTypeText},
					{Name: "sort_order", Label: "Sort Order", Type: types.FieldTypeNumber},
					{Name: "description", Label: "Description", Type: types.FieldTypeTextarea},
					{Name: "published", Label: "Published", Type: types.FieldTypeBoolean},
				},
			},

			"doc_section": {
				Name:      "doc_section",
				Label:     "Section",
				Type:      types.NewNullableString("doc_component"),
				ParentRef: "documentation",
				FieldRefs: []FieldDef{
					{Name: "heading", Label: "Heading", Type: types.FieldTypeText},
					{Name: "content", Label: "Content", Type: types.FieldTypeRichText},
				},
			},

			"code_block": {
				Name:      "code_block",
				Label:     "Code Block",
				Type:      types.NewNullableString("doc_component"),
				ParentRef: "documentation",
				FieldRefs: []FieldDef{
					{Name: "language", Label: "Language", Type: types.FieldTypeSelect, Data: types.NewNullableString(`{"options":["go","javascript","typescript","html","css","bash","sql","json","yaml"]}`)},
					{Name: "code", Label: "Code", Type: types.FieldTypeTextarea},
					{Name: "caption", Label: "Caption", Type: types.FieldTypeText},
				},
			},

			"doc_image": {
				Name:      "doc_image",
				Label:     "Image",
				Type:      types.NewNullableString("doc_component"),
				ParentRef: "documentation",
				FieldRefs: []FieldDef{
					{Name: "image", Label: "Image", Type: types.FieldTypeMedia},
					{Name: "alt_text", Label: "Alt Text", Type: types.FieldTypeText},
					{Name: "caption", Label: "Caption", Type: types.FieldTypeText},
				},
			},

			"doc_reference": {
				Name:      "doc_reference",
				Label:     "Reference",
				Type:      types.NewNullableString("doc_component"),
				ParentRef: "documentation",
				FieldRefs: []FieldDef{
					{Name: "label", Label: "Label", Type: types.FieldTypeText},
					{Name: "url", Label: "URL", Type: types.FieldTypeURL},
					{Name: "description", Label: "Description", Type: types.FieldTypeTextarea},
				},
			},

			"step_header": {
				Name:      "step_header",
				Label:     "Step Header",
				Type:      types.NewNullableString("doc_component"),
				ParentRef: "documentation",
				FieldRefs: []FieldDef{
					{Name: "step_number", Label: "Step Number", Type: types.FieldTypeNumber},
					{Name: "title", Label: "Title", Type: types.FieldTypeText},
					{Name: "description", Label: "Description", Type: types.FieldTypeTextarea},
				},
			},

			// ──────────────────────────────────────
			// Root: Menu
			// ──────────────────────────────────────

			"menu": {
				Name:  "menu",
				Label: "Menu",
				Type:  types.NewNullableString(string(types.DatatypeTypeRoot)),
				FieldRefs: []FieldDef{
					{Name: "title", Label: "Title", Type: types.FieldTypeTitle},
					{Name: "slug", Label: "Slug", Type: types.FieldTypeSlug},
					{Name: "position", Label: "Position", Type: types.FieldTypeSelect, Data: types.NewNullableString(`{"options":["header","sidebar"]}`)},
				},
			},

			"menu_link": {
				Name:      "menu_link",
				Label:     "Menu Link",
				Type:      types.NewNullableString("menu_component"),
				ParentRef: "menu",
				FieldRefs: []FieldDef{
					{Name: "label", Label: "Label", Type: types.FieldTypeText},
					{Name: "url", Label: "URL", Type: types.FieldTypeURL},
					{Name: "target", Label: "Target", Type: types.FieldTypeSelect, Data: types.NewNullableString(`{"options":["_self","_blank"]}`)},
				},
			},

			"menu_icon_link": {
				Name:      "menu_icon_link",
				Label:     "Menu Icon Link",
				Type:      types.NewNullableString("menu_component"),
				ParentRef: "menu",
				FieldRefs: []FieldDef{
					{Name: "label", Label: "Label", Type: types.FieldTypeText},
					{Name: "url", Label: "URL", Type: types.FieldTypeURL},
					{Name: "target", Label: "Target", Type: types.FieldTypeSelect, Data: types.NewNullableString(`{"options":["_self","_blank"]}`)},
					{Name: "icon", Label: "Icon", Type: types.FieldTypeText},
				},
			},

			"menu_list": {
				Name:      "menu_list",
				Label:     "Menu List",
				Type:      types.NewNullableString("menu_component"),
				ParentRef: "menu",
				FieldRefs: []FieldDef{
					{Name: "label", Label: "Label", Type: types.FieldTypeText},
				},
			},

			"menu_list_link": {
				Name:      "menu_list_link",
				Label:     "Menu List Link",
				Type:      types.NewNullableString("menu_component"),
				ParentRef: "menu_list",
				FieldRefs: []FieldDef{
					{Name: "label", Label: "Label", Type: types.FieldTypeText},
					{Name: "url", Label: "URL", Type: types.FieldTypeURL},
					{Name: "target", Label: "Target", Type: types.FieldTypeSelect, Data: types.NewNullableString(`{"options":["_self","_blank"]}`)},
				},
			},

			"menu_list_icon_link": {
				Name:      "menu_list_icon_link",
				Label:     "Menu List Icon Link",
				Type:      types.NewNullableString("menu_component"),
				ParentRef: "menu_list",
				FieldRefs: []FieldDef{
					{Name: "label", Label: "Label", Type: types.FieldTypeText},
					{Name: "url", Label: "URL", Type: types.FieldTypeURL},
					{Name: "target", Label: "Target", Type: types.FieldTypeSelect, Data: types.NewNullableString(`{"options":["_self","_blank"]}`)},
					{Name: "icon", Label: "Icon", Type: types.FieldTypeText},
				},
			},

			"menu_nested_list": {
				Name:      "menu_nested_list",
				Label:     "Menu Nested List",
				Type:      types.NewNullableString("menu_component"),
				ParentRef: "menu_list",
				FieldRefs: []FieldDef{
					{Name: "label", Label: "Label", Type: types.FieldTypeText},
				},
			},

			"menu_nested_link": {
				Name:      "menu_nested_link",
				Label:     "Menu Nested Link",
				Type:      types.NewNullableString("menu_component"),
				ParentRef: "menu_nested_list",
				FieldRefs: []FieldDef{
					{Name: "label", Label: "Label", Type: types.FieldTypeText},
					{Name: "url", Label: "URL", Type: types.FieldTypeURL},
					{Name: "target", Label: "Target", Type: types.FieldTypeSelect, Data: types.NewNullableString(`{"options":["_self","_blank"]}`)},
				},
			},

			"menu_nested_icon_link": {
				Name:      "menu_nested_icon_link",
				Label:     "Menu Nested Icon Link",
				Type:      types.NewNullableString("menu_component"),
				ParentRef: "menu_nested_list",
				FieldRefs: []FieldDef{
					{Name: "label", Label: "Label", Type: types.FieldTypeText},
					{Name: "url", Label: "URL", Type: types.FieldTypeURL},
					{Name: "target", Label: "Target", Type: types.FieldTypeSelect, Data: types.NewNullableString(`{"options":["_self","_blank"]}`)},
					{Name: "icon", Label: "Icon", Type: types.FieldTypeText},
				},
			},

			// ──────────────────────────────────────
			// Root: Footer
			// ──────────────────────────────────────

			"footer": {
				Name:  "footer",
				Label: "Footer",
				Type:  types.NewNullableString(string(types.DatatypeTypeRoot)),
				FieldRefs: []FieldDef{
					{Name: "title", Label: "Title", Type: types.FieldTypeTitle},
					{Name: "slug", Label: "Slug", Type: types.FieldTypeSlug},
					{Name: "copyright", Label: "Copyright", Type: types.FieldTypeText},
				},
			},

			"footer_column": {
				Name:      "footer_column",
				Label:     "Footer Column",
				Type:      types.NewNullableString("footer_component"),
				ParentRef: "footer",
				FieldRefs: []FieldDef{
					{Name: "heading", Label: "Heading", Type: types.FieldTypeText},
				},
			},

			"footer_link": {
				Name:      "footer_link",
				Label:     "Footer Link",
				Type:      types.NewNullableString("footer_component"),
				ParentRef: "footer_column",
				FieldRefs: []FieldDef{
					{Name: "label", Label: "Label", Type: types.FieldTypeText},
					{Name: "url", Label: "URL", Type: types.FieldTypeURL},
					{Name: "target", Label: "Target", Type: types.FieldTypeSelect, Data: types.NewNullableString(`{"options":["_self","_blank"]}`)},
				},
			},

			"footer_social": {
				Name:      "footer_social",
				Label:     "Footer Social",
				Type:      types.NewNullableString("footer_component"),
				ParentRef: "footer",
				FieldRefs: []FieldDef{
					{Name: "platform", Label: "Platform", Type: types.FieldTypeSelect, Data: types.NewNullableString(`{"options":["github","twitter","linkedin","youtube","discord","instagram","facebook","mastodon"]}`)},
					{Name: "url", Label: "URL", Type: types.FieldTypeURL},
				},
			},

			"footer_text": {
				Name:      "footer_text",
				Label:     "Footer Text",
				Type:      types.NewNullableString("footer_component"),
				ParentRef: "footer",
				FieldRefs: []FieldDef{
					{Name: "content", Label: "Content", Type: types.FieldTypeRichText},
				},
			},
		},
	})
}
