package tui

import (
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// HomepageMenuInit initializes the menu for the homepage with main navigation pages.
// In remote mode, DATABASEPAGE is excluded because it requires raw SQL access.
func (m Model) HomepageMenuInit() []Page {
	pages := []Page{
		m.PageMap[CMSPAGE],
		m.PageMap[ADMINCMSPAGE],
	}
	if !m.IsRemote {
		pages = append(pages, m.PageMap[DATABASEPAGE])
	}
	pages = append(pages,
		m.PageMap[QUICKSTARTPAGE],
		m.PageMap[CONFIGPAGE],
		m.PageMap[ACTIONSPAGE],
		m.PageMap[PLUGINSPAGE],
		m.PageMap[PIPELINESPAGE],
		m.PageMap[DEPLOYPAGE],
		m.PageMap[WEBHOOKSPAGE],
	)
	return pages
}

// CmsMenuInit initializes the menu for CMS navigation with content management pages.
func (m Model) CmsMenuInit() []Page {
	CmsMenu := []Page{
		m.PageMap[CONTENT],
		m.PageMap[DATATYPES],
		m.PageMap[ROUTES],
		m.PageMap[MEDIA],
		m.PageMap[USERSADMIN],
		m.PageMap[FIELDTYPES],
	}
	return CmsMenu
}

// AdminCmsMenuInit initializes the menu for admin CMS navigation.
func (m Model) AdminCmsMenuInit() []Page {
	AdminCmsMenu := []Page{
		m.PageMap[ADMINCONTENT],
		m.PageMap[ADMINDATATYPES],
		m.PageMap[ADMINROUTES],
		m.PageMap[ADMINFIELDTYPES],
	}
	return AdminCmsMenu
}

// ContentMenuInit initializes the menu for content navigation.
func (m Model) ContentMenuInit() []Page {
	ContentMenu := []Page{
		m.PageMap[ROUTES],
	}
	return ContentMenu
}

// BuildDatatypeMenu builds a menu of _root datatype labels from the provided datatypes.
func (m Model) BuildDatatypeMenu(datatypes []db.Datatypes) []string {
	out := make([]string, 0)
	for _, item := range datatypes {
		if types.DatatypeType(item.Type).IsRootType() {
			out = append(out, item.Label)
		}
	}
	return out
}

// ConfigCategoryMenuInit returns the category labels used on the CONFIGPAGE.
func ConfigCategoryMenuInit() []string {
	categories := config.AllCategories()
	items := make([]string, 0, len(categories)+1)
	for _, cat := range categories {
		items = append(items, config.CategoryLabel(cat))
	}
	items = append(items, "View Raw JSON")
	return items
}
