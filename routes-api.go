package main

import (
	"fmt"
	"net/http"
	"time"
)



func apiCreatePost(w http.ResponseWriter,r *http.Request)string{
			err := r.ParseForm()
			if err != nil {
				http.Error(w, "Error parsing form", http.StatusBadRequest)
				return "Couldn't parse form"
			}
			db, err := getDb()
			if err != nil {
				return "Couldn't get db"
			}
			fmt.Print(r.FormValue("title"))
			title := r.FormValue("title")
			slug := r.FormValue("slug")
			content := r.FormValue("content")
			now := time.Now().Unix()
			post := Post{Slug: slug, Title: title, Status: 0, DateCreated: now, DateModified: now, Content: content, Template: "page.html"}
			_, err = createPost(db, post)
			message := "created successfully"
			if err != nil {
				message = "error creating post"
			}
    return message
}
func apiGetPost(){}
func apiGetAllPosts(){}
func apiUpdatePost(){}
func apiDeletePost(){}


func apiCreateField(){}
func apiGetField(){}
func apiGetAllFields(){}
func apiUpdateField(){}
func apiDeleteField(){}


func apiCreateMetadata(){}
func apiGetMetadata(){}
func apiGetAllMetadatas(){}
func apiUpdateMetadata(){}
func apiDeleteMetadata(){}


func apiCreateUser(){}
func apiGetUser(){}
func apiGetAllUsers(){}
func apiUpdateUser(){}
func apiDeleteUser(){}


func apiCreateMedia(){}
func apiGetMedia(){}
func apiGetAllMedias(){}
func apiUpdateMedia(){}
func apiDeleteMedia(){}
