package tui

func init() {
	RegisterFieldInput(FieldInputEntry{
		Key:         "url",
		Label:       "URL",
		Description: "URL input",
		NewBubble:   func() FieldBubble { return NewURLBubble() },
	})
}

func NewURLBubble() *TextInputBubble {
	return NewTextInputBubble("URL", "https://...", 512, nil)
}
