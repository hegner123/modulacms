# cli

The cli package implements the Terminal User Interface for ModulaCMS using Charmbracelet Bubbletea following the Elm Architecture pattern. It provides interactive management of the CMS through SSH or direct terminal access.

## Overview

This package implements a full-featured TUI for ModulaCMS content management. The architecture follows the Bubbletea pattern with Model, Update, and View functions. The package handles user provisioning, database operations, content tree management, form dialogs, and CMS administration.

## Core Model

### Model

Main application state struct following the Bubbletea tea.Model interface. Contains all state for the TUI including database driver, configuration, UI state, dialog models, content tree, and user session data.

Fields include DB driver, config, logger, status, dimensions, paginator, page state, form state, table state, focus tracking, dialog models, content tree root, route and datatype lists, media list, users list, and SSH user provisioning state.

### ModelInterface

Defines interface for interacting with CLI model. Provides GetConfig, GetRoot, SetRoot, and SetError methods.

### FocusKey

Type alias for focus tracking. Values are PAGEFOCUS, TABLEFOCUS, FORMFOCUS, DIALOGFOCUS.

### ApplicationState

Type alias for application status. Values are OK, EDITING, DELETING, WARN, ERROR.

### FilePickerPurpose

Purpose enumeration for file picker. Values are FILEPICKER_MEDIA for media upload and FILEPICKER_RESTORE for backup restore.

## Initialization and Lifecycle

### CliRun

Runs the CLI program with Bubbletea tea.Program. Takes model pointer, creates program with alt screen, runs program, and returns continuation flag CliContinue.

### InitialModel

Creates initial Model with provided verbose flag, config, database driver, and logger. Sets up paginator, spinner, viewport, page map, form model, table model, focus state, and panel focus. Returns Model and command to fetch database tables.

### ModelPostInit

Returns batch command for post-initialization tasks. Logs test menu init message and sets page menu to homepage menu.

### Init

Implements Bubbletea Init method. Returns spinner tick command only if in loading state, otherwise returns nil.

### Update

Main update function implementing Bubbletea Update method. Routes messages through specialized update handlers in order: UpdateProvisioning, UpdateLog, UpdateTea, UpdateState, UpdateNavigation, UpdateFetch, UpdateAdminFetch, UpdateForm, UpdateDialog, UpdateDatabase, UpdateCms, UpdateAdminCms. Handles file picker, dialog, form dialog, user dialog, and database dialog routing when active.

### View

Main view function implementing Bubbletea View method. Shows user provisioning form if needed. Routes to panel layout for CMS pages. Routes to page-specific views for HOMEPAGE, DATABASEPAGE, BUCKETPAGE, OAUTHPAGE, CONFIGPAGE, TABLEPAGE, CREATEPAGE, READPAGE, READSINGLEPAGE, UPDATEPAGE, UPDATEFORMPAGE, DELETEPAGE, DEVELOPMENT, DYNAMICPAGE, DATATYPE, ACTIONSPAGE. Returns overlay views for file picker, dialogs, form dialogs, content form dialogs, user dialogs, and database dialogs when active.

## Page Management

### Page

Represents a navigable page with index and label.

### PageIndex

Enumeration of page types. Values include HOMEPAGE, CMSPAGE, ADMINCMSPAGE, DATABASEPAGE, BUCKETPAGE, OAUTHPAGE, CONFIGPAGE, TABLEPAGE, CREATEPAGE, READPAGE, UPDATEPAGE, DELETEPAGE, UPDATEFORMPAGE, READSINGLEPAGE, DYNAMICPAGE, DATATYPES, DATATYPESMENU, FIELDS, DEVELOPMENT, DATATYPE, USERSADMIN, MEDIA, CONTENT, PICKCONTENT, EDITCONTENT, ACTIONSPAGE, ROUTES, ADMINROUTES, ADMINDATATYPES, ADMINCONTENT.

### NewPage

Creates new Page with given index and label.

### NewDatatypePage

Creates Page with DATATYPE index and provided label.

### NewDynamicPage

Creates Page with DYNAMICPAGE index and provided label.

### NewPickContentPage

Creates Page with PICKCONTENT index and provided label.

### InitPages

Initializes map of all page definitions keyed by PageIndex.

### PageHistory

Stores cursor position, page, and menu for navigation history stack.

## Menu Initialization

### HomepageMenuInit

Returns slice of Pages for homepage menu including CMS, Admin CMS, Database, Bucket, OAuth, Config, Actions.

### CmsMenuInit

