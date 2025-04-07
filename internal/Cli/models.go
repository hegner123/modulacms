package cli

import (
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	utility "github.com/hegner123/modulacms/internal/Utility"
)

type ThemeComponent struct {
	titleFont int
	titles    []string
	txtStyle  lipgloss.Style
	quitStyle lipgloss.Style
}

func NewThemeComponent() ThemeComponent {
	fs, err := TitleFile.ReadDir("titles")
	if err != nil {
		utility.DefaultLogger.Fatal("", err)
	}
	fonts := ParseTitleFonts(fs)

	return ThemeComponent{
		titleFont: 0,
		titles:    LoadTitles(fonts),
		txtStyle:  lipgloss.NewStyle(),
		quitStyle: lipgloss.NewStyle(),
	}
}

type NavigationComponent struct {
	cursor     int
	page       Page
	pageMenu   []*Page
	pages      []Page
	history    []Page
	controller CliInterface
}

func (n *NavigationComponent) PushHistory(page Page) {
	n.history = append(n.history, page)
}

func (n *NavigationComponent) PopHistory() *Page {
	if len(n.history) == 0 {
		return &n.page
	}
	lastIndex := len(n.history) - 1
	lastPage := n.history[lastIndex]
	n.history = n.history[:lastIndex]
	return &lastPage
}

func NewNavigationComponent() NavigationComponent {
	return NavigationComponent{
		cursor: 0,
		pageMenu: []*Page{
			cmsPage,
			databasePage,
			bucketPage,
			oauthPage,
			configPage,
		},
		pages: []Page{
			*homePage,
			*cmsPage,
			*databasePage,
			*bucketPage,
			*oauthPage,
			*configPage,
			*tablePage,
			*createPage,
			*readPage,
			*updatePage,
			*deletePage,
			*updateFormPage,
			*readSinglePage,
			*dynamicPage,
			*defineDatatype,
		},
		controller: pageInterface,
	}

}

type TableComponent struct {
	cursor    int
	table     string
	tables    []string
	headers   []string
	rows      [][]string
	maxRows   int
	paginator paginator.Model
	pageMod   int
}

func (t *TableComponent) GetIDRow() int64 {
	if len(t.rows) == 0 || t.cursor >= len(t.rows) {
		return 0
	}
	row := t.rows[t.cursor]
	rowCol := row[0]
	id, err := strconv.ParseInt(rowCol, 10, 64)
	if err != nil {
		return 0
	}
	return id
}

func NewTableComponent() TableComponent {
	p := paginator.New()
	p.Type = paginator.Dots

	p.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render("•")
	p.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).Render("•")
	return TableComponent{
		paginator: p,
		maxRows:   10,
		pageMod:   0,
		tables:    GetTables(),
		table:     "",
		headers:   make([]string, 0),
	}

}

type FormComponent struct {
	form       *huh.Form
    formGroups []huh.Group
    formFields []huh.Field
	formLen    int
	formMap    []string
	formValues []*string
	formSubmit bool
}

func NewFormComponent() FormComponent {
	return FormComponent{
        form: &huh.Form{},
        formGroups: make([]huh.Group, 0),
        formFields: make([]huh.Field, 0),
        formLen: 0,
        formMap: make([]string, 0),
        formValues: make([]*string, 0),
        formSubmit: false,

    }

}

type ContentComponent struct {
	title    string
	header   string
	body     string
	footer   string
	content  string
	viewport viewport.Model
	ready    bool
}

func NewContentComponent() ContentComponent {
	return ContentComponent{
		title:    "",
		header:   "",
		body:     "",
		footer:   "",
		content:  "",
		viewport: viewport.Model{},
		ready:    false,
	}
}

type Model struct {
	Theme      ThemeComponent
	Navigation NavigationComponent
	Table      TableComponent
	Form       FormComponent
	Content    ContentComponent
	Error      string
	Verbose    bool
	Time       time.Time
}

func InitModel(v *bool) Model {
	verbose := false
	if v != nil {
		verbose = *v
	}

	return Model{
		Theme:      NewThemeComponent(),
		Navigation: NewNavigationComponent(),
		Table:      NewTableComponent(),
		Form:       NewFormComponent(),
		Content:    NewContentComponent(),
		Verbose:    verbose,
		Time:       time.Now(),
	}
}
