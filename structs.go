package main

type Media struct {
    Id int32
    Name string
    DisplayName string
    Alt string
    Caption string
    Description string
    Class string
    CreatedBy int32
    DateCreated string
    DateModified string
    Url string
    MimeType string
    Dimensions string
    OptimizedMobile string
    OptimizedTablet string
    OptimizedDesktop string
    OptimizedUltrawide string
}

type Field struct {
    ID          int
    Type        string


}
type Config struct {
	DB_URL          string `json:"db_url"`
	DB_NAME         string `json:"db_name"`
	DB_PASSWORD     string `json:"db_password"`
	Bucket_URL      string `json:"bucket_url"`
	Bucket_PASSWORD string `json:"bucket_password"`
}
type User struct {
	ID       int
	UserName string
	Name     string
	Email    string
	Hash     string
	Role     string
}



type BlogPageData struct {
	Title       string
	Heading     string
	Description string
	Posts       []Post
}

type Page404 struct {
	Title   string
	Message string
}

type Routes struct {
	Title string
	Pages []Post
}
