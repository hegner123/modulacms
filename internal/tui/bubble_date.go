package tui

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/hegner123/modulacms/internal/config"
)

func init() {
	RegisterFieldInput(FieldInputEntry{
		Key:         "date",
		Label:       "Date",
		Description: "Date picker with calendar",
		NewBubble:   func() FieldBubble { return NewDatePickerBubble() },
	})
}

// DatePickerMode determines which parts of the date/time are editable.
type DatePickerMode int

const (
	DateOnlyMode DatePickerMode = iota
	DateTimeMode
	TimeOnlyMode
)

// datePickerFocus tracks which sub-field has focus within the bubble.
type datePickerFocus int

const (
	dpFocusCalendar datePickerFocus = iota
	dpFocusHour
	dpFocusMinute
)

// DatePickerBubble is a FieldBubble that displays a navigable calendar grid
// rendered with a lipgloss table. Supports date-only, datetime, and time-only modes.
//
// Calendar navigation: arrow keys or h/j/k/l for days/weeks, H/L for months, t for today.
// Time navigation: up/down to change value, left/right to switch hour/minute.
// Datetime sub-focus: ) moves calendar→time, ( moves time→calendar.
type DatePickerBubble struct {
	mode     DatePickerMode
	cursor   time.Time
	today    time.Time
	focused  bool
	width    int
	hour     int
	minute   int
	subFocus datePickerFocus
}

// NewDatePickerBubble creates a date-only picker initialized to today.
func NewDatePickerBubble() *DatePickerBubble {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	return &DatePickerBubble{
		mode:   DateOnlyMode,
		cursor: today,
		today:  today,
		hour:   now.Hour(),
		minute: now.Minute(),
	}
}

// NewDateTimePickerBubble creates a datetime picker initialized to now.
func NewDateTimePickerBubble() *DatePickerBubble {
	b := NewDatePickerBubble()
	b.mode = DateTimeMode
	return b
}

// NewTimePickerBubble creates a time-only picker initialized to current time.
func NewTimePickerBubble() *DatePickerBubble {
	b := NewDatePickerBubble()
	b.mode = TimeOnlyMode
	b.subFocus = dpFocusHour
	return b
}

// Update handles key messages for calendar and time navigation.
func (b *DatePickerBubble) Update(msg tea.Msg) (FieldBubble, tea.Cmd) {
	if !b.focused {
		return b, nil
	}
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return b, nil
	}
	switch b.subFocus {
	case dpFocusCalendar:
		b.updateCalendar(keyMsg)
	case dpFocusHour:
		b.updateHour(keyMsg)
	case dpFocusMinute:
		b.updateMinute(keyMsg)
	}
	return b, nil
}

func (b *DatePickerBubble) updateCalendar(msg tea.KeyPressMsg) {
	switch msg.String() {
	case "left", "h":
		b.cursor = b.cursor.AddDate(0, 0, -1)
	case "right", "l":
		b.cursor = b.cursor.AddDate(0, 0, 1)
	case "up", "k":
		b.cursor = b.cursor.AddDate(0, 0, -7)
	case "down", "j":
		b.cursor = b.cursor.AddDate(0, 0, 7)
	case "H":
		b.prevMonth()
	case "L":
		b.nextMonth()
	case "t":
		b.cursor = b.today
	case ")":
		if b.mode == DateTimeMode {
			b.subFocus = dpFocusHour
		}
	}
}

func (b *DatePickerBubble) updateHour(msg tea.KeyPressMsg) {
	switch msg.String() {
	case "up", "k":
		b.hour = (b.hour + 1) % 24
	case "down", "j":
		b.hour = (b.hour + 23) % 24
	case "left", "h":
		if b.mode == DateTimeMode {
			b.subFocus = dpFocusCalendar
		}
	case "right", "l":
		b.subFocus = dpFocusMinute
	default:
		b.typeDigitHour(msg)
	}
}

func (b *DatePickerBubble) updateMinute(msg tea.KeyPressMsg) {
	switch msg.String() {
	case "up", "k":
		b.minute = (b.minute + 1) % 60
	case "down", "j":
		b.minute = (b.minute + 59) % 60
	case "left", "h":
		b.subFocus = dpFocusHour
	case "right", "l":
		if b.mode == DateTimeMode {
			b.subFocus = dpFocusCalendar
		}
	default:
		b.typeDigitMinute(msg)
	}
}

func (b *DatePickerBubble) typeDigitHour(msg tea.KeyPressMsg) {
	if len(msg.Text) != 1 {
		return
	}
	r := rune(msg.Text[0])
	if r < '0' || r > '9' {
		return
	}
	d := int(r - '0')
	h := b.hour*10 + d
	if h > 23 {
		h = d
	}
	b.hour = h
}

func (b *DatePickerBubble) typeDigitMinute(msg tea.KeyPressMsg) {
	if len(msg.Text) != 1 {
		return
	}
	r := rune(msg.Text[0])
	if r < '0' || r > '9' {
		return
	}
	d := int(r - '0')
	m := b.minute*10 + d
	if m > 59 {
		m = d
	}
	b.minute = m
}

// prevMonth moves the cursor to the same day of the previous month,
// clamping to the last day if the target month is shorter.
func (b *DatePickerBubble) prevMonth() {
	y, m, d := b.cursor.Date()
	m--
	if m < time.January {
		m = time.December
		y--
	}
	max := daysInMonth(y, m)
	if d > max {
		d = max
	}
	b.cursor = time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}