Returns slice of Pages for CMS menu including Content, Datatypes, Routes, Media, Users.

### AdminCmsMenuInit

Returns slice of Pages for Admin CMS menu including Admin Content, Admin Datatypes, Admin Routes.

### ContentMenuInit

Returns slice of Pages for content menu including Routes.

### DatabaseMenuInit

Returns slice of Pages for database menu including Create, Read, Update, Delete.

### BuildDatatypeMenu

Builds menu of Pages from slice of Datatypes, filtering for ROOT type.

### BuildContentDataMenu

Builds menu of Pages from slice of ContentData filtered by root ContentID.

## History Management

### PushHistory

Pushes PageHistory entry onto history stack.

### PopHistory

Pops and returns PageHistory from history stack, returns nil if empty.

### Peek

Returns pointer to last PageHistory entry without removing, returns nil and false if empty.

## Form State

### FormModel

Encapsulates form-related state extracted from Model. Groups form lifecycle fields including Form, FormLen, FormMap, FormValues, FormSubmit, FormGroups, FormFields, FormOptions.

### NewFormModel

Creates new FormModel with safe defaults. FormMap initialized as empty slice, FormSubmit is false.

### FormOptionsMap

Type alias for map of string to slice of huh.Option strings. Used for form dropdown options.

## Table State

### TableModel

Encapsulates table-related state for database viewing. Groups fields including Table name, Headers, Rows, Columns, ColumnTypes, Selected row map, and current Row pointer.

### NewTableModel

Creates new TableModel with safe defaults. Headers and Rows initialized as empty slices, Selected initialized as empty map.

## Dialogs

### DialogAction

Enumeration of dialog action types. Values include DIALOGGENERIC, DIALOGDELETE, DIALOGACTIONCONFIRM, DIALOGINITCONTENT, DIALOGQUITCONFIRM, DIALOGDELETECONTENT, DIALOGDELETEDATATYPE, DIALOGDELETEFIELD, DIALOGDELETEROUTE, DIALOGDELETEMEDIA, DIALOGDELETEUSER, DIALOGDELETECONTENTFIELD, DIALOGDELETEADMINROUTE, DIALOGDELETEADMINDATATYPE, DIALOGDELETEADMINFIELD, DIALOGDELETEADMINCONTENT, DIALOGBACKUPRESTORE.

### DialogModel

Represents modal dialog overlay with title, message, width, height, button text, cancel visibility, ready state, action type, action index, help model, and style fields.

### NewDialog

Creates new DialogModel with given title, message, showCancel flag, and action type. Initializes with default dimensions, button text, help model, and styles.

### SetSize

Sets dialog width and height.

### SetButtons

Sets dialog OK and Cancel button text.

### Update

Handles user input for DialogModel. Routes to ToggleControls for action-based dialogs, dismisses on enter or esc for generic dialogs.

### ToggleControls

Handles tab navigation and enter/esc for action dialogs. Returns DialogAcceptMsg or DialogCancelMsg based on focus and key.

### Render

Renders DialogModel with title, message, and button bar. Applies border style. Returns string for overlay positioning.

### DialogOverlay

Positions dialog over existing content using layer compositing. Centers dialog in window, returns composited string.

## Form Dialogs

### FormDialogAction

Enumeration of form dialog types. Values include FORMDIALOGCREATEDATATYPE, FORMDIALOGEDITDATATYPE, FORMDIALOGCREATEFIELD, FORMDIALOGEDITFIELD, FORMDIALOGCREATEROUTE, FORMDIALOGEDITROUTE, FORMDIALOGCREATEROUTEWITHCONTENT, FORMDIALOGINITIALIZEROUTECONTENT, FORMDIALOGCHILDDATATYPE, FORMDIALOGCREATECONTENT, FORMDIALOGEDITCONTENT, FORMDIALOGMOVECONTENT, FORMDIALOGCREATEUSER, FORMDIALOGEDITUSER, FORMDIALOGEDIITSINGLEFIELD, FORMDIALOGADDCONTENTFIELD, FORMDIALOGCREATEADMINROUTE, FORMDIALOGEDITADMINROUTE, FORMDIALOGCREATEADMINDATATYPE, FORMDIALOGEDITADMINDATATYPE, FORMDIALOGCREATEADMINFIELD, FORMDIALOGEDITADMINFIELD, FORMDIALOGDBINSERT, FORMDIALOGDBUPDATE.

### FormDialogModel

Form dialog with text inputs, type selector carousel, parent selector carousel, and button navigation. Contains title, width, action, entity ID, label input, type input, type options, parent options, and focus index.

