package cli

func (m *model) PushHistory(entry Page) {
	m.history = append(m.history, entry)
}

func (m *model) PopHistory() *Page {
	if len(m.history) == 0 {
		return nil
	}
	// Get the last element.
	index := len(m.history) - 1
	page := m.history[index]
	m.history = m.history[:index]
	return &page
}

func (m *model) Peek() (*Page, bool) {
	if len(m.history) == 0 {
		return nil, false
	}
	return &m.history[len(m.history)-1], true
}
