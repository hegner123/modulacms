package main

import (
	"net/http"
	"strconv"
	"time"
)

func apiCreatePost(w http.ResponseWriter, r *http.Request) string {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return "Couldn't parse form"
	}
	db, err := getDb()
	if err != nil {
		return "Couldn't get db"
	}
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
func apiGetAllPosts() ([]Post, error) {
	fetchedPosts := []Post{}
	db, err := getDb()
	if err != nil {
		return fetchedPosts, err
	}

	fetchedPosts, err = getAllPosts(db)

	return fetchedPosts, nil
}
func apiGetPost(w http.ResponseWriter, r *http.Request) (Post, error) {
	fetchedPost := Post{}
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return fetchedPost, err
	}
	db, err := getDb()
	if err != nil {
		return fetchedPost, err
	}
	postIdForm := r.FormValue("postid")
	postId, err := strconv.ParseInt(postIdForm, 10, 32)
	if err != nil {
		return fetchedPost, err
	}
	fetchedPost, err = getPostById(db, int(postId))
	if err != nil {
		return fetchedPost, err
	}
	return fetchedPost, nil
}
func apiUpdatePost() {}
func apiDeletePost() {}

func apiCreateField()         {}
func apiGetField()            {}
func apiGetAllFieldsForPost() {}
func apiUpdateField()         {}
func apiDeleteField()         {}

func apiCreateUser()  {}
func apiGetUser()     {}
func apiAuthUser()    {}
func apiGetAllUsers() {}
func apiUpdateUser()  {}
func apiDeleteUser()  {}

func apiCreateMedia()  {}
func apiGetMedia()     {}
func apiGetAllMedias() {}
func apiUpdateMedia()  {}
func apiDeleteMedia()  {}
