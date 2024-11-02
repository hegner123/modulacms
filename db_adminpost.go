package main

import (
	"database/sql"
	"fmt"
)


func createAdminPost(db *sql.DB, post Post) (int64, error) {
	result, err := db.Exec("INSERT INTO adminposts (slug, title, status, datecreated, datemodified, content, template) VALUES (?,?,?,?,?,?,?)",
		post.Slug, post.Title, post.Status, post.DateCreated, post.DateModified, post.Content, post.Template)
	if err != nil {
		fmt.Print(err)
		return 0, err
	}
	return result.LastInsertId()
}

func adminPostExists(db *sql.DB, name string) bool {

	query := `SELECT id FROM adminposts WHERE title LIKE '%' || ? || '%'`

	rows, err := db.Query(query, name)
	if err != nil {
		return false

	}
	defer rows.Close()
	return true

}

func matchAdminSlugToRoute(db *sql.DB, slug string) (Post, error) {
	var route Post
	err := db.QueryRow(`SELECT template FROM adminposts WHERE slug LIKE ?;`, slug).Scan(&route.Template)
	if err != nil {
		return route, err
	}
	return route, nil

}

func getAdminPostById(db *sql.DB, id int) (Post, error) {
	var post Post
	err := db.QueryRow("SELECT id, name FROM adminposts WHERE id = ?", id).Scan(&post.ID, &post.Title)
	return post, err
}

func getAllAdminPosts(db *sql.DB) ([]Post, error) {
	var posts []Post
	// Query only the fields we need (slug, title, and template)
	rows, err := db.Query("SELECT slug, title, template FROM adminposts")
	if err != nil {
		return posts, err
	}
	defer rows.Close()

	for rows.Next() {
		post := Post{}
		// Only scan into the selected fields
		if err := rows.Scan(&post.Slug, &post.Title, &post.Template); err != nil {
			return posts, err
		}
		posts = append(posts, post)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return posts, err
	}

	return posts, nil
}

func updateAdminPostById(db *sql.DB, post Post) error {
	_, err := db.Exec("UPDATE adminposts SET title = ?, status = ?,  WHERE id = ?",
		post.Title, post.Status)
	return err
}

func deleteAdminPostById(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM adminposts WHERE id = ?", id)
	return err
}
