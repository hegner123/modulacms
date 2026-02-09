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
		t := msg.Table
		dbt := db.StringDBTable(t)
		d := db.ConfigDB(msg.Config)

		columns := db.GenericHeaders(dbt)
		if columns == nil {
			return m, ErrorSetCmd(fmt.Errorf("unknown table: %s", t))
		}
		listRows, err := db.GenericList(dbt, d)
		if err != nil {
			m.Logger.Ferror("", err)
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
		clm := db.GenericHeaders(dbt)
		if clm == nil {
			return m, ErrorSetCmd(fmt.Errorf("unknown table: %s", msg.Table))
		}
		return m, ColumnsSetCmd(&clm)
	case DatatypesFetchMsg:
		return m, tea.Batch(
			LoadingStartCmd(),
			DatabaseListCmd(DATATYPEMENU, db.Datatype),
		)

	case DatatypesFetchResultsMsg:
		m.Logger.Finfo("tableFetchedMsg returned")
		newMenu := m.BuildDatatypeMenu(msg.Data)
		m.Logger.Finfo("newMenu", newMenu)

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
		if d == nil {
			return m, func() tea.Msg {
				return FetchErrMsg{Error: fmt.Errorf("database not connected")}
			}
		}
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
			dtFields, err := d.ListDatatypeFieldByDatatypeID(datatypeID)
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if dtFields == nil || len(*dtFields) == 0 {
				return DatatypeFieldsFetchResultsMsg{Fields: []db.Fields{}}
			}

			// Fetch actual field details for each field ID
			var fields []db.Fields
			for _, dtf := range *dtFields {
				if !dtf.FieldID.IsZero() {
					field, err := d.GetField(dtf.FieldID)
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

	case UsersFetchMsg:
		d := m.DB
		return m, func() tea.Msg {
			users, err := d.ListUsers()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if users == nil {
				return UsersFetchResultsMsg{Data: []db.Users{}}
			}
			return UsersFetchResultsMsg{Data: *users}
		}

	case UsersFetchResultsMsg:
		return m, tea.Batch(
			UsersListSetCmd(msg.Data),
			LoadingStopCmd(),
		)

	case RootContentSummaryFetchMsg:
		d := m.DB
		return m, func() tea.Msg {
			summary, err := d.ListRootContentSummary()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if summary == nil {
				return RootContentSummaryFetchResultsMsg{Data: []db.RootContentSummary{}}
			}
			return RootContentSummaryFetchResultsMsg{Data: *summary}
		}

	case RootContentSummaryFetchResultsMsg:
		return m, tea.Batch(
			RootContentSummarySetCmd(msg.Data),
			LoadingStopCmd(),
		)

	case FetchChildDatatypesMsg:
		d := m.DB
		parentID := msg.ParentDatatypeID
		routeID := msg.RouteID
		return m, func() tea.Msg {
			children, err := d.ListDatatypeChildren(parentID)
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if children == nil || len(*children) == 0 {
				// No child datatypes, cannot create child content
				return ActionResultMsg{
					Title:   "No Child Types",
					Message: "This datatype has no child types defined.",
				}
			}
			return ShowChildDatatypeDialogMsg{
				ChildDatatypes: *children,
				RouteID:        string(routeID),
			}
		}

	case FetchContentFieldsMsg:
		d := m.DB
		datatypeID := msg.DatatypeID
		routeID := msg.RouteID
		parentID := msg.ParentID
		title := msg.Title
		utility.DefaultLogger.Finfo(fmt.Sprintf("FetchContentFieldsMsg received: DatatypeID=%s, RouteID=%s", datatypeID, routeID))
		return m, func() tea.Msg {
			utility.DefaultLogger.Finfo("FetchContentFieldsMsg command executing...")
			// Fetch field IDs from the datatypes_fields join table
			dtFields, err := d.ListDatatypeFieldByDatatypeID(datatypeID)
			if err != nil {
				utility.DefaultLogger.Ferror("ListDatatypeFieldByDatatypeID error", err)
				return FetchErrMsg{Error: err}
			}

			if dtFields == nil || len(*dtFields) == 0 {
				utility.DefaultLogger.Finfo("No datatype fields found")
				return ActionResultMsg{
					Title:   "No Fields",
					Message: "This datatype has no fields defined.",
				}
			}

			utility.DefaultLogger.Finfo(fmt.Sprintf("Found %d datatype fields", len(*dtFields)))

			// Fetch actual field details for each field ID
			var fields []db.Fields
			for _, dtf := range *dtFields {
				if !dtf.FieldID.IsZero() {
					field, err := d.GetField(dtf.FieldID)
					if err == nil && field != nil {
						fields = append(fields, *field)
					}
				}
			}

			if len(fields) == 0 {
				utility.DefaultLogger.Finfo("No valid fields found after fetching")
				return ActionResultMsg{
					Title:   "No Fields",
					Message: "No valid fields found for this datatype.",
				}
			}

			utility.DefaultLogger.Finfo(fmt.Sprintf("Returning ShowContentFormDialogMsg with %d fields", len(fields)))
			return ShowContentFormDialogMsg{
				Action:     FORMDIALOGCREATECONTENT,
				Title:      title,
				DatatypeID: datatypeID,
				RouteID:    routeID,
				ParentID:   parentID,
				Fields:     fields,
			}
		}

	case FetchErrMsg:
		// Handle an error from data fetching - show dialog to user
		return m, tea.Batch(
			ErrorSetCmd(msg.Error),
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Database fetch error: %s", msg.Error.Error())),
			func() tea.Msg {
				return ActionResultMsg{
					Title:   "Fetch Error",
					Message: msg.Error.Error(),
				}
			},
		)
	}
	return m, nil
}
