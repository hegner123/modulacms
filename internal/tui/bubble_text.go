package tui

func init() {
	RegisterFieldInput(FieldInputEntry{
		Key:         "text",
		Label:       "Text",
		Description: "Single-line text input",
		NewBubble:   func() FieldBubble { return NewTextBubble() },
	})
}

func NewTextBubble() *TextInputBubble {
	return NewTextInputBubble("Text", "Enter text...", 256, nil)
}
