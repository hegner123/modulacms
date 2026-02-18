package cli

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
	TABLEPAGE
	CREATEPAGE
	READPAGE
	UPDATEPAGE
	DELETEPAGE
	UPDATEFORMPAGE
	READSINGLEPAGE
	DYNAMICPAGE
	DATATYPES
	DATATYPESMENU
	FIELDS
	DEVELOPMENT
	DATATYPE
	USERSADMIN
	MEDIA
	CONTENT
	PICKCONTENT
	EDITCONTENT
	ACTIONSPAGE
	ROUTES
	ADMINROUTES
	ADMINDATATYPES
	ADMINCONTENT
	PLUGINSPAGE
	PLUGINDETAILPAGE
	CONFIGCATEGORYPAGE
)

// NewDatatypePage creates a new datatype page with the specified label.
func NewDatatypePage(label string) Page {
	return Page{
		Index: DATATYPE,
		Label: label,
	}
}

// NewDynamicPage creates a new dynamic page with the specified label.
func NewDynamicPage(label string) Page {
	return Page{
		Index: DYNAMICPAGE,
		Label: label,
	}
}

// NewPickContentPage creates a new content picker page with the specified label.
func NewPickContentPage(label string) Page {
	return Page{
		Index: PICKCONTENT,
		Label: label,
	}
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
	tableActionsPage := NewPage(TABLEPAGE, "Table Actions")
	createPage := NewPage(CREATEPAGE, "Create")
	readPage := NewPage(READPAGE, "Read")
	updatePage := NewPage(UPDATEPAGE, "Update")
	deletePage := NewPage(DELETEPAGE, "Delete")
	updateFormPage := NewPage(UPDATEFORMPAGE, "Update Form")
	readSinglePage := NewPage(READSINGLEPAGE, "Read Single")
	dynamicPage := NewPage(DYNAMICPAGE, "Dynamic Page")
	datatypesPage := NewPage(DATATYPES, "Datatypes")
	datatypesMenuPage := NewPage(DATATYPESMENU, "Datatypes Menu")
	addFields := NewPage(FIELDS, "Fields")
	developmentPage := NewPage(DEVELOPMENT, "Development")
	usersAdminPage := NewPage(USERSADMIN, "Users")
	mediaPage := NewPage(MEDIA, "Media")
	contentPage := NewPage(CONTENT, "Content")
	editContentPage := NewPage(EDITCONTENT, "Edit")
	actionsPage := NewPage(ACTIONSPAGE, "Actions")
	routesPage := NewPage(ROUTES, "Routes")
	datatypePage := NewPage(DATATYPE, "Define Datatype")
	adminRoutesPage := NewPage(ADMINROUTES, "Admin Routes")
	adminDatatypesPage := NewPage(ADMINDATATYPES, "Admin Datatypes")
	adminContentPage := NewPage(ADMINCONTENT, "Admin Content")
	pluginsPage := NewPage(PLUGINSPAGE, "Plugins")
	pluginDetailPage := NewPage(PLUGINDETAILPAGE, "Plugin Detail")
	configCategoryPage := NewPage(CONFIGCATEGORYPAGE, "Config Category")

	p := make(map[PageIndex]Page, 0)
	p[HOMEPAGE] = homePage
	p[CMSPAGE] = cmsPage
	p[ADMINCMSPAGE] = adminCmsPage
	p[DATABASEPAGE] = databasePage
	p[CONFIGPAGE] = configPage
	p[TABLEPAGE] = tableActionsPage
	p[CREATEPAGE] = createPage
	p[READPAGE] = readPage
	p[UPDATEPAGE] = updatePage
	p[DELETEPAGE] = deletePage
	p[UPDATEFORMPAGE] = updateFormPage
	p[READSINGLEPAGE] = readSinglePage
	p[DYNAMICPAGE] = dynamicPage
	p[DATATYPES] = datatypesPage
	p[DATATYPESMENU] = datatypesMenuPage
	p[FIELDS] = addFields
	p[DEVELOPMENT] = developmentPage
	p[USERSADMIN] = usersAdminPage
	p[MEDIA] = mediaPage
	p[CONTENT] = contentPage
	p[EDITCONTENT] = editContentPage
	p[ACTIONSPAGE] = actionsPage
	p[ROUTES] = routesPage
	p[DATATYPE] = datatypePage
	p[ADMINROUTES] = adminRoutesPage
	p[ADMINDATATYPES] = adminDatatypesPage
	p[ADMINCONTENT] = adminContentPage
	p[PLUGINSPAGE] = pluginsPage
	p[PLUGINDETAILPAGE] = pluginDetailPage
	p[CONFIGCATEGORYPAGE] = configCategoryPage
	return &p
}
