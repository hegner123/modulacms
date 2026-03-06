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

// defaultLayout is the fallback 3-panel 25/50/25 layout.
var defaultLayout = PageLayout{
	Panels: 3,
	Ratios: [3]float64{0.25, 0.50, 0.25},
	Titles: [3]string{"Tree", "Content", "Route"},
}

// pageLayouts maps each page to its preferred panel configuration.
// Pages not in the map use defaultLayout.
var pageLayouts = map[PageIndex]PageLayout{
	// Single-panel pages
	HOMEPAGE:       {1, [3]float64{0, 1, 0}, [3]string{"", "Home", ""}},
	QUICKSTARTPAGE: {1, [3]float64{0, 1, 0}, [3]string{"", "Quickstart", ""}},
	// ACTIONSPAGE uses GridScreen, no legacy layout needed
	PLUGINDETAILPAGE: {1, [3]float64{0, 1, 0}, [3]string{"", "Plugin", ""}},

	// Two-panel pages
	CONFIGPAGE:   {2, [3]float64{0.30, 0.70, 0}, [3]string{"Categories", "Fields", "Detail"}},
	DATABASEPAGE: {2, [3]float64{0.30, 0.70, 0}, [3]string{"Tables", "Actions", "Info"}},
	// MEDIA uses GridScreen, no legacy layout needed

	// Three-panel pages
	// CONTENT uses GridScreen, no legacy layout needed
	DATATYPES:      {3, [3]float64{0.25, 0.40, 0.35}, [3]string{"Datatypes", "Fields", "Actions"}},
	ROUTES:         {3, [3]float64{0.25, 0.50, 0.25}, [3]string{"Routes", "Details", "Actions"}},
	USERSADMIN:     {3, [3]float64{0.25, 0.50, 0.25}, [3]string{"Users", "Details", "Permissions"}},
	ADMINROUTES:    {3, [3]float64{0.25, 0.50, 0.25}, [3]string{"Admin Routes", "Details", "Actions"}},
	ADMINDATATYPES: {3, [3]float64{0.25, 0.40, 0.35}, [3]string{"Admin Datatypes", "Fields", "Actions"}},
	// ADMINCONTENT uses GridScreen, no legacy layout needed
	PLUGINSPAGE:        {3, [3]float64{0.25, 0.50, 0.25}, [3]string{"Plugins", "Details", "Info"}},
	FIELDTYPES:         {3, [3]float64{0.25, 0.50, 0.25}, [3]string{"Field Types", "Details", "Actions"}},
	ADMINFIELDTYPES:    {3, [3]float64{0.25, 0.50, 0.25}, [3]string{"Admin Field Types", "Details", "Actions"}},
	DEPLOYPAGE:         {3, [3]float64{0.25, 0.50, 0.25}, [3]string{"Environments", "Details", "Actions"}},
	PIPELINESPAGE:      {3, [3]float64{0.25, 0.50, 0.25}, [3]string{"Pipelines", "Entries", "Info"}},
	PIPELINEDETAILPAGE: {3, [3]float64{0.25, 0.50, 0.25}, [3]string{"Pipelines", "Configuration", "Status"}},
	WEBHOOKSPAGE:       {3, [3]float64{0.25, 0.50, 0.25}, [3]string{"Webhooks", "Details", "Info"}},
	CMSPAGE:            {3, [3]float64{0.25, 0.50, 0.25}, [3]string{"System", "Navigation", "Info"}},
	ADMINCMSPAGE:       {3, [3]float64{0.25, 0.50, 0.25}, [3]string{"System", "Navigation", "Info"}},
	READPAGE:           {3, [3]float64{0.20, 0.55, 0.25}, [3]string{"Mode", "Data", "Detail"}},
}

// layoutForPage returns the PageLayout for a given page index,
// falling back to defaultLayout if the page is not in the map.
func layoutForPage(idx PageIndex) PageLayout {
	if l, ok := pageLayouts[idx]; ok {
		return l
	}
	return defaultLayout
}

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
