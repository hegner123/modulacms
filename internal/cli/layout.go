package cli

import (
	"github.com/charmbracelet/lipgloss"
)

// NewHorizontalGroup joins a slice of strings horizontally at the given position.
func NewHorizontalGroup(p lipgloss.Position, s []string) string {
	return lipgloss.JoinHorizontal(p, s...)
}

// NewVerticalGroup joins a slice of strings vertically at the given position.
func NewVerticalGroup(p lipgloss.Position, s []string) string {
	return lipgloss.JoinVertical(p, s...)
}

// Row represents a horizontal layout containing columns.
type Row struct {
	Position lipgloss.Position
	Items    []Column
}

// AddColumn appends a column to the row and returns the updated row.
func (r *Row) AddColumn(c Column) Row {
	r.Items = append(r.Items, c)
	return *r
}

// SetPosition sets the horizontal alignment position for the row and returns it.
func (r *Row) SetPosition(p lipgloss.Position) Row {
	r.Position = p
	return *r
}

// NewRow creates a new row with the given position and optional columns.
func NewRow(p lipgloss.Position, c ...Column) Row {
	if len(c) > 0 {
		return Row{
			Position: p,
			Items:    c,
		}
	}
	return Row{
		Position: p,
		Items:    make([]Column, 0),
	}
}

// Build renders the row by joining its columns horizontally.
func (r Row) Build() string {
	s := make([]string, len(r.Items))
	for _, c := range r.Items {
		s = append(s, c.Build())
	}
	return lipgloss.JoinHorizontal(r.Position, s...)
}

// Column represents a vertical layout containing strings.
type Column struct {
	Position lipgloss.Position
	Items    []string
}

// NewColumn creates a new column with the given position and optional strings.
func NewColumn(p lipgloss.Position, s ...string) Column {
	if len(s) > 0 {
		return Column{
			Position: p,
			Items:    s,
		}
	}
	return Column{
		Position: p,
		Items:    make([]string, 0),
	}
}

// AddLine appends a string to the column and returns the updated column.
func (c *Column) AddLine(s string) Column {
	c.Items = append(c.Items, s)
	return *c
}

// SetPosition sets the vertical alignment position for the column and returns it.
func (c *Column) SetPosition(p lipgloss.Position) Column {
	c.Position = p
	return *c
}

// Build renders the column by joining its items vertically.
func (c Column) Build() string {
	return lipgloss.JoinVertical(c.Position, c.Items...)
}
