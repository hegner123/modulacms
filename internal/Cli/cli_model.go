package cli

import (
	"log"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type CliPage struct {
	Index      int
	Controller CliInterface
	Label      string
	Parent     *CliPage
	Children   []*CliPage
}

type OptionList struct {
	Key  string
	List []Option
}

type Option struct {
	Index     int
	InputType InputType
	Key       string
	Value     string
}

type CliInterface string
type InputType string
type errMsg error

type model struct {
	page       CliPage
	pages      []CliPage
	menu       []*CliPage
	table      string
	tables     []string
	Options    []OptionList
	cursor     int
	selected   map[int]struct{}
	header     string
	body       string
	textarea   textarea.Model
	footer     string
	controller CliInterface
	err        error
}

var (
	txtarea InputType = "textarea"
)

var (
	homePage     *CliPage = &CliPage{Index: 0, Controller: pageInterface, Label: "Home", Parent: nil, Children: []*CliPage{cmsPage, databasePage, bucketPage, oauthPage, configPage}}
	cmsPage      *CliPage = &CliPage{Index: 1, Controller: pageInterface, Label: "CMS", Parent: nil, Children: []*CliPage{contentPage, mediaPage, usersPage}}
	databasePage *CliPage = &CliPage{Index: 2, Controller: pageInterface, Label: "Database", Parent: nil, Children: []*CliPage{createPage, readPage, updatePage, deletePage}}
	bucketPage   *CliPage = &CliPage{Index: 3, Controller: pageInterface, Label: "Bucket", Parent: nil, Children: []*CliPage{}}
	oauthPage    *CliPage = &CliPage{Index: 4, Controller: pageInterface, Label: "oAuth", Parent: nil, Children: []*CliPage{}}
	configPage   *CliPage = &CliPage{Index: 5, Controller: pageInterface, Label: "Configuration", Parent: nil, Children: []*CliPage{}}
	createPage   *CliPage = &CliPage{Index: 6, Controller: tableInterface, Label: "Create", Parent: nil, Children: nil}
	readPage     *CliPage = &CliPage{Index: 7, Controller: tableInterface, Label: "Read", Parent: nil, Children: nil}
	updatePage   *CliPage = &CliPage{Index: 8, Controller: tableInterface, Label: "Update", Parent: nil, Children: nil}
	deletePage   *CliPage = &CliPage{Index: 9, Controller: tableInterface, Label: "Delete", Parent: nil, Children: nil}
	tablePage    *CliPage = &CliPage{Index: 10, Controller: tableInterface, Label: "Table", Parent: nil, Children: nil}
	contentPage  *CliPage = &CliPage{Index: 11, Controller: pageInterface, Label: "Content", Parent: nil, Children: nil}
	mediaPage    *CliPage = &CliPage{Index: 12, Controller: pageInterface, Label: "Media", Parent: nil, Children: nil}
	usersPage    *CliPage = &CliPage{Index: 13, Controller: pageInterface, Label: "Users", Parent: nil, Children: nil}
)
var (
	createInterface CliInterface = "CreateInterface"
	readInterface   CliInterface = "ReadInterface"
	updateInterface CliInterface = "UpdateInterface"
	deleteInterface CliInterface = "DeleteInterface"
	tableInterface  CliInterface = "TableInterface"
	pageInterface   CliInterface = "PageInterface"
)

func CliRun() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

}

func initialModel() model {
	t1 := textarea.New()
	t1.Placeholder = ""
	return model{
		page:   *homePage,
		tables: GetTables(""),
		table:  "",
		menu: []*CliPage{
			cmsPage,
			databasePage,
			bucketPage,
			oauthPage,
			configPage,
		},
		pages: []CliPage{
			*homePage,
			*cmsPage,
			*databasePage,
			*bucketPage,
			*oauthPage,
			*configPage,
			*createPage,
			*readPage,
			*updatePage,
			*deletePage,
			*tablePage,
		},
		selected:   make(map[int]struct{}),
		controller: pageInterface,
		textarea:   t1,
	}
}
