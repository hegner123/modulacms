package tui

import (
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// HomepageMenuInit builds the flattened home navigation menu.
// CMS items (Content, Datatypes, etc.) are top-level; admin mode
// is resolved at navigation time via AdminPageIndex, not via submenus.
func (m Model) HomepageMenuInit() []Page {
	// Daily workflow
	pages := []Page{
		m.PageMap[CONTENT],
		m.PageMap[MEDIA],
		m.PageMap[ROUTES],
	}

	// Schema / structure
	pages = append(pages,
		m.PageMap[DATATYPES],
		m.PageMap[FIELDTYPES],
		m.PageMap[VALIDATIONS],
		m.PageMap[MEDIADIMENSIONSPAGE],
		m.PageMap[USERSADMIN],
	)

	// System
	pages = append(pages,
		m.PageMap[PLUGINSPAGE],
		m.PageMap[PIPELINESPAGE],
		m.PageMap[WEBHOOKSPAGE],
		m.PageMap[TOKENSPAGE],
		m.PageMap[SESSIONSPAGE],
		m.PageMap[CONFIGPAGE],
		m.PageMap[DEPLOYPAGE],
	)

	// Power user
	pages = append(pages, m.PageMap[IMPORTPAGE])
	pages = append(pages, m.PageMap[ACTIONSPAGE])
	if !m.IsRemote {
		pages = append(pages, m.PageMap[DATABASEPAGE])
	}
	pages = append(pages, m.PageMap[QUICKSTARTPAGE])

	return pages
}

// AdminPageIndex maps a client page index to its admin variant when
// admin mode is active. Pages without an admin variant return unchanged.
func AdminPageIndex(idx PageIndex) PageIndex {
	switch idx {
	case CONTENT:
		return ADMINCONTENT
	case DATATYPES:
		return ADMINDATATYPES
	case ROUTES:
		return ADMINROUTES
	case FIELDTYPES:
		return ADMINFIELDTYPES
	case VALIDATIONS:
		return ADMINVALIDATIONS
	default:
		return idx
	}
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
		m.PageMap[VALIDATIONS],
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
		m.PageMap[ADMINVALIDATIONS],
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
