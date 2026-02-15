package cli

// PageHistory represents a saved page state in navigation history.
type PageHistory struct {
	Cursor int
	Page   Page
	Menu   []Page
}

// PushHistory adds an entry to the navigation history stack.
func (m *Model) PushHistory(entry PageHistory) {
	m.History = append(m.History, entry)
}

// PopHistory removes and returns the last entry from the navigation history stack.
func (m *Model) PopHistory() *PageHistory {
	if len(m.History) == 0 {
		return nil
	}
	index := len(m.History) - 1
	pageHistory := m.History[index]
	m.History = m.History[:index]
	return &pageHistory
}

// Peek returns the last entry from the navigation history stack without removing it.
func (m *Model) Peek() (*PageHistory, bool) {
	if len(m.History) == 0 {
		return nil, false
	}
	return &m.History[len(m.History)-1], true
}
