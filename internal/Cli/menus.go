package cli

type formAction string

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

const (
	edit   formAction = "Edit"
	submit formAction = "Submit"
	reset  formAction = "Reset"
	cancel formAction = "Cancel"
)