### NewFormDialog

Creates form dialog for datatype creation with label input, type input, and parent selector. Returns FormDialogModel.

### NewFieldFormDialog

Creates form dialog for field creation with label input and type selector from registry. No parent selector.

### NewRouteFormDialog

Creates form dialog for route creation with title input and slug input using LabelInput and TypeInput fields.

### NewEditDatatypeDialog

Creates form dialog for editing datatype with pre-populated values. Excludes self from parent options to prevent circular reference.

### NewEditFieldDialog

Creates form dialog for editing field with pre-populated values. Type index set from registry.

### NewEditRouteDialog

Creates form dialog for editing route with pre-populated title and slug values.

### NewRouteWithContentDialog

Creates form dialog for creating route with initial content. Stores datatypeID in EntityID field.

### NewChildDatatypeDialog

Creates selection dialog for child datatypes. Renders as vertical list.

### NewMoveContentDialog

Creates selection dialog for move content target. Renders as vertical list.

### HasParentSelector

Returns true if dialog has parent options.

### HasTypeSelector

Returns true if dialog has type options carousel.

### Update

Handles input for FormDialogModel. Special vertical list handling for child datatype and move content. Tab navigation, arrow keys for selectors, enter to confirm, esc to cancel.

### Render

Renders form dialog with title, fields, selectors, and buttons. Special rendering for child datatype and move content as vertical lists.

### FormDialogOverlay

Positions form dialog over existing content. Centers dialog, returns composited string.

## Content Form Dialogs

### ContentFieldInput

Represents single field input in content form with FieldID, Label, Type, and textinput.Model.

### ContentFormDialogModel

Form dialog with dynamic content fields. Contains title, width, action, datatype ID, route ID, content ID for edit mode, parent ID for child creation, field inputs, and focus index.

### NewContentFormDialog

Creates content form dialog with dynamic fields from slice of Fields. Initializes text inputs for each field.

### NewContentFormDialogWithParent

Creates content form for child content creation with parent ID set.

### NewEditContentFormDialog

Creates content form dialog pre-populated with existing values for editing.

### ButtonCancelIndex

Returns focus index of Cancel button equal to field count.

### ButtonConfirmIndex

Returns focus index of Confirm button equal to field count plus one.

### Update

Handles input for ContentFormDialogModel. Tab navigation, enter to confirm or move to next field, esc to cancel.

### Render

Renders content form dialog with title, all field inputs, and buttons.

### ContentFormDialogOverlay

Positions content form dialog over existing content. Centers dialog, returns composited string.

### ExistingContentField

Represents field with current value for editing. Contains ContentFieldID, FieldID, Label, Type, Value.

## User Form Dialogs

### UserFormDialogModel

Form dialog for user CRUD operations. Contains title, width, action, entity ID, username input, name input, email input, role input, and focus index.

### NewUserFormDialog

Creates user form dialog for creating new user with empty inputs.

### NewEditUserFormDialog

Creates user form dialog pre-populated with existing user values.

### Update

Handles input for UserFormDialogModel. Tab navigation, enter to confirm or move to next field, esc to cancel.

### Render

Renders user form dialog with title, all user fields, and buttons.

### UserFormDialogOverlay

Positions user form dialog over existing content. Centers dialog, returns composited string.

## Database Commands

### DatabaseCMD

Type alias for database operation types. Values are INSERT, SELECT, UPDATE, DELETE, BATCH.

### ForeignKeyReference

Stores foreign key relationship with From, Table, and Column fields.

### GetTablesCMD

Command that fetches database tables and returns TablesSet message with table labels.

### DatabaseInsert

Executes INSERT operation using query builder. Converts column values to map, returns DbResultCmd or error log.

### DatabaseUpdate

Executes UPDATE operation using query builder. Updates row by ID, returns DbResultCmd or error log.

### DatabaseGet

Executes SELECT for single row by ID. Parses result, returns DatabaseGetRowsCmd.

### DatabaseList

Executes SELECT for all rows. Parses result, returns DatabaseListRowsCmd.

### DatabaseFilteredList

Executes SELECT with WHERE filter. Parses result, returns DatabaseListRowsCmd.

### DatabaseDelete

Executes DELETE by current row ID. Returns DbResultCmd or error log.

## Content Operations

### CreateContentWithFields

Performs atomic content creation using typed DbDriver methods. Creates ContentData first, then creates associated ContentFields using returned ID. Returns ContentCreatedMsg or ContentCreatedWithErrorsMsg.

