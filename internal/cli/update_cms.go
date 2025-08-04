package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/db"
)

type CmsUpdate struct{}

func NewCmsUpdate() tea.Cmd {
	return func() tea.Msg {
		return CmsUpdate{}
	}
}

func (m Model) UpdateCms(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case BuildTreeFromRouteMsg:
		cdColumns := []string{"ContentDataID",
			"ParentID",
			"RouteID",
			"DatatypeID",
			"AuthorID",
			"DateCreated",
			"DateModified",
			"History",
		}
		cfColumns := []string{
			"ContentFieldID",
			"RouteID",
			"ContentDataID",
			"FieldID",
			"FieldValue",
			"AuthorID",
			"DateCreated",
			"DateModified",
			"History",
		}

		return m, tea.Batch(
			DatabaseListFilteredCmd(BUILDTREE, db.Content_data, cdColumns, "RouteID", msg.RouteID),
			DatabaseListFilteredCmd(BUILDTREE, db.Content_fields, cfColumns, "RouteID", msg.RouteID),
			LogMessageCmd(fmt.Sprintln(msg)),
		)
	case ProcessContentDataRowsFromRoute:
		cmds := make([]tea.Cmd, 0)
		for _, v := range msg.Rows {
			c := GetDatatypeFromContentDataCMD(v.DatatypeID)
			cmds = append(cmds, c)
		}
		return m, tea.Batch(cmds...)
	case ProcessContentFieldRowsFromRoute:
		cmds := make([]tea.Cmd, 0)
		for _, v := range msg.Rows {
			c := ListFieldsFromContentFieldsCmd(v.FieldID)
			cmds = append(cmds, c)
		}
		return m, tea.Batch(cmds...)
	case ProcessDatatypeFromContentDataID:
		return m, LogMessageCmd(msg.Row.Label)
	case ProcessFieldsFromContentFieldID:
		cmds := make([]tea.Cmd, 0)
		for _, v := range msg.Rows {
			cmds = append(cmds, LogMessageCmd(v.Label))
		}

		return m, tea.Batch(cmds...)
	case GetDatatypeFromContentData:
		return m, tea.Batch(
			DatabaseGetCmd(BUILDTREE, db.Datatype, msg.DatatypeID),
		)
	case ListFieldsFromContentFields:
		columns := []string{
			"FieldID",
			"ParentID",
			"Label",
			"Data",
			"Type",
			"AuthorID",
			"DateCreated",
			"DateModified",
			"History",
		}
		return m, tea.Batch(
			DatabaseListFilteredCmd(BUILDTREE, db.Datatype, columns, "FieldID", msg.FieldID),
		)

	case ListContentDataMsg:
		columns := []string{
			"ContentDataID",
			"ParentID",
			"RouteID",
			"DatatypeID",
		}
		return m, tea.Batch(
			DatabaseListFilteredCmd(PICKCONTENTDATA, db.Content_data, columns, "ParentID", nil),
		)
	default:
		return m, nil
	}

}
