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
		defer rows.Close()
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
			TableHeadersRowsFetchedCmd(columns, listRows),
			LogMessageCmd(fmt.Sprintf("Table %s headers fetched: %s", m.Table, strings.Join(columns, ", "))),
		)
	case TableHeadersRowsFetchedMsg:
		return m, tea.Batch(
			HeadersSetCmd(msg.Headers),
			RowsSetCmd(msg.Rows),
			PaginatorUpdateCmd(m.MaxRows, len(msg.Rows)),
			CursorMaxSetCmd(m.Paginator.ItemsOnPage(len(msg.Rows))),
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
		defer rows.Close()
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

	case DatatypesFetchedMsg:
		utility.DefaultLogger.Finfo("tableFetchedMsg returned")
		newMenu := m.BuildDatatypeMenu(msg.data)
		utility.DefaultLogger.Finfo("newMenu", newMenu)

		datatypeMenuLabels := make([]string, 0, len(newMenu))
		for _, item := range newMenu {
			datatypeMenuLabels = append(datatypeMenuLabels, item.Label)
			utility.DefaultLogger.Finfo("item", item)
		}

		return m, tea.Batch(
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
