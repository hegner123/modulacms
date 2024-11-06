package main

import (
	"database/sql"
	"fmt"
)

func createRoute(db *sql.DB, route Route) (int64, error) {
	result, err := db.Exec("INSERT INTO routes (slug, title, status, datecreated, datemodified, content, template) VALUES (?,?,?,?,?,?,?)",
		route.Slug, route.Title, route.Status, route.DateCreated, route.DateModified, route.Content, route.Template)
	if err != nil {
		fmt.Print(err)
		return 0, err
	}
	return result.LastInsertId()
}

func routeExists(db *sql.DB, name string) bool {

	query := `SELECT id FROM routes WHERE title LIKE '%' || ? || '%'`

	rows, err := db.Query(query, name)
	if err != nil {
		return false

	}
	defer rows.Close()
	return true

}

func matchSlugToRoute(db *sql.DB, slug string) (Route, error) {
	var route Route
	err := db.QueryRow(`SELECT template FROM routes WHERE slug LIKE ?;`, slug).Scan(&route.Template)
	if err != nil {
		return route, err
	}
	return route, nil

}

func getRouteById(db *sql.DB, id int) (Route, error) {
	var route Route
	err := db.QueryRow("SELECT id, name FROM routes WHERE id = ?", id).Scan(&route.ID, &route.Title)
	return route, err
}

func getAllRoutes(db *sql.DB) ([]Route, error) {
	var routes []Route
	rows, err := db.Query("SELECT slug, title, template FROM routes")
	if err != nil {
		return routes, err
	}
	defer rows.Close()

	for rows.Next() {
		route := Route{}
		if err := rows.Scan(&route.Slug, &route.Title, &route.Template); err != nil {
			return routes, err
		}
		routes = append(routes, route)
	}

	if err := rows.Err(); err != nil {
		return routes, err
	}

	return routes, nil
}

func updateRouteById(db *sql.DB, route Route) error {
	_, err := db.Exec("UPDATE routes SET title = ?, status = ?,  WHERE id = ?",
		route.Title, route.Status)
	return err
}

func deleteRouteById(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM routes WHERE id = ?", id)
	return err
}
