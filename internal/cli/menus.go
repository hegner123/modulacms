package cli

import (
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// HomepageMenuInit initializes the menu for the homepage with main navigation pages.
func (m Model) HomepageMenuInit() []Page {
	HomepageMenu := []Page{
		m.PageMap[CMSPAGE],
		m.PageMap[ADMINCMSPAGE],
		m.PageMap[DATABASEPAGE],
		m.PageMap[BUCKETPAGE],
		m.PageMap[OAUTHPAGE],
		m.PageMap[CONFIGPAGE],
		m.PageMap[ACTIONSPAGE],
	}

	return HomepageMenu
}

// CmsMenuInit initializes the menu for CMS navigation with content management pages.
func (m Model) CmsMenuInit() []Page {
	CmsMenu := []Page{
		m.PageMap[CONTENT],
		m.PageMap[DATATYPES],
		m.PageMap[ROUTES],
		m.PageMap[MEDIA],
		m.PageMap[USERSADMIN],
	}
	return CmsMenu
}

// AdminCmsMenuInit initializes the menu for admin CMS navigation.
func (m Model) AdminCmsMenuInit() []Page {
	AdminCmsMenu := []Page{
		m.PageMap[ADMINCONTENT],
		m.PageMap[ADMINDATATYPES],
		m.PageMap[ADMINROUTES],
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

// DatabaseMenuInit initializes the menu for database operations navigation.
func (m Model) DatabaseMenuInit() []Page {
	DatabaseMenu := []Page{
		m.PageMap[CREATEPAGE],
		m.PageMap[READPAGE],
		m.PageMap[UPDATEPAGE],
		m.PageMap[DELETEPAGE],
	}
	return DatabaseMenu
}

// BuildDatatypeMenu builds a menu of ROOT datatype pages from the provided datatypes.
func (m Model) BuildDatatypeMenu(datatypes []db.Datatypes) []Page {
	out := make([]Page, 0)
	for _, item := range datatypes {
		if item.Type == "ROOT" {
			page := NewDatatypePage(item.Label)
			out = append(out, page)
		}
	}
	return out
}

// BuildContentDataMenu builds a menu of content data pages that are children of the specified root.
func (m Model) BuildContentDataMenu(contentData []db.ContentData, root types.ContentID) []Page {
	out := make([]Page, 0)
	for _, item := range contentData {
		if item.ParentID.Valid && item.ParentID.ID == root {
			out = append(out, NewDynamicPage(fmt.Sprint(item.ContentDataID)))

		}
	}

	return out
}
