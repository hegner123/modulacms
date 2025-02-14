package cli

func (m *model) PushHistory(entry CliPage) {
	m.history = append(m.history, entry)
}

func (m *model) PopHistory() *CliPage {
	if len(m.history) == 0 {
		return nil
	}
	// Get the last element.
	index := len(m.history) - 1
	page := m.history[index]
	m.history = m.history[:index]
	return &page
}

func (m *model) Peek() (*CliPage, bool) {
	if len(m.history) == 0 {
		return nil, false
	}
	return &m.history[len(m.history)-1], true
}
