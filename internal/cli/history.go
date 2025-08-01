package cli

type PageHistory struct {
	Cursor int
	Page   Page
}

func (m *Model) PushHistory(entry PageHistory) {
	m.History = append(m.History, entry)
}

func (m *Model) PopHistory() *PageHistory {
	if len(m.History) == 0 {
		return nil
	}
	index := len(m.History) - 1
	pageHistory := m.History[index]
	m.History = m.History[:index]
	return &pageHistory
}

func (m *Model) Peek() (*PageHistory, bool) {
	if len(m.History) == 0 {
		return nil, false
	}
	return &m.History[len(m.History)-1], true
}