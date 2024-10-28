package main

import (
	"database/sql"
)

type Post struct {
	ID           int
    Slug         string
	Title        string
	Status       int
	DateCreated  int64
	DateModified int64
    Content      string
    Template     string
}

func createPost(db *sql.DB, post Post) (int64, error) {
	result, err := db.Exec("INSERT INTO posts (slug,title, status, dateCreated, dateModified,template) VALUES (?,?,?,?,?,?)",
		post.Slug,post.Title, post.Status, post.DateCreated, post.DateModified, post.Template)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func postExists(db *sql.DB, name string) bool {

	query := `SELECT id FROM posts WHERE title LIKE '%' || ? || '%'`

	rows, err := db.Query(query, name)
	if err != nil {
		return false

	}
	defer rows.Close()
	return true

}

func getPostById(db *sql.DB, id int) (Post, error) {
	var post Post
	err := db.QueryRow("SELECT id, name FROM posts WHERE id = ?", id).Scan(&post.ID, &post.Title)
	return post, err
}

func getAllPosts(db *sql.DB) ([]Post, error) {
    var posts []Post
    // Query only the fields we need (slug, title, and template)
    rows, err := db.Query("SELECT slug, title, template FROM posts")
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

func updatePostById(db *sql.DB, post Post) error {
	_, err := db.Exec("UPDATE posts SET title = ?, status = ?,  WHERE id = ?",
		post.Title, post.Status)
	return err
}

func deletePostById(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM posts WHERE id = ?", id)
	return err
}
