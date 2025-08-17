package cli

type PageIndex int

type Page struct {
	Index PageIndex
	Label string
}

const (
	HOMEPAGE PageIndex = iota
	CMSPAGE
	ADMINCMSPAGE
	DATABASEPAGE
	BUCKETPAGE
	OAUTHPAGE
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
)

func NewDatatypePage(label string) Page {
	return Page{
		Index: DATATYPE,
		Label: label,
	}
}

func NewDynamicPage(label string) Page {
	return Page{
		Index: DYNAMICPAGE,
		Label: label,
	}
}

func NewPickContentPage(label string) Page {
	return Page{
		Index: PICKCONTENT,
		Label: label,
	}
}

func NewPage(index PageIndex, label string) Page {
	return Page{Index: index, Label: label}
}

func InitPages() *map[PageIndex]Page {
	homePage := NewPage(HOMEPAGE, "Home")
	cmsPage := NewPage(CMSPAGE, "CMS")
	adminCmsPage := NewPage(ADMINCMSPAGE, "Admin CMS")
	databasePage := NewPage(DATABASEPAGE, "Database")
	bucketPage := NewPage(BUCKETPAGE, "Bucket Settings")
	oauthPage := NewPage(OAUTHPAGE, "Oauth Settings")
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

	p := make(map[PageIndex]Page, 0)
	p[HOMEPAGE] = homePage
	p[CMSPAGE] = cmsPage
	p[ADMINCMSPAGE] = adminCmsPage
	p[DATABASEPAGE] = databasePage
	p[BUCKETPAGE] = bucketPage
	p[OAUTHPAGE] = oauthPage
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
	return &p
}
