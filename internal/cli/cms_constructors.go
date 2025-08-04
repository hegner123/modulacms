package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/db"
)

func ListContentDataCMD() tea.Cmd {
	return func() tea.Msg {
		return ListContentDataMsg{}
	}
}
func BuildTreeFromRouteCMD(id int64) tea.Cmd {
	return func() tea.Msg {
		return BuildTreeFromRouteMsg{
			RouteID: id,
		}
	}
}

func GetDatatypeFromContentDataCMD(id int64) tea.Cmd {
	return func() tea.Msg {
		return GetDatatypeFromContentData{
			DatatypeID: id,
		}
	}
}

func ListFieldsFromContentFieldsCmd(id int64) tea.Cmd {
	return func() tea.Msg {
		return ListFieldsFromContentFields{
			FieldID: id,
		}
	}
}
func ProcessContentDataRowsFromRouteCMD(rows []db.ContentData) tea.Cmd {
	return func() tea.Msg {
		return ProcessContentDataRowsFromRoute{
			Rows: rows,
		}
	}
}
func ProcessContentFieldRowsFromRouteCMD(rows []db.ContentFields) tea.Cmd {
	return func() tea.Msg {
		return ProcessContentFieldRowsFromRoute{
			Rows: rows,
		}
	}
}
func ProcessDatatypeFromContentDataIdCMD(row db.Datatypes) tea.Cmd {
	return func() tea.Msg {
		return ProcessDatatypeFromContentDataID{
			Row: row,
		}
	}
}
func ProcessFieldsFromContentFieldIdCMD(rows []db.Fields) tea.Cmd {
	return func() tea.Msg {
		return ProcessFieldsFromContentFieldID{
			Rows: rows,
		}
	}
}
