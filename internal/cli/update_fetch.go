package cli

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/utility"
)

type FetchUpdate struct{}

func NewFetchUpdate() tea.Cmd {
	return func() tea.Msg {
		return FetchUpdate{}
	}
}

func (m Model) UpdateFetch(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case FetchHeadersRows:
		c := msg.Config
		t := msg.Table
		dbt := db.StringDBTable(t)
		query := "SELECT * FROM"
		d := db.ConfigDB(c)
		rows, err := d.ExecuteQuery(query, dbt)
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return m, ErrorSetCmd(err)
		}
		defer utility.HandleRowsCloseDeferErr(rows)
		columns, err := rows.Columns()
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return m, ErrorSetCmd(err)
		}
		listRows, err := db.GenericList(dbt, d)
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return m, ErrorSetCmd(err)
		}

		return m, tea.Batch(
			TableHeadersRowsFetchedCmd(columns, listRows, msg.Page),
			LogMessageCmd(fmt.Sprintf("Table %s headers fetched: %s", m.Table, strings.Join(columns, ", "))),
		)
	case TableHeadersRowsFetchedMsg:
		s := strings.Builder{}
		for _, v := range m.Headers {
			s.WriteString(v)
			s.WriteString("\n")

		}
		return m, tea.Batch(
			HeadersSetCmd(msg.Headers),
			RowsSetCmd(msg.Rows),
			PaginatorUpdateCmd(m.MaxRows, len(msg.Rows)),
			CursorMaxSetCmd(m.Paginator.ItemsOnPage(len(msg.Rows))),
			PageSetCmd(*msg.Page),
			LogMessageCmd(s.String()),
			LoadingStopCmd(),
		)
	case GetColumns:
		dbt := db.StringDBTable(msg.Table)
		query := "SELECT * FROM"
		d := db.ConfigDB(msg.Config)
		rows, err := d.ExecuteQuery(query, dbt)
		if err != nil {
			return m, ErrorSetCmd(err)
		}
		defer utility.HandleRowsCloseDeferErr(rows)
		clm, err := rows.Columns()
		if err != nil {
			return m, ErrorSetCmd(err)
		}
		ct, err := rows.ColumnTypes()
		if err != nil {
			return m, ErrorSetCmd(err)
		}
		return m,
			tea.Batch(
				ColumnsSetCmd(&clm),
				ColumnTypesSetCmd(&ct),
			)
	case DatatypesFetchMsg:
		return m, DatabaseListCmd(DATATYPEMENU, db.Datatype)

	case DatatypesFetchResultsMsg:
		utility.DefaultLogger.Finfo("tableFetchedMsg returned")
		newMenu := m.BuildDatatypeMenu(msg.Data)
		utility.DefaultLogger.Finfo("newMenu", newMenu)

		datatypeMenuLabels := make([]string, 0, len(newMenu))
		for _, item := range newMenu {
			datatypeMenuLabels = append(datatypeMenuLabels, item.Label)
		}

		return m, tea.Batch(
			LogMessageCmd(fmt.Sprintln(datatypeMenuLabels)),
			DatatypeMenuSetCmd(datatypeMenuLabels),
			PageMenuSetCmd(newMenu),
		)

	case TablesFetch:
		return m, GetTablesCMD(m.Config)
	case ColumnsFetched:
		return m, tea.Batch(
			LoadingStopCmd(),
			ColumnTypesSetCmd(msg.ColumnTypes),
			ColumnsSetCmd(msg.Columns),
		)
	case FetchErrMsg:
		// Handle an error from data fetching.
		return m, tea.Batch(
			ErrorSetCmd(msg.Error),
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Database fetch error for table %s: %s", m.Table, msg.Error.Error())),
		)
	}
	return m, nil
}
