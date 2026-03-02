package tui

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

// systemDatatypeTypes are datatype types that should not appear in the child picker.
// Matches the admin panel's SYSTEM_TYPES in block-editor-src/cache.js.
var systemDatatypeTypes = map[string]bool{
	"_root":        true,
	"_nested_root": true,
	"_system_log":  true,
	"_reference":   true,
}

// filterChildDatatypes filters datatypes using the same 3-category logic as the
// admin panel block editor (cache.js fetchDatatypesGrouped):
//  1. Descendants of the root datatype
//  2. Children of _collection datatypes
//  3. _global datatypes and their children
//
// System types (_root, _nested_root, _system_log, _reference) are excluded.
func filterChildDatatypes(all []db.Datatypes, rootDatatypeID types.DatatypeID) []db.Datatypes {
	// Build lookup maps
	byID := make(map[types.DatatypeID]db.Datatypes, len(all))
	childrenOf := make(map[types.DatatypeID][]db.Datatypes)
	for _, dt := range all {
		byID[dt.DatatypeID] = dt
		pid := dt.ParentID
		if pid.Valid {
			childrenOf[pid.ID] = append(childrenOf[pid.ID], dt)
		}
	}

	// Recursively collect descendants, excluding system types
	var collectDescendants func(parentID types.DatatypeID) []db.Datatypes
	collectDescendants = func(parentID types.DatatypeID) []db.Datatypes {
		var result []db.Datatypes
		for _, kid := range childrenOf[parentID] {
			if systemDatatypeTypes[kid.Type] {
				continue
			}
			result = append(result, kid)
			result = append(result, collectDescendants(kid.DatatypeID)...)
		}
		return result
	}

	seen := make(map[types.DatatypeID]bool)
	var filtered []db.Datatypes
	addUnique := func(dts []db.Datatypes) {
		for _, dt := range dts {
			if !seen[dt.DatatypeID] {
				seen[dt.DatatypeID] = true
				filtered = append(filtered, dt)
			}
		}
	}

	// Category 1: Descendants of root datatype
	if _, ok := byID[rootDatatypeID]; ok {
		addUnique(collectDescendants(rootDatatypeID))
	}

	// Category 2: Children of _collection datatypes
	for _, dt := range all {
		if dt.Type == "_collection" {
			addUnique(collectDescendants(dt.DatatypeID))
		}
	}

	// Category 3: _global datatypes and their children
	for _, dt := range all {
		if dt.Type == "_global" && !systemDatatypeTypes[dt.Type] {
			if !seen[dt.DatatypeID] {
				seen[dt.DatatypeID] = true
				filtered = append(filtered, dt)
			}
			addUnique(collectDescendants(dt.DatatypeID))
		}
	}

	return filtered
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
		datatypeMenuLabels := m.BuildDatatypeMenu(msg.Data)
		m.Logger.Finfo("newMenu", datatypeMenuLabels)

		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintln(datatypeMenuLabels)),
			DatatypeMenuSetCmd(datatypeMenuLabels),
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
		rootDatatypeID := msg.ParentDatatypeID
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
			filtered := filterChildDatatypes(*all, rootDatatypeID)
			if len(filtered) == 0 {
				return ActionResultMsg{
					Title:   "No Datatypes",
					Message: "No eligible child datatypes for this root type.",
				}
			}
			return ShowChildDatatypeDialogMsg{
				ChildDatatypes: filtered,
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
		return m, NewStateUpdate()

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
		return m, NewStateUpdate()

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

	case WebhooksFetchMsg:
		driver := m.DB
		return m, func() tea.Msg {
			list, err := driver.ListWebhooks()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			data := make([]db.Webhook, 0)
			if list != nil {
				data = *list
			}
			return WebhooksFetchResultsMsg{Data: data}
		}

	case WebhooksFetchResultsMsg:
		m.WebhooksList = msg.Data
		return m, LoadingStopCmd()

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
