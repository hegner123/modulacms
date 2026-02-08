package cli

import (
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

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

func (m Model) AdminCmsMenuInit() []Page {
	AdminCmsMenu := []Page{
		m.PageMap[ADMINCONTENT],
		m.PageMap[ADMINDATATYPES],
		m.PageMap[ADMINROUTES],
	}
	return AdminCmsMenu
}

func (m Model) ContentMenuInit() []Page {
	ContentMenu := []Page{
		m.PageMap[ROUTES],
	}
	return ContentMenu
}

func (m Model) DatabaseMenuInit() []Page {
	DatabaseMenu := []Page{
		m.PageMap[CREATEPAGE],
		m.PageMap[READPAGE],
		m.PageMap[UPDATEPAGE],
		m.PageMap[DELETEPAGE],
	}
	return DatabaseMenu
}

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

func (m Model) BuildContentDataMenu(contentData []db.ContentData, root types.ContentID) []Page {
	out := make([]Page, 0)
	for _, item := range contentData {
		if item.ParentID.Valid && item.ParentID.ID == root {
			out = append(out, NewDynamicPage(fmt.Sprint(item.ContentDataID)))

		}
	}

	return out
}
