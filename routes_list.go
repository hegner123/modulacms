package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)

func apiListRoutes() ([]byte, error) {
	db, ctx, err := getDb(Database{})
	if err != nil {
		return nil, err
	}
	defer db.Close()
	fetchedRoutes := dbListRoute(db, ctx)
	routes, err := json.Marshal(fetchedRoutes)
	if err != nil {
		logError("failed to Marshal json ", err)
	}
	return routes, nil
}
func apiListFieldsForRoute(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	var routeid int64
	db, ctx, err := getDb(Database{})
	if err != nil {
		return nil, err
	}
	fs := []mdb.ListFieldJoinRow{}
	key := "routid"
	params, err := parseQueryParams(r)
	if err != nil {
		logError("failed to parseQueryParams : ", err)
	}

	if value, exists := params[key]; exists {
		fmt.Fprintf(w, "Key '%s' exists with value: %s\n", key, value)
		routeid, err = strconv.ParseInt(params[key], 10, 64)
	} else {
		fmt.Fprintf(w, "Key '%s' does not exist\n", key)
	}
    fs=dbListFieldsByRoute(db, ctx, routeid)
	fields, err := json.Marshal(fs)
	return fields, nil
}
func apiListUsers(w http.ResponseWriter, r *http.Request) ([]mdb.User, error) {
	us := []mdb.User{}
	return us, nil
}
func apiListMedia(w http.ResponseWriter, r *http.Request) ([]mdb.Media, error) {
	ms := []mdb.Media{}
	return ms, nil
}
