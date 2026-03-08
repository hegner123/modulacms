package tui

import (
	"github.com/hegner123/modulacms/internal/model"
)

// LogModelMsg requests model state logging with optional include/exclude filters.
type LogModelMsg struct {
	Include *[]string
	Exclude *[]string
}

// ClearScreen requests clearing the terminal screen.
type ClearScreen struct{}

// SetReadyMsg sets the application ready state.
type SetReadyMsg struct {
	Ready bool
}

// TitleFontMsg cycles the title font forward or backward.
type TitleFontMsg struct {
	Forward bool
}

// SetLoadingMsg sets the loading state.
type SetLoadingMsg struct {
	Loading bool
}

// CursorAction represents the type of cursor movement.
type CursorAction int

const (
	CursorMoveUp CursorAction = iota
	CursorMoveDown
	CursorMoveReset
	CursorMoveSet
)

// CursorMsg requests a cursor movement.
type CursorMsg struct {
	Action CursorAction
	Index  int
}

// UpdateMaxCursorMsg updates the maximum cursor position.
type UpdateMaxCursorMsg struct {
	CursorMax int
}

// PageModMsg navigates pagination forward or backward.
type PageModMsg struct {
	Forward bool
}

// PageSet sets the current page.
type PageSet struct {
	Page Page
}

// UpdatePagination triggers recalculation of pagination state.
type UpdatePagination struct{}

// FocusSet sets the focus to a specific UI element.
type FocusSet struct {
	Focus FocusKey
}

// CursorMaxSet sets the maximum cursor value for the current view.
type CursorMaxSet struct {
	CursorMax int
}

// PaginatorUpdate updates paginator configuration.
type PaginatorUpdate struct {
	PerPage    int
	TotalPages int
}

// StatusSet sets the application status.
type StatusSet struct {
	Status ApplicationState
}

// ErrorSet sets the current error state.
type ErrorSet struct {
	Err error
}

// LogMsg requests logging a message.
type LogMsg struct {
	Message string
}

// SetPageContent sets the content to display on the current page.
type SetPageContent struct {
	Content string
}

// SetViewportContent sets the content for the viewport display.
type SetViewportContent struct {
	Content string
}

// RootSet sets the root model state.
type RootSet struct {
	Root model.Root
}

// DatatypeMenuSet sets the datatype menu options.
type DatatypeMenuSet struct {
	DatatypeMenu []string
}

// PageMenuSet sets the page menu items.
type PageMenuSet struct {
	PageMenu []Page
}

// DialogReadyOKSet sets whether the dialog OK button is ready.
type DialogReadyOKSet struct {
	Ready bool
}
