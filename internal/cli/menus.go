package cli


var homepageMenu []*Page = []*Page{
    developmentPage,
	cmsPage,
	selectTablePage,
	bucketPage,
	oauthPage,
	configPage,
}

var cmsMenu []*Page = []*Page{
    definedDatatypePage,
	contentPage,
	mediaPage,
	usersPage,
}
var tableMenu []*Page = []*Page{
	createPage,
	readPage,
	updatePage,
	deletePage,
}

