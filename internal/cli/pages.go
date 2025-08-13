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
	PICKCONTENT
)

var (
	homePage            *Page = &Page{Index: HOMEPAGE, Label: "Home"}
	cmsPage             *Page = &Page{Index: CMSPAGE, Label: "CMS"}
	adminCmsPage        *Page = &Page{Index: ADMINCMSPAGE, Label: "ADMIN CMS"}
	selectTablePage     *Page = &Page{Index: DATABASEPAGE, Label: "Database"}
	bucketPage          *Page = &Page{Index: BUCKETPAGE, Label: "Bucket"}
	oauthPage           *Page = &Page{Index: OAUTHPAGE, Label: "Oauth"}
	configPage          *Page = &Page{Index: CONFIGPAGE, Label: "Config"}
	tableActionsPage    *Page = &Page{Index: TABLEPAGE, Label: "Table Actions"}
	createPage          *Page = &Page{Index: CREATEPAGE, Label: "Create"}
	readPage            *Page = &Page{Index: READPAGE, Label: "Read"}
	updatePage          *Page = &Page{Index: UPDATEPAGE, Label: "Update"}
	deletePage          *Page = &Page{Index: DELETEPAGE, Label: "Delete"}
	updateFormPage      *Page = &Page{Index: UPDATEFORMPAGE, Label: "UpdateForm"}
	readSinglePage      *Page = &Page{Index: READSINGLEPAGE, Label: "ReadSingle"}
	dynamicPage         *Page = &Page{Index: DYNAMICPAGE, Label: "Dynamic"}
	definedDatatypePage *Page = &Page{Index: DEFINEDATATYPE, Label: "DefineDatatype"}
	developmentPage     *Page = &Page{Index: DEVELOPMENT, Label: "Development"}
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
