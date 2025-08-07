package cli

import (
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
)

var HomepageMenu []*Page = []*Page{
	developmentPage,
	cmsPage,
	selectTablePage,
	bucketPage,
	oauthPage,
	configPage,
}

var CmsMenu []*Page = []*Page{
	definedDatatypePage,
	// Removing undefined references
	// contentPage,
	// mediaPage,
	// usersPage,
}
var TableMenu []*Page = []*Page{
	createPage,
	readPage,
	updatePage,
	deletePage,
}

func (m Model) BuildDatatypeMenu(datatypes []db.Datatypes) []*Page {
	out := make([]*Page, 0)
	for _, item := range datatypes {
		if item.Type == "ROOT" {
			page := NewDatatypePage(item.Label)
			out = append(out, page)
		}
	}
	return out
}

func (m Model) BuildContentDataMenu(contentData []db.ContentData, root int) []*Page {
	out := make([]*Page, 0)
	for _, item := range contentData {
		if int(item.ParentID.Int64) == root {
			out = append(out, NewDynamicPage(fmt.Sprint(item.ContentDataID)))

		}
	}

	return out
}
