```go

  // BasePage provides common functionality for all pages
  type BasePage struct {
      header    string
      body      string
      footer    string
      renderUI  func(header, body, footer string) string
  }

  func NewBasePage(header string, renderFunc func(header, body, footer string) string) *BasePage {
      return &BasePage{
          header:   header,
          renderUI: renderFunc,
      }
  }

  func (b *BasePage) SetHeader(header string) {
      b.header = header
  }

  func (b *BasePage) AppendToBody(content string) {
      b.body += content
  }

  func (b *BasePage) View() string {
      return b.renderUI(b.header, b.body, b.footer)
  }


  type HomePage struct {
      *BasePage
      menu *MenuComponent
  }

  func NewHomePage(menuItems []*Page, cursor int, renderFunc func(header, body, footer string) string) *HomePage {
      basePage := NewBasePage("MAIN MENU", renderFunc)
      return &HomePage{
          BasePage: basePage,
          menu: &MenuComponent{
              items:    menuItems,
              cursor:   cursor,
              basePage: basePage,
          },
      }
  }

  func (h *HomePage) Render() string {
      h.menu.RenderMenu()
      return h.View()
  }

func (m *model) View() string {
      // Create the appropriate page type based on current state
      switch m.page.Index {
      case HOMEPAGE:
          homePage := NewHomePage(m.pageMenu, m.cursor, m.RenderUI)
          return homePage.Render()
      case READPAGE:
          h, r, _ := GetColumnsRows(m.table)
          readPage := NewReadPage(m.table, h, r, m.cursor, m.paginator, m.maxRows, m.RenderUI)
          return readPage.Render()
      }
      // Default fallback
      return "Page not found"
  }
```


This approach eliminates duplicate code, improves maintainability, and allows each component to focus on its specific responsibility. 
You can easily create new page types by combining existing components.
