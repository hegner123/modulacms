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
	_ // 5 -- was TABLEPAGE (deprecated Phase 1)
	_ // 6 -- was CREATEPAGE (deprecated Phase 4)
	READPAGE
	_ // 8 -- was UPDATEPAGE (deprecated Phase 3)
	_ // 9 -- was DELETEPAGE (deprecated Phase 3)
	_ // 10 -- was UPDATEFORMPAGE (deprecated Phase 4)
	_ // 11 -- was READSINGLEPAGE (deprecated Phase 3)
	_ // 12 -- was DYNAMICPAGE (deprecated Phase 0)
	DATATYPES
	_ // 14 -- was DATATYPESMENU (deprecated Phase 0)
	_ // 15 -- was FIELDS (deprecated Phase 0)
	_ // 16 -- was DEVELOPMENT (deprecated Phase 0)
	_ // 17 -- was DATATYPE (deprecated Phase 4)
	USERSADMIN
	MEDIA
	CONTENT
	_ // 21 -- was PICKCONTENT (deprecated Phase 0)
	_ // 22 -- was EDITCONTENT (deprecated Phase 0)
	ACTIONSPAGE
	ROUTES
	ADMINROUTES
	ADMINDATATYPES
	ADMINCONTENT
	PLUGINSPAGE
	PLUGINDETAILPAGE
	_ // 30 -- was CONFIGCATEGORYPAGE (deprecated Phase 2)
	QUICKSTARTPAGE
	FIELDTYPES
	ADMINFIELDTYPES
	DEPLOYPAGE
	PIPELINESPAGE
	PIPELINEDETAILPAGE
	WEBHOOKSPAGE
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
	return &p
}
