package cli

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
)

// FetchUpdate signals a data fetch operation.
type FetchUpdate struct{}

// NewFetchUpdate returns a command that creates a FetchUpdate message.
func NewFetchUpdate() tea.Cmd {
	return func() tea.Msg {
		return FetchUpdate{}
	}
}

// UpdateFetch handles data fetching pipelines for routes, datatypes, content, media, and users.
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
			// List fields by parent datatype ID
			fields, err := d.ListFieldsByDatatypeID(types.NullableDatatypeID{ID: datatypeID, Valid: true})
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if fields == nil || len(*fields) == 0 {
				return DatatypeFieldsFetchResultsMsg{Fields: []db.Fields{}}
			}
			return DatatypeFieldsFetchResultsMsg{Fields: *fields}
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
			users, err := d.ListUsersWithRoleLabel()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if users == nil {
				return UsersFetchResultsMsg{Data: []db.UserWithRoleLabelRow{}}
			}
			return UsersFetchResultsMsg{Data: *users}
		}

	case UsersFetchResultsMsg:
		return m, tea.Batch(
			UsersListSetCmd(msg.Data),
			LoadingStopCmd(),
		)

	case RolesFetchMsg:
		d := m.DB
		return m, func() tea.Msg {
			roles, err := d.ListRoles()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if roles == nil {
				return RolesFetchResultsMsg{Data: []db.Roles{}}
			}
			return RolesFetchResultsMsg{Data: *roles}
		}

	case RolesFetchResultsMsg:
		return m, RolesListSetCmd(msg.Data)

	case RootContentSummaryFetchMsg:
		d := m.DB
		return m, func() tea.Msg {
			summary, err := d.ListContentDataTopLevelPaginated(db.PaginationParams{Limit: 10000, Offset: 0})
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if summary == nil {
				return RootContentSummaryFetchResultsMsg{Data: []db.ContentDataTopLevel{}}
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
		routeID := msg.RouteID
		return m, func() tea.Msg {
			all, err := d.ListDatatypes()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if all == nil || len(*all) == 0 {
				return ActionResultMsg{
					Title:   "No Datatypes",
					Message: "No datatypes are defined.",
				}
			}
			return ShowChildDatatypeDialogMsg{
				ChildDatatypes: *all,
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
			// Fetch fields by parent datatype ID
			fieldList, err := d.ListFieldsByDatatypeID(types.NullableDatatypeID{ID: datatypeID, Valid: true})
			if err != nil {
				utility.DefaultLogger.Ferror("ListFieldsByDatatypeID error", err)
				return FetchErrMsg{Error: err}
			}

			var fields []db.Fields
			if fieldList != nil {
				fields = *fieldList
			}

			if len(fields) == 0 {
				utility.DefaultLogger.Finfo("No fields found for datatype")
				return ActionResultMsg{
					Title:   "No Fields",
					Message: "This datatype has no fields defined.",
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

	case PluginsFetchMsg:
		mgr := m.PluginManager
		if mgr == nil {
			return m, func() tea.Msg {
				return PluginsFetchResultsMsg{Data: []PluginDisplay{}}
			}
		}
		return m, func() tea.Msg {
			instances := mgr.ListPlugins()
			displays := make([]PluginDisplay, 0, len(instances))
			for _, inst := range instances {
				cbState := "closed"
				if inst.CB != nil {
					cbState = inst.CB.State().String()
				}
				displays = append(displays, PluginDisplay{
					Name:             inst.Info.Name,
					Version:          inst.Info.Version,
					State:            inst.State.String(),
					CBState:          cbState,
					Description:      inst.Info.Description,
					ManifestDrift:    inst.ManifestDrift,
					CapabilityDrifts: len(inst.CapabilityDrift),
					SchemaDrifts:     len(inst.SchemaDrift),
				})
			}
			return PluginsFetchResultsMsg{Data: displays}
		}

	case PluginsFetchResultsMsg:
		return m, tea.Batch(
			PluginsListSetCmd(msg.Data),
			LoadingStopCmd(),
		)

	case PipelinesFetchMsg:
		mgr := m.PluginManager
		if mgr == nil {
			return m, func() tea.Msg {
				return PipelinesFetchResultsMsg{Data: []PipelineDisplay{}}
			}
		}
		return m, func() tea.Msg {
			results := mgr.DryRunAllPipelines()
			displays := make([]PipelineDisplay, 0, len(results))
			for _, r := range results {
				displays = append(displays, PipelineDisplay{
					Key:       r.Table + "." + r.Phase + "_" + r.Operation,
					Table:     r.Table,
					Operation: r.Operation,
					Phase:     r.Phase,
					Count:     len(r.Entries),
				})
			}
			return PipelinesFetchResultsMsg{Data: displays}
		}

	case PipelinesFetchResultsMsg:
		return m, tea.Batch(
			PipelinesListSetCmd(msg.Data),
			LoadingStopCmd(),
		)

	case PipelinesListSet:
		m.PipelinesList = msg.PipelinesList
		return m, nil

	case PipelineEntriesFetchMsg:
		mgr := m.PluginManager
		if mgr == nil {
			return m, func() tea.Msg {
				return PipelineEntriesFetchResultsMsg{Entries: []PipelineEntryDisplay{}}
			}
		}
		key := msg.Key
		return m, func() tea.Msg {
			results := mgr.DryRunAllPipelines()
			var matchedEntries []PipelineEntryDisplay
			for _, r := range results {
				rKey := r.Table + "." + r.Phase + "_" + r.Operation
				if rKey == key {
					for _, e := range r.Entries {
						matchedEntries = append(matchedEntries, PipelineEntryDisplay{
							PipelineID: e.PipelineID,
							PluginName: e.PluginName,
							Handler:    e.Handler,
							Priority:   e.Priority,
							Enabled:    e.Enabled,
						})
					}
					break
				}
			}
			return PipelineEntriesFetchResultsMsg{Entries: matchedEntries}
		}

	case PipelineEntriesFetchResultsMsg:
		return m, PipelineEntriesSetCmd(msg.Entries)

	case PipelineEntriesSet:
		m.PipelineEntries = msg.PipelineEntries
		return m, nil

	case PluginSyncCapabilitiesRequestMsg:
		mgr := m.PluginManager
		name := msg.Name
		adminUser := m.AdminUsername
		if mgr == nil {
			return m, func() tea.Msg {
				return PluginSyncCapabilitiesResultMsg{Name: name, Err: fmt.Errorf("plugin manager not available")}
			}
		}
		return m, func() tea.Msg {
			ctx := context.Background()
			err := mgr.SyncCapabilities(ctx, name, adminUser)
			return PluginSyncCapabilitiesResultMsg{Name: name, Err: err}
		}

	case PluginSyncCapabilitiesResultMsg:
		if msg.Err != nil {
			return m, ShowDialog("Sync Failed", fmt.Sprintf("Failed to sync capabilities for %q: %s", msg.Name, msg.Err.Error()), false)
		}
		return m, tea.Batch(
			ShowDialog("Sync Complete", fmt.Sprintf("Capabilities synced for plugin %q", msg.Name), false),
			PluginsFetchCmd(),
		)

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
