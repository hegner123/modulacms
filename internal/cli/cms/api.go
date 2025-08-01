package cms

import (
	"github.com/charmbracelet/huh"
)


// CMS message types
type NewRootMSG struct{}
type NewNodeMSG struct {
	ParentID   int
	DatatypeID int
	ContentID  int
}
type LoadPageMSG struct {
	ContentID int
}
type SavePageMSG struct{}

// FormResult contains a built form and form-related information
type FormResult struct {
	Form        *huh.Form
	FieldsCount int
}
