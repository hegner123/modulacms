package cli

func (m model) View() string {
	var ui string
	switch m.page.Index {
	case homePage.Index:
		ui = m.PageHome()
	case databasePage.Index:
		ui = m.PageDatabase()
	case cmsPage.Index:
		ui = m.PageCMS()
	case bucketPage.Index:
		ui = m.PageBucket()
	case tablePage.Index:
		ui = m.PageTable()
	case createPage.Index:
		ui = m.PageCreate()
	case readPage.Index:
		ui = m.PageRead()
	case updatePage.Index:
		ui = m.PageUpdate()
	case updateFormPage.Index:
		ui = m.PageUpdateForm()
	case deletePage.Index:
		ui = m.PageDelete()
	default:
		ui = m.Page404()
	}
	return ui
}