### HandleCreateContentFromDialog

Creates content from dialog values with parent support. Returns ContentCreatedFromDialogMsg or error.

### HandleFetchContentForEdit

Fetches existing content fields and shows edit dialog. Returns ShowEditContentFormDialogMsg with existing fields in creation order.

### HandleUpdateContentFromDialog

Updates existing content fields from dialog values. Updates existing fields or creates new fields added to datatype after content creation.

### HandleDeleteContent

Deletes content and updates tree structure. Detaches from siblings, updates parent first_child if needed, cascades to content_fields.

### HandleMoveContent

Detaches source from current position and attaches as last child of target node. Updates all affected sibling and parent pointers.

### ReloadContentTree

Fetches tree data from database and loads into Root. Returns TreeLoadedMsg with stats and root node.

## Content Field Operations

### LoadContentFieldsCmd

Fetches content fields for specific content node. Resolves field labels from datatype field definitions via junction table. Returns fields in sort_order from junction table.

### HandleEditSingleField

Updates one content field value. Returns ContentFieldUpdatedMsg or error.

### HandleDeleteContentField

Deletes content field record. Returns ContentFieldDeletedMsg or error.

### HandleAddContentField

Creates new content field record for unpopulated field. Returns ContentFieldAddedMsg or error.

### HandleReorderField

Swaps sort_order between two junction records. Updates both records, returns FieldReorderedMsg or error.

## Datatype and Field Operations

### HandleDeleteDatatype

Deletes datatype and its junction records. Checks for child datatypes first, deletes all junction records, then deletes datatype.

### HandleDeleteField

Deletes field and its junction record. Finds junction records linking field to datatype, deletes junction, then deletes field.

## Route Operations

### HandleDeleteRoute

Deletes route record. Returns RouteDeletedMsg or error.

## Media Operations

### HandleMediaUpload

Runs media upload pipeline asynchronously. Creates placeholder DB record, creates temp directory, runs optimize and S3 upload, updates DB with final data.

### HandleDeleteMedia

Deletes media item. Returns MediaDeletedMsg or error.

## Sibling Reordering

### ReorderSiblingCmd

Creates command to reorder content among siblings. Direction is up or down.

### HandleReorderSibling

Swaps node with prev or next sibling in linked list. Updates all affected prev, next, and parent first_child pointers.

### CopyContentCmd

Creates command to copy content node as new sibling.

### HandleCopyContent

Duplicates content node and its fields as new sibling. Creates new ContentData with same values, copies all fields, inserts after source in sibling list.

### TogglePublishCmd

Creates command to toggle content publish status.

### HandleTogglePublish

Toggles content node between draft and published status.

### ArchiveContentCmd

Creates command to toggle content archive status.

### HandleArchiveContent

Toggles content node between archived and draft status. If archived, reverts to draft. Otherwise, sets to archived.

## User Operations

### HandleCreateUserFromDialog

Processes user creation request. Validates email, creates user record, returns UserCreatedFromDialogMsg or error.

### HandleUpdateUserFromDialog

Processes user update request. Fetches existing user to preserve hash, updates record, returns UserUpdatedFromDialogMsg or error.

### HandleDeleteUser

Deletes user record. Returns UserDeletedMsg or error.

## Actions System

### ActionParams

Groups context needed by action commands including Config, UserID, SSH fingerprint, key type, and public key.

### ActionItem

Describes single action with label, description, and destructive flag.

### ActionsMenu

Returns ordered list of ActionItem structs. Index matches cursor position on Actions page. Items include DB Init, DB Wipe, DB Wipe and Redeploy, DB Reset, DB Export, Generate Certs, Check for Updates, Validate Config, Generate API Token, Register SSH Key, Create Backup, Restore Backup.

### ActionsMenuLabels

Returns slice of label strings from ActionsMenu for menu rendering.

### ActionResultMsg

Message returned by action commands with title, message, error flag, and optional width override.

### ActionConfirmMsg

Sent when destructive action needs confirmation. Contains action index.

### ActionConfirmedMsg

Sent when user confirms destructive action. Contains action index.

### RunActionCmd

Executes non-destructive action by index. Routes to appropriate run function.

### RunDestructiveActionCmd

Executes destructive action by index after confirmation. Routes to appropriate run function.

## Message Types

### LogModelMsg

Message for logging model state with include and exclude field filters.

### ClearScreen

Message to clear screen.

### ReadyTrue

Message to set ready state true.

### ReadyFalse

