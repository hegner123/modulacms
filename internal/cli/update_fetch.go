package cli

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
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
			LogMessageCmd(fmt.Sprintf("Table %s headers fetched: %s", m.TableState.Table, strings.Join(columns, ", "))),
		)
	case TableHeadersRowsFetchedMsg:
		s := strings.Builder{}
		for _, v := range m.TableState.Headers {
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
		return m, tea.Batch(
			LoadingStartCmd(),
			DatabaseListCmd(DATATYPEMENU, db.Datatype),
		)

	case DatatypesFetchResultsMsg:
		utility.DefaultLogger.Finfo("tableFetchedMsg returned")
		newMenu := m.BuildDatatypeMenu(msg.Data)
		utility.DefaultLogger.Finfo("newMenu", newMenu)

		datatypeMenuLabels := make([]string, 0, len(newMenu))
		for _, item := range newMenu {
			datatypeMenuLabels = append(datatypeMenuLabels, item.Label)
		}

		return m, tea.Batch(
			LoadingStopCmd(),
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
	case RoutesFetchMsg:
		d := m.DB
		return m, func() tea.Msg {
			routes, err := d.ListRoutes()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if routes == nil {
				return RoutesFetchResultsMsg{Data: []db.Routes{}}
			}
			return RoutesFetchResultsMsg{Data: *routes}
		}

	case RoutesFetchResultsMsg:
		return m, tea.Batch(
			RoutesSetCmd(msg.Data),
			LoadingStopCmd(),
		)

	case RootDatatypesFetchMsg:
		d := m.DB
		return m, func() tea.Msg {
			datatypes, err := d.ListDatatypesRoot()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if datatypes == nil {
				return RootDatatypesFetchResultsMsg{Data: []db.Datatypes{}}
			}
			return RootDatatypesFetchResultsMsg{Data: *datatypes}
		}

	case RootDatatypesFetchResultsMsg:
		return m, tea.Batch(
			RootDatatypesSetCmd(msg.Data),
			LoadingStopCmd(),
		)

	case AllDatatypesFetchMsg:
		d := m.DB
		return m, func() tea.Msg {
			datatypes, err := d.ListDatatypes()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if datatypes == nil {
				return AllDatatypesFetchResultsMsg{Data: []db.Datatypes{}}
			}
			return AllDatatypesFetchResultsMsg{Data: *datatypes}
		}

	case AllDatatypesFetchResultsMsg:
		cmds := []tea.Cmd{
			AllDatatypesSetCmd(msg.Data),
			LoadingStopCmd(),
		}
		// Fetch fields for the first datatype (cursor position 0)
		if len(msg.Data) > 0 {
			cmds = append(cmds, DatatypeFieldsFetchCmd(msg.Data[0].DatatypeID))
		}
		return m, tea.Batch(cmds...)

	case DatatypeFieldsFetchMsg:
		d := m.DB
		datatypeID := msg.DatatypeID
		return m, func() tea.Msg {
			// Get field IDs from the join table
			dtID := types.NullableDatatypeID{ID: datatypeID, Valid: true}
			dtFields, err := d.ListDatatypeFieldByDatatypeID(dtID)
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if dtFields == nil || len(*dtFields) == 0 {
				return DatatypeFieldsFetchResultsMsg{Fields: []db.Fields{}}
			}

			// Fetch actual field details for each field ID
			var fields []db.Fields
			for _, dtf := range *dtFields {
				if dtf.FieldID.Valid {
					field, err := d.GetField(dtf.FieldID.ID)
					if err == nil && field != nil {
						fields = append(fields, *field)
					}
				}
			}
			return DatatypeFieldsFetchResultsMsg{Fields: fields}
		}

	case DatatypeFieldsFetchResultsMsg:
		return m, DatatypeFieldsSetCmd(msg.Fields)

	case RoutesByDatatypeFetchMsg:
		d := m.DB
		datatypeID := msg.DatatypeID
		return m, func() tea.Msg {
			routes, err := d.ListRoutesByDatatype(datatypeID)
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if routes == nil {
				return RoutesFetchResultsMsg{Data: []db.Routes{}}
			}
			return RoutesFetchResultsMsg{Data: *routes}
		}

	case MediaFetchMsg:
		d := m.DB
		return m, func() tea.Msg {
			media, err := d.ListMedia()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if media == nil {
				return MediaFetchResultsMsg{Data: []db.Media{}}
			}
			return MediaFetchResultsMsg{Data: *media}
		}

	case MediaFetchResultsMsg:
		return m, tea.Batch(
			MediaListSetCmd(msg.Data),
			LoadingStopCmd(),
		)

	case FetchErrMsg:
		// Handle an error from data fetching.
		return m, tea.Batch(
			ErrorSetCmd(msg.Error),
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Database fetch error for table %s: %s", m.TableState.Table, msg.Error.Error())),
		)
	}
	return m, nil
}
