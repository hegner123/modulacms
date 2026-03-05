package tui

// ConfigCategorySelectMsg navigates to a config category detail view.
type ConfigCategorySelectMsg struct {
	Category string
}

// ConfigFieldUpdateMsg requests updating a config field value.
type ConfigFieldUpdateMsg struct {
	Key   string
	Value string
}

// ConfigUpdateResultMsg carries the result of a config update.
type ConfigUpdateResultMsg struct {
	RestartRequired []string
	Err             error
}
