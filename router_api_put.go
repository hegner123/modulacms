package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func apiPutHandler(w http.ResponseWriter, r *http.Request, segments []string) {
	switch {
	case checkPath(segments, DBMETHOD, "adminroute"):
		res := fmt.Sprintf("updated adminroute %v successfully\n", r.FormValue("slug"))
		err := apiUpdateAdminRoute(w, r)
		if err != nil {
			res = fmt.Sprintf("Error updating adminroute:%v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(map[string]string{"result": res})
		if err != nil {
			fmt.Printf("\nerror: %s", err)
			return
		}
	case checkPath(segments, DBMETHOD, "datatype"):
		res := fmt.Sprintf("updated datatype  %v successfully\n", r.FormValue("id"))
		err := apiUpdateDatatype(w, r)
		if err != nil {
			res = fmt.Sprintf("Error updating datatype:%v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(map[string]string{"result": res})
		if err != nil {
			fmt.Printf("\nerror: %s", err)
			return
		}
	case checkPath(segments, DBMETHOD, "field"):
		res := fmt.Sprintf("updated field %v successfully\n", r.FormValue("id"))
        err := apiUpdateField(w, r)
		if err != nil {
			res = fmt.Sprintf("Error updating field:%v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(map[string]string{"result": res})
		if err != nil {
			fmt.Printf("\nerror: %s", err)
			return
		}
	case checkPath(segments, DBMETHOD, "media"):
		res := fmt.Sprintf("updated media %v successfully\n", r.FormValue("id"))
		err := apiUpdateMedia(w, r)
		if err != nil {
			res = fmt.Sprintf("Error updating media:%v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(map[string]string{"result": res})
		if err != nil {
			fmt.Printf("\nerror: %s", err)
			return
		}
	case checkPath(segments, DBMETHOD, "mediadimension"):
		res := fmt.Sprintf("updated mediadimension %v successfully\n", r.FormValue("id"))
		err := apiUpdateMediaDimension(w, r)
		if err != nil {
			res = fmt.Sprintf("Error updating mediadimension:%v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(map[string]string{"result": res})
		if err != nil {
			fmt.Printf("\nerror: %s", err)
			return
		}
	case checkPath(segments, DBMETHOD, "route"):
		res := fmt.Sprintf("updated route %v successfully\n", r.FormValue("slug"))
		err := apiUpdateRoute(w, r)
		if err != nil {
			res = fmt.Sprintf("Error updating Route :%v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(map[string]string{"result": res})
		if err != nil {
			fmt.Printf("\nerror: %s", err)
			return
		}
	case checkPath(segments, DBMETHOD, "table"):
		res := fmt.Sprintf("updated table %v successfully\n", r.FormValue("id"))
		err := apiUpdateTables(w, r)
		if err != nil {
			res = fmt.Sprintf("Error updating table:%v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(map[string]string{"result": res})
		if err != nil {
			fmt.Printf("\nerror: %s", err)
			return
		}
	case checkPath(segments, DBMETHOD, "token"):
		res := fmt.Sprintf("updated token %v successfully\n", r.FormValue("id"))
		err := apiUpdateToken(w, r)
		if err != nil {
			res = fmt.Sprintf("Error updating token:%v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(map[string]string{"result": res})
		if err != nil {
			fmt.Printf("\nerror: %s", err)
			return
		}
	case checkPath(segments, DBMETHOD, "user"):
		res := fmt.Sprintf("updated user %v successfully\n", r.FormValue("id"))
		err := apiUpdateUser(w, r)
		if err != nil {
			res = fmt.Sprintf("Error updating user:%v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(map[string]string{"result": res})
		if err != nil {
			fmt.Printf("\nerror: %s", err)
			return
		}
	}
}
