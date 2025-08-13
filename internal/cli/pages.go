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
	DEFINEDATATYPE
	DEVELOPMENT
	DATATYPE
	USERSADMIN
	MEDIA
	CONTENT
	PICKCONTENT
)

var (
	homePage           *Page = &Page{Index: HOMEPAGE, Label: "Home"}
	cmsPage            *Page = &Page{Index: CMSPAGE, Label: "CMS"}
	adminCmsPage       *Page = &Page{Index: ADMINCMSPAGE, Label: "ADMIN CMS"}
	selectTablePage    *Page = &Page{Index: DATABASEPAGE, Label: "Database"}
	bucketPage         *Page = &Page{Index: BUCKETPAGE, Label: "Bucket"}
	oauthPage          *Page = &Page{Index: OAUTHPAGE, Label: "Oauth"}
	configPage         *Page = &Page{Index: CONFIGPAGE, Label: "Config"}
	tableActionsPage   *Page = &Page{Index: TABLEPAGE, Label: "Table Actions"}
	createPage         *Page = &Page{Index: CREATEPAGE, Label: "Create"}
	readPage           *Page = &Page{Index: READPAGE, Label: "Read"}
	updatePage         *Page = &Page{Index: UPDATEPAGE, Label: "Update"}
	deletePage         *Page = &Page{Index: DELETEPAGE, Label: "Delete"}
	updateFormPage     *Page = &Page{Index: UPDATEFORMPAGE, Label: "UpdateForm"}
	readSinglePage     *Page = &Page{Index: READSINGLEPAGE, Label: "ReadSingle"}
	dynamicPage        *Page = &Page{Index: DYNAMICPAGE, Label: "Dynamic"}
	defineDatatypePage *Page = &Page{Index: DEFINEDATATYPE, Label: "DefineDatatype"}
	developmentPage    *Page = &Page{Index: DEVELOPMENT, Label: "Development"}
	usersAdminPage     *Page = &Page{Index: USERSADMIN, Label: "Users"}
	mediaPage          *Page = &Page{Index: MEDIA, Label: "Media"}
	contentPage        *Page = &Page{Index: CONTENT, Label: "Content"}
)

func NewDatatypePage(label string) *Page {
	return &Page{
		Index: DATATYPE,
		Label: label,
	}
}

func NewDynamicPage(label string) *Page {
	return &Page{
		Index: DYNAMICPAGE,
		Label: label,
	}
}

func NewPickContentPage(label string) *Page {
	return &Page{
		Index: PICKCONTENT,
		Label: label,
	}
}

func InitPages() *map[PageIndex]Page {
	p := make(map[PageIndex]Page, 0)
	p[HOMEPAGE] = *homePage
	p[CMSPAGE] = *cmsPage
	p[ADMINCMSPAGE] = *adminCmsPage
	p[DATABASEPAGE] = *selectTablePage
	p[BUCKETPAGE] = *bucketPage
	p[OAUTHPAGE] = *oauthPage
	p[CONFIGPAGE] = *configPage
	p[TABLEPAGE] = *tableActionsPage
	p[CREATEPAGE] = *createPage
	p[READPAGE] = *readPage
	p[UPDATEPAGE] = *updatePage
	p[DELETEPAGE] = *deletePage
	p[UPDATEFORMPAGE] = *updateFormPage
	p[READSINGLEPAGE] = *readSinglePage
	p[DYNAMICPAGE] = *dynamicPage
	p[DEFINEDATATYPE] = *defineDatatypePage
	p[DEVELOPMENT] = *developmentPage
	p[USERSADMIN] = *usersAdminPage
	p[MEDIA] = *mediaPage
	p[CONTENT] = *contentPage
	return &p
}
