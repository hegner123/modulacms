package definitions

import "github.com/hegner123/modulacms/internal/db/types"

func init() {
	Register(SchemaDefinition{
		Name:        "tailwind-marketing",
		Label:       "Tailwind Marketing",
		Description: "Marketing site schema derived from Tailwind UI marketing templates. Section-based page builder with hero, features, pricing, testimonials, team, blog, FAQ, stats, CTA, contact, newsletter, and navigation sections. Each section type maps to a Tailwind UI template category.",
		Format:      "modula",
		Datatypes: map[string]DatatypeDef{

			// ──────────────────────────────────────
			// Root: Marketing Page
			// ──────────────────────────────────────

			"marketing_page": {
				Name:  "marketing_page",
				Label: "Marketing Page",
				Type:  types.NewNullableString(string(types.DatatypeTypeRoot)),
				FieldRefs: []FieldDef{
					{Name: "title", Label: "Title", Type: types.FieldTypeTitle},
					{Name: "description", Label: "Description", Type: types.FieldTypeTextarea},
					{Name: "meta_title", Label: "Meta Title", Type: types.FieldTypeText},
					{Name: "meta_description", Label: "Meta Description", Type: types.FieldTypeTextarea},
					{Name: "og_image", Label: "OG Image", Type: types.FieldTypeMedia},
				},
			},

			// ──────────────────────────────────────
			// Navigation
			// ──────────────────────────────────────

			"header": {
				Name:      "header",
				Label:     "Header",
				Type:      types.NewNullableString("navigation"),
				ParentRef: "marketing_page",
				FieldRefs: []FieldDef{
					{Name: "logo", Label: "Logo", Type: types.FieldTypeMedia},
					{Name: "logo_dark", Label: "Logo (Dark)", Type: types.FieldTypeMedia},
					{Name: "login_text", Label: "Login Text", Type: types.FieldTypeText},
					{Name: "login_url", Label: "Login URL", Type: types.FieldTypeURL},
					{Name: "cta_text", Label: "CTA Text", Type: types.FieldTypeText},
					{Name: "cta_url", Label: "CTA URL", Type: types.FieldTypeURL},
				},
			},

			"nav_link": {
				Name:      "nav_link",
				Label:     "Nav Link",
				Type:      types.NewNullableString("navigation"),
				ParentRef: "header",
				FieldRefs: []FieldDef{
					{Name: "label", Label: "Label", Type: types.FieldTypeText},
					{Name: "url", Label: "URL", Type: types.FieldTypeURL},
				},
			},

			"flyout_menu": {
				Name:      "flyout_menu",
				Label:     "Flyout Menu",
				Type:      types.NewNullableString("navigation"),
				ParentRef: "header",
				FieldRefs: []FieldDef{
					{Name: "label", Label: "Trigger Label", Type: types.FieldTypeText},
				},
			},

			"flyout_link": {
				Name:      "flyout_link",
				Label:     "Flyout Link",
				Type:      types.NewNullableString("navigation"),
				ParentRef: "flyout_menu",
				FieldRefs: []FieldDef{
					{Name: "label", Label: "Label", Type: types.FieldTypeText},
					{Name: "description", Label: "Description", Type: types.FieldTypeTextarea},
					{Name: "url", Label: "URL", Type: types.FieldTypeURL},
					{Name: "icon", Label: "Icon SVG", Type: types.FieldTypeTextarea},
				},
			},

			"footer": {
				Name:      "footer",
				Label:     "Footer",
				Type:      types.NewNullableString("navigation"),
				ParentRef: "marketing_page",
				FieldRefs: []FieldDef{
					{Name: "copyright", Label: "Copyright Text", Type: types.FieldTypeText},
					{Name: "newsletter_heading", Label: "Newsletter Heading", Type: types.FieldTypeText},
					{Name: "newsletter_description", Label: "Newsletter Description", Type: types.FieldTypeTextarea},
				},
			},

			"footer_column": {
				Name:      "footer_column",
				Label:     "Footer Column",
				Type:      types.NewNullableString("navigation"),
				ParentRef: "footer",
				FieldRefs: []FieldDef{
					{Name: "heading", Label: "Column Heading", Type: types.FieldTypeText},
				},
			},

			"footer_link": {
				Name:      "footer_link",
				Label:     "Footer Link",
				Type:      types.NewNullableString("navigation"),
				ParentRef: "footer_column",
				FieldRefs: []FieldDef{
					{Name: "label", Label: "Label", Type: types.FieldTypeText},
					{Name: "url", Label: "URL", Type: types.FieldTypeURL},
				},
			},

			"social_link": {
				Name:      "social_link",
				Label:     "Social Link",
				Type:      types.NewNullableString("navigation"),
				ParentRef: "footer",
				FieldRefs: []FieldDef{
					{Name: "platform", Label: "Platform", Type: types.FieldTypeText},
					{Name: "url", Label: "URL", Type: types.FieldTypeURL},
					{Name: "icon", Label: "Icon SVG", Type: types.FieldTypeTextarea},
				},
			},

			// ──────────────────────────────────────
			// Hero Sections
			// ──────────────────────────────────────

			"hero_section": {
				Name:      "hero_section",
				Label:     "Hero Section",
				Type:      types.NewNullableString("section"),
				ParentRef: "marketing_page",
				FieldRefs: []FieldDef{
					{Name: "heading", Label: "Heading", Type: types.FieldTypeText},
					{Name: "description", Label: "Description", Type: types.FieldTypeTextarea},
					{Name: "image", Label: "Hero Image", Type: types.FieldTypeMedia},
					{Name: "image_dark", Label: "Hero Image (Dark)", Type: types.FieldTypeMedia},
					{Name: "primary_cta_text", Label: "Primary CTA Text", Type: types.FieldTypeText},
					{Name: "primary_cta_url", Label: "Primary CTA URL", Type: types.FieldTypeURL},
					{Name: "secondary_cta_text", Label: "Secondary CTA Text", Type: types.FieldTypeText},
					{Name: "secondary_cta_url", Label: "Secondary CTA URL", Type: types.FieldTypeURL},
					{Name: "announcement_text", Label: "Announcement Text", Type: types.FieldTypeText},
					{Name: "announcement_url", Label: "Announcement URL", Type: types.FieldTypeURL},
				},
			},

			// ──────────────────────────────────────
			// Feature Sections
			// ──────────────────────────────────────

			"feature_section": {
				Name:      "feature_section",
				Label:     "Feature Section",
				Type:      types.NewNullableString("section"),
				ParentRef: "marketing_page",
				FieldRefs: []FieldDef{
					{Name: "eyebrow", Label: "Eyebrow Text", Type: types.FieldTypeText},
					{Name: "heading", Label: "Heading", Type: types.FieldTypeText},
					{Name: "description", Label: "Description", Type: types.FieldTypeTextarea},
					{Name: "image", Label: "Section Image", Type: types.FieldTypeMedia},
					{Name: "image_dark", Label: "Section Image (Dark)", Type: types.FieldTypeMedia},
				},
			},

			"feature_item": {
				Name:      "feature_item",
				Label:     "Feature Item",
				Type:      types.NewNullableString("content"),
				ParentRef: "feature_section",
				FieldRefs: []FieldDef{
					{Name: "title", Label: "Title", Type: types.FieldTypeText},
					{Name: "description", Label: "Description", Type: types.FieldTypeTextarea},
					{Name: "icon", Label: "Icon SVG", Type: types.FieldTypeTextarea},
					{Name: "link_url", Label: "Link URL", Type: types.FieldTypeURL},
				},
			},

			// ──────────────────────────────────────
			// CTA Sections
			// ──────────────────────────────────────

			"cta_section": {
				Name:      "cta_section",
				Label:     "CTA Section",
				Type:      types.NewNullableString("section"),
				ParentRef: "marketing_page",
				FieldRefs: []FieldDef{
					{Name: "heading", Label: "Heading", Type: types.FieldTypeText},
					{Name: "description", Label: "Description", Type: types.FieldTypeTextarea},
					{Name: "image", Label: "Image", Type: types.FieldTypeMedia},
					{Name: "cta_text", Label: "CTA Text", Type: types.FieldTypeText},
					{Name: "cta_url", Label: "CTA URL", Type: types.FieldTypeURL},
				},
			},

			"cta_benefit": {
				Name:      "cta_benefit",
				Label:     "CTA Benefit",
				Type:      types.NewNullableString("content"),
				ParentRef: "cta_section",
				FieldRefs: []FieldDef{
					{Name: "text", Label: "Benefit Text", Type: types.FieldTypeText},
				},
			},

			// ──────────────────────────────────────
			// Pricing Sections
			// ──────────────────────────────────────

			"pricing_section": {
				Name:      "pricing_section",
				Label:     "Pricing Section",
				Type:      types.NewNullableString("section"),
				ParentRef: "marketing_page",
				FieldRefs: []FieldDef{
					{Name: "eyebrow", Label: "Eyebrow Text", Type: types.FieldTypeText},
					{Name: "heading", Label: "Heading", Type: types.FieldTypeText},
					{Name: "description", Label: "Description", Type: types.FieldTypeTextarea},
				},
			},

			"pricing_tier": {
				Name:      "pricing_tier",
				Label:     "Pricing Tier",
				Type:      types.NewNullableString("content"),
				ParentRef: "pricing_section",
				FieldRefs: []FieldDef{
					{Name: "name", Label: "Tier Name", Type: types.FieldTypeText},
					{Name: "description", Label: "Description", Type: types.FieldTypeTextarea},
					{Name: "price_monthly", Label: "Monthly Price", Type: types.FieldTypeText},
					{Name: "price_annual", Label: "Annual Price", Type: types.FieldTypeText},
					{Name: "price_period", Label: "Price Period", Type: types.FieldTypeText},
					{Name: "featured", Label: "Featured", Type: types.FieldTypeBoolean},
					{Name: "badge_text", Label: "Badge Text", Type: types.FieldTypeText},
					{Name: "cta_text", Label: "CTA Text", Type: types.FieldTypeText},
					{Name: "cta_url", Label: "CTA URL", Type: types.FieldTypeURL},
				},
			},

			"pricing_feature": {
				Name:      "pricing_feature",
				Label:     "Pricing Feature",
				Type:      types.NewNullableString("content"),
				ParentRef: "pricing_tier",
				FieldRefs: []FieldDef{
					{Name: "text", Label: "Feature Text", Type: types.FieldTypeText},
				},
			},

			// ──────────────────────────────────────
			// Testimonials
			// ──────────────────────────────────────

			"testimonial_section": {
				Name:      "testimonial_section",
				Label:     "Testimonial Section",
				Type:      types.NewNullableString("section"),
				ParentRef: "marketing_page",
				FieldRefs: []FieldDef{
					{Name: "eyebrow", Label: "Eyebrow Text", Type: types.FieldTypeText},
					{Name: "heading", Label: "Heading", Type: types.FieldTypeText},
				},
			},

			"testimonial": {
				Name:      "testimonial",
				Label:     "Testimonial",
				Type:      types.NewNullableString("content"),
				ParentRef: "testimonial_section",
				FieldRefs: []FieldDef{
					{Name: "quote", Label: "Quote", Type: types.FieldTypeTextarea},
					{Name: "author_name", Label: "Author Name", Type: types.FieldTypeText},
					{Name: "author_handle", Label: "Author Handle", Type: types.FieldTypeText},
					{Name: "author_avatar", Label: "Author Avatar", Type: types.FieldTypeMedia},
					{Name: "company_logo", Label: "Company Logo", Type: types.FieldTypeMedia},
					{Name: "company_logo_dark", Label: "Company Logo (Dark)", Type: types.FieldTypeMedia},
					{Name: "featured", Label: "Featured", Type: types.FieldTypeBoolean},
				},
			},

			// ──────────────────────────────────────
			// Team Sections
			// ──────────────────────────────────────

			"team_section": {
				Name:      "team_section",
				Label:     "Team Section",
				Type:      types.NewNullableString("section"),
				ParentRef: "marketing_page",
				FieldRefs: []FieldDef{
					{Name: "heading", Label: "Heading", Type: types.FieldTypeText},
					{Name: "description", Label: "Description", Type: types.FieldTypeTextarea},
				},
			},

			"team_member": {
				Name:      "team_member",
				Label:     "Team Member",
				Type:      types.NewNullableString("content"),
				ParentRef: "team_section",
				FieldRefs: []FieldDef{
					{Name: "name", Label: "Name", Type: types.FieldTypeText},
					{Name: "role", Label: "Role", Type: types.FieldTypeText},
					{Name: "bio", Label: "Bio", Type: types.FieldTypeTextarea},
					{Name: "photo", Label: "Photo", Type: types.FieldTypeMedia},
				},
			},

			"team_social_link": {
				Name:      "team_social_link",
				Label:     "Team Social Link",
				Type:      types.NewNullableString("content"),
				ParentRef: "team_member",
				FieldRefs: []FieldDef{
					{Name: "platform", Label: "Platform", Type: types.FieldTypeText},
					{Name: "url", Label: "URL", Type: types.FieldTypeURL},
					{Name: "icon", Label: "Icon SVG", Type: types.FieldTypeTextarea},
				},
			},

			// ──────────────────────────────────────
			// Blog Sections
			// ──────────────────────────────────────

			"blog_section": {
				Name:      "blog_section",
				Label:     "Blog Section",
				Type:      types.NewNullableString("section"),
				ParentRef: "marketing_page",
				FieldRefs: []FieldDef{
					{Name: "heading", Label: "Heading", Type: types.FieldTypeText},
					{Name: "description", Label: "Description", Type: types.FieldTypeTextarea},
				},
			},

			"blog_post_card": {
				Name:      "blog_post_card",
				Label:     "Blog Post Card",
				Type:      types.NewNullableString("content"),
				ParentRef: "blog_section",
				FieldRefs: []FieldDef{
					{Name: "title", Label: "Title", Type: types.FieldTypeText},
					{Name: "excerpt", Label: "Excerpt", Type: types.FieldTypeTextarea},
					{Name: "date", Label: "Date", Type: types.FieldTypeDate},
					{Name: "category", Label: "Category", Type: types.FieldTypeText},
					{Name: "category_url", Label: "Category URL", Type: types.FieldTypeURL},
					{Name: "post_url", Label: "Post URL", Type: types.FieldTypeURL},
					{Name: "featured_image", Label: "Featured Image", Type: types.FieldTypeMedia},
					{Name: "author_name", Label: "Author Name", Type: types.FieldTypeText},
					{Name: "author_role", Label: "Author Role", Type: types.FieldTypeText},
					{Name: "author_avatar", Label: "Author Avatar", Type: types.FieldTypeMedia},
				},
			},

			// ──────────────────────────────────────
			// Contact Sections
			// ──────────────────────────────────────

			"contact_section": {
				Name:      "contact_section",
				Label:     "Contact Section",
				Type:      types.NewNullableString("section"),
				ParentRef: "marketing_page",
				FieldRefs: []FieldDef{
					{Name: "heading", Label: "Heading", Type: types.FieldTypeText},
					{Name: "description", Label: "Description", Type: types.FieldTypeTextarea},
					{Name: "submit_text", Label: "Submit Button Text", Type: types.FieldTypeText},
				},
			},

			"contact_detail": {
				Name:      "contact_detail",
				Label:     "Contact Detail",
				Type:      types.NewNullableString("content"),
				ParentRef: "contact_section",
				FieldRefs: []FieldDef{
					{Name: "type", Label: "Type", Type: types.FieldTypeSelect, Data: types.NewNullableString(`{"options":["address","phone","email"]}`)},
					{Name: "value", Label: "Value", Type: types.FieldTypeText},
					{Name: "url", Label: "URL", Type: types.FieldTypeURL},
					{Name: "icon", Label: "Icon SVG", Type: types.FieldTypeTextarea},
				},
			},

			// ──────────────────────────────────────
			// Content Sections
			// ──────────────────────────────────────

			"content_section": {
				Name:      "content_section",
				Label:     "Content Section",
				Type:      types.NewNullableString("section"),
				ParentRef: "marketing_page",
				FieldRefs: []FieldDef{
					{Name: "eyebrow", Label: "Eyebrow Text", Type: types.FieldTypeText},
					{Name: "heading", Label: "Heading", Type: types.FieldTypeText},
					{Name: "body", Label: "Body", Type: types.FieldTypeRichText},
					{Name: "image", Label: "Background Image", Type: types.FieldTypeMedia},
					{Name: "cta_text", Label: "CTA Text", Type: types.FieldTypeText},
					{Name: "cta_url", Label: "CTA URL", Type: types.FieldTypeURL},
				},
			},

			// ──────────────────────────────────────
			// FAQ Sections
			// ──────────────────────────────────────

			"faq_section": {
				Name:      "faq_section",
				Label:     "FAQ Section",
				Type:      types.NewNullableString("section"),
				ParentRef: "marketing_page",
				FieldRefs: []FieldDef{
					{Name: "heading", Label: "Heading", Type: types.FieldTypeText},
					{Name: "support_text", Label: "Support Text", Type: types.FieldTypeTextarea},
					{Name: "support_url", Label: "Support Link URL", Type: types.FieldTypeURL},
				},
			},

			"faq_item": {
				Name:      "faq_item",
				Label:     "FAQ Item",
				Type:      types.NewNullableString("content"),
				ParentRef: "faq_section",
				FieldRefs: []FieldDef{
					{Name: "question", Label: "Question", Type: types.FieldTypeText},
					{Name: "answer", Label: "Answer", Type: types.FieldTypeRichText},
				},
			},

			// ──────────────────────────────────────
			// Stats Sections
			// ──────────────────────────────────────

			"stats_section": {
				Name:      "stats_section",
				Label:     "Stats Section",
				Type:      types.NewNullableString("section"),
				ParentRef: "marketing_page",
				FieldRefs: []FieldDef{
					{Name: "eyebrow", Label: "Eyebrow Text", Type: types.FieldTypeText},
					{Name: "heading", Label: "Heading", Type: types.FieldTypeText},
					{Name: "description", Label: "Description", Type: types.FieldTypeTextarea},
				},
			},

			"stat_item": {
				Name:      "stat_item",
				Label:     "Stat Item",
				Type:      types.NewNullableString("content"),
				ParentRef: "stats_section",
				FieldRefs: []FieldDef{
					{Name: "value", Label: "Value", Type: types.FieldTypeText},
					{Name: "label", Label: "Label", Type: types.FieldTypeText},
					{Name: "description", Label: "Description", Type: types.FieldTypeTextarea},
				},
			},

			// ──────────────────────────────────────
			// Newsletter Sections
			// ──────────────────────────────────────

			"newsletter_section": {
				Name:      "newsletter_section",
				Label:     "Newsletter Section",
				Type:      types.NewNullableString("section"),
				ParentRef: "marketing_page",
				FieldRefs: []FieldDef{
					{Name: "heading", Label: "Heading", Type: types.FieldTypeText},
					{Name: "description", Label: "Description", Type: types.FieldTypeTextarea},
					{Name: "placeholder", Label: "Email Placeholder", Type: types.FieldTypeText},
					{Name: "submit_text", Label: "Submit Button Text", Type: types.FieldTypeText},
				},
			},

			"newsletter_detail": {
				Name:      "newsletter_detail",
				Label:     "Newsletter Detail",
				Type:      types.NewNullableString("content"),
				ParentRef: "newsletter_section",
				FieldRefs: []FieldDef{
					{Name: "title", Label: "Title", Type: types.FieldTypeText},
					{Name: "description", Label: "Description", Type: types.FieldTypeTextarea},
					{Name: "icon", Label: "Icon SVG", Type: types.FieldTypeTextarea},
				},
			},

			// ──────────────────────────────────────
			// Logo Cloud
			// ──────────────────────────────────────

			"logo_cloud_section": {
				Name:      "logo_cloud_section",
				Label:     "Logo Cloud Section",
				Type:      types.NewNullableString("section"),
				ParentRef: "marketing_page",
				FieldRefs: []FieldDef{
					{Name: "heading", Label: "Heading", Type: types.FieldTypeText},
					{Name: "description", Label: "Description", Type: types.FieldTypeTextarea},
					{Name: "cta_text", Label: "CTA Text", Type: types.FieldTypeText},
					{Name: "cta_url", Label: "CTA URL", Type: types.FieldTypeURL},
				},
			},

			"logo_item": {
				Name:      "logo_item",
				Label:     "Logo Item",
				Type:      types.NewNullableString("content"),
				ParentRef: "logo_cloud_section",
				FieldRefs: []FieldDef{
					{Name: "company_name", Label: "Company Name", Type: types.FieldTypeText},
					{Name: "logo", Label: "Logo", Type: types.FieldTypeMedia},
					{Name: "logo_dark", Label: "Logo (Dark)", Type: types.FieldTypeMedia},
				},
			},

			// ──────────────────────────────────────
			// Header Section (Page Headers / Banners)
			// ──────────────────────────────────────

			"header_section": {
				Name:      "header_section",
				Label:     "Header Section",
				Type:      types.NewNullableString("section"),
				ParentRef: "marketing_page",
				FieldRefs: []FieldDef{
					{Name: "heading", Label: "Heading", Type: types.FieldTypeText},
					{Name: "description", Label: "Description", Type: types.FieldTypeTextarea},
					{Name: "background_image", Label: "Background Image", Type: types.FieldTypeMedia},
					{Name: "background_image_dark", Label: "Background Image (Dark)", Type: types.FieldTypeMedia},
				},
			},

			"header_action_link": {
				Name:      "header_action_link",
				Label:     "Header Action Link",
				Type:      types.NewNullableString("content"),
				ParentRef: "header_section",
				FieldRefs: []FieldDef{
					{Name: "label", Label: "Label", Type: types.FieldTypeText},
					{Name: "url", Label: "URL", Type: types.FieldTypeURL},
				},
			},

			// ──────────────────────────────────────
			// Banners
			// ──────────────────────────────────────

			"banner": {
				Name:      "banner",
				Label:     "Banner",
				Type:      types.NewNullableString("section"),
				ParentRef: "marketing_page",
				FieldRefs: []FieldDef{
					{Name: "text", Label: "Announcement Text", Type: types.FieldTypeText},
					{Name: "highlight", Label: "Highlight Text", Type: types.FieldTypeText},
					{Name: "cta_text", Label: "CTA Text", Type: types.FieldTypeText},
					{Name: "cta_url", Label: "CTA URL", Type: types.FieldTypeURL},
					{Name: "dismissible", Label: "Dismissible", Type: types.FieldTypeBoolean},
				},
			},

			// ──────────────────────────────────────
			// Bento Grid
			// ──────────────────────────────────────

			"bento_grid_section": {
				Name:      "bento_grid_section",
				Label:     "Bento Grid Section",
				Type:      types.NewNullableString("section"),
				ParentRef: "marketing_page",
				FieldRefs: []FieldDef{
					{Name: "eyebrow", Label: "Eyebrow Text", Type: types.FieldTypeText},
					{Name: "heading", Label: "Heading", Type: types.FieldTypeText},
				},
			},

			"bento_cell": {
				Name:      "bento_cell",
				Label:     "Bento Cell",
				Type:      types.NewNullableString("content"),
				ParentRef: "bento_grid_section",
				FieldRefs: []FieldDef{
					{Name: "title", Label: "Title", Type: types.FieldTypeText},
					{Name: "description", Label: "Description", Type: types.FieldTypeTextarea},
					{Name: "image", Label: "Image", Type: types.FieldTypeMedia},
					{Name: "image_dark", Label: "Image (Dark)", Type: types.FieldTypeMedia},
					{Name: "span", Label: "Grid Span", Type: types.FieldTypeSelect, Data: types.NewNullableString(`{"options":["1","2","3"]}`)},
				},
			},

			// ──────────────────────────────────────
			// 404 Page
			// ──────────────────────────────────────

			"error_page": {
				Name:      "error_page",
				Label:     "Error Page",
				Type:      types.NewNullableString("section"),
				ParentRef: "marketing_page",
				FieldRefs: []FieldDef{
					{Name: "error_code", Label: "Error Code", Type: types.FieldTypeText},
					{Name: "heading", Label: "Heading", Type: types.FieldTypeText},
					{Name: "description", Label: "Description", Type: types.FieldTypeTextarea},
					{Name: "home_text", Label: "Home Link Text", Type: types.FieldTypeText},
					{Name: "home_url", Label: "Home Link URL", Type: types.FieldTypeURL},
				},
			},

			"popular_page": {
				Name:      "popular_page",
				Label:     "Popular Page Link",
				Type:      types.NewNullableString("content"),
				ParentRef: "error_page",
				FieldRefs: []FieldDef{
					{Name: "title", Label: "Title", Type: types.FieldTypeText},
					{Name: "description", Label: "Description", Type: types.FieldTypeTextarea},
					{Name: "url", Label: "URL", Type: types.FieldTypeURL},
					{Name: "icon", Label: "Icon SVG", Type: types.FieldTypeTextarea},
				},
			},

			// ──────────────────────────────────────
			// Embedded Testimonial (for content sections)
			// ──────────────────────────────────────

			"inline_testimonial": {
				Name:      "inline_testimonial",
				Label:     "Inline Testimonial",
				Type:      types.NewNullableString("content"),
				ParentRef: "content_section",
				FieldRefs: []FieldDef{
					{Name: "quote", Label: "Quote", Type: types.FieldTypeTextarea},
					{Name: "author_name", Label: "Author Name", Type: types.FieldTypeText},
					{Name: "author_role", Label: "Author Role", Type: types.FieldTypeText},
					{Name: "company_logo", Label: "Company Logo", Type: types.FieldTypeMedia},
				},
			},
		},
	})
}
