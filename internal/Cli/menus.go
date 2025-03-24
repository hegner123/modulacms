package cli


var homepageMenu []*Page = []*Page{
	cmsPage,
	databasePage,
	bucketPage,
	oauthPage,
	configPage,
}

var cmsMenu []*Page = []*Page{
    defineDatatype,
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

