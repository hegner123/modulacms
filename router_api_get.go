package main

import (
	"net/http"
)

func apiGetHandler(w http.ResponseWriter, r *http.Request, segments []string) {
	switch {
	case checkPath(segments, DBMETHOD, "get"):
		getRouter(w, r, segments)
	case checkPath(segments, DBMETHOD, "list"):
		listRouter(w, r, segments)
	}
}

func getRouter(w http.ResponseWriter, r *http.Request, segments []string) {
	switch {
	case checkPath(segments, TABLE, "admindatatype"):
		err := apiGetAdminDatatype(w, r)
		if err != nil {
			logError("failed to get adminroute", err)
		}
	case checkPath(segments, TABLE, "adminfields"):
		err := apiGetAdminField(w, r)
		if err != nil {
			logError("failed to get adminroute", err)
		}
	case checkPath(segments, TABLE, "adminroute"):
		err := apiGetAdminRoute(w, r)
		if err != nil {
			logError("failed to get adminroute", err)
		}
	case checkPath(segments, TABLE, "datatype"):
		err := apiGetDatatype(w, r)
		if err != nil {
			logError("failed to get datatype", err)
		}
	case checkPath(segments, TABLE, "field"):
		err := apiGetField(w, r)
		if err != nil {
			logError("failed to get field ", err)
		}
	case checkPath(segments, TABLE, "media"):
		err := apiGetMedia(w, r)
		if err != nil {
			logError("failed to get media ", err)
		}
	case checkPath(segments, TABLE, "mediadimension"):
		err := apiGetMediaDimension(w, r)
		if err != nil {
			logError("failed to get mediadimension", err)
		}
	case checkPath(segments, TABLE, "route"):
		err := apiGetRoute(w, r)
		if err != nil {
			logError("failed to get route", err)
		}
	case checkPath(segments, TABLE, "table"):
		err := apiGetTable(w, r)
		if err != nil {
			logError("failed to get table", err)
		}
	case checkPath(segments, TABLE, "token"):
		err := apiGetToken(w, r)
		if err != nil {
			logError("failed to get token", err)
		}
	case checkPath(segments, TABLE, "user"):
		err := apiGetUser(w, r)
		if err != nil {
			logError("failed to get user", err)
		}
	}
}

func listRouter(w http.ResponseWriter, r *http.Request, segments []string) {
	switch {
	case checkPath(segments, TABLE, "admindatatype"):
		err := apiListAdminDatatypes(w, r)
		if err != nil {
			logError("failed to list admindatatype", err)
		}
	case checkPath(segments, TABLE, "adminfield"):
		err := apiListAdminFields(w, r)
		if err != nil {
			logError("failed to list adminfield", err)
		}
	case checkPath(segments, TABLE, "adminroutes"):
		err := apiListAdminRoutes(w, r)
		if err != nil {
			logError("failed to list adminroute", err)
		}
	case checkPath(segments, TABLE, "datatypes"):
		err := apiListDatatypes(w, r)
		if err != nil {
			logError("failed to list datatype", err)
		}
	case checkPath(segments, TABLE, "fields"):
		err := apiListFields(w, r)
		if err != nil {
			logError("failed to list field", err)
		}
	case checkPath(segments, TABLE, "media"):
		err := apiListMedia(w, r)
		if err != nil {
			logError("failed to list media", err)
		}
	case checkPath(segments, TABLE, "mediadimensions"):
		err := apiListMediaDimensions(w, r)
		if err != nil {
			logError("failed to list mediadimension", err)
		}
	case checkPath(segments, TABLE, "tables"):
		err := apiListTables(w, r)
		if err != nil {
			logError("failed to list table", err)
		}
	case checkPath(segments, TABLE, "tokens"):
	case checkPath(segments, TABLE, "users"):
		err := apiListUsers(w, r)
		if err != nil {
			logError("failed to list user", err)
		}
	case checkPath(segments, TABLE, "routes"):
		err := apiListRoutes(w, r)
		if err != nil {
			logError("failed to list routes", err)
		}
	case checkPath(segments, TABLE, "fieldsbyroute"):
		err := apiListFieldsForRoute(w, r)
		if err != nil {
			logError("failed to get fieldsbyrouter : ", err)
		}

	}
}
