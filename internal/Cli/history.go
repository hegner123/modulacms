package cli

type PageHistory struct {
	Cursor int
	Page   Page
}

func (m *model) PushHistory(entry PageHistory) {
	m.history = append(m.history, entry)
}

func (m *model) PopHistory() *PageHistory {
	if len(m.history) == 0 {
		return nil
	}
	index := len(m.history) - 1
	pageHistory := m.history[index]
	m.history = m.history[:index]
	return &pageHistory
}

func (m *model) Peek() (*PageHistory, bool) {
	if len(m.history) == 0 {
		return nil, false
	}
	return &m.history[len(m.history)-1], true
}
