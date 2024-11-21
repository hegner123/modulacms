package main

type ModulaAdminMenu struct {
	DestinationLinks []AdminLink
	ToolbarPrimary   []AdminLink
	ToolbarSecondary []AdminLink
	Style            Theme
}


type AdminLink struct {
	Name          string
	Href          string
	Target        bool
	Icon          IconSrc
	ListClasses   string
	AnchorClasses string
	Type          string
	Children      bool
	ChildLinks    []AdminLink
}

type IconSrc struct {
	Svg string
	Src string
}


type Theme struct {
	PrimaryColor    string `json:"primarycolor"`
	SecondaryColor  string `json:"secondarycolor"`
	BackgroundColor string `json:"backgroundcolor"`
	ForegroundColor string `json:"foregroundcolor"`
	BorderColor     string `json:"bordercolor"`
	Color           string `json:"color"`
	FontFamily      string `json:"fontfamily"`
	FontSize        string `json:"fontsize"`
	FontWeight      string `json:"fontweight"`
	LineHeight      string `json:"lineheight"`
	Margin          string `json:"margin"`
	Padding         string `json:"padding"`
	Gap             string `json:"gap"`
	MaxWidth        string `json:"maxwidth"`
	GridTemplate    string `json:"gridtemplate"`
	FlexDirection   string `json:"flexdirection"`
	BorderRadius    string `json:"borderradius"`
	BoxShadow       string `json:"boxshadow"`
	Transition      string `json:"transition"`
	Opacity         string `json:"opacity"`
}

type ModulaSidebarMenu struct {
    Sections []SidebarSection
}


type SidebarSection struct{
    Label string
    Links []AdminLink

}



func initAdmin(){
    adminMenu := ModulaAdminMenu{}
    destinationLinks :=[]AdminLink{}
    toolbarsSecondary :=[]AdminLink{}
    
    visitSite:= AdminLink{Name: "Visit Site",Href: "/",Target: false,ListClasses:"modula-admin-menu-primary",Type: "main" }
    logout := AdminLink{Name:"Logout", Href:"/admin/logout", Target: false, ListClasses: "modula-admin-logout", Type: "link"}
    

    destinationLinks[0]=visitSite
    toolbarsSecondary[0]=logout
    adminMenu.DestinationLinks = destinationLinks
    adminMenu.ToolbarSecondary = toolbarsSecondary
}







