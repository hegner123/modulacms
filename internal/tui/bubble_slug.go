package tui

func init() {
	RegisterFieldInput(FieldInputEntry{
		Key:         "slug",
		Label:       "Slug",
		Description: "URL slug input",
		NewBubble:   func() FieldBubble { return NewSlugBubble() },
	})
}

func NewSlugBubble() *TextInputBubble {
	return NewTextInputBubble("Slug", "my-page-slug", 256, nil)
}