// nextMonth moves the cursor to the same day of the next month,
// clamping to the last day if the target month is shorter.
func (b *DatePickerBubble) nextMonth() {
	y, m, d := b.cursor.Date()
	m++
	if m > time.December {
		m = time.January
		y++
	}
	max := daysInMonth(y, m)
	if d > max {
		d = max
	}
	b.cursor = time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()
}

// View renders the calendar grid and optional time fields.
func (b *DatePickerBubble) View() string {
	if b.mode == TimeOnlyMode {
		return b.renderTime()
	}
	cal := b.renderCalendar()
	if b.mode == DateTimeMode {
		return cal + "\n" + b.renderTime()
	}
	return cal
}

func (b *DatePickerBubble) renderCalendar() string {
	year, month, _ := b.cursor.Date()
	first := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	startWeekday := int(first.Weekday())
	days := daysInMonth(year, month)

	cursorDay := b.cursor.Day()
	cursorRow, cursorCol := dayPosition(cursorDay, startWeekday)

	ty, tm, td := b.today.Date()
	todayRow, todayCol := -1, -1
	if ty == year && tm == month {
		todayRow, todayCol = dayPosition(td, startWeekday)
	}

	rows := buildCalendarRows(days, startWeekday)

	cursorStyle := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Primary).
		Background(config.DefaultStyle.Accent).
		Bold(true).
		Align(lipgloss.Center)
	todayStyle := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Accent2).
		Bold(true).
		Align(lipgloss.Center)
	headerStyle := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Secondary).
		Bold(true).
		Align(lipgloss.Center)
	cellStyle := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Tertiary).
		Align(lipgloss.Center)

	calFocused := b.focused && b.subFocus == dpFocusCalendar

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(config.DefaultStyle.PrimaryBorder)).
		StyleFunc(func(r, c int) lipgloss.Style {
			if r == table.HeaderRow {
				return headerStyle
			}
			if calFocused && r == cursorRow && c == cursorCol {
				return cursorStyle
			}
			if r == todayRow && c == todayCol {
				return todayStyle
			}
			return cellStyle
		}).
		Headers("Su", "Mo", "Tu", "We", "Th", "Fr", "Sa").
		Rows(rows...)

	monthLabel := fmt.Sprintf("  ◂ %s %d ▸", month.String(), year)
	monthStyle := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Secondary).
		Bold(true)

	return monthStyle.Render(monthLabel) + "\n" + t.String()
}

func buildCalendarRows(days, startWeekday int) [][]string {
	var rows [][]string
	row := make([]string, 7)
	for i := 0; i < startWeekday; i++ {
		row[i] = "  "
	}
	for day := 1; day <= days; day++ {
		col := (startWeekday + day - 1) % 7
		row[col] = fmt.Sprintf("%2d", day)
		if col == 6 || day == days {
			if day == days && col < 6 {
				for c := col + 1; c < 7; c++ {
					row[c] = "  "
				}
			}
			rows = append(rows, row)
			row = make([]string, 7)
		}
	}
	return rows
}

func (b *DatePickerBubble) renderTime() string {
	hStyle := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Tertiary).
		Padding(0, 1)
	mStyle := hStyle
	sepStyle := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Secondary)
	activeStyle := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Primary).
		Background(config.DefaultStyle.Accent).
		Padding(0, 1).
		Bold(true)

	if b.focused && b.subFocus == dpFocusHour {
		hStyle = activeStyle
	}
	if b.focused && b.subFocus == dpFocusMinute {
		mStyle = activeStyle
	}

	return hStyle.Render(fmt.Sprintf("%02d", b.hour)) +
		sepStyle.Render(":") +
		mStyle.Render(fmt.Sprintf("%02d", b.minute))
}

// dayPosition maps a 1-based day number to table row and column indices.
func dayPosition(day, startWeekday int) (row, col int) {
	idx := startWeekday + day - 1
	return idx / 7, idx % 7
}

// Value returns the date/time as a formatted string.
func (b *DatePickerBubble) Value() string {
	switch b.mode {
	case DateTimeMode:
		return fmt.Sprintf("%sT%02d:%02d", b.cursor.Format("2006-01-02"), b.hour, b.minute)
	case TimeOnlyMode:
		return fmt.Sprintf("%02d:%02d", b.hour, b.minute)
	default:
		return b.cursor.Format("2006-01-02")
	}
}

// SetValue parses a date/time string and sets the bubble state.
func (b *DatePickerBubble) SetValue(v string) {
	if v == "" {
		return
	}
	switch b.mode {
	case DateTimeMode:
		if t, err := time.Parse("2006-01-02T15:04", v); err == nil {
			b.cursor = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
			b.hour = t.Hour()
			b.minute = t.Minute()
			return
		}
		if t, err := time.Parse("2006-01-02", v); err == nil {
			b.cursor = t
			return
		}
	case TimeOnlyMode:
		if t, err := time.Parse("15:04", v); err == nil {
			b.hour = t.Hour()
			b.minute = t.Minute()
			return
		}
	default:
		if t, err := time.Parse("2006-01-02", v); err == nil {
			b.cursor = t
			return
		}
		if t, err := time.Parse("2006-01-02T15:04", v); err == nil {
			b.cursor = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
			return
		}
	}
}

// Focus gives the bubble input focus.
func (b *DatePickerBubble) Focus() tea.Cmd {
	b.focused = true
	return nil
}

// Blur removes input focus.
func (b *DatePickerBubble) Blur() { b.focused = false }

// Focused returns whether the bubble currently has focus.
func (b *DatePickerBubble) Focused() bool { return b.focused }

// SetWidth sets the display width for layout.
func (b *DatePickerBubble) SetWidth(w int) { b.width = w }
