package tui

import (
	"charm.land/huh/v2"
)

// FormCreate requests creation of a new form of the specified type.
type FormCreate struct {
	FormType FormIndex
}

// FormSet sets the active form and its initial values.
type FormSet struct {
	Form   huh.Form
	Values []*string
}

// FormValuesSet updates the form values.
type FormValuesSet struct {
	Values []*string
}

// FormAborted signals that a form operation was aborted.
type FormAborted struct {
	Action DatabaseCMD
	Table  string
}

// FormSubmitMsg requests form submission.
type FormSubmitMsg struct{}

// FormCompletedMsg signals form completion with optional destination page.
type FormCompletedMsg struct {
	DestinationPage *Page // Optional - if nil, will try history pop, then home
}

// FormActionMsg requests a database action based on form data.
type FormActionMsg struct {
	Action  DatabaseCMD
	Table   string
	Columns []string
	Values  []*string
}

// FormCancelMsg signals form cancellation.
type FormCancelMsg struct{}

// FormOptionsSet sets the options map for form select fields.
type FormOptionsSet struct {
	Options *FormOptionsMap
}

// FormInitOptionsMsg requests initialization of form options for a specific form and table.
type FormInitOptionsMsg struct {
	Form  string
	Table string
}

// FormLenSet sets the length of the current form.
type FormLenSet struct {
	FormLen int
}

// FormMapSet sets the form field mapping.
type FormMapSet struct {
	FormMap []string
}
