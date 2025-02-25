package cli

var homepageMenu []*CliPage = []*CliPage{
	cmsPage,
	databasePage,
	bucketPage,
	oauthPage,
	configPage,
}

var cmsMenu []*CliPage = []*CliPage{
	contentPage,
	mediaPage,
	usersPage,
}
var tableMenu []*CliPage = []*CliPage{
	createPage,
	readPage,
	updatePage,
	deletePage,
}
