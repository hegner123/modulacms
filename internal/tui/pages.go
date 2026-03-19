package tui

// PageIndex represents a page identifier in the application.
type PageIndex int

// Page represents a navigable page with an index and label.
type Page struct {
	Index PageIndex
	Label string
}

// Page index constants define all available pages in the application.
const (
	HOMEPAGE PageIndex = iota
	CMSPAGE
	ADMINCMSPAGE
	DATABASEPAGE
	CONFIGPAGE
	READPAGE
	DATATYPES
	USERSADMIN
	MEDIA
	CONTENT
	ACTIONSPAGE
	ROUTES
	ADMINROUTES
	ADMINDATATYPES
	ADMINCONTENT
	PLUGINSPAGE
	PLUGINDETAILPAGE
	QUICKSTARTPAGE
	FIELDTYPES
	ADMINFIELDTYPES
	DEPLOYPAGE
	PIPELINESPAGE
	PIPELINEDETAILPAGE
	WEBHOOKSPAGE
	PLUGINTUIPAGE
	VALIDATIONS
	ADMINVALIDATIONS
	TOKENSPAGE
	SESSIONSPAGE
	MEDIADIMENSIONSPAGE
	IMPORTPAGE
)

// NewPage creates a new page with the specified index and label.
func NewPage(index PageIndex, label string) Page {
	return Page{Index: index, Label: label}
}

// InitPages initializes and returns a map of all application pages.
func InitPages() *map[PageIndex]Page {
	homePage := NewPage(HOMEPAGE, "Home")
	cmsPage := NewPage(CMSPAGE, "CMS")
	adminCmsPage := NewPage(ADMINCMSPAGE, "Admin CMS")
	databasePage := NewPage(DATABASEPAGE, "Database")
	configPage := NewPage(CONFIGPAGE, "Config")
	readPage := NewPage(READPAGE, "Read")
	datatypesPage := NewPage(DATATYPES, "Datatypes")
	usersAdminPage := NewPage(USERSADMIN, "Users")
	mediaPage := NewPage(MEDIA, "Media")
	contentPage := NewPage(CONTENT, "Content")
	actionsPage := NewPage(ACTIONSPAGE, "Actions")
	routesPage := NewPage(ROUTES, "Routes")
	adminRoutesPage := NewPage(ADMINROUTES, "Admin Routes")
	adminDatatypesPage := NewPage(ADMINDATATYPES, "Admin Datatypes")
	adminContentPage := NewPage(ADMINCONTENT, "Admin Content")
	pluginsPage := NewPage(PLUGINSPAGE, "Plugins")
	pluginDetailPage := NewPage(PLUGINDETAILPAGE, "Plugin Detail")
	quickstartPage := NewPage(QUICKSTARTPAGE, "Quickstart")
	fieldTypesPage := NewPage(FIELDTYPES, "Field Types")
	adminFieldTypesPage := NewPage(ADMINFIELDTYPES, "Admin Field Types")
	deployPage := NewPage(DEPLOYPAGE, "Deploy")
	pipelinesPage := NewPage(PIPELINESPAGE, "Pipelines")
	pipelineDetailPage := NewPage(PIPELINEDETAILPAGE, "Pipeline Detail")
	webhooksPage := NewPage(WEBHOOKSPAGE, "Webhooks")
	pluginTuiPage := NewPage(PLUGINTUIPAGE, "Plugin")
	validationsPage := NewPage(VALIDATIONS, "Validations")
	adminValidationsPage := NewPage(ADMINVALIDATIONS, "Admin Validations")
	tokensPage := NewPage(TOKENSPAGE, "Tokens")
	sessionsPage := NewPage(SESSIONSPAGE, "Sessions")
	mediaDimensionsPage := NewPage(MEDIADIMENSIONSPAGE, "Media Dimensions")
	importPage := NewPage(IMPORTPAGE, "Import")

	p := make(map[PageIndex]Page, 0)
	p[HOMEPAGE] = homePage
	p[CMSPAGE] = cmsPage
	p[ADMINCMSPAGE] = adminCmsPage
	p[DATABASEPAGE] = databasePage
	p[CONFIGPAGE] = configPage
	p[READPAGE] = readPage
	p[DATATYPES] = datatypesPage
	p[USERSADMIN] = usersAdminPage
	p[MEDIA] = mediaPage
	p[CONTENT] = contentPage
	p[ACTIONSPAGE] = actionsPage
	p[ROUTES] = routesPage
	p[ADMINROUTES] = adminRoutesPage
	p[ADMINDATATYPES] = adminDatatypesPage
	p[ADMINCONTENT] = adminContentPage
	p[PLUGINSPAGE] = pluginsPage
	p[PLUGINDETAILPAGE] = pluginDetailPage

	p[QUICKSTARTPAGE] = quickstartPage
	p[FIELDTYPES] = fieldTypesPage
	p[ADMINFIELDTYPES] = adminFieldTypesPage
	p[DEPLOYPAGE] = deployPage
	p[PIPELINESPAGE] = pipelinesPage
	p[PIPELINEDETAILPAGE] = pipelineDetailPage
	p[WEBHOOKSPAGE] = webhooksPage
	p[PLUGINTUIPAGE] = pluginTuiPage
	p[VALIDATIONS] = validationsPage
	p[ADMINVALIDATIONS] = adminValidationsPage
	p[TOKENSPAGE] = tokensPage
	p[SESSIONSPAGE] = sessionsPage
	p[MEDIADIMENSIONSPAGE] = mediaDimensionsPage
	p[IMPORTPAGE] = importPage
	return &p
}
