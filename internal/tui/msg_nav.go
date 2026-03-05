package tui

// NavigateToPage requests navigation to a specific page with optional menu.
type NavigateToPage struct {
	Page Page
	Menu []*Page
}

// NavigateToDatabaseCreate requests navigation to the database create page.
type NavigateToDatabaseCreate struct{}

// HistoryPop pops the last page from navigation history.
type HistoryPop struct{}

// HistoryPush pushes a page onto the navigation history stack.
type HistoryPush struct {
	Page PageHistory
}

// SelectTable selects a table for viewing or editing.
type SelectTable struct {
	Table string
}
