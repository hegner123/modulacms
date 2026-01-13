package cli

import "github.com/charmbracelet/huh"

// FormModel encapsulates all form-related state extracted from Model.
// This groups the form lifecycle fields together following the DialogModel pattern.
type FormModel struct {
	Form        *huh.Form
	FormLen     int
	FormMap     []string
	FormValues  []*string
	FormSubmit  bool
	FormGroups  []huh.Group
	FormFields  []huh.Field
	FormOptions *FormOptionsMap
}

// NewFormModel creates a new FormModel with safe defaults.
// FormMap is initialized as an empty slice, FormSubmit is false.
// Other fields are initialized to their zero values.
func NewFormModel() *FormModel {
	return &FormModel{
		FormMap:    make([]string, 0),
		FormSubmit: false,
	}
}