Message to set ready state false.

### TitleFontNext

Message to cycle title font forward.

### TitleFontPrevious

Message to cycle title font backward.

### TablesFetch

Message to trigger fetching database tables.

### TablesSet

Message to set tables list with slice of table names.

### LoadingTrue

Message to set loading state true.

### LoadingFalse

Message to set loading state false.

### CursorUp

Message to move cursor up.

### CursorDown

Message to move cursor down.

### CursorReset

Message to reset cursor to zero.

### CursorSet

Message to set cursor to specific index.

### UpdateMaxCursorMsg

Message to update max cursor value with CursorMax field.

### PageModNext

Message to move to next page module.

### PageModPrevious

Message to move to previous page module.

### PageSet

Message to set current page with Page field.

### UpdatePagination

Message to trigger pagination update.

### TableSet

Message to set current table with Table name string.

### SetPageContent

Message to set page content with Content string.

### SetViewportContent

Message to set viewport content with Content string.

### FocusSet

Message to set focus with FocusKey field.

### FormCreate

Message to create form with FormType field.

### FormSet

Message to set form with Form and Values fields.

### FormValuesSet

Message to set form values with Values field.

### FormAborted

Message when form is aborted with Action and Table fields.

### FormSubmitMsg

Message when form is submitted.

### FormCompletedMsg

Message when form is completed with optional DestinationPage.

### FormActionMsg

Message to trigger form action with Action, Table, Columns, Values fields.

### FormCancelMsg

Message when form is cancelled.

### FormOptionsSet

Message to set form options with Options map.

### FormInitOptionsMsg

Message to initialize form options with Form and Table names.

### HistoryPop

Message to pop history stack.

### HistoryPush

Message to push PageHistory onto stack.

### NavigateToPage

Message to navigate to page with Page and Menu fields.

### NavigateToDatabaseCreate

Message to navigate to database create page.

### SelectTable

Message to select table with Table name string.

### DatabaseDeleteEntry

Message to delete database entry with Id and Table fields.

### DatabaseInsertEntry

Message to insert database entry with Table, Columns, Values fields.

### DatabaseUpdateEntry

Message to update database entry with Table, RowID, Values map.

### DatabaseGetMsg

Message to get single database row with Source, Table, ID fields.

### DatabaseListFilteredMsg

Message to list filtered database rows with Source, Table, Columns, WhereColumn, Value fields.

### DatabaseListMsg

Message to list all database rows with Source and Table fields.

### DatabaseGetRowMsg

Message carrying single database row result with Source, Table, Rows fields.

### DatabaseListFilteredRowsMsg

Message carrying filtered database rows with Source, Table, Rows fields.

### DatabaseListRowsMsg

Message carrying all database rows with Source, Table, Rows fields.

### ColumnsFetched

Message when columns are fetched with Columns and ColumnTypes fields.

### ColumnsSet

Message to set columns with Columns field.

### ColumnTypesSet

Message to set column types with ColumnTypes field.

### HeadersSet

Message to set headers with Headers slice.

### RowsSet

Message to set rows with Rows slice of string slices.

### CursorMaxSet

Message to set cursor max with CursorMax field.

### PaginatorUpdate

Message to update paginator with PerPage and TotalPages fields.

### FormLenSet

Message to set form length with FormLen field.

### FormMapSet

Message to set form map with FormMap slice.

### ErrorSet

Message to set error with Err field.

### StatusSet

Message to set application status with Status field.

### DialogSet

Message to set dialog with Dialog pointer.

### DialogActiveSet

Message to set dialog active state with DialogActive bool.

### RootSet

Message to set root with Root field.

### DatatypeMenuSet

Message to set datatype menu with DatatypeMenu slice.

### PageMenuSet

Message to set page menu with PageMenu slice.

### DialogReadyOKSet

Message to set dialog ready OK state with Ready bool.

### DbResMsg

Message carrying database result with Result and Table fields.

### DbErrMsg

Message carrying database error with Error field.

### ReadMsg

Message carrying read result with Result, Error, RType fields.

### DatatypesFetchMsg

Message to trigger fetching datatypes.

### DatatypesFetchResultsMsg

Message carrying datatypes with Data slice.

### DataFetchErrorMsg

Message carrying data fetch error with Error field.

### LogMsg

Message carrying log message with Message string.

### FetchHeadersRows

Message to fetch table headers and rows with Config, Table, Page fields.

### TableHeadersRowsFetchedMsg

Message carrying fetched headers and rows with Headers, Rows, Page fields.

### GetColumns

