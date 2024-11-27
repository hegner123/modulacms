package main

import (
	"fmt"
	"net/http"
)

func apiGetHandler(w http.ResponseWriter, r *http.Request, apiRoute string) {
	getRoute, err := stripGetPath(apiRoute)
	if err != nil {
		fmt.Print("UM, this ain't a url bud.")
		fmt.Printf("\nerror: %s", err)
		return
	}
    fmt.Println(getRoute)
	switch {
	case matchesPath(getRoute, "adminroute"):
		err := apiGetAdminRoute(w, r)
		if err != nil {
			logError("failed to list Routes: ", err)
		}
	case matchesPath(getRoute, "datatype"):
		err := apiGetDatatype(w, r)
		if err != nil {
			logError("failed to list Routes: ", err)
		}
	case matchesPath(getRoute, "field"):
		err := apiGetField(w, r)
		if err != nil {
			logError("failed to get field ", err)
		}
	case matchesPath(getRoute, "media"):
		err := apiGetMedia(w, r)
		if err != nil {
			logError("failed to get media ", err)
		}
	case matchesPath(getRoute, "mediadimension"):
		err := apiGetMediaDimension(w, r)
		if err != nil {
			logError("failed to list Routes: ", err)
		}
	case matchesPath(getRoute, "route"):
		err := apiGetRoute(w, r)
		if err != nil {
			logError("failed to list Routes: ", err)
		}
	case matchesPath(getRoute, "table"):
		err := apiGetTable(w, r)
		if err != nil {
			logError("failed to list Routes: ", err)
		}
	case matchesPath(getRoute, "token"):
		err := apiGetToken(w, r)
		if err != nil {
			logError("failed to list Routes: ", err)
		}
	case matchesPath(getRoute, "user"):
		err := apiGetUser(w, r)
		if err != nil {
			logError("failed to list Routes: ", err)
		}
	case matchesPath(getRoute, "list/adminroute"):
		err := apiListAdminRoutes(w, r)
		if err != nil {
			logError("failed to list Routes: ", err)
		}
	case matchesPath(getRoute, "list/datatype"):
		err := apiListDatatypes(w, r)
		if err != nil {
			logError("failed to list Routes: ", err)
		}
	case matchesPath(getRoute, "list/field"):
		err := apiListFields(w, r)
		if err != nil {
			logError("failed to list Routes: ", err)
		}
	case matchesPath(getRoute, "list/media"):
		err := apiListMedia(w, r)
		if err != nil {
			logError("failed to list Routes: ", err)
		}
	case matchesPath(getRoute, "list/mediadimension"):
		err := apiListMediaDimensions(w, r)
		if err != nil {
			logError("failed to list Routes: ", err)
		}
	case matchesPath(getRoute, "list/table"):
		err := apiListTables(w, r)
		if err != nil {
			logError("failed to list Routes: ", err)
		}
	case matchesPath(getRoute, "list/token"):
	case matchesPath(getRoute, "list/user"):
		err := apiListUsers(w, r)
		if err != nil {
			logError("failed to list Routes: ", err)
		}
	case matchesPath(apiRoute, "list/routes"):
		err := apiListRoutes(w, r)
		if err != nil {
			logError("failed to list Routes: ", err)
		}
	case matchesPath(apiRoute, "list/fieldsbyroute"):
		err := apiListFieldsForRoute(w, r)
		if err != nil {
			logError("failed to get fields : ", err)
		}

	}
}
