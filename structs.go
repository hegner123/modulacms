package main

type Database struct {
	DB string
}

type Media struct {
	Id                 int32  `json:"id"`
	Name               string `json:"name"`
	DisplayName        string `json:"displayName"`
	Alt                string `json:"alt"`
	Caption            string `json:"caption"`
	Description        string `json:"description"`
	Class              string `json:"class"`
	Author             string `json:"author"`
	AuthorID           int32  `json:"authorid"`
	DateCreated        string `json:"datecreated"`
	DateModified       string `json:"datemodified"`
	Url                string `json:"url"`
	MimeType           string `json:"mimeType"`
	Dimensions         string `json:"dimensions"`
	OptimizedMobile    string `json:"optimizedMobile"`
	OptimizedTablet    string `json:"optimizedTablet"`
	OptimizedDesktop   string `json:"optimizedDesktop"`
	OptimizedUltrawide string `json:"optimizedUltrawide"`
}

type Post struct {
	ID           int    `json:"id"`
	Author       string `json:"author"`
	AuthorID     string `json:"authorId"`
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Status       int    `json:"status"`
	DateCreated  int64  `json:"datecreated"`
	DateModified int64  `json:"datemodified"`
	Content      string `json:"content"`
	Template     string `json:"template"`
}
type AdminPost struct {
	ID           int    `json:"id"`
	Author       string `json:"author"`
	AuthorID     string `json:"authorId"`
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Status       int    `json:"status"`
	DateCreated  int64  `json:"datecreated"`
	DateModified int64  `json:"datemodified"`
	Content      string `json:"content"`
	Template     string `json:"template"`
}

type Field struct {
	ID           int    `json:"id"`
	PostID       int    `json:"postId"`
	Author       string `json:"author"`
	AuthorID     string `json:"authorId"`
	Key          string `json:"key"`
	Data         string `json:"data"`
	DateCreated  string `json:"datecreated"`
	DateModified string `json:"datemodified"`
	Component    string `json:"component"`
	Tags         string `json:"tags"`
	Parent       string `json:"parent"`
}

type Config struct {
	Port            string `json:"port"`
	SSLPort         string `json:"ssl_port"`
	ClientSite      string `json:"client_site"`
	DB_URL          string `json:"db_url"`
	DB_NAME         string `json:"db_name"`
	DB_PASSWORD     string `json:"db_password"`
	Bucket_URL      string `json:"bucket_url"`
	Bucket_PASSWORD string `json:"bucket_password"`
}

type User struct {
	ID           int    `json:"id"`
	DateCreated  string `json:"datecreated"`
	DateModified string `json:"datemodified"`
	UserName     string `json:"username"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Hash         string `json:"hash"`
	Role         string `json:"role"`
}

type Routes struct {
	Title string
	Pages []Post
}