Message to get columns with Config and Table fields.

### BuildTreeFromRows

Message to build tree from rows with Rows slice.

### GetFullTreeResMsg

Message carrying full tree rows with Rows slice.

### DatabaseTreeMsg

Message to trigger database tree operation.

### CmsDefineDatatypeLoadMsg

Message to load define datatype form.

### CmsDefineDatatypeReadyMsg

Message when define datatype is ready.

### CmsBuildDefineDatatypeFormMsg

Message to build define datatype form.

### CmsDefineDatatypeFormMsg

Message to show define datatype form.

### CmsEditDatatypeLoadMsg

Message to load edit datatype with Datatype field.

### CmsEditDatatypeFormMsg

Message to show edit datatype form with Datatype field.

### DatatypeUpdateSaveMsg

Message to save datatype update with DatatypeID, Parent, Label, Type fields.

### DatatypeUpdatedMsg

Message when datatype is updated with DatatypeID and Label fields.

### DatatypeUpdateFailedMsg

Message when datatype update fails with Error field.

### CmsGetDatatypeParentOptionsMsg

Message to get datatype parent options with Admin bool flag.

### CmsAddNewContentDataMsg

Message to add new content data with Datatype ID field.

### CmsAddNewContentFieldsMsg

Message to add new content fields with Datatype int64 field.

### ContentCreatedMsg

Message when content is created with ContentDataID, RouteID, FieldCount fields.

### ContentCreatedWithErrorsMsg

Message when content created with errors. Contains ContentDataID, RouteID, CreatedFields count, FailedFields slice.

### TreeLoadedMsg

Message when tree is loaded with RouteID, Stats, RootNode fields.

### BuildContentFormMsg

Message to build content form with DatatypeID and RouteID fields.

### RoutesFetchMsg

Message to trigger fetching routes.

### RoutesFetchResultsMsg

Message carrying routes with Data slice.

### RouteSelectedMsg

Message when route is selected with Route field.

### RoutesSet

Message to set routes with Routes slice.

### RootDatatypesFetchMsg

Message to trigger fetching root datatypes.

### RootDatatypesFetchResultsMsg

Message carrying root datatypes with Data slice.

### RootDatatypesSet

Message to set root datatypes with RootDatatypes slice.

### AllDatatypesFetchMsg

Message to trigger fetching all datatypes.

### AllDatatypesFetchResultsMsg

Message carrying all datatypes with Data slice.

### AllDatatypesSet

Message to set all datatypes with AllDatatypes slice.

### DatatypeFieldsFetchMsg

Message to fetch datatype fields with DatatypeID field.

### DatatypeFieldsFetchResultsMsg

Message carrying datatype fields with Fields slice.

### DatatypeFieldsSet

Message to set datatype fields with Fields slice.

### RoutesByDatatypeFetchMsg

Message to fetch routes by datatype with DatatypeID field.

### SelectedDatatypeSet

Message to set selected datatype with DatatypeID field.

### MediaFetchMsg

Message to trigger fetching media.

### MediaFetchResultsMsg

Message carrying media with Data slice.

### MediaListSet

Message to set media list with MediaList slice.

### RootContentSummaryFetchMsg

Message to trigger fetching root content summary.

### RootContentSummaryFetchResultsMsg

Message carrying root content summary with Data slice.

### RootContentSummarySet

Message to set root content summary with RootContentSummary slice.

### MediaUploadStartMsg

Message to trigger async media upload with FilePath field.

### MediaUploadedMsg

Message when media is uploaded with Name field.

### ReorderSiblingRequestMsg

Message to request sibling reorder with ContentID, RouteID, Direction fields.

### ContentReorderedMsg

Message when content is reordered with ContentID, RouteID, Direction fields.

### CopyContentRequestMsg

Message to request content copy with SourceContentID and RouteID fields.

### ContentCopiedMsg

Message when content is copied with SourceContentID, NewContentID, RouteID, FieldCount fields.

### TogglePublishRequestMsg

Message to toggle publish status with ContentID and RouteID fields.

### ContentPublishToggledMsg

Message when publish is toggled with ContentID, RouteID, NewStatus fields.

### ArchiveContentRequestMsg

Message to archive content with ContentID and RouteID fields.

### ContentArchivedMsg

Message when content is archived with ContentID, RouteID, NewStatus fields.

### PanelFocusReset

Message to reset panel focus.

### UsersFetchMsg

Message to trigger fetching users.

### UsersFetchResultsMsg

Message carrying users with Data slice.

