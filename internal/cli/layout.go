package cli

import (
	"github.com/charmbracelet/lipgloss"
)

func NewHorizontalGroup(p lipgloss.Position, s []string) string {
	return lipgloss.JoinHorizontal(p, s...)
}

func NewVerticalGroup(p lipgloss.Position, s []string) string {
	return lipgloss.JoinVertical(p, s...)
}

type Row struct {
	Position lipgloss.Position
	Items    []Column
}

func (r *Row) AddColumn(c Column) Row {
	r.Items = append(r.Items, c)
	return *r
}

func (r *Row) SetPosition(p lipgloss.Position) Row {
	r.Position = p
	return *r
}

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

func (r Row) Build() string {
	s := make([]string, len(r.Items))
	for _, c := range r.Items {
		s = append(s, c.Build())
	}
	return lipgloss.JoinHorizontal(r.Position, s...)
}

type Column struct {
	Position lipgloss.Position
	Items    []string
}

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

func (c *Column) AddLine(s string) Column {
	c.Items = append(c.Items, s)
	return *c
}

func (c *Column) SetPosition(p lipgloss.Position) Column {
	c.Position = p
	return *c
}

func (c Column) Build() string {
	return lipgloss.JoinVertical(c.Position, c.Items...)
}
