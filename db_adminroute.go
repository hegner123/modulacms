package main

import (
	"database/sql"
	"fmt"
)


func createAdminRoute(db *sql.DB, route AdminRoute) (int64, error) {
	result, err := db.Exec("INSERT INTO adminroutes (slug, title, status, datecreated, datemodified, content, template) VALUES (?,?,?,?,?,?,?)",
		route.Slug, route.Title, route.Status, route.DateCreated, route.DateModified, route.Content, route.Template)
	if err != nil {
		fmt.Print(err)
		return 0, err
	}
	return result.LastInsertId()
}

func adminRouteExists(db *sql.DB, name string) bool {

	query := `SELECT id FROM adminroutes WHERE title LIKE '%' || ? || '%'`

	rows, err := db.Query(query, name)
	if err != nil {
		return false

	}
	defer rows.Close()
	return true

}

func matchAdminSlugToRoute(db *sql.DB, slug string) (Route, error) {
	var route Route
	err := db.QueryRow(`SELECT template FROM adminroutes WHERE slug LIKE ?;`, slug).Scan(&route.Template)
	if err != nil {
		return route, err
	}
	return route, nil

}

func getAdminRouteById(db *sql.DB, id int) (Route, error) {
	var route Route
	err := db.QueryRow("SELECT id, name FROM adminroutes WHERE id = ?", id).Scan(&route.ID, &route.Title)
	return route, err
}

func getAllAdminRoutes(db *sql.DB) ([]Route, error) {
	var routes []Route
	// Query only the fields we need (slug, title, and template)
	rows, err := db.Query("SELECT slug, title, template FROM adminroutes")
	if err != nil {
		return routes, err
	}
	defer rows.Close()

	for rows.Next() {
		route := Route{}
		// Only scan into the selected fields
		if err := rows.Scan(&route.Slug, &route.Title, &route.Template); err != nil {
			return routes, err
		}
		routes = append(routes, route)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return routes, err
	}

	return routes, nil
}

func updateAdminRouteById(db *sql.DB, route Route) error {
	_, err := db.Exec("UPDATE adminroutes SET title = ?, status = ?,  WHERE id = ?",
		route.Title, route.Status)
	return err
}

func deleteAdminRouteById(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM adminroutes WHERE id = ?", id)
	return err
}