### UsersListSet

Message to set users list with UsersList slice.

### OpenFilePickerForRestoreMsg

Message to open file picker for restore operation.

### RestoreBackupFromPathMsg

Message to restore backup from path with Path field.

### BackupRestoreCompleteMsg

Message when backup restore is complete with Path field.

### BuildTreeFromRouteMsg

Message to build tree from route with RouteID int64 field.

### DialogAcceptMsg

Message when dialog is accepted with Action field.

### DialogCancelMsg

Message when dialog is cancelled.

### DialogReadyOK

Message when dialog ready OK is set.

### ShowDialogMsg

Message to show dialog with Title, Message, ShowCancel fields.

### ShowQuitConfirmDialogMsg

Message to show quit confirmation dialog.

### ShowDeleteContentDialogMsg

Message to show delete content dialog with ContentID, ContentName, HasChildren fields.

### DeleteContentRequestMsg

Message to delete content with ContentID and RouteID fields.

### ContentDeletedMsg

Message when content is deleted with ContentID and RouteID fields.

### FormDialogAcceptMsg

Message when form dialog is accepted with Action, EntityID, Label, Type, ParentID fields.

### FormDialogCancelMsg

Message when form dialog is cancelled.

### ShowFormDialogMsg

Message to show form dialog with Action, Title, Parents fields.

### ShowFieldFormDialogMsg

Message to show field form dialog with Action and Title fields.

### ShowRouteFormDialogMsg

Message to show route form dialog with Action and Title fields.

### FormDialogSetMsg

Message to set form dialog with Dialog pointer.

### FormDialogActiveSetMsg

Message to set form dialog active state with Active bool.

### ShowEditDatatypeDialogMsg

Message to show edit datatype dialog with Datatype and Parents fields.

### ShowEditFieldDialogMsg

Message to show edit field dialog with Field field.

### ShowEditRouteDialogMsg

Message to show edit route dialog with Route field.

### ShowCreateRouteWithContentDialogMsg

Message to show create route with content dialog with DatatypeID field.

### ShowInitializeRouteContentDialogMsg

Message to show initialize route content dialog with Route and DatatypeID fields.

### ShowChildDatatypeDialogMsg

Message to show child datatype dialog with ParentDatatypeID, RouteID, ChildDatatypes fields.

### FetchChildDatatypesMsg

Message to fetch child datatypes with ParentDatatypeID and RouteID fields.

### ChildDatatypeSelectedMsg

Message when child datatype is selected with DatatypeID and RouteID fields.

### ShowMoveContentDialogMsg

Message to show move content dialog with SourceNode, RouteID, ValidTargets fields.

### MoveContentRequestMsg

Message to move content with SourceContentID, TargetContentID, RouteID fields.

### ContentMovedMsg

Message when content is moved with SourceContentID, TargetContentID, RouteID fields.

### LoadContentFieldsMsg

Message carrying content fields for right panel with Fields slice.

### ContentFieldUpdatedMsg

Message when content field is updated with ContentID, DatatypeID, RouteID fields.

### ContentFieldDeletedMsg

Message when content field is deleted with ContentID, DatatypeID, RouteID fields.

### ContentFieldAddedMsg

Message when content field is added with ContentID, DatatypeID, RouteID fields.

### FieldReorderedMsg

Message when field is reordered with DatatypeID, ContentID, RouteID, Direction fields.

### ContentFormDialogAcceptMsg

Message when content form dialog is accepted with Action, DatatypeID, RouteID, ContentID, ParentID, FieldValues fields.

### ContentFormDialogCancelMsg

Message when content form dialog is cancelled.

### ShowContentFormDialogMsg

Message to show content form dialog with Action, Title, DatatypeID, RouteID, ParentID, Fields fields.

### ContentFormDialogSetMsg

Message to set content form dialog with Dialog pointer.

### ContentFormDialogActiveSetMsg

Message to set content form dialog active state with Active bool.

### CreateContentFromDialogRequestMsg

Message to create content from dialog with DatatypeID, RouteID, ParentID, FieldValues fields.

### ContentCreatedFromDialogMsg

Message when content created from dialog with ContentID, DatatypeID, RouteID, FieldCount fields.

### FetchContentFieldsMsg

Message to fetch content fields with DatatypeID, RouteID, ParentID, Title fields.

### ShowEditContentFormDialogMsg

Message to show edit content form dialog with Title, ContentID, DatatypeID, RouteID, ExistingFields fields.

### FetchContentForEditMsg

