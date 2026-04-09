package tui

func init() {
	RegisterFieldInput(FieldInputEntry{
		Key:         "email",
		Label:       "Email",
		Description: "email address input",
		NewBubble:   func() FieldBubble { return NewEmailBubble() },
	})
}

func NewEmailBubble() *TextInputBubble {
	return NewTextInputBubble("Email", "user@example.com", 256, nil)
}