Message to fetch content for edit with ContentID, DatatypeID, RouteID, Title fields.

### UpdateContentFromDialogRequestMsg

Message to update content from dialog with ContentID, DatatypeID, RouteID, FieldValues fields.

### ContentUpdatedFromDialogMsg

Message when content updated from dialog with ContentID, DatatypeID, RouteID, UpdatedCount fields.

### UserFormDialogAcceptMsg

Message when user form dialog is accepted with Action, EntityID, Username, Name, Email, Role fields.

### UserFormDialogCancelMsg

Message when user form dialog is cancelled.

### ShowUserFormDialogMsg

Message to show user form dialog with Title field.

### ShowEditUserDialogMsg

Message to show edit user dialog with User field.

### UserFormDialogSetMsg

Message to set user form dialog with Dialog pointer.

### UserFormDialogActiveSetMsg

Message to set user form dialog active state with Active bool.

### ShowEditSingleFieldDialogMsg

Message to show edit single field dialog with Field, ContentID, RouteID, DatatypeID fields.

### EditSingleFieldAcceptMsg

Message when edit single field is accepted with ContentFieldID, ContentID, FieldID, NewValue, RouteID, DatatypeID fields.

### ShowAddContentFieldDialogMsg

Message to show add content field dialog with Options, ContentID, RouteID, DatatypeID fields.

### DeleteContentFieldContext

Stores context for deleting content field with ContentFieldID, ContentID, RouteID, DatatypeID fields.

### ShowDeleteContentFieldDialogMsg

Message to show delete content field dialog with Field, ContentID, RouteID, DatatypeID fields.

### DeleteDatatypeRequestMsg

Message to delete datatype with DatatypeID field.

### DatatypeDeletedMsg

Message when datatype is deleted with DatatypeID field.

### DeleteFieldRequestMsg

Message to delete field with FieldID and DatatypeID fields.

### FieldDeletedMsg

Message when field is deleted with FieldID and DatatypeID fields.

### DeleteRouteRequestMsg

Message to delete route with RouteID field.

### RouteDeletedMsg

Message when route is deleted with RouteID field.

### DeleteMediaRequestMsg

Message to delete media with MediaID field.

### MediaDeletedMsg

Message when media is deleted with MediaID field.

### DeleteUserRequestMsg

Message to delete user with UserID field.

### UserDeletedMsg

Message when user is deleted with UserID field.

### CreateUserFromDialogRequestMsg

Message to create user from dialog with Username, Name, Email, Role fields.

### UserCreatedFromDialogMsg

Message when user created from dialog with UserID and Username fields.

### UpdateUserFromDialogRequestMsg

Message to update user from dialog with UserID, Username, Name, Email, Role fields.

### UserUpdatedFromDialogMsg

Message when user updated from dialog with UserID and Username fields.

### AdminContentFieldDisplay

Represents admin content field for right panel display. Contains ContentFieldID, DatatypeFieldID, FieldID, Label, Type, Value fields.

### ContentFieldDisplay

Represents content field for right panel display. Contains ContentFieldID, DatatypeFieldID, FieldID, Label, Type, Value fields.

## Command Constructors

The package exports over 100 command constructor functions following the pattern FunctionNameCmd that return tea.Cmd. These wrap operations in commands for the Bubbletea message passing pattern. Key categories include logging, state management, pagination, focus control, form management, history management, database operations, dialog management, CMS operations, route operations, datatype operations, field operations, media operations, user operations, backup operations, and content operations.

## Static Variables

### CliContinue

Global bool flag indicating whether CLI should continue running after exit. Defaults to false.

### TitleFile

Embedded filesystem containing ASCII art title files in titles directory.

## Utility Functions

### ParseTitles

Parses title filenames from embedded filesystem, extracting font names from title_FONTNAME.txt pattern.

### LoadTitles

Loads all title ASCII art from embedded filesystem using font names slice, returns slice of title strings.

### GetStatus

Returns styled status bar string based on current ApplicationState. Styles include EDIT, DELETE, WARN, ERROR, OK with appropriate colors and formatting.

### GetConfig

Returns config pointer from Model implementing ModelInterface.

### ShowDialog

Creates command to show dialog with title, message, and showCancel flag. Returns ShowDialogMsg.

### FilePickerOverlay

Renders file picker as full-screen overlay with title, picker view, and hint text. Uses lipgloss Place for centering.

### ViewPageMenus

Builds string from page menu labels with spaces.

### RenderUI

Renders default UI with status bar in docStyle container. Returns dialog overlay if dialog is active.
